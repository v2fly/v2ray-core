package server

// this file defines methods to pad packets
// with a given number of bytes, and to unpack the padding from a padded packet.
// The packet format is as follows if the desired output length is greater than
// 4 bytes:
// | data | padding | data length |
// The data length is a 32-bit big-endian integer that represents the length of
// the data in bytes.
// If the desired output length is 4 bytes or less, the packet format is as
// follows:
// | padding |
// No payload will be included in the packet.

import "encoding/binary"

// Pack pads the given data with the given number of bytes, and appends the
// length of the data to the end of the data. The returned byte slice
// contains the padded data.
// This generates a packet with a length of
// len(data_OWNERSHIP_RELINQUISHED) + padding + 2
// @param data_OWNERSHIP_RELINQUISHED - The payload, this reference is consumed and should not be used after this call.
// @param padding - The number of padding bytes to add to the data.
func Pack(data_OWNERSHIP_RELINQUISHED []byte, paddingLength int) []byte {
	data := append(data_OWNERSHIP_RELINQUISHED, make([]byte, paddingLength)...)
	dataLength := len(data_OWNERSHIP_RELINQUISHED)
	data = binary.BigEndian.AppendUint32(data, uint32(dataLength))
	return data
}

// Pad returns a padding packet of padding length.
// If the padding length is less than 0, nil is returned.
// @param padding - The number of padding bytes to add to the data.
func Pad(paddingLength int) []byte {
	if assertPaddingLengthIsNotNegative := paddingLength < 0; assertPaddingLengthIsNotNegative {
		return nil
	}
	switch paddingLength {
	case 0:
		return []byte{}
	case 1:
		return []byte{0}
	case 2:
		return []byte{0, 0}
	case 3:
		return []byte{0, 0, 0}
	case 4:
		return []byte{0, 0, 0, 0}
	default:
		return append(make([]byte, paddingLength-4), byte(paddingLength>>24), byte(paddingLength>>16), byte(paddingLength>>8), byte(paddingLength))
	}

}

// Unpack extracts the data and padding from the given padded data. It
// returns the data and the number of padding bytes.
// the data may be nil.
// @param wrappedData_OWNERSHIP_RELINQUISHED - The packet, this reference is consumed and should not be used after this call.
func Unpack(wrappedData_OWNERSHIP_RELINQUISHED []byte) ([]byte, int) {
	dataLength := len(wrappedData_OWNERSHIP_RELINQUISHED)
	if dataLength < 4 {
		return nil, dataLength
	}

	dataLen := int(binary.BigEndian.Uint32(wrappedData_OWNERSHIP_RELINQUISHED[dataLength-4:]))
	if dataLen > len(wrappedData_OWNERSHIP_RELINQUISHED)-4 {
		return nil, 0
	}
	paddingLength := dataLength - dataLen - 4
	if paddingLength < 0 {
		return nil, paddingLength
	}

	return wrappedData_OWNERSHIP_RELINQUISHED[:dataLen], paddingLength
}
