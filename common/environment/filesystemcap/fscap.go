package filesystemcap

import "github.com/v2fly/v2ray-core/v5/common/platform/filesystem/fsifce"

type FileSystemCapabilitySet interface {
	OpenFileForReadSeek() fsifce.FileSeekerFunc
	OpenFileForRead() fsifce.FileReaderFunc
	OpenFileForWrite() fsifce.FileWriterFunc
}
