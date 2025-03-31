package api

import (
	"context"
	"net/http"

	endpoint "github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/api/errors"
	"github.com/pureapi/pureapi-framework/defaults"
	"github.com/pureapi/pureapi-framework/jsonapi"
	"github.com/pureapi/pureapi-framework/util/apimapper"
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
//   - endpointtypes.Definition: The definition of the health endpoint.
func HealthDefinition(systemID string) endpointtypes.Definition {
	handler := HealthEndpointHandler(
		apimapper.NewMapInputHandler(
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
//   - endpointtypes.Handler: The handler for the health endpoint.
func HealthEndpointHandler(
	inputHandler endpointtypes.InputHandler[HealthInput],
	emitterLogger utiltypes.EmitterLogger,
	systemID string,
) endpointtypes.Handler[HealthInput] {
	return endpoint.NewHandler(
		inputHandler,
		func(w http.ResponseWriter, r *http.Request, i *HealthInput,
		) (any, error) {
			// Only return empty output to indicate success.
			return NewHealthOutput(), nil
		},
		errors.NewErrorHandler(errors.NewExpectedErrorBuilder(systemID).
			WithErrors(errors.GenericErrors()).
			Build(),
		).WithSystemID(&systemID),
		jsonapi.NewJSONOutput(emitterLogger, systemID),
	).WithEmitterLogger(emitterLogger)
}

// SendHealthRequest sends a request to the health endpoint.
//
// Parameters:
//   - ctx: The context.
//   - host: The host.
//
// Returns:
//   - *jsonapi.Response[jsonapi.APIOutput[HealthOutput]]: The response.
//   - error: The error.
func SendHealthRequest(
	ctx context.Context, host string,
) (*jsonapi.Response[jsonapi.APIOutput[HealthOutput]], error) {
	return jsonapi.SendRequest[HealthInput, jsonapi.APIOutput[HealthOutput]](
		ctx, host, HealthURL, HealthMethod, NewHealthInput(),
	)
}
