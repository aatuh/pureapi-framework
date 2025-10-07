package registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// RenderFunc renders the payload and returns the body bytes and content type.
type RenderFunc func(ctx context.Context, status int, payload any) ([]byte, string, error)

// Registry stores renderers keyed by content type.
type Registry struct {
	renderers map[string]RenderFunc
	defaultCT string
}

// New creates a registry with the provided default renderer.
func New(defaultType string, defaultRenderer RenderFunc) *Registry {
	r := &Registry{
		renderers: make(map[string]RenderFunc),
		defaultCT: canonicalContentType(defaultType),
	}
	if defaultRenderer != nil {
		r.renderers[r.defaultCT] = defaultRenderer
	}
	return r
}

// Register adds or overrides a renderer for the given content type.
func (r *Registry) Register(contentType string, renderer RenderFunc) {
	if r == nil || renderer == nil {
		return
	}
	r.renderers[canonicalContentType(contentType)] = renderer
}

// Clone returns a shallow copy of the registry.
func (r *Registry) Clone() *Registry {
	if r == nil {
		return nil
	}
	clone := &Registry{
		renderers: make(map[string]RenderFunc, len(r.renderers)),
		defaultCT: r.defaultCT,
	}
	for k, v := range r.renderers {
		clone.renderers[k] = v
	}
	return clone
}

// Render picks a renderer based on Accept header. Returns content type and body.
func (r *Registry) Render(ctx context.Context, w http.ResponseWriter, req *http.Request, status int, payload any) error {
	if r == nil {
		return fmt.Errorf("renderer registry is nil")
	}
	var acceptHeader string
	if req != nil {
		acceptHeader = req.Header.Get("Accept")
	}
	accepted := parseAccept(acceptHeader)
	for _, ct := range accepted {
		if renderer, ok := r.renderers[ct]; ok {
			return r.render(ctx, w, status, payload, ct, renderer)
		}
	}
	if renderer, ok := r.renderers[r.defaultCT]; ok {
		return r.render(ctx, w, status, payload, r.defaultCT, renderer)
	}
	return fmt.Errorf("no renderer registered")
}

func (r *Registry) render(ctx context.Context, w http.ResponseWriter, status int, payload any, ct string, renderFn RenderFunc) error {
	data, contentType, err := renderFn(ctx, status, payload)
	if err != nil {
		return err
	}
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if len(data) == 0 {
		data = []byte("null")
	}
	_, err = w.Write(data)
	return err
}

func parseAccept(header string) []string {
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	out := make([]string, 0, len(parts))
	wildcard := false
	for _, part := range parts {
		ct := strings.TrimSpace(strings.Split(part, ";")[0])
		ct = canonicalContentType(ct)
		if ct == "*/*" {
			wildcard = true
			continue
		}
		if ct != "" {
			out = append(out, ct)
		}
	}
	if wildcard {
		out = append(out, "*/*")
	}
	return out
}

func canonicalContentType(ct string) string {
	return strings.ToLower(strings.TrimSpace(ct))
}
