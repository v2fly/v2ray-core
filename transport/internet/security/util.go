package security

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func CreateSecurityEngineFromSettings(context context.Context, settings *internet.MemoryStreamConfig) (Engine, error) {
	if settings == nil || settings.SecurityType == "" {
		return nil, nil
	}
	securityEngine, err := common.CreateObject(context, settings.SecuritySettings)
	if err != nil {
		return nil, newError("unable to create security engine from security settings").Base(err)
	}
	securityEngineTyped, ok := securityEngine.(Engine)
	if !ok {
		return nil, newError("type assertion error when create security engine from security settings")
	}
	return securityEngineTyped, nil
}
