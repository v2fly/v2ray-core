package salamander

import (
	"errors"
	"net"
	"sync"
	"syscall"
)

const udpBufferSize = 2048

type obfsPacketConn struct {
	net.PacketConn
	Obfs *SalamanderObfuscator

	readBuf    []byte
	readMutex  sync.Mutex
	writeBuf   []byte
	writeMutex sync.Mutex
}

func WrapPacketConn(conn net.PacketConn, obfs *SalamanderObfuscator) net.PacketConn {
	return &obfsPacketConn{
		PacketConn: conn,
		Obfs:       obfs,
		readBuf:    make([]byte, udpBufferSize),
		writeBuf:   make([]byte, udpBufferSize),
	}
}

func (c *obfsPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	c.readMutex.Lock()
	defer c.readMutex.Unlock()

	n, addr, err = c.PacketConn.ReadFrom(c.readBuf)
	if err != nil {
		return n, addr, err
	}

	if n < smSaltLen {
		return 0, addr, nil
	}

	if len(p) < n-smSaltLen {
		return 0, addr, nil
	}

	c.Obfs.Deobfuscate(c.readBuf[:n], p)

	return n - smSaltLen, addr, nil
}

func (c *obfsPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if len(p)+smSaltLen > udpBufferSize {
		return 0, nil
	}

	c.Obfs.Obfuscate(p, c.writeBuf)

	_, err = c.PacketConn.WriteTo(c.writeBuf[:len(p)+smSaltLen], addr)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (c *obfsPacketConn) SyscallConn() (syscall.RawConn, error) {
	sc, ok := c.PacketConn.(syscall.Conn)
	if !ok {
		return nil, errors.New("not supported")
	}
	return sc.SyscallConn()
}
