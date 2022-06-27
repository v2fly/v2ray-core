package fsifce

import "io"

type FileSeekerFunc func(path string) (io.ReadSeekCloser, error)

type FileReaderFunc func(path string) (io.ReadCloser, error)

type FileWriterFunc func(path string) (io.WriteCloser, error)
