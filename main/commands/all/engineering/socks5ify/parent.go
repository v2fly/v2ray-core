//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

func runParent(opts parentOptions, child childConfig) error {
	socketPair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET, 0)
	if err != nil {
		return err
	}
	parentSock := socketPair[0]
	childSock := socketPair[1]
	defer unix.Close(parentSock)

	encodedChildConfig, err := encodeChildConfig(child)
	if err != nil {
		unix.Close(childSock)
		return err
	}

	childFile := os.NewFile(uintptr(childSock), "socks5ify-control")
	childCmd := exec.Command(os.Args[0], "engineering", "socks5ify")
	childCmd.Stdin = os.Stdin
	childCmd.Stdout = os.Stdout
	childCmd.Stderr = os.Stderr
	childCmd.Env = append(os.Environ(),
		childConfigEnv+"="+encodedChildConfig,
		childSockFDEnv+"=3",
	)
	childCmd.ExtraFiles = []*os.File{childFile}
	childCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUSER | unix.CLONE_NEWNS | unix.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
		GidMappingsEnableSetgroups: false,
	}

	if err := childCmd.Start(); err != nil {
		childFile.Close()
		return err
	}
	childFile.Close()

	tunFD, err := recvFileDescriptor(parentSock)
	if err != nil {
		_ = childCmd.Wait()
		return err
	}

	server, err := startV2RayWithTun(opts, child, tunFD)
	if err != nil {
		_ = childCmd.Process.Kill()
		_ = childCmd.Wait()
		return err
	}
	defer server.Close()

	if _, err := unix.Write(parentSock, []byte{1}); err != nil {
		_ = childCmd.Process.Kill()
		_ = childCmd.Wait()
		return fmt.Errorf("failed to release child process: %w", err)
	}

	return waitForChild(childCmd)
}

func waitForChild(childCmd *exec.Cmd) error {
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- childCmd.Wait()
	}()

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		select {
		case err := <-waitDone:
			if err == nil {
				return nil
			}
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 0 {
					return nil
				}
			}
			return err
		case sig := <-signals:
			if sig == os.Interrupt {
				continue
			}
			_ = childCmd.Process.Signal(sig)
		}
	}
}
