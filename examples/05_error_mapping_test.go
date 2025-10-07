package examples_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

var errNotFound = errors.New("not found")

type emInput struct{}
type emOutput struct {
	OK bool `json:"ok"`
}

// Demonstrates custom ErrorCatalog/ErrorMapper with RegisterIs and RegisterType.
func Test_ErrorMapping_CustomCatalog(t *testing.T) {
	catalog, _ := framework.NewErrorCatalog(
		framework.CatalogEntry{ID: "internal_error", Status: http.StatusInternalServerError, Message: "Internal"},
		framework.CatalogEntry{ID: "not_found", Status: http.StatusNotFound, Message: "Not found"},
		framework.CatalogEntry{ID: "boom", Status: http.StatusBadRequest, Message: "Boom"},
	)
	mapper, _ := framework.NewErrorMapper(catalog, "internal_error")
	if err := mapper.RegisterIs(errNotFound, "not_found"); err != nil {
		t.Fatalf("register is: %v", err)
	}
	// Map BindError explicitly (though engine already registers it globally)
	_ = mapper.RegisterType((*framework.BindError)(nil), "boom")

	engine := framework.NewEngine(framework.WithErrorMapper(mapper))

	ep := framework.Endpoint[emInput, emOutput](
		engine,
		http.MethodGet,
		"/em1",
		func(ctx context.Context, in emInput) (emOutput, error) {
			return emOutput{}, errNotFound
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	req := httptest.NewRequest(http.MethodGet, "/em1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
