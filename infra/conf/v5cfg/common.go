package v5cfg

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v4/common/environment/envimpl"
	"github.com/v2fly/v2ray-core/v4/common/registry"
)

func loadHeterogeneousConfigFromRawJson(interfaceType, name string, rawJson json.RawMessage) (proto.Message, error) {
	fsdef := envimpl.NewDefaultFileSystemDefaultImpl()
	ctx := envctx.ContextWithEnvironment(context.TODO(), fsdef)
	if rawJson == nil || len(rawJson) == 0 {
		rawJson = []byte("{}")
	}
	return registry.LoadImplementationByAlias(ctx, interfaceType, name, []byte(rawJson))
}

// LoadHeterogeneousConfigFromRawJson private API
func LoadHeterogeneousConfigFromRawJson(ctx context.Context, interfaceType, name string, rawJson json.RawMessage) (proto.Message, error) {
	return loadHeterogeneousConfigFromRawJson(interfaceType, name, rawJson)
}
