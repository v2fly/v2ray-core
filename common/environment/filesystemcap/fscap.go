package filesystemcap

import "github.com/ghxhy/v2ray-core/v5/common/platform/filesystem/fsifce"

type FileSystemCapabilitySet interface {
	OpenFileForReadSeek() fsifce.FileSeekerFunc
	OpenFileForRead() fsifce.FileReaderFunc
	OpenFileForWrite() fsifce.FileWriterFunc
	ReadDir() fsifce.FileReadDirFunc
	RemoveFile() fsifce.FileRemoveFunc
}
