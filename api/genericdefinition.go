package api

import (
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-framework/api/errutil"
	"github.com/aatuh/pureapi-framework/api/json"
	"github.com/aatuh/pureapi-framework/defaults"
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
//   - endpoint.Definition: The endpoint definition.
func GenericEndpointDefinition[Input any](
	systemID string,
	url string,
	method string,
	inputHandler endpoint.InputHandler[Input],
	expectedErrors errutil.ExpectedErrors,
	invokeFn endpoint.HandlerLogicFn[Input],
) endpoint.Definition {
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
//   - endpoint.Handler: The endpoint handler.
func GenericEndpointHandler[Input any](
	systemID string,
	inputHandler endpoint.InputHandler[Input],
	expectedErrors errutil.ExpectedErrors,
	invokeFn endpoint.HandlerLogicFn[Input],
) endpoint.Handler[Input] {
	emitterLogger := defaults.EmitterLogger()
	return endpoint.NewHandler(
		inputHandler,
		invokeFn,
		errutil.NewErrorHandler(
			errutil.NewExpectedErrorBuilder(systemID).
				WithErrors(expectedErrors).
				Build(),
		).WithSystemID(&systemID),
		json.NewJSONOutput(emitterLogger, systemID),
	).WithEmitterLogger(emitterLogger)
}
