package hooks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
	"github.com/aatuh/pureapi-framework/hooks"
	"github.com/aatuh/validate/v3"
)

type validatorInput struct {
	Name string `query:"name" validate:"string;min=3"`
}

type validatorOutput struct {
	Status string `json:"status"`
}

func TestInputHook_WithValidator(t *testing.T) {
	engine := framework.NewEngine()

	validator := validate.New()
	validationHook := hooks.NewInputHook(func(ctx context.Context, in *validatorInput) error {
		if in == nil {
			return nil
		}
		if err := validator.ValidateStruct(*in); err != nil {
			if es, ok := err.(validate.Errors); ok {
				fields := make([]framework.FieldError, 0, len(es))
				for _, fe := range es {
					message := fe.Msg
					if message == "" {
						message = fe.Code
					}
					fields = append(fields, framework.NewFieldError(fe.Path, framework.SourceQuery, message))
				}
				return framework.NewBindError("validation failed", fields).WithCause(err)
			}
			return err
		}
		return nil
	})

	endpoint := framework.Endpoint[validatorInput, validatorOutput](
		engine,
		http.MethodGet,
		"/users",
		func(ctx context.Context, in validatorInput) (validatorOutput, error) {
			return validatorOutput{Status: "ok"}, nil
		},
		framework.WithEndpointInputHooks[validatorInput, validatorOutput](validationHook),
	)

	handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(handler, endpoint)

	t.Run("invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("valid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?name=Jane", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}
