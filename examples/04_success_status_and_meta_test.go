package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type metaInput struct{}

type metaOutput struct {
	ID string `json:"id"`
}

// Demonstrates WithMeta and WithSuccessStatus.
func Test_SuccessStatusAndMeta(t *testing.T) {
	engine := framework.NewEngine()

	ep := framework.Endpoint[metaInput, metaOutput](
		engine,
		http.MethodPost,
		"/resources",
		func(ctx context.Context, in metaInput) (metaOutput, error) {
			return metaOutput{ID: "123"}, nil
		},
		framework.WithMeta[metaInput, metaOutput](framework.EndpointMeta{
			Summary:     "Create resource",
			Description: "Creates a resource and returns its id",
			OperationID: "createResource",
			Tags:        []string{"resources"},
		}),
		framework.WithSuccessStatus[metaInput, metaOutput](http.StatusAccepted),
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, ep)

	req := httptest.NewRequest(http.MethodPost, "/resources", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
