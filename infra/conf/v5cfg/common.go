package v5cfg

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/registry"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"strings"
)

func loadHeterogeneousConfigFromRawJson(interfaceType, name string, rawJson json.RawMessage) (proto.Message, error) {
	var implementationFullName string
	if strings.HasPrefix(name, "#") {
		// skip resolution for full name
		implementationFullName = name
	} else {
		registryResult, err := registry.FindImplementationByAlias(interfaceType, name)
		if err != nil {
			return nil, newError("unable to find implementation").Base(err)
		}
		implementationFullName = registryResult
	}
	implementationConfigInstance, err := serial.GetInstance(implementationFullName)
	if err != nil {
		return nil, newError("unable to create implementation config instance").Base(err)
	}

	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: false}
	err = unmarshaler.Unmarshal(bytes.NewReader([]byte(rawJson)), implementationConfigInstance.(proto.Message))
	if err != nil {
		return nil, newError("unable to parse json content").Base(err)
	}

	return implementationConfigInstance.(proto.Message), nil
}
