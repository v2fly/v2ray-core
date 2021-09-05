package envctx

import "context"

type environmentContextKey int

const (
	environmentKey environmentContextKey = iota
)

func ContextWithEnvironment(ctx context.Context, environment interface{}) context.Context {
	return context.WithValue(ctx, environmentKey, environment)
}

func EnvironmentFromContext(ctx context.Context) interface{} {
	if environment, ok := ctx.Value(environmentKey).(interface{}); ok {
		return environment
	}
	return nil
}
