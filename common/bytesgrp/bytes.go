package bytesgrp

import (
	"encoding/binary"
)

func Pack(data [][]byte) []byte {
	var merged []byte
	for _, b := range data {
		length := make([]byte, 8)
		offset := binary.PutUvarint(length, uint64(len(b)))
		merged = append(merged, length[0:offset]...)
		merged = append(merged, b...)
	}
	return merged
}

func UnPack(pack []byte) [][]byte {
	var data [][]byte
	dataLength := len(pack)
	index := 0
	for index < dataLength {
		length, skip := binary.Uvarint(pack[index:])
		index += skip + int(length)
		if dataLength-index < 0 {
			// err
			return [][]byte{pack}
		}
		data = append(data, pack[index-int(length):index])
	}
	return data
}
