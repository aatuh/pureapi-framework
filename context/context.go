package context

import "context"

type contextKey string

const pathParamsContextKey contextKey = "pureapi-framework/path-params"

// WithPathParams annotates context with path parameters for custom routers.
func WithPathParams(ctx context.Context, params map[string]string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if params == nil {
		return ctx
	}
	copy := make(map[string]string, len(params))
	for k, v := range params {
		copy[k] = v
	}
	return context.WithValue(ctx, pathParamsContextKey, copy)
}

// PathParamsFromContext extracts path parameters stored via WithPathParams.
func PathParamsFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(pathParamsContextKey); v != nil {
		if params, ok := v.(map[string]string); ok {
			copy := make(map[string]string, len(params))
			for k, val := range params {
				copy[k] = val
			}
			return copy
		}
	}
	return nil
}
