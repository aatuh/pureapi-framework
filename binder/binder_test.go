package binder_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type binderTestBody struct {
	Name string `json:"name"`
}

type binderTestInput struct {
	ID      string         `path:"id" required:"true"`
	Cursor  int            `query:"cursor"`
	Sort    []string       `query:"sort"`
	Client  string         `header:"X-Client" required:"true"`
	Session string         `cookie:"session_id"`
	Body    binderTestBody `body:"" required:"true"`
}

type binderTestOutput struct {
	OK bool `json:"ok"`
}

func TestDefaultBinder_BindsAllSources(t *testing.T) {
	engine := framework.NewEngine()

	var captured binderTestInput
	decl := framework.Endpoint[binderTestInput, binderTestOutput](
		engine,
		http.MethodGet,
		"/widgets/{id}",
		func(ctx context.Context, in binderTestInput) (binderTestOutput, error) {
			captured = in
			return binderTestOutput{OK: true}, nil
		},
	)

	handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(handler, decl)

	req := httptest.NewRequest(http.MethodGet, "/widgets/abc?cursor=42&sort=a&sort=b", strings.NewReader("{\"name\":\"demo\"}"))
	req.Header.Set("X-Client", "console")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "cookie-1"})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if captured.ID != "abc" {
		t.Errorf("expected path param binding, got %q", captured.ID)
	}
	if captured.Cursor != 42 {
		t.Errorf("expected query int, got %d", captured.Cursor)
	}
	if want := []string{"a", "b"}; len(captured.Sort) != 2 || captured.Sort[0] != want[0] || captured.Sort[1] != want[1] {
		t.Errorf("expected query slice, got %#v", captured.Sort)
	}
	if captured.Client != "console" {
		t.Errorf("expected header binding, got %q", captured.Client)
	}
	if captured.Session != "cookie-1" {
		t.Errorf("expected cookie binding, got %q", captured.Session)
	}
	if captured.Body.Name != "demo" {
		t.Errorf("expected body binding, got %+v", captured.Body)
	}

	var out binderTestOutput
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected renderer output true")
	}
}

func TestDefaultBinder_RequiredFieldMissing(t *testing.T) {
	engine := framework.NewEngine()

	decl := framework.Endpoint[binderTestInput, binderTestOutput](
		engine,
		http.MethodGet,
		"/widgets/{id}",
		func(ctx context.Context, in binderTestInput) (binderTestOutput, error) {
			return binderTestOutput{OK: true}, nil
		},
	)

	handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(handler, decl)

	req := httptest.NewRequest(http.MethodGet, "/widgets/abc", strings.NewReader(`{}`))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var wire struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Data    struct {
			Fields []framework.FieldError `json:"fields"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &wire); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if wire.ID != "invalid_request" {
		t.Fatalf("expected error id invalid_request, got %s", wire.ID)
	}
	if len(wire.Data.Fields) == 0 {
		t.Fatalf("expected field errors in response")
	}
}
