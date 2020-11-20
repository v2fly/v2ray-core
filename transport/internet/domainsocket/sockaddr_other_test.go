// +build !confonly !linux

package domainsocket

import (
	"testing"

	"net"
	"os"
)

func TestUnixSockaddrDial103(t *testing.T) {
	name := "/tmp/len103_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.sock"
	os.Remove(name)
	addr := &net.UnixAddr{Net: "unix", Name: name}
	if _, err := net.ListenUnix("unix", addr); err != nil {
		t.Error(err)
	}
	_, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		t.Error(err)
	}
}

func TestUnixSockaddrDial104(t *testing.T) {
	name := "/tmp/len104_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.sock"
	os.Remove(name)
	addr := &net.UnixAddr{Net: "unix", Name: name}
	if _, err := net.ListenUnix("unix", addr); err == nil {
		t.Error(err)
	}
	_, err := net.DialUnix("unix", nil, addr)
	if err == nil {
		t.Error(err)
	}
}
