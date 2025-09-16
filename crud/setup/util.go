package setup

import (
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-framework/api/errutil"
	"github.com/aatuh/pureapi-framework/api/json"
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
// adds generic errors along with the additionalapi.
func defaultErrorHandlerFactory(
	systemID string,
	additionalErrors errutil.ExpectedErrors,
) func() endpoint.ErrorHandler {
	return func() endpoint.ErrorHandler {
		return errutil.NewErrorHandler(
			errutil.NewExpectedErrorBuilder(systemID).
				WithErrors(errutil.GenericErrors()).
				WithErrors(additionalErrors).
				Build(),
		).WithSystemID(&systemID)
	}
}

// defaultOutputHandlerFactory returns a default output handler factory.
func defaultOutputHandlerFactory(
	systemID string,
	emitterLogger event.EmitterLogger,
) func() endpoint.OutputHandler {
	return func() endpoint.OutputHandler {
		return json.NewJSONOutput(emitterLogger, systemID)
	}
}

// newDefinition creates a new endpoint definition.
func newDefinition[Input any](
	url, method string,
	stack endpoint.Stack,
	handler *endpoint.DefaultHandler[Input],
) *endpoint.DefaultDefinition {
	return endpoint.NewDefinition(url, method, stack, handler.Handle)
}

// newHandler is a helper to build a new endpoint handler.
func newHandler[Input any](
	inputHandler endpoint.InputHandler[Input],
	handlerLogic endpoint.HandlerLogicFn[Input],
	errorHandler endpoint.ErrorHandler,
	outputHandler endpoint.OutputHandler,
	emitterLogger event.EmitterLogger,
) *endpoint.DefaultHandler[Input] {
	return endpoint.NewHandler(
		inputHandler, handlerLogic, errorHandler, outputHandler,
	).WithEmitterLogger(emitterLogger)
}
