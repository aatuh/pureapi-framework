package examples_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
	"github.com/aatuh/validate/v3"
)

type userInput struct {
	Name string `query:"name" validate:"string;min=3"`
}

type userOutput struct {
	Greeting string `json:"greeting"`
}

// Input hook illustrates plugging in validation without coupling the framework.
func Test_InputHookValidation(t *testing.T) {
	engine := framework.NewEngine()

	validator := validate.New()
	validationHook := framework.NewInputHook(func(ctx context.Context, in *userInput) error {
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

	endpoint := framework.Endpoint[userInput, userOutput](
		engine,
		http.MethodGet,
		"/welcome",
		func(ctx context.Context, in userInput) (userOutput, error) {
			return userOutput{Greeting: "hello " + in.Name}, nil
		},
		framework.WithEndpointInputHooks[userInput, userOutput](validationHook),
	)

	handler := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(handler, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/welcome", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
