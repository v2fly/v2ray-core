package shadowsocks2022

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

type testUDPCachedStateContainer struct {
	client map[string]UDPClientPacketProcessorCachedState
	server map[string]UDPClientPacketProcessorCachedState
}

func newTestUDPCachedStateContainer() *testUDPCachedStateContainer {
	return &testUDPCachedStateContainer{
		client: make(map[string]UDPClientPacketProcessorCachedState),
		server: make(map[string]UDPClientPacketProcessorCachedState),
	}
}

func (c *testUDPCachedStateContainer) GetCachedState(sessionID string) UDPClientPacketProcessorCachedState {
	return c.client[sessionID]
}

func (c *testUDPCachedStateContainer) PutCachedState(sessionID string, cache UDPClientPacketProcessorCachedState) {
	c.client[sessionID] = cache
}

func (c *testUDPCachedStateContainer) GetCachedServerState(serverSessionID string) UDPClientPacketProcessorCachedState {
	return c.server[serverSessionID]
}

func (c *testUDPCachedStateContainer) PutCachedServerState(serverSessionID string, cache UDPClientPacketProcessorCachedState) {
	c.server[serverSessionID] = cache
}

func newTestAESUDPClientPacketProcessor(t *testing.T) *AESUDPClientPacketProcessor {
	t.Helper()
	key := make([]byte, 16)
	requestBlock, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	responseBlock, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	return NewAESUDPClientPacketProcessor(requestBlock, responseBlock, func([]byte) cipher.AEAD {
		block, err := aes.NewCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			t.Fatal(err)
		}
		return aead
	}, nil)
}

// Truncated UDP responses used to panic instead of being rejected: packets
// shorter than the 16 byte separate header panicked inside crypto/aes, and
// packets shorter than the separate header plus the AEAD overhead panicked
// while extending the destination buffer by a negative amount.
func TestDecodeUDPRespTruncatedPacket(t *testing.T) {
	processor := newTestAESUDPClientPacketProcessor(t)

	// 16 (separate header) + 16 (GCM overhead) is the smallest valid length.
	for length := 0; length < 32; length++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic decoding a %d byte packet: %v", length, r)
				}
			}()
			if err := processor.DecodeUDPResp(make([]byte, length), &UDPResponse{},
				newTestUDPCachedStateContainer()); err == nil {
				t.Errorf("expected an error decoding a %d byte packet", length)
			}
		}()
	}
}
