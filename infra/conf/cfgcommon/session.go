package cfgcommon

import (
	"context"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/infra/conf/geodata"
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

type ConfigureLoadingEnvironment interface {
	GetGeoLoader() geodata.Loader
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
