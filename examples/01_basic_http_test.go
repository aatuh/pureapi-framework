package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type basicInput struct {
	ID   string `path:"id"`
	Name string `query:"name"`
}

type basicOutput struct {
	Greeting string `json:"greeting"`
}

// Basic server with a single endpoint.
func Test_BasicHTTP(t *testing.T) {
	engine := framework.NewEngine()

	endpoint := framework.Endpoint[basicInput, basicOutput](
		engine,
		http.MethodGet,
		"/hello/{id}",
		func(ctx context.Context, in basicInput) (basicOutput, error) {
			return basicOutput{Greeting: "hello " + in.Name + " (#" + in.ID + ")"}, nil
		},
	)

	handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(handler, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/hello/42?name=World", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "{\"greeting\":\"hello World (#42)\"}" {
		t.Fatalf("unexpected body: %s", body)
	}
}
