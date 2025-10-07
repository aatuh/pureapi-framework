package examples_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	framework "github.com/aatuh/pureapi-framework"
)

type strictBody struct {
	A int `json:"a"`
}

type strictInput struct {
	Payload *strictBody `body:"" required:"true"`
}

type strictOutput struct {
	OK bool `json:"ok"`
}

// Demonstrates strict JSON decoding (disallow unknown fields) and body size limits.
func Test_JSONStrict_And_BodyLimit(t *testing.T) {
	// Custom binder with strict JSON and small size limit
	b := framework.NewDefaultBinder()
	b.BodyDecoder = framework.JSONBodyDecoder{DisallowUnknown: true}
	b.MaxBodyBytes = 8 // small limit to trigger too large
	b.ReadTimeout = 100 * time.Millisecond

	engine := framework.NewEngine(framework.WithBinder(b))

	ep := framework.Endpoint[strictInput, strictOutput](
		engine,
		http.MethodPost,
		"/strict",
		func(ctx context.Context, in strictInput) (strictOutput, error) {
			return strictOutput{OK: true}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	// Unknown field should fail with 400 due to DisallowUnknown
	bad := map[string]any{"a": 1, "extra": true}
	badJSON, _ := json.Marshal(bad)
	req1 := httptest.NewRequest(http.MethodPost, "/strict", bytes.NewReader(badJSON))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown fields, got %d", rec1.Code)
	}

	// Body too large triggers 400 from binder mapping
	large := bytes.Repeat([]byte("x"), 32)
	req2 := httptest.NewRequest(http.MethodPost, "/strict", bytes.NewReader(large))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for too large body, got %d", rec2.Code)
	}
}
