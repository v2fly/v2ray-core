package environment

import (
	"github.com/v2fly/v2ray-core/v4/common/log"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/tagged"
)

type BaseEnvironmentCapabilitySet interface {
	FeaturesLookupCapabilitySet
	LogCapabilitySet
}

type BaseEnvironment interface {
	BaseEnvironmentCapabilitySet
	doNotImpl()
}

type SystemNetworkCapabilitySet interface {
	Dialer() internet.SystemDialer
	Listener() internet.SystemListener
}

type InstanceNetworkCapabilitySet interface {
	OutboundDialer() tagged.DialFunc
}

type FeaturesLookupCapabilitySet interface {
	RequireFeatures(callback interface{}) error
}

type LogCapabilitySet interface {
	RecordLog(msg log.Message)
}

type FileSystemCapabilitySet interface {
	OpenFileForReadSeek() filesystem.FileSeekerFunc
	OpenFileForRead() filesystem.FileReaderFunc
	OpenFileForWrite() filesystem.FileWriterFunc
}
