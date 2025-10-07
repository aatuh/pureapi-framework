package json

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-framework/renderer/registry"
)

func TestRenderer_Render(t *testing.T) {
	r := Renderer{}
	registry := registry.New("application/json", r.RenderFunc())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	if err := registry.Render(context.Background(), rec, req, http.StatusOK, map[string]string{"ok": "yes"}); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if rec.Body.String() != "{\"ok\":\"yes\"}" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
