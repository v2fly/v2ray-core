package scenarios

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/proxy"

	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/retry"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/units"
)

func xor(b []byte) []byte {
	r := make([]byte, len(b))
	for i, v := range b {
		r[i] = v ^ 'c'
	}
	return r
}

// idleReader refreshes the read deadline before every read, so that the
// timeout bounds the time the peer may stay silent rather than the duration of
// the whole transfer. A busy CI machine can make a large but perfectly healthy
// transfer exceed a whole-transfer deadline, while a genuinely stalled proxy
// still trips the idle deadline.
type idleReader struct {
	conn    net.Conn
	timeout time.Duration
}

func (r *idleReader) Read(b []byte) (int, error) {
	if err := r.conn.SetReadDeadline(time.Now().Add(r.timeout)); err != nil {
		return 0, err
	}
	return r.conn.Read(b)
}

func readFrom(conn net.Conn, timeout time.Duration, length int) []byte {
	b := make([]byte, length)
	n, err := io.ReadFull(&idleReader{conn: conn, timeout: timeout}, b[:length])
	conn.SetReadDeadline(time.Time{}) //nolint: errcheck
	if err != nil {
		fmt.Println("Unexpected error from readFrom:", err)
	}
	return b[:n]
}

func readFrom2(conn net.Conn, timeout time.Duration, length int) ([]byte, error) {
	b := make([]byte, length)
	n, err := io.ReadFull(&idleReader{conn: conn, timeout: timeout}, b[:length])
	conn.SetReadDeadline(time.Time{}) //nolint: errcheck
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

// writeChunkSize bounds how much is handed to a single Write so that the write
// deadline can be refreshed as the transfer progresses. It is larger than any
// datagram used by the UDP tests, so those payloads are still sent as one
// datagram by a single Write.
const writeChunkSize = 64 * 1024

// writeWithIdleTimeout writes payload to conn, allowing timeout for every
// chunk rather than for the payload as a whole. See idleReader for why.
func writeWithIdleTimeout(conn net.Conn, payload []byte, timeout time.Duration) (int, error) {
	written := 0
	for written < len(payload) {
		end := written + writeChunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return written, err
		}
		n, err := conn.Write(payload[written:end])
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func InitializeServerConfigs(configs ...*core.Config) ([]*exec.Cmd, error) {
	servers := make([]*exec.Cmd, 0, 10)

	for _, config := range configs {
		server, err := InitializeServerConfig(config)
		if err != nil {
			CloseAllServers(servers)
			return nil, err
		}
		servers = append(servers, server)
	}

	time.Sleep(time.Second * 2)

	return servers, nil
}

func InitializeServerConfig(config *core.Config) (*exec.Cmd, error) {
	err := BuildV2Ray()
	if err != nil {
		return nil, err
	}

	config = withDefaultApps(config)
	configBytes, err := proto.Marshal(config)
	if err != nil {
		return nil, err
	}
	proc := RunV2RayProtobuf(configBytes)

	if err := proc.Start(); err != nil {
		return nil, err
	}

	return proc, nil
}

var (
	testBinaryPath    string
	testBinaryPathGen sync.Once
)

func genTestBinaryPath() {
	testBinaryPathGen.Do(func() {
		var tempDir string
		common.Must(retry.Timed(5, 100).On(func() error {
			dir, err := os.MkdirTemp("", "v2ray")
			if err != nil {
				return err
			}
			tempDir = dir
			return nil
		}))
		file := filepath.Join(tempDir, "v2ray.test")
		if runtime.GOOS == "windows" {
			file += ".exe"
		}
		testBinaryPath = file
		fmt.Printf("Generated binary path: %s\n", file)
	})
}

func GetSourcePath() string {
	return filepath.Join("github.com", "v2fly", "v2ray-core", "v5", "main")
}

func CloseAllServers(servers []*exec.Cmd) {
	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "Closing all servers.",
	})
	for _, server := range servers {
		if runtime.GOOS == "windows" {
			server.Process.Kill()
		} else {
			server.Process.Signal(syscall.SIGTERM)
		}
	}
	for _, server := range servers {
		// Cmd.Wait, unlike Process.Wait, also releases the pipe that feeds the
		// configuration into the process and waits for the goroutine copying
		// it, so repeated test runs do not leak file descriptors.
		server.Wait()
	}
	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "All server closed.",
	})
}

func CloseServer(server *exec.Cmd) {
	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "Closing server.",
	})
	if runtime.GOOS == "windows" {
		server.Process.Kill()
	} else {
		server.Process.Signal(syscall.SIGTERM)
	}
	server.Wait()
	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "Server closed.",
	})
}

func withDefaultApps(config *core.Config) *core.Config {
	config.App = append(config.App, serial.ToTypedMessage(&dispatcher.Config{}))
	config.App = append(config.App, serial.ToTypedMessage(&proxyman.InboundConfig{}))
	config.App = append(config.App, serial.ToTypedMessage(&proxyman.OutboundConfig{}))
	return config
}

func testTCPConnViaSocks(socksPort, testPort net.Port, payloadSize int, timeout time.Duration) func() error { //nolint: unparam
	return func() error {
		socksDialer, err := proxy.SOCKS5("tcp", "127.0.0.1:"+socksPort.String(), nil, nil)
		if err != nil {
			return err
		}
		destAddr := &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(testPort),
		}
		conn, err := socksDialer.Dial("tcp", destAddr.String())
		if err != nil {
			return err
		}
		defer conn.Close()

		return testTCPConn2(conn, payloadSize, timeout)()
	}
}

func testTCPConn(port net.Port, payloadSize int, timeout time.Duration) func() error {
	return func() error {
		conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		return testTCPConn2(conn, payloadSize, timeout)()
	}
}

func testUDPConn(port net.Port, payloadSize int, timeout time.Duration) func() error { // nolint: unparam
	return func() error {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		return testTCPConn2(conn, payloadSize, timeout)()
	}
}

func testTCPConn2(conn net.Conn, payloadSize int, timeout time.Duration) func() error {
	return func() (err1 error) {
		start := time.Now()
		defer func() {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			fmt.Println("testConn finishes:", time.Since(start).Milliseconds(), "ms\t",
				err1, "\tAlloc =", units.ByteSize(m.Alloc).String(),
				"\tTotalAlloc =", units.ByteSize(m.TotalAlloc).String(),
				"\tSys =", units.ByteSize(m.Sys).String(),
				"\tNumGC =", m.NumGC)
		}()
		payload := make([]byte, payloadSize)
		common.Must2(rand.Read(payload))

		nBytes, err := writeWithIdleTimeout(conn, payload, timeout)
		conn.SetWriteDeadline(time.Time{}) //nolint: errcheck
		if err != nil {
			return err
		}
		if nBytes != len(payload) {
			return errors.New("expect ", len(payload), " written, but actually ", nBytes)
		}

		response, err := readFrom2(conn, timeout, payloadSize)
		if err != nil {
			return err
		}

		if !bytes.Equal(response, xor(payload)) {
			return errors.New("unexpected response of ", len(response), " bytes, expected the ", payloadSize, " bytes sent to be echoed back")
		}

		return nil
	}
}
