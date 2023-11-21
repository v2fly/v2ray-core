package jsonfieldarray

import (
	"encoding/json"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/common"
)

// NewJSONFieldArrayParser internal api
func NewJSONFieldArrayParser() containers.SubscriptionContainerDocumentParser {
	return newJSONFieldArrayParser()
}

func newJSONFieldArrayParser() containers.SubscriptionContainerDocumentParser {
	return &parser{}
}

type parser struct{}

type jsonDocument map[string]json.RawMessage

func (p parser) ParseSubscriptionContainerDocument(rawConfig []byte) (*containers.Container, error) {
	result := &containers.Container{}
	result.Kind = "JsonFieldArray"
	result.Metadata = make(map[string]string)

	var doc jsonDocument
	if err := json.Unmarshal(rawConfig, &doc); err != nil {
		return nil, newError("failed to parse as json").Base(err)
	}

	for key, value := range doc {
		switch value[0] {
		case '[':
			parsedArray, err := p.parseArray(value, "JsonFieldArray+"+key)
			if err != nil {
				return nil, newError("failed to parse as json array").Base(err)
			}
			result.ServerSpecs = append(result.ServerSpecs, parsedArray...)
		case '{':
			fallthrough
		default:
			result.Metadata[key] = string(value)
		}
	}

	return result, nil
}

func (p parser) parseArray(rawConfig []byte, kindHint string) ([]containers.UnparsedServerConf, error) {
	var result []json.RawMessage
	if err := json.Unmarshal(rawConfig, &result); err != nil {
		return nil, newError("failed to parse as json array").Base(err)
	}
	var ret []containers.UnparsedServerConf
	for _, value := range result {
		ret = append(ret, containers.UnparsedServerConf{
			KindHint: kindHint,
			Content:  []byte(value),
		})
	}
	return ret, nil
}

func init() {
	common.Must(containers.RegisterParser("JsonFieldArray", newJSONFieldArrayParser()))
}
