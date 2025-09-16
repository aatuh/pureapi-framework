package api

import (
	"context"
	"net/http"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-framework/api/errutil"
	"github.com/aatuh/pureapi-framework/api/input"
	"github.com/aatuh/pureapi-framework/api/json"
	"github.com/aatuh/pureapi-framework/defaults"
)

// Health endpoint constants.
const (
	HealthURL    = "/health"
	HealthMethod = http.MethodGet
)

// HealthInput is the input of the health endpoint.
type HealthInput struct{}

// NewHealthInput returns a new instance of HealthInput.
//
// Returns:
//   - *HealthInput: A new instance of HealthInput.
func NewHealthInput() *HealthInput {
	return &HealthInput{}
}

// HealthOutput is the output of the health endpoint.
type HealthOutput struct{}

// NewHealthOutput returns a new instance of HealthOutput.
//
// Returns:
//   - *HealthOutput: A new instance of HealthOutput.
func NewHealthOutput() *HealthOutput {
	return &HealthOutput{}
}

// HealthDefinition returns the definition of a simple health endpoint.
// It uses default handlers to handle the request.
//
// Parameters:
//   - systemID: The system ID.
//
// Returns:
//   - endpoint.Definition: The definition of the health endpoint.
func HealthDefinition(systemID string) endpoint.Definition {
	handler := HealthEndpointHandler(
		input.NewMapInputHandler(
			nil,
			defaults.InputConversionRules(),
			defaults.ValidationRules(),
			NewHealthInput,
		),
		defaults.EmitterLogger(),
		systemID,
	)
	return endpoint.NewDefinition(
		HealthURL,
		HealthMethod,
		defaults.NewStackBuilder().Build(),
		handler.Handle,
	)
}

// HealthEndpointHandler returns the handler for a simple health endpoint.
// It uses default handlers to handle the request.
//
// Parameters:
//   - inputHandler: The input handler for the endpoint.
//   - emitterLogger: The emitter logger for the endpoint.
//   - systemID: The system ID.
//
// Returns:
//   - endpoint.Handler: The handler for the health endpoint.
func HealthEndpointHandler(
	inputHandler endpoint.InputHandler[HealthInput],
	emitterLogger event.EmitterLogger,
	systemID string,
) endpoint.Handler[HealthInput] {
	return endpoint.NewHandler(
		inputHandler,
		func(w http.ResponseWriter, r *http.Request, i *HealthInput,
		) (any, error) {
			// Only return empty output to indicate success.
			return NewHealthOutput(), nil
		},
		errutil.NewErrorHandler(errutil.NewExpectedErrorBuilder(systemID).
			WithErrors(errutil.GenericErrors()).
			Build(),
		).WithSystemID(&systemID),
		json.NewJSONOutput(emitterLogger, systemID),
	).WithEmitterLogger(emitterLogger)
}

// SendHealthRequest sends a request to the health endpoint.
//
// Parameters:
//   - ctx: The context.
//   - host: The host.
//
// Returns:
//   - *apijson.Response[apijson.APIOutput[HealthOutput]]: The response.
//   - error: The error.
func SendHealthRequest(
	ctx context.Context, host string,
) (*json.Response[json.APIOutput[HealthOutput]], error) {
	return json.SendRequest[HealthInput, json.APIOutput[HealthOutput]](
		ctx,
		host,
		HealthURL,
		HealthMethod,
		NewHealthInput(),
		defaults.CtxLogger,
	)
}
