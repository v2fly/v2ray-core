package bytesgrp

import (
	"bytes"
	"testing"
)

var data = [][]byte{{1, 1, 4}, {5, 1, 4}, {1, 9}, {1, 9, 8}, {1, 0}}

func TestGroup(t *testing.T) {
	test := UnPack(Pack(data))

	for i, b := range test {
		if !bytes.Equal(data[i], b) {
			t.Error("encode failed")
		}
	}
}
