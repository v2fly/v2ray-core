package filesystem

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"

	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/platform"
)

// FileReaderFunc defines a function type of FileReader
type FileReaderFunc func(path string) (io.ReadCloser, error)

// NewFileReader returns a FileReader
var NewFileReader FileReaderFunc = func(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// ReadFile reads a file from path
func ReadFile(path string) ([]byte, error) {
	reader, err := NewFileReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return buf.ReadAllToBytes(reader)
}

// ReadGzipFile reads a gzip file from path
func ReadGzipFile(path string) ([]byte, error) {
	reader, err := NewFileReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	gunzipBytes, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	return gunzipBytes, nil
}

// ReadAsset reads geofiles
func ReadAsset(file string) ([]byte, error) {
	return ReadGzipFile(platform.GetAssetLocation(file))
}

// CopyFile copies a file from src to dst
func CopyFile(dst string, src string) error {
	bytes, err := ReadFile(src)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes)
	return err
}
