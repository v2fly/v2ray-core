package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem/fsifce"
)

var NewFileSeeker fsifce.FileSeekerFunc = func(path string) (io.ReadSeekCloser, error) {
	return os.Open(path)
}

var NewFileReader fsifce.FileReaderFunc = func(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

var NewFileWriter fsifce.FileWriterFunc = func(path string) (io.WriteCloser, error) {
	basePath := filepath.Dir(path)
	if err := os.MkdirAll(basePath, 0o700); err != nil {
		return nil, err
	}
	return os.Create(path)
}

var NewFileRemover fsifce.FileRemoveFunc = os.Remove

var NewFileReadDir fsifce.FileReadDirFunc = os.ReadDir

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

func CopyFile(dst string, src string, perm os.FileMode) error {
	bytes, err := ReadFile(src)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes)
	return err
}
