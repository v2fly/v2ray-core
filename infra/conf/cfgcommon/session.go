package cfgcommon

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
)

type configureLoadingContext int

const confContextKey = configureLoadingContext(1)

type configureLoadingEnvironment struct {
	geoLoader geodata.Loader
}

func (c *configureLoadingEnvironment) GetGeoLoader() geodata.Loader {
	if c.geoLoader == nil {
		var err error
		c.geoLoader, err = geodata.GetGeoDataLoader("standard")
		common.Must(err)
	}
	return c.geoLoader
}

func (c *configureLoadingEnvironment) doNotImpl() {}

type ConfigureLoadingEnvironmentCapabilitySet interface {
	GetGeoLoader() geodata.Loader
}

type ConfigureLoadingEnvironment interface {
	ConfigureLoadingEnvironmentCapabilitySet
	doNotImpl()

	// TODO environment.BaseEnvironmentCapabilitySet
	// TODO environment.FileSystemCapabilitySet
}

func NewConfigureLoadingContext(ctx context.Context) context.Context {
	environment := &configureLoadingEnvironment{}
	return context.WithValue(ctx, confContextKey, environment)
}

func GetConfigureLoadingEnvironment(ctx context.Context) ConfigureLoadingEnvironment {
	return ctx.Value(confContextKey).(ConfigureLoadingEnvironment)
}

func SetGeoDataLoader(ctx context.Context, loader geodata.Loader) {
	GetConfigureLoadingEnvironment(ctx).(*configureLoadingEnvironment).geoLoader = loader
}
