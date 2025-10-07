package hooks

import (
	"context"
	"net/http"
)

// ContextEnricher can attach values (e.g., auth principals) to the request context
// before binding and handler execution.
type ContextEnricher interface {
	Enrich(ctx context.Context, r *http.Request) (context.Context, error)
}

// ContextEnricherFunc lifts a function into a ContextEnricher.
type ContextEnricherFunc func(ctx context.Context, r *http.Request) (context.Context, error)

// Enrich implements ContextEnricher.
func (f ContextEnricherFunc) Enrich(ctx context.Context, r *http.Request) (context.Context, error) {
	return f(ctx, r)
}

// NewContextEnricher wraps a function into a ContextEnricher implementation.
func NewContextEnricher(fn func(ctx context.Context, r *http.Request) (context.Context, error)) ContextEnricher {
	if fn == nil {
		return nil
	}
	return ContextEnricherFunc(fn)
}
