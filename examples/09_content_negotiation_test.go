package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type negotiationOutput struct {
	Message string `json:"message"`
}

// Demonstrates codec registry content negotiation.
func Test_ContentNegotiation(t *testing.T) {
	textRenderer := func(ctx context.Context, status int, payload any) ([]byte, string, error) {
		msg := ""
		if v, ok := payload.(negotiationOutput); ok {
			msg = v.Message
		}
		return []byte(msg), "text/plain", nil
	}

	engine := framework.NewEngine(
		framework.WithRenderer("text/plain", textRenderer),
	)

	endpoint := framework.Endpoint[struct{}, negotiationOutput](
		engine,
		http.MethodGet,
		"/hello",
		func(ctx context.Context, _ struct{}) (negotiationOutput, error) {
			return negotiationOutput{Message: "hello"}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	// Plain text request
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("expected text/plain renderer")
	}
	if body := rec.Body.String(); body != "hello" {
		t.Fatalf("unexpected body: %s", body)
	}

	// JSON fallback when Accept prefers JSON
	reqJSON := httptest.NewRequest(http.MethodGet, "/hello", nil)
	reqJSON.Header.Set("Accept", "application/json")
	recJSON := httptest.NewRecorder()
	h.ServeHTTP(recJSON, reqJSON)

	if recJSON.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON renderer")
	}
	if recJSON.Body.String() != "{\"message\":\"hello\"}" {
		t.Fatalf("unexpected JSON payload: %s", recJSON.Body.String())
	}
}
