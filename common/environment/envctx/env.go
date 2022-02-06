package envctx

import "context"

const (
	environmentKey string = "v2.environment"
)

func ContextWithEnvironment(ctx context.Context, environment interface{}) context.Context {
	return context.WithValue(ctx, environmentKey, environment)
}

func EnvironmentFromContext(ctx context.Context) interface{} {
	return ctx.Value(environmentKey)
}
