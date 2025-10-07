package json

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aatuh/pureapi-framework/renderer/registry"
)

// Renderer renders JSON responses.
type Renderer struct {
	Pretty bool
}

// Render implements renderer.RenderFunc and returns JSON bytes and content type.
func (r Renderer) Render(ctx context.Context, status int, payload any) ([]byte, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}

	var (
		data []byte
		err  error
	)
	if payload == nil {
		data = []byte("null")
	} else if r.Pretty {
		data, err = json.MarshalIndent(payload, "", "  ")
	} else {
		data, err = json.Marshal(payload)
	}
	if err != nil {
		return nil, "", fmt.Errorf("render json: %w", err)
	}
	return data, "application/json", nil
}

// RenderFunc returns a renderer.RenderFunc compatible closure.
func (r Renderer) RenderFunc() registry.RenderFunc {
	return func(ctx context.Context, status int, payload any) ([]byte, string, error) {
		return r.Render(ctx, status, payload)
	}
}
