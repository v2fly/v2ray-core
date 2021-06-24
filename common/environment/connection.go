package environment

import "github.com/v2fly/v2ray-core/v4/common/log"

type ConnectionCapabilitySet interface {
	ConnectionLogCapabilitySet
}

type ConnectionEnvironment interface {
	ConnectionCapabilitySet

	doNotImpl()
}

type ConnectionLogCapabilitySet interface {
	RecordConnectionLog(msg log.Message)
}
