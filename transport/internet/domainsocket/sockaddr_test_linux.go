// +build !confonly linux

package domainsocket

import (
	"testing"

	"net"
	"os"
)

func TestUnixSockaddrDial107(t *testing.T) {
	name := "/tmp/len107_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.sock"
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

func TestUnixSockaddrDial108(t *testing.T) {
	name := "/tmp/len108_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.sock"
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
