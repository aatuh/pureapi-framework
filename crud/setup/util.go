package setup

import (
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/api/errors"
	"github.com/pureapi/pureapi-framework/jsonapi"
)

// withDefaultFactory returns the provided factory if non-nil, or the
// defaultFactory.
func withDefaultFactory[T any](
	factory func() T, defaultFactory func() T,
) func() T {
	if factory != nil {
		return factory
	}
	return defaultFactory
}

// defaultErrorHandlerFactory returns a default error handler factory that
// adds generic errors along with the additionalErrors.
func defaultErrorHandlerFactory(
	systemID string,
	additionalErrors errors.ExpectedErrors,
) func() endpointtypes.ErrorHandler {
	return func() endpointtypes.ErrorHandler {
		return errors.NewErrorHandler(
			errors.NewExpectedErrorBuilder(systemID).
				WithErrors(errors.GenericErrors()).
				WithErrors(additionalErrors).
				Build(),
		).WithSystemID(&systemID)
	}
}

// defaultOutputHandlerFactory returns a default output handler factory.
func defaultOutputHandlerFactory(
	systemID string,
	emitterLogger utiltypes.EmitterLogger,
) func() endpointtypes.OutputHandler {
	return func() endpointtypes.OutputHandler {
		return jsonapi.NewJSONOutput(emitterLogger, systemID)
	}
}

// newDefinition creates a new endpoint definition.
func newDefinition[Input any](
	url, method string,
	stack endpointtypes.Stack,
	handler *endpoint.DefaultHandler[Input],
) *endpoint.DefaultDefinition {
	return endpoint.NewDefinition(url, method, stack, handler.Handle)
}

// newHandler is a helper to build a new endpoint handler.
func newHandler[Input any](
	inputHandler endpointtypes.InputHandler[Input],
	handlerLogic endpoint.HandlerLogicFn[Input],
	errorHandler endpointtypes.ErrorHandler,
	outputHandler endpointtypes.OutputHandler,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[Input] {
	return endpoint.NewHandler(
		inputHandler, handlerLogic, errorHandler, outputHandler,
	).WithEmitterLogger(emitterLogger)
}
