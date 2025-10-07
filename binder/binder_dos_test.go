package binder

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultBinder_RespectsBodyLimit(t *testing.T) {
	type bodyInput struct {
		Payload map[string]any `body:"" required:"true"`
	}

	b := NewDefaultBinder()
	b.MaxBodyBytes = 16

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 32)))
	req.Header.Set("Content-Type", "application/json")

	var dst bodyInput
	err := b.Bind(req.Context(), req, &dst)
	if err == nil {
		t.Fatalf("expected body limit error")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

func TestDefaultBinder_HonorsCanceledContext(t *testing.T) {
	type bodyInput struct {
		Name string `body:""`
	}

	b := NewDefaultBinder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	ctx, cancel := context.WithCancel(req.Context())
	cancel()

	var dst bodyInput
	err := b.Bind(ctx, req, &dst)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestReadAllWithContextCanceled(t *testing.T) {
	r := strings.NewReader("{}")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := readAllWithContext(ctx, r, 0)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDefaultBinder_StrictJSONRejectsUnknown(t *testing.T) {
	b := NewDefaultBinder()
	b.SetStrictJSONBodies(true)

	type requestBody struct {
		Name string `json:"name"`
	}

	type input struct {
		Payload requestBody `body:""`
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ok","extra":1}`))
	req.Header.Set("Content-Type", "application/json")

	var dst input
	err := b.Bind(req.Context(), req, &dst)
	if err == nil {
		t.Fatalf("expected error for unknown field")
	}
	bindErr, ok := err.(*BindError)
	if !ok {
		t.Fatalf("expected BindError, got %T", err)
	}
	if bindErr.Message() != "Unknown field in request body" {
		t.Fatalf("unexpected bind error message: %s fields=%v", bindErr.Message(), bindErr.Fields())
	}
	if cause := errors.Unwrap(bindErr); cause == nil || !strings.Contains(cause.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", cause)
	}
}
