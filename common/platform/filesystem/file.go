package filesystem

import (
	"io"
	"os"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/platform"
)

type FileSeekerFunc func(path string) (io.ReadSeekCloser, error)

type FileReaderFunc func(path string) (io.ReadCloser, error)

type FileWriterFunc func(path string) (io.WriteCloser, error)

var NewFileSeeker FileSeekerFunc = func(path string) (io.ReadSeekCloser, error) {
	return os.Open(path)
}

var NewFileReader FileReaderFunc = func(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

var NewFileWriter FileWriterFunc = func(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func ReadFile(path string) ([]byte, error) {
	reader, err := NewFileReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return buf.ReadAllToBytes(reader)
}

func WriteFile(path string, payload []byte) error {
	writer, err := NewFileWriter(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	return buf.WriteAllBytes(writer, payload)
}

func ReadAsset(file string) ([]byte, error) {
	return ReadFile(platform.GetAssetLocation(file))
}

func CopyFile(dst string, src string) error {
	bytes, err := ReadFile(src)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes)
	return err
}
