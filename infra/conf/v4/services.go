package v4

import (
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/common/serial"
)

func (c *Config) BuildServices(service map[string]*json.RawMessage) ([]*anypb.Any, error) {
	var ret []*anypb.Any
	for k, v := range service {
		message, err := desc.LoadMessageDescriptor(k)
		if err != nil || message == nil {
			return nil, newError("Cannot find service", k, "").Base(err)
		}

		serviceConfig := dynamic.NewMessage(message)

		if err := serviceConfig.UnmarshalJSONPB(&jsonpb.Unmarshaler{AllowUnknownFields: false}, *v); err != nil {
			return nil, newError("Cannot interpret service configure file", k, "").Base(err)
		}

		ret = append(ret, serial.ToTypedMessage(serviceConfig))
	}
	return ret, nil
}
