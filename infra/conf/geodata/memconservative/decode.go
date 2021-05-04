package memconservative

import (
	"errors"
	"io"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
)

var (
	errFailedToReadBytes            = errors.New("failed to read bytes")
	errFailedToReadExpectedLenBytes = errors.New("failed to read expected length of bytes")
	errInvalidGeodataFile           = errors.New("invalid geodata file")
	errInvalidGeodataVarintLength   = errors.New("invalid geodata varint length")
	errCodeNotFound                 = errors.New("code not found")
)

func emitBytes(f io.ReadSeeker, code string) ([]byte, error) {
	count := 1
	isInner := false
	tempContainer := make([]byte, 0, 5)

	var result []byte
	var advancedN uint64 = 1
	var geoDataVarintLength, codeVarintLength, varintLenByteLen uint64 = 0, 0, 0

Loop:
	for {
		container := make([]byte, advancedN)
		bytesRead, err := f.Read(container)
		if err == io.EOF {
			return nil, errCodeNotFound
		}
		if err != nil {
			return nil, errFailedToReadBytes
		}
		if bytesRead != len(container) {
			return nil, errFailedToReadExpectedLenBytes
		}

		switch count {
		case 1, 3: // data type ((field_number << 3) | wire_type)
			if container[0] != 10 { // byte `0A` equals to `10` in decimal
				return nil, errInvalidGeodataFile
			}
			advancedN = 1
			count++
		case 2, 4: // data length
			tempContainer = append(tempContainer, container...)
			if container[0] > 127 { // max one-byte-length byte `7F`(0FFF FFFF) equals to `127` in decimal
				advancedN = 1
				goto Loop
			}
			lenVarint, n := protowire.ConsumeVarint(tempContainer)
			if n < 0 {
				return nil, errInvalidGeodataVarintLength
			}
			tempContainer = nil
			if !isInner {
				isInner = true
				geoDataVarintLength = lenVarint
				advancedN = 1
			} else {
				isInner = false
				codeVarintLength = lenVarint
				varintLenByteLen = uint64(n)
				advancedN = codeVarintLength
			}
			count++
		case 5: // data value
			if strings.EqualFold(string(container), code) {
				count++
				offset := -(1 + int64(varintLenByteLen) + int64(codeVarintLength))
				f.Seek(offset, 1)               // back to the start of GeoIP or GeoSite varint
				advancedN = geoDataVarintLength // the number of bytes to be read in next round
			} else {
				count = 1
				offset := int64(geoDataVarintLength) - int64(codeVarintLength) - int64(varintLenByteLen) - 1
				f.Seek(offset, 1) // skip the unmatched GeoIP or GeoSite varint
				advancedN = 1     // the next round will be the start of another GeoIPList or GeoSiteList
			}
		case 6: // matched GeoIP or GeoSite varint
			result = container
			break Loop
		}
	}
	return result, nil
}

func Decode(filename, code string) ([]byte, error) {
	f, err := filesystem.NewFileSeeker(filename)
	if err != nil {
		return nil, newError("failed to open file: ", filename).Base(err)
	}
	defer f.Close()

	geoBytes, err := emitBytes(f, code)
	if err != nil {
		return nil, err
	}
	return geoBytes, nil
}
