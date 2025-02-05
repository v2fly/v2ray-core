package filesystemimpl

import (
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem/fsifce"
)

func NewDefaultFileSystemDefaultImpl() environment.FileSystemCapabilitySet {
	return fsCapImpl{}
}

type fsCapImpl struct{}

func (f fsCapImpl) OpenFileForReadSeek() fsifce.FileSeekerFunc {
	return filesystem.NewFileSeeker
}

func (f fsCapImpl) OpenFileForRead() fsifce.FileReaderFunc {
	return filesystem.NewFileReader
}

func (f fsCapImpl) OpenFileForWrite() fsifce.FileWriterFunc {
	return filesystem.NewFileWriter
}

func (f fsCapImpl) ReadDir() fsifce.FileReadDirFunc {
	return filesystem.NewFileReadDir
}

func (f fsCapImpl) RemoveFile() fsifce.FileRemoveFunc {
	return filesystem.NewFileRemover
}
