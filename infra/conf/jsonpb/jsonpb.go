package jsonpb

import (
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
	"github.com/v2fly/v2ray-core/v5/common/serial"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func loadJSONPB(data []byte) (*core.Config, error) {
	coreconf := &core.Config{}
	jsonpbloader := protojson.UnmarshalOptions{Resolver: serial.GetResolver()}
	err := jsonpbloader.Unmarshal(data, coreconf)
	if err != nil {
		return nil, err
	}
	return coreconf, nil
}

func dumpJSONPb(config proto.Message, w io.Writer) error {
	jsonpbdumper := protojson.MarshalOptions{Resolver: serial.GetResolver()}
	data, err := jsonpbdumper.Marshal(config)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func DumpJSONPb(config proto.Message, w io.Writer) error {
	return dumpJSONPb(config, w)
}

const FormatProtobufJSONPB = "jsonpb"

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{FormatProtobufJSONPB},
		Extension: []string{".pb.json", ".pbjson"},
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
				return loadJSONPB(data)
			case []byte:
				return loadJSONPB(v)
			case io.Reader:
				data, err := buf.ReadAllToBytes(v)
				if err != nil {
					return nil, err
				}
				return loadJSONPB(data)
			default:
				return nil, newError("unknown type")
			}
		},
	}))
}
