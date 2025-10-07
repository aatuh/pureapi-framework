package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type outHookInput struct{}

type outHookOutput struct {
	Message string `json:"message"`
}

// Demonstrates output hook that normalizes/enriches the response prior to render.
func Test_OutputHook(t *testing.T) {
	engine := framework.NewEngine()

	// Hook to enforce suffix
	hook := framework.NewOutputHook(func(ctx context.Context, out *outHookOutput) error {
		if out == nil {
			return nil
		}
		if out.Message != "" && out.Message[len(out.Message)-1] != '!' {
			out.Message += "!"
		}
		return nil
	})

	ep := framework.Endpoint[outHookInput, outHookOutput](
		engine,
		http.MethodGet,
		"/oh",
		func(ctx context.Context, in outHookInput) (outHookOutput, error) {
			return outHookOutput{Message: "hello"}, nil
		},
		framework.WithEndpointOutputHooks[outHookInput, outHookOutput](hook),
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	req := httptest.NewRequest(http.MethodGet, "/oh", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"message\":\"hello!\"}" {
		t.Fatalf("unexpected body: %s", got)
	}
}
