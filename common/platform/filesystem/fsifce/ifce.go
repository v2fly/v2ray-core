package fsifce

import (
	"io"
	"io/fs"
)

type FileSeekerFunc func(path string) (io.ReadSeekCloser, error)

type FileReaderFunc func(path string) (io.ReadCloser, error)

type FileWriterFunc func(path string) (io.WriteCloser, error)

type FileReadDirFunc func(path string) ([]fs.DirEntry, error)

type FileRemoveFunc func(path string) error
