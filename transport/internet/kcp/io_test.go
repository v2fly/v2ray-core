package kcp_test

import (
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/transport/internet/kcp"
)

func TestKCPPacketReader(t *testing.T) {
	reader := KCPPacketReader{
		Security: &SimpleAuthenticator{},
	}

	testCases := []struct {
		Input  []byte
		Output []Segment
	}{
		{
			Input:  []byte{},
			Output: nil,
		},
		{
			Input:  []byte{1},
			Output: nil,
		},
	}

	for _, testCase := range testCases {
		seg := reader.Read(testCase.Input)
		if testCase.Output == nil && seg != nil {
			t.Errorf("Expect nothing returned, but actually %v", seg)
		} else if testCase.Output != nil && seg == nil {
			t.Errorf("Expect some output, but got nil")
		}
	}
}
