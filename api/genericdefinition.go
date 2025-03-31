package api

import (
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	"github.com/pureapi/pureapi-framework/api/errors"
	"github.com/pureapi/pureapi-framework/defaults"
	"github.com/pureapi/pureapi-framework/jsonapi"
)

// GenericEndpointDefinition builds the endpoint definition for any operation.
// It uses default handlers to handle the request.
//
// Parameters:
//   - systemID: The system ID.
//   - url: The URL of the endpoint.
//   - method: The HTTP method of the endpoint.
//   - inputHandler: The input handler for the endpoint.
//   - expectedErrors: The expected errors for the endpoint.
//   - invokeFn: The function to invoke the endpoint.
//
// Returns:
//   - endpointtypes.Definition: The endpoint definition.
func GenericEndpointDefinition[Input any](
	systemID string,
	url string,
	method string,
	inputHandler endpointtypes.InputHandler[Input],
	expectedErrors errors.ExpectedErrors,
	invokeFn endpoint.HandlerLogicFn[Input],
) endpointtypes.Definition {
	handler := GenericEndpointHandler(
		systemID, inputHandler, expectedErrors, invokeFn,
	)
	return endpoint.NewDefinition(
		url, method, defaults.NewStackBuilder().Build(), handler.Handle,
	)
}

// GenericEndpointHandler builds the endpoint handler for any operation.
// It uses default handlers to handle the request.
//
// Parameters:
//   - systemID: The system ID.
//   - inputHandler: The input handler for the endpoint.
//   - expectedErrors: The expected errors for the endpoint.
//   - invokeFn: The function to invoke the endpoint.
//
// Returns:
//   - endpointtypes.Handler: The endpoint handler.
func GenericEndpointHandler[Input any](
	systemID string,
	inputHandler endpointtypes.InputHandler[Input],
	expectedErrors errors.ExpectedErrors,
	invokeFn endpoint.HandlerLogicFn[Input],
) endpointtypes.Handler[Input] {
	emitterLogger := defaults.EmitterLogger()
	return endpoint.NewHandler(
		inputHandler,
		invokeFn,
		errors.NewErrorHandler(
			errors.
				NewExpectedErrorBuilder(systemID).
				WithErrors(expectedErrors).
				Build(),
		).WithSystemID(&systemID),
		jsonapi.NewJSONOutput(emitterLogger, systemID),
	).WithEmitterLogger(emitterLogger)
}
