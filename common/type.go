package common

import (
	"context"
	"reflect"

	"github.com/ghxhy/v2ray-core/v5/common/registry"
)

// ConfigCreator is a function to create an object by a config.
type ConfigCreator func(ctx context.Context, config interface{}) (interface{}, error)

var typeCreatorRegistry = make(map[reflect.Type]ConfigCreator)

// RegisterConfig registers a global config creator. The config can be nil but must have a type.
func RegisterConfig(config interface{}, configCreator ConfigCreator) error {
	configType := reflect.TypeOf(config)
	if _, found := typeCreatorRegistry[configType]; found {
		return newError(configType.Name() + " is already registered").AtError()
	}
	typeCreatorRegistry[configType] = configCreator

	registry.RegisterImplementation(config, nil)
	return nil
}

// CreateObject creates an object by its config. The config type must be registered through RegisterConfig().
func CreateObject(ctx context.Context, config interface{}) (interface{}, error) {
	configType := reflect.TypeOf(config)
	creator, found := typeCreatorRegistry[configType]
	if !found {
		return nil, newError(configType.String() + " is not registered").AtError()
	}
	return creator(ctx, config)
}
