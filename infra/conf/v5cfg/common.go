package v5cfg

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/registry"
)

func loadHeterogeneousConfigFromRawJson(interfaceType, name string, rawJson json.RawMessage) (proto.Message, error) {
	return registry.LoadImplementationByAlias(interfaceType, name, []byte(rawJson))
}
