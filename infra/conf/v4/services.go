package v4

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/common/serial"
)

func (c *Config) BuildServices(service map[string]*json.RawMessage) ([]*anypb.Any, error) {
	var ret []*anypb.Any
	for k, v := range service {
		mt, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(k))
		if err != nil {
			return nil, newError("Cannot find service", k, "").Base(err)
		}

		message, ok := mt.(protoreflect.MessageDescriptor)
		if !ok {
			return nil, newError("Cannot find service", k, "").Base(err)
		}

		serviceConfig := dynamicpb.NewMessage(message)

		unmarshalOpt := protojson.UnmarshalOptions{DiscardUnknown: false}

		if err := unmarshalOpt.Unmarshal(*v, serviceConfig); err != nil {
			return nil, newError("Cannot interpret service configure file", k, "").Base(err)
		}

		ret = append(ret, serial.ToTypedMessage(serviceConfig))
	}
	return ret, nil
}
