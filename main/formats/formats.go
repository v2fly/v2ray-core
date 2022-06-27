package formats

import (
	"bytes"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v5/infra/conf/mergers"
	"github.com/v2fly/v2ray-core/v5/infra/conf/serial"
)

func init() {
	for _, formatName := range mergers.GetAllNames() {
		loader, err := makeMergeLoader(formatName)
		if err != nil {
			panic(err)
		}
		if formatName == core.FormatAuto {
			loader.Extension = nil
		}
		common.Must(core.RegisterConfigLoader(loader))
	}
}

func makeMergeLoader(formatName string) (*core.ConfigFormat, error) {
	extensions, err := mergers.GetExtensions(formatName)
	if err != nil {
		return nil, err
	}
	return &core.ConfigFormat{
		Name:      []string{formatName},
		Extension: extensions,
		Loader:    makeLoaderFunc(formatName),
	}, nil
}

func makeLoaderFunc(formatName string) core.ConfigLoader {
	return func(input interface{}) (*core.Config, error) {
		m := make(map[string]interface{})
		err := mergers.MergeAs(formatName, input, m)
		if err != nil {
			return nil, err
		}
		data, err := merge.FromMap(m)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(data)
		cf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			return nil, err
		}
		return cf.Build()
	}
}
