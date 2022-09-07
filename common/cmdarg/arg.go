package cmdarg

import (
	"bytes"
	"io"
	"io/ioutil"
)

// LoadArg loads one arg, maybe an remote url, or local file path
func LoadArg(arg string) (out io.Reader, err error) {
	bs, err := LoadArgToBytes(arg)
	if err != nil {
		return nil, err
	}
	out = bytes.NewBuffer(bs)
	return
}

// LoadArgToBytes loads one arg to []byte, maybe an remote url, or local file path
func LoadArgToBytes(arg string) (out []byte, err error) {
	out, err = ioutil.ReadFile(arg)
	if err != nil {
		return
	}
	return
}
