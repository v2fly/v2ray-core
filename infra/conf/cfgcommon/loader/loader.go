package loader

import (
	"encoding/json"
	"strings"
)

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

type ConfigCreator func() interface{}

type ConfigCreatorCache map[string]ConfigCreator

func (v ConfigCreatorCache) RegisterCreator(id string, creator ConfigCreator) error {
	if _, found := v[id]; found {
		return newError(id, " already registered.").AtError()
	}

	v[id] = creator
	return nil
}

func (v ConfigCreatorCache) CreateConfig(id string) (interface{}, error) {
	creator, found := v[id]
	if !found {
		return nil, newError("unknown config id: ", id)
	}
	return creator(), nil
}

type JSONConfigLoader struct {
	cache     ConfigCreatorCache
	idKey     string
	configKey string
}

func NewJSONConfigLoader(cache ConfigCreatorCache, idKey string, configKey string) *JSONConfigLoader {
	return &JSONConfigLoader{
		idKey:     idKey,
		configKey: configKey,
		cache:     cache,
	}
}

func (v *JSONConfigLoader) LoadWithID(raw []byte, id string) (interface{}, error) {
	id = strings.ToLower(id)
	config, err := v.cache.CreateConfig(id)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (v *JSONConfigLoader) Load(raw []byte) (interface{}, string, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, "", err
	}
	rawID, found := obj[v.idKey]
	if !found {
		return nil, "", newError(v.idKey, " not found in JSON context").AtError()
	}
	var id string
	if err := json.Unmarshal(rawID, &id); err != nil {
		return nil, "", err
	}
	rawConfig := json.RawMessage(raw)
	if len(v.configKey) > 0 {
		configValue, found := obj[v.configKey]
		if found {
			rawConfig = configValue
		} else {
			// Default to empty json object.
			rawConfig = json.RawMessage([]byte("{}"))
		}
	}
	config, err := v.LoadWithID([]byte(rawConfig), id)
	if err != nil {
		return nil, id, err
	}
	return config, id, nil
}
