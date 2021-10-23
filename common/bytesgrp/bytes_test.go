package bytesgrp_test

import (
	"bytes"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common/bytesgrp"
)

func TestGroup(t *testing.T) {
	data := [][]byte{{1, 1, 4}, {5, 1, 4}, {1, 9}, {1, 9, 8}, {1, 0}}

	test := bytesgrp.UnPack(bytesgrp.Pack(data))

	for i, b := range test {
		if !bytes.Equal(data[i], b) {
			t.Error("encode failed")
		}
	}
}
