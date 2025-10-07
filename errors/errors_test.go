package errors_test

import (
	"errors"
	"testing"

	frameworkerrors "github.com/aatuh/pureapi-framework/errors"
)

type customWireError struct {
	id      string
	message string
	data    any
}

func (e customWireError) Error() string       { return e.message }
func (e customWireError) CatalogID() string   { return e.id }
func (e customWireError) WireMessage() string { return e.message }
func (e customWireError) WireData() any       { return e.data }

func TestErrorMapper_MapByCatalogError(t *testing.T) {
	catalog, err := frameworkerrors.NewErrorCatalog(
		frameworkerrors.CatalogEntry{ID: "internal_error", Status: 500, Message: "internal"},
		frameworkerrors.CatalogEntry{ID: "not_found", Status: 404, Message: "missing"},
	)
	if err != nil {
		t.Fatalf("failed to build catalog: %v", err)
	}
	mapper, err := frameworkerrors.NewErrorMapper(catalog, "internal_error")
	if err != nil {
		t.Fatalf("failed to build mapper: %v", err)
	}

	wire := customWireError{id: "not_found", message: "resource missing", data: map[string]string{"id": "42"}}
	mapped := mapper.Map(wire)

	if mapped.Entry.ID != "not_found" {
		t.Fatalf("expected mapped id not_found, got %s", mapped.Entry.ID)
	}
	if mapped.Message != "resource missing" {
		t.Fatalf("expected custom wire message, got %s", mapped.Message)
	}
	if mapped.Data == nil {
		t.Fatalf("expected custom wire data")
	}
}

func TestErrorMapper_MapByRegistration(t *testing.T) {
	catalog, _ := frameworkerrors.NewErrorCatalog(
		frameworkerrors.CatalogEntry{ID: "internal_error", Status: 500, Message: "internal"},
		frameworkerrors.CatalogEntry{ID: "not_found", Status: 404, Message: "missing"},
	)
	mapper, _ := frameworkerrors.NewErrorMapper(catalog, "internal_error")

	sentinel := errors.New("sentinel not found")
	if err := mapper.RegisterIs(sentinel, "not_found"); err != nil {
		t.Fatalf("register is: %v", err)
	}

	err := errors.Join(errors.New("wrap"), sentinel)
	mapped := mapper.Map(err)
	if mapped.Entry.ID != "not_found" {
		t.Fatalf("expected mapped id not_found, got %s", mapped.Entry.ID)
	}
}

func TestRenderError(t *testing.T) {
	entry := frameworkerrors.CatalogEntry{ID: "invalid_request", Status: 400, Message: "invalid"}
	mapped := frameworkerrors.MappedError{Entry: entry, Message: "bad", Data: map[string]string{"field": "id"}}
	payload := frameworkerrors.RenderError(mapped)
	if payload.ID() != "invalid_request" {
		t.Fatalf("expected id invalid_request, got %s", payload.ID())
	}
	if payload.Message() != "bad" {
		t.Fatalf("expected message bad, got %s", payload.Message())
	}
	if data, ok := payload.Data().(map[string]string); !ok || data["field"] != "id" {
		t.Fatalf("unexpected payload data: %#v", payload.Data())
	}
}
