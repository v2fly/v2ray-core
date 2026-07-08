//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func runChildFromEnv() error {
	sockFD, err := childSocketFD()
	if err != nil {
		return err
	}

	cfg, err := decodeChildConfig(os.Getenv(childConfigEnv))
	if err != nil {
		sendChildError(sockFD, err)
		return err
	}

	tunFD, err := setupChildNamespace(cfg)
	if err != nil {
		sendChildError(sockFD, err)
		return err
	}

	if err := sendFileDescriptor(sockFD, tunFD); err != nil {
		_ = unix.Close(tunFD)
		return err
	}
	_ = unix.Close(tunFD)

	ready := []byte{0}
	if _, err := unix.Read(sockFD, ready); err != nil {
		return fmt.Errorf("parent did not finish setup: %w", err)
	}
	if ready[0] != 1 {
		return fmt.Errorf("parent reported setup failure")
	}

	command := cfg.Command
	if len(command) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		command = []string{shell}
	}

	env := filteredChildEnv(os.Environ())
	argv0 := command[0]
	if !strings.ContainsRune(argv0, '/') {
		resolved, err := exec.LookPath(argv0)
		if err != nil {
			return err
		}
		argv0 = resolved
	}
	if cfg.KeepUID {
		return runCommandWithCallerIdentity(cfg, argv0, command, env)
	}
	return syscall.Exec(argv0, command, env)
}

func runCommandWithCallerIdentity(cfg childConfig, argv0 string, command []string, env []string) error {
	if cfg.CallerUID <= 0 {
		return fmt.Errorf("-keep-uid requires a non-root caller UID, got %d", cfg.CallerUID)
	}
	if cfg.CallerGID < 0 {
		return fmt.Errorf("-keep-uid requires a valid caller GID, got %d", cfg.CallerGID)
	}

	childCmd := exec.Command(argv0, command[1:]...)
	childCmd.Args[0] = command[0]
	childCmd.Stdin = os.Stdin
	childCmd.Stdout = os.Stdout
	childCmd.Stderr = os.Stderr
	childCmd.Env = env
	childCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: cfg.CallerUID, HostID: 0, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: cfg.CallerGID, HostID: 0, Size: 1},
		},
		GidMappingsEnableSetgroups: false,
		Credential: &syscall.Credential{
			Uid:         uint32(cfg.CallerUID),
			Gid:         uint32(cfg.CallerGID),
			NoSetGroups: true,
		},
	}
	if err := childCmd.Start(); err != nil {
		return fmt.Errorf("start command with caller UID/GID: %w", err)
	}
	waitForNestedCommand(childCmd)
	return nil
}

func waitForNestedCommand(childCmd *exec.Cmd) {
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- childCmd.Wait()
	}()

	signals := make(chan os.Signal, 4)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		select {
		case err := <-waitDone:
			exitWithCommandStatus(err)
		case sig := <-signals:
			if sig == os.Interrupt {
				continue
			}
			_ = childCmd.Process.Signal(sig)
		}
	}
}

func exitWithCommandStatus(err error) {
	if err == nil {
		os.Exit(0)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Exited() {
				os.Exit(status.ExitStatus())
			}
			if status.Signaled() {
				os.Exit(128 + int(status.Signal()))
			}
		}
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func setupChildNamespace(cfg childConfig) (int, error) {
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return -1, fmt.Errorf("failed to make mounts private: %w", err)
	}
	if err := applyMountOverrides(cfg); err != nil {
		return -1, err
	}

	tunFD, err := createTun(cfg.TunName)
	if err != nil {
		return -1, err
	}
	if err := configureInterface(cfg.TunName, cfg.MTU, cfg.IPv4, cfg.IPv6, cfg.IPv6Config); err != nil {
		_ = unix.Close(tunFD)
		return -1, err
	}
	return tunFD, nil
}

func filteredChildEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, item := range env {
		if strings.HasPrefix(item, childConfigEnv+"=") || strings.HasPrefix(item, childSockFDEnv+"=") {
			continue
		}
		out = append(out, item)
	}
	return out
}

func applyMountOverrides(cfg childConfig) error {
	if len(cfg.DNS) > 0 {
		resolvFile, err := writeGeneratedResolvConf(cfg.DNS)
		if err != nil {
			return err
		}
		defer os.Remove(resolvFile)
		if err := bindMountFile(resolvFile, "/etc/resolv.conf"); err != nil {
			return err
		}
	}
	if cfg.ResolvConf != "" {
		if err := bindMountFile(cfg.ResolvConf, "/etc/resolv.conf"); err != nil {
			return err
		}
	}
	for _, bind := range cfg.BindFiles {
		if err := bindMountFile(bind.Source, bind.Target); err != nil {
			return err
		}
	}
	return nil
}

func writeGeneratedResolvConf(servers []string) (string, error) {
	var builder strings.Builder
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if server == "" {
			continue
		}
		builder.WriteString("nameserver ")
		builder.WriteString(server)
		builder.WriteByte('\n')
	}
	if builder.Len() == 0 {
		return "", fmt.Errorf("-dns did not contain any DNS server")
	}

	file, err := os.CreateTemp("", "socks5ify-resolv-*.conf")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(builder.String()); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func bindMountFile(source string, target string) error {
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("bind source %q: %w", source, err)
	}
	targetInfo, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("bind target %q: %w", target, err)
	}
	if targetInfo.IsDir() {
		return fmt.Errorf("bind target %q is a directory", target)
	}
	if err := unix.Mount(source, target, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("bind mount %q onto %q: %w", source, target, err)
	}
	return nil
}

func createTun(name string) (int, error) {
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	ifr, err := unix.NewIfreq(name)
	if err != nil {
		_ = unix.Close(fd)
		return -1, err
	}
	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)
	if err := unix.IoctlIfreq(fd, unix.TUNSETIFF, ifr); err != nil {
		_ = unix.Close(fd)
		return -1, err
	}
	return fd, nil
}

func configureInterface(name string, mtu int, ipv4 tunProtocolConfig, enableIPv6 bool, ipv6 tunProtocolConfig) error {
	if err := setLinkMTU(name, mtu); err != nil {
		return err
	}
	if err := setLinkUp("lo"); err != nil {
		return err
	}
	if err := setLinkUp(name); err != nil {
		return err
	}
	iface, err := interfaceIndex(name)
	if err != nil {
		return err
	}
	if err := addAddress(iface, ipv4.Guest, ipv4.Prefix); err != nil {
		return err
	}
	if err := addDefaultRoute(iface, false); err != nil {
		return err
	}
	if enableIPv6 {
		if err := addAddress(iface, ipv6.Guest, ipv6.Prefix); err != nil {
			return err
		}
		if err := addDefaultRoute(iface, true); err != nil {
			return err
		}
	}
	return nil
}

func setLinkMTU(name string, mtu int) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	ifr, err := unix.NewIfreq(name)
	if err != nil {
		return err
	}
	ifr.SetUint32(uint32(mtu))
	if err := unix.IoctlIfreq(fd, unix.SIOCSIFMTU, ifr); err != nil {
		return fmt.Errorf("set MTU on %s to %s: %w", name, strconv.Itoa(mtu), err)
	}
	return nil
}

func setLinkUp(name string) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	ifr, err := unix.NewIfreq(name)
	if err != nil {
		return err
	}
	if err := unix.IoctlIfreq(fd, unix.SIOCGIFFLAGS, ifr); err != nil {
		return fmt.Errorf("get flags for %s: %w", name, err)
	}
	ifr.SetUint16(ifr.Uint16() | unix.IFF_UP)
	if err := unix.IoctlIfreq(fd, unix.SIOCSIFFLAGS, ifr); err != nil {
		return fmt.Errorf("set %s up: %w", name, err)
	}
	return nil
}
