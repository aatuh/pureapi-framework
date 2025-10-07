package hooks

import (
	"context"
	"fmt"
)

type valueHook interface {
	Process(ctx context.Context, value any) error
}

type valueHookFunc func(ctx context.Context, value any) error

func (f valueHookFunc) Process(ctx context.Context, value any) error {
	return f(ctx, value)
}

// InputHook processes the bound input before the handler executes.
type InputHook interface {
	valueHook
}

// OutputHook processes the handler output before rendering.
type OutputHook interface {
	valueHook
}

// NewInputHook wraps a strongly-typed function into a generic input hook.
func NewInputHook[T any](fn func(ctx context.Context, value *T) error) InputHook {
	if fn == nil {
		return nil
	}
	return valueHookFunc(func(ctx context.Context, value any) error {
		if value == nil {
			return fn(ctx, nil)
		}
		typed, ok := value.(*T)
		if !ok {
			return fmt.Errorf("input hook: expected *%T, got %T", new(T), value)
		}
		return fn(ctx, typed)
	})
}

// NewOutputHook wraps a strongly-typed function into a generic output hook.
func NewOutputHook[T any](fn func(ctx context.Context, value *T) error) OutputHook {
	if fn == nil {
		return nil
	}
	return valueHookFunc(func(ctx context.Context, value any) error {
		if value == nil {
			return fn(ctx, nil)
		}
		typed, ok := value.(*T)
		if !ok {
			return fmt.Errorf("output hook: expected *%T, got %T", new(T), value)
		}
		return fn(ctx, typed)
	})
}
