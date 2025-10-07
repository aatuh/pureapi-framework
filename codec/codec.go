package codec

import (
	"context"
)

// Codec serializes/deserializes payloads.
type Codec interface {
	ContentType() string
	Marshal(ctx context.Context, v any) ([]byte, error)
}

// Renderer wraps a Codec into a renderer-compatible interface.
type Renderer interface {
	Render(ctx context.Context, status int, payload any) ([]byte, string, error)
}

// CodecRenderer adapts a Codec to Renderer.
type CodecRenderer struct {
	Codec Codec
}

func (c CodecRenderer) Render(ctx context.Context, status int, payload any) ([]byte, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	data, err := c.Codec.Marshal(ctx, payload)
	if err != nil {
		return nil, "", err
	}
	return data, c.Codec.ContentType(), nil
}
