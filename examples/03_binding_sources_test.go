package examples_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type bindingBody struct {
	Note string `json:"note"`
}

type bindingInput struct {
	ID      int          `path:"id" required:"true"`
	Cursor  int          `query:"cursor"`
	Sort    []string     `query:"sort"`
	Client  string       `header:"X-Client" required:"true"`
	Session string       `cookie:"session_id"`
	Body    *bindingBody `body:"" required:"true"`
}

type bindingOutput struct {
	OK bool `json:"ok"`
}

// Demonstrates binding from path, query, header, cookie, and JSON body.
func Test_BindingSources(t *testing.T) {
	engine := framework.NewEngine()

	ep := framework.Endpoint[bindingInput, bindingOutput](
		engine,
		http.MethodPost,
		"/items/{id}",
		func(ctx context.Context, in bindingInput) (bindingOutput, error) {
			if in.ID <= 0 || in.Client == "" || in.Body == nil || in.Body.Note == "" {
				t.Fatalf("handler received unexpected input: %+v", in)
			}
			return bindingOutput{OK: true}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	payload := map[string]string{"note": "hello"}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(
		http.MethodPost,
		"/items/7?cursor=10&sort=a&sort=b",
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client", "web")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated { // POST defaults to 201
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

// Missing required values should yield 400 via BindError mapping.
func Test_BindingSources_MissingRequired(t *testing.T) {
	engine := framework.NewEngine()

	ep := framework.Endpoint[bindingInput, bindingOutput](
		engine,
		http.MethodPost,
		"/items/{id}",
		func(ctx context.Context, in bindingInput) (bindingOutput, error) {
			return bindingOutput{OK: true}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	// Missing body and header and invalid id
	req := httptest.NewRequest(http.MethodPost, "/items/0", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
