package conf

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/v2fly/v2ray-core/v4/common/serial"
)

func (c *Config) BuildServices(service map[string]*json.RawMessage) ([]*serial.TypedMessage, error) {
	var ret []*serial.TypedMessage
	for k, v := range service {
		message, err := desc.LoadMessageDescriptor(k)
		if err != nil {
			return nil, newError("Cannot find service", k, "").Base(err)
		}

		serviceConfig := dynamic.NewMessage(message)

		if err := jsonpb.Unmarshal(bytes.NewReader(*v), serviceConfig); err != nil {
			return nil, newError("Cannot interpret service configure file", k, "").Base(err)
		}

		ret = append(ret, serial.ToTypedMessage(serviceConfig))
	}
	return ret, nil
}
