package headers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-framework/security/headers"
)

func TestMiddlewareSetsDefaultHeaders(t *testing.T) {
	mw := headers.Default()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	h.ServeHTTP(rec, req)

	tests := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "0",
		"Referrer-Policy":        "no-referrer",
	}

	for header, expected := range tests {
		if got := rec.Header().Get(header); got != expected {
			t.Fatalf("expected %s=%s, got %s", header, expected, got)
		}
	}
}

func TestMiddlewareWithCustomConfig(t *testing.T) {
	mw := headers.Middleware(headers.Config{
		NoSniff:       false,
		FrameOptions:  "SAMEORIGIN",
		XSSProtection: "1; mode=block",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Content-Type-Options") != "" {
		t.Fatalf("expected no X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatalf("unexpected frame options header")
	}
	if rec.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Fatalf("unexpected XSS protection header")
	}
}
