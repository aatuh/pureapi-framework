package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type mwInput struct{}
type mwOutput struct {
	ID string `json:"id"`
}

// Demonstrates global/per-endpoint middleware and RequestID propagation.
func Test_Middleware_RequestID(t *testing.T) {
	// Simple middleware to set a header
	addHeader := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-App", "test")
			next.ServeHTTP(w, r)
		})
	}

	engine := framework.NewEngine(
		framework.WithGlobalMiddlewares(framework.RequestIDMiddleware(), addHeader),
	)

	ep := framework.Endpoint[mwInput, mwOutput](
		engine,
		http.MethodGet,
		"/mw",
		func(ctx context.Context, in mwInput) (mwOutput, error) {
			return mwOutput{ID: "ok"}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-App") != "test" {
		t.Fatalf("expected X-App header")
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("expected X-Request-ID header")
	}
}
