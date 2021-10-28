package jsonpb

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/common/serial"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func loadJsonPb(data io.Reader) (*core.Config, error) {
	coreconf := &core.Config{}
	jsonpbloader := &jsonpb.Unmarshaler{AnyResolver: serial.GetResolver()}
	err := jsonpbloader.Unmarshal(data, coreconf)
	if err != nil {
		return nil, err
	}
	return coreconf, nil
}

func dumpJsonPb(config proto.Message, w io.Writer) error {
	jsonpbdumper := &jsonpb.Marshaler{AnyResolver: serial.GetResolver()}
	err := jsonpbdumper.Marshal(w, config)
	if err != nil {
		return err
	}
	return nil
}

func DumpJsonPb(config proto.Message, w io.Writer) error {
	return dumpJsonPb(config, w)
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
				return loadJsonPb(bytes.NewReader(data))
			case io.Reader:
				data, err := buf.ReadAllToBytes(v)
				if err != nil {
					return nil, err
				}
				return loadJsonPb(bytes.NewReader(data))
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
