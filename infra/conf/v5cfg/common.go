package v5cfg

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/environment/envimpl"
	"github.com/v2fly/v2ray-core/v5/common/registry"
)

func loadHeterogeneousConfigFromRawJSON(interfaceType, name string, rawJSON json.RawMessage) (proto.Message, error) {
	fsdef := envimpl.NewDefaultFileSystemDefaultImpl()
	ctx := envctx.ContextWithEnvironment(context.TODO(), fsdef)
	if len(rawJSON) == 0 {
		rawJSON = []byte("{}")
	}
	return registry.LoadImplementationByAlias(ctx, interfaceType, name, []byte(rawJSON))
}

// LoadHeterogeneousConfigFromRawJSON private API
func LoadHeterogeneousConfigFromRawJSON(ctx context.Context, interfaceType, name string, rawJSON json.RawMessage) (proto.Message, error) {
	return loadHeterogeneousConfigFromRawJSON(interfaceType, name, rawJSON)
}
