package v2jsonpb

import (
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"io"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func loadV2JsonPb(data []byte) (*core.Config, error) {
	coreconf := &core.Config{}
	jsonpbloader := &protojson.UnmarshalOptions{}
	err := jsonpbloader.Unmarshal(data, coreconf)
	if err != nil {
		return nil, err
	}
	return coreconf, nil
}

func dumpV2JsonPb(config proto.Message) ([]byte, error) {
	jsonpbdumper := &protojson.MarshalOptions{Resolver: resolver2{serial.GetResolver()}, AllowPartial: true}
	bytew, err := jsonpbdumper.Marshal(&V2JsonProtobufFollower{config.ProtoReflect()})
	if err != nil {
		return nil, err
	}
	return bytew, nil
}

func DumpV2JsonPb(config proto.Message) ([]byte, error) {
	return dumpV2JsonPb(config)
}

const FormatProtobufV2JSONPB = "v2jsonpb"

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{FormatProtobufV2JSONPB},
		Extension: []string{".v2pb.json", ".v2pbjson"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case string:
				r, err := cmdarg.LoadArg(v)
				if err != nil {
					return nil, err
				}
				data, err := buf.ReadAllToBytes(r)
				if err != nil {
					return nil, err
				}
				return loadV2JsonPb(data)
			case io.Reader:
				data, err := buf.ReadAllToBytes(v)
				if err != nil {
					return nil, err
				}
				return loadV2JsonPb(data)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
