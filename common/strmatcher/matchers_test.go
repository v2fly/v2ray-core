package strmatcher_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/v2fly/v2ray-core/v5/common"
	. "github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

func TestMatcher(t *testing.T) {
	cases := []struct {
		pattern string
		mType   Type
		input   string
		output  bool
	}{
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "www.v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "www.v3fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "2fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "xv2fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Full,
			input:   "v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Full,
			input:   "xv2fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Regex,
			input:   "v2flyxorg",
			output:  true,
		},
	}
	for _, test := range cases {
		matcher, err := test.mType.New(test.pattern)
		common.Must(err)
		if m := matcher.Match(test.input); m != test.output {
			t.Error("unexpected output: ", m, " for test case ", test)
		}
	}
}

func TestToDomain(t *testing.T) {
	{ // Test normal ASCII domain, which should not trigger new string data allocation
		input := "v2fly.org"
		domain, err := ToDomain(input)
		if err != nil {
			t.Error("unexpected error: ", err)
		}
		if domain != input {
			t.Error("unexpected output: ", domain, " for test case ", input)
		}
		if (*reflect.StringHeader)(unsafe.Pointer(&input)).Data != (*reflect.StringHeader)(unsafe.Pointer(&domain)).Data {
			t.Error("different string data of output: ", domain, " and test case ", input)
		}
	}
	{ // Test ASCII domain containing upper case letter, which should be converted to lower case
		input := "v2FLY.oRg"
		domain, err := ToDomain(input)
		if err != nil {
			t.Error("unexpected error: ", err)
		}
		if domain != "v2fly.org" {
			t.Error("unexpected output: ", domain, " for test case ", input)
		}
	}
	{ // Test internationalized domain, which should be translated to ASCII punycode
		input := "v2fly.公益"
		domain, err := ToDomain(input)
		if err != nil {
			t.Error("unexpected error: ", err)
		}
		if domain != "v2fly.xn--55qw42g" {
			t.Error("unexpected output: ", domain, " for test case ", input)
		}
	}
	{ // Test internationalized domain containing upper case letter
		input := "v2FLY.公益"
		domain, err := ToDomain(input)
		if err != nil {
			t.Error("unexpected error: ", err)
		}
		if domain != "v2fly.xn--55qw42g" {
			t.Error("unexpected output: ", domain, " for test case ", input)
		}
	}
	{ // Test domain name of invalid character, which should return with error
		input := "{"
		_, err := ToDomain(input)
		if err == nil {
			t.Error("unexpected non error for test case ", input)
		}
	}
	{ // Test domain name containing a space, which should return with error
		input := "Mijia Cloud"
		_, err := ToDomain(input)
		if err == nil {
			t.Error("unexpected non error for test case ", input)
		}
	}
	{ // Test domain name containing an underscore, which should return with error
		input := "Mijia_Cloud.com"
		_, err := ToDomain(input)
		if err == nil {
			t.Error("unexpected non error for test case ", input)
		}
	}
	{ // Test internationalized domain containing invalid character
		input := "Mijia Cloud.公司"
		_, err := ToDomain(input)
		if err == nil {
			t.Error("unexpected non error for test case ", input)
		}
	}
}
