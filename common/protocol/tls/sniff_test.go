package tls_test

import (
	"testing"

	. "github.com/v2fly/v2ray-core/common/protocol/tls"
)

func TestTLSHeaders(t *testing.T) {
	cases := []struct {
		input  []byte
		domain string
		err    bool
	}{
		{
			input: []byte{
				0x16, 0x03, 0x01, 0x00, 0xc8, 0x01, 0x00, 0x00,
				0xc4, 0x03, 0x03, 0x1a, 0xac, 0xb2, 0xa8, 0xfe,
				0xb4, 0x96, 0x04, 0x5b, 0xca, 0xf7, 0xc1, 0xf4,
				0x2e, 0x53, 0x24, 0x6e, 0x34, 0x0c, 0x58, 0x36,
				0x71, 0x97, 0x59, 0xe9, 0x41, 0x66, 0xe2, 0x43,
				0xa0, 0x13, 0xb6, 0x00, 0x00, 0x20, 0x1a, 0x1a,
				0xc0, 0x2b, 0xc0, 0x2f, 0xc0, 0x2c, 0xc0, 0x30,
				0xcc, 0xa9, 0xcc, 0xa8, 0xcc, 0x14, 0xcc, 0x13,
				0xc0, 0x13, 0xc0, 0x14, 0x00, 0x9c, 0x00, 0x9d,
				0x00, 0x2f, 0x00, 0x35, 0x00, 0x0a, 0x01, 0x00,
				0x00, 0x7b, 0xba, 0xba, 0x00, 0x00, 0xff, 0x01,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00,
				0x14, 0x00, 0x00, 0x11, 0x63, 0x2e, 0x73, 0x2d,
				0x6d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x6f, 0x66,
				0x74, 0x2e, 0x63, 0x6f, 0x6d, 0x00, 0x17, 0x00,
				0x00, 0x00, 0x23, 0x00, 0x00, 0x00, 0x0d, 0x00,
				0x14, 0x00, 0x12, 0x04, 0x03, 0x08, 0x04, 0x04,
				0x01, 0x05, 0x03, 0x08, 0x05, 0x05, 0x01, 0x08,
				0x06, 0x06, 0x01, 0x02, 0x01, 0x00, 0x05, 0x00,
				0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x12,
				0x00, 0x00, 0x00, 0x10, 0x00, 0x0e, 0x00, 0x0c,
				0x02, 0x68, 0x32, 0x08, 0x68, 0x74, 0x74, 0x70,
				0x2f, 0x31, 0x2e, 0x31, 0x00, 0x0b, 0x00, 0x02,
				0x01, 0x00, 0x00, 0x0a, 0x00, 0x0a, 0x00, 0x08,
				0xaa, 0xaa, 0x00, 0x1d, 0x00, 0x17, 0x00, 0x18,
				0xaa, 0xaa, 0x00, 0x01, 0x00,
			},
			domain: "c.s-microsoft.com",
			err:    false,
		},
		{
			input: []byte{
				0x16, 0x03, 0x01, 0x00, 0xee, 0x01, 0x00, 0x00,
				0xea, 0x03, 0x03, 0xe7, 0x91, 0x9e, 0x93, 0xca,
				0x78, 0x1b, 0x3c, 0xe0, 0x65, 0x25, 0x58, 0xb5,
				0x93, 0xe1, 0x0f, 0x85, 0xec, 0x9a, 0x66, 0x8e,
				0x61, 0x82, 0x88, 0xc8, 0xfc, 0xae, 0x1e, 0xca,
				0xd7, 0xa5, 0x63, 0x20, 0xbd, 0x1c, 0x00, 0x00,
				0x8b, 0xee, 0x09, 0xe3, 0x47, 0x6a, 0x0e, 0x74,
				0xb0, 0xbc, 0xa3, 0x02, 0xa7, 0x35, 0xe8, 0x85,
				0x70, 0x7c, 0x7a, 0xf0, 0x00, 0xdf, 0x4a, 0xea,
				0x87, 0x01, 0x14, 0x91, 0x00, 0x20, 0xea, 0xea,
				0xc0, 0x2b, 0xc0, 0x2f, 0xc0, 0x2c, 0xc0, 0x30,
				0xcc, 0xa9, 0xcc, 0xa8, 0xcc, 0x14, 0xcc, 0x13,
				0xc0, 0x13, 0xc0, 0x14, 0x00, 0x9c, 0x00, 0x9d,
				0x00, 0x2f, 0x00, 0x35, 0x00, 0x0a, 0x01, 0x00,
				0x00, 0x81, 0x9a, 0x9a, 0x00, 0x00, 0xff, 0x01,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x18, 0x00,
				0x16, 0x00, 0x00, 0x13, 0x77, 0x77, 0x77, 0x30,
				0x37, 0x2e, 0x63, 0x6c, 0x69, 0x63, 0x6b, 0x74,
				0x61, 0x6c, 0x65, 0x2e, 0x6e, 0x65, 0x74, 0x00,
				0x17, 0x00, 0x00, 0x00, 0x23, 0x00, 0x00, 0x00,
				0x0d, 0x00, 0x14, 0x00, 0x12, 0x04, 0x03, 0x08,
				0x04, 0x04, 0x01, 0x05, 0x03, 0x08, 0x05, 0x05,
				0x01, 0x08, 0x06, 0x06, 0x01, 0x02, 0x01, 0x00,
				0x05, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x12, 0x00, 0x00, 0x00, 0x10, 0x00, 0x0e,
				0x00, 0x0c, 0x02, 0x68, 0x32, 0x08, 0x68, 0x74,
				0x74, 0x70, 0x2f, 0x31, 0x2e, 0x31, 0x75, 0x50,
				0x00, 0x00, 0x00, 0x0b, 0x00, 0x02, 0x01, 0x00,
				0x00, 0x0a, 0x00, 0x0a, 0x00, 0x08, 0x9a, 0x9a,
				0x00, 0x1d, 0x00, 0x17, 0x00, 0x18, 0x8a, 0x8a,
				0x00, 0x01, 0x00,
			},
			domain: "www07.clicktale.net",
			err:    false,
		},
		{
			input: []byte{
				0x16, 0x03, 0x01, 0x00, 0xe6, 0x01, 0x00, 0x00, 0xe2, 0x03, 0x03, 0x81, 0x47, 0xc1,
				0x66, 0xd5, 0x1b, 0xfa, 0x4b, 0xb5, 0xe0, 0x2a, 0xe1, 0xa7, 0x87, 0x13, 0x1d, 0x11, 0xaa, 0xc6,
				0xce, 0xfc, 0x7f, 0xab, 0x94, 0xc8, 0x62, 0xad, 0xc8, 0xab, 0x0c, 0xdd, 0xcb, 0x20, 0x6f, 0x9d,
				0x07, 0xf1, 0x95, 0x3e, 0x99, 0xd8, 0xf3, 0x6d, 0x97, 0xee, 0x19, 0x0b, 0x06, 0x1b, 0xf4, 0x84,
				0x0b, 0xb6, 0x8f, 0xcc, 0xde, 0xe2, 0xd0, 0x2d, 0x6b, 0x0c, 0x1f, 0x52, 0x53, 0x13, 0x00, 0x08,
				0x13, 0x02, 0x13, 0x03, 0x13, 0x01, 0x00, 0xff, 0x01, 0x00, 0x00, 0x91, 0x00, 0x00, 0x00, 0x0c,
				0x00, 0x0a, 0x00, 0x00, 0x07, 0x64, 0x6f, 0x67, 0x66, 0x69, 0x73, 0x68, 0x00, 0x0b, 0x00, 0x04,
				0x03, 0x00, 0x01, 0x02, 0x00, 0x0a, 0x00, 0x0c, 0x00, 0x0a, 0x00, 0x1d, 0x00, 0x17, 0x00, 0x1e,
				0x00, 0x19, 0x00, 0x18, 0x00, 0x23, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x17, 0x00, 0x00,
				0x00, 0x0d, 0x00, 0x1e, 0x00, 0x1c, 0x04, 0x03, 0x05, 0x03, 0x06, 0x03, 0x08, 0x07, 0x08, 0x08,
				0x08, 0x09, 0x08, 0x0a, 0x08, 0x0b, 0x08, 0x04, 0x08, 0x05, 0x08, 0x06, 0x04, 0x01, 0x05, 0x01,
				0x06, 0x01, 0x00, 0x2b, 0x00, 0x07, 0x06, 0x7f, 0x1c, 0x7f, 0x1b, 0x7f, 0x1a, 0x00, 0x2d, 0x00,
				0x02, 0x01, 0x01, 0x00, 0x33, 0x00, 0x26, 0x00, 0x24, 0x00, 0x1d, 0x00, 0x20, 0x2f, 0x35, 0x0c,
				0xb6, 0x90, 0x0a, 0xb7, 0xd5, 0xc4, 0x1b, 0x2f, 0x60, 0xaa, 0x56, 0x7b, 0x3f, 0x71, 0xc8, 0x01,
				0x7e, 0x86, 0xd3, 0xb7, 0x0c, 0x29, 0x1a, 0x9e, 0x5b, 0x38, 0x3f, 0x01, 0x72,
			},
			domain: "dogfish",
			err:    false,
		},
		{
			input: []byte{
				0x16, 0x03, 0x01, 0x01, 0x03, 0x01, 0x00, 0x00,
				0xff, 0x03, 0x03, 0x3d, 0x89, 0x52, 0x9e, 0xee,
				0xbe, 0x17, 0x63, 0x75, 0xef, 0x29, 0xbd, 0x14,
				0x6a, 0x49, 0xe0, 0x2c, 0x37, 0x57, 0x71, 0x62,
				0x82, 0x44, 0x94, 0x8f, 0x6e, 0x94, 0x08, 0x45,
				0x7f, 0xdb, 0xc1, 0x00, 0x00, 0x3e, 0xc0, 0x2c,
				0xc0, 0x30, 0x00, 0x9f, 0xcc, 0xa9, 0xcc, 0xa8,
				0xcc, 0xaa, 0xc0, 0x2b, 0xc0, 0x2f, 0x00, 0x9e,
				0xc0, 0x24, 0xc0, 0x28, 0x00, 0x6b, 0xc0, 0x23,
				0xc0, 0x27, 0x00, 0x67, 0xc0, 0x0a, 0xc0, 0x14,
				0x00, 0x39, 0xc0, 0x09, 0xc0, 0x13, 0x00, 0x33,
				0x00, 0x9d, 0x00, 0x9c, 0x13, 0x02, 0x13, 0x03,
				0x13, 0x01, 0x00, 0x3d, 0x00, 0x3c, 0x00, 0x35,
				0x00, 0x2f, 0x00, 0xff, 0x01, 0x00, 0x00, 0x98,
				0x00, 0x00, 0x00, 0x10, 0x00, 0x0e, 0x00, 0x00,
				0x0b, 0x31, 0x30, 0x2e, 0x34, 0x32, 0x2e, 0x30,
				0x2e, 0x32, 0x34, 0x33, 0x00, 0x0b, 0x00, 0x04,
				0x03, 0x00, 0x01, 0x02, 0x00, 0x0a, 0x00, 0x0a,
				0x00, 0x08, 0x00, 0x1d, 0x00, 0x17, 0x00, 0x19,
				0x00, 0x18, 0x00, 0x23, 0x00, 0x00, 0x00, 0x0d,
				0x00, 0x20, 0x00, 0x1e, 0x04, 0x03, 0x05, 0x03,
				0x06, 0x03, 0x08, 0x04, 0x08, 0x05, 0x08, 0x06,
				0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x02, 0x03,
				0x02, 0x01, 0x02, 0x02, 0x04, 0x02, 0x05, 0x02,
				0x06, 0x02, 0x00, 0x16, 0x00, 0x00, 0x00, 0x17,
				0x00, 0x00, 0x00, 0x2b, 0x00, 0x09, 0x08, 0x7f,
				0x14, 0x03, 0x03, 0x03, 0x02, 0x03, 0x01, 0x00,
				0x2d, 0x00, 0x03, 0x02, 0x01, 0x00, 0x00, 0x28,
				0x00, 0x26, 0x00, 0x24, 0x00, 0x1d, 0x00, 0x20,
				0x13, 0x7c, 0x6e, 0x97, 0xc4, 0xfd, 0x09, 0x2e,
				0x70, 0x2f, 0x73, 0x5a, 0x9b, 0x57, 0x4d, 0x5f,
				0x2b, 0x73, 0x2c, 0xa5, 0x4a, 0x98, 0x40, 0x3d,
				0x75, 0x6e, 0xb4, 0x76, 0xf9, 0x48, 0x8f, 0x36,
			},
			domain: "10.42.0.243",
			err:    false,
		},
	}

	for _, test := range cases {
		header, err := SniffTLS(test.input)
		if test.err {
			if err == nil {
				t.Errorf("Exepct error but nil in test %v", test)
			}
		} else {
			if err != nil {
				t.Errorf("Expect no error but actually %s in test %v", err.Error(), test)
			}
			if header.Domain() != test.domain {
				t.Error("expect domain ", test.domain, " but got ", header.Domain())
			}
		}
	}
}
