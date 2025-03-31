package errors

import (
	"net/http"

	"github.com/pureapi/pureapi-core/util"
	"github.com/pureapi/pureapi-core/util/types"
)

// InternalServerError represents an internal server error.
var InternalServerError = util.NewAPIError("INTERNAL_SERVER_ERROR")

// errorHandler handles errors and maps them to appropriate HTTP responses.
type errorHandler struct {
	expectedErrs []ExpectedError // Expected errors to handle.
	systemID     *string
}

// NewErrorHandler creates a new ErrorHandler.
//
// Parameters:
//   - expectedErrs: A slice of ExpectedError objects that define how to handle
//     specific errors.
//
// Returns:
//   - *ErrorHandler: A new ErrorHandler.
func NewErrorHandler(
	expectedErrs []ExpectedError) *errorHandler {
	return &errorHandler{
		expectedErrs: expectedErrs,
	}
}

// WithSystemID adds a system ID to the handler.
//
// Parameters:
//   - systemID: The optional system ID. It is used to add the system ID to any
//     APIError instances passing through this handler. If the system ID is nil,
//     or the error alread has a system ID, no system ID is set.
//
// Returns:
//   - *handler: A new handler instance.
func (h *errorHandler) WithSystemID(
	systemID *string,
) *errorHandler {
	new := *h
	new.systemID = systemID
	return &new
}

// Handle processes an error and returns the corresponding HTTP status code and
// API error. It checks if the error is an *apierror.Error and handles it
// accordingly
//
// Parameters:
//   - err: The error to handle.
//
// Returns:
//   - int: The HTTP status code.
//   - *api.APIError: The mapped API error.
func (e errorHandler) Handle(err error) (int, types.APIError) {
	apiError, ok := err.(types.APIError)
	if !ok {
		return http.StatusInternalServerError, InternalServerError
	}
	status, apiError := e.handleAPIError(apiError)

	return status, apiError
}

// handleAPIError maps an API error to an HTTP status code and API error.
func (e *errorHandler) handleAPIError(
	apiError types.APIError,
) (int, types.APIError) {
	// Add system ID to error if not set and if system ID is available
	if apiError.Origin() == "" && e.systemID != nil {
		apiError = util.APIErrorFrom(apiError).WithOrigin(*e.systemID)
	}

	expectedError := e.getExpectedError(apiError)
	if expectedError == nil {
		return http.StatusInternalServerError, InternalServerError
	}
	return expectedError.maskAPIError(apiError)
}

// getExpectedError finds the ExpectedError that matches the given API error.
// It returns nil if no match is found.
func (e *errorHandler) getExpectedError(
	apiError types.APIError,
) *ExpectedError {
	for i := range e.expectedErrs {
		if apiError.ID() == e.expectedErrs[i].ID &&
			apiError.Origin() == e.expectedErrs[i].Origin {

			return &e.expectedErrs[i]
		}
	}
	return nil
}
