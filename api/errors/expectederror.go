package errors

import (
	"github.com/pureapi/pureapi-core/util"
	"github.com/pureapi/pureapi-core/util/types"
)

// ExpectedError represents an expected error configuration.
// It defines how to handle specific errors that are anticipated.
type ExpectedError struct {
	ID         string // The ID of the expected error.
	MaskedID   string // An optional ID to mask the original error ID in the response.
	Status     int    // The HTTP status code to return for this error.
	PublicData bool   // Whether to include the error data in the response.
	Origin     string // The orign of the error.
}

// NewExpectedError creates a new ExpectedError.
//
// Parameters:
//   - id: The ID of the expected error.
//   - status: The HTTP status code to return for this error.
//   - origin: The origin of the error.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func NewExpectedError(id string, status int, origin string) ExpectedError {
	return ExpectedError{
		ID:     id,
		Status: status,
		Origin: origin,
	}
}

// WithID returns a new ExpectedError with the given ID.
//
// Parameters:
//   - id: The ID to use in the response.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithID(id string) ExpectedError {
	e.ID = id
	return e
}

// WithMaskedID returns a new ExpectedError with the given masked ID.
//
// Parameters:
//   - maskedID: The masked ID to use in the response.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithMaskedID(maskedID string) ExpectedError {
	e.MaskedID = maskedID
	return e
}

// WithStatus returns a new ExpectedError with the given status.
//
// Parameters:
//   - status: The status to use in the response.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithStatus(status int) ExpectedError {
	e.Status = status
	return e
}

// WithPublicData returns a new ExpectedError with the public data flag set.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithPublicData(isPublic bool) ExpectedError {
	e.PublicData = isPublic
	return e
}

// WithOrigin returns a new ExpectedError with the given origin.
//
// Parameters:
//   - origin: The origin to use in the response.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithOrigin(origin string) ExpectedError {
	e.Origin = origin
	return e
}

// maskAPIError masks the ID and data of the given API error based on the
// configuration of the ExpectedError.
func (e *ExpectedError) maskAPIError(
	apiError types.APIError,
) (int, types.APIError) {
	// If a masked ID is defined, use it. Otherwise, use the original ID.
	var useErrorID string
	if e.MaskedID != "" {
		useErrorID = e.MaskedID
	} else {
		useErrorID = e.ID
	}

	// If the error data is public, use it. Otherwise, use nil.
	var useData any
	if e.PublicData {
		useData = apiError.Data()
	} else {
		useData = nil
	}

	return e.Status, util.APIErrorFrom(apiError).
		WithID(useErrorID).WithData(useData)
}

// ExpectedErrors is a slice of ExpectedError.
type ExpectedErrors []ExpectedError

// WithErrors returns a new slice with the errors appended to the slice.
//
// Parameters:
//   - errs: The errors to append.
//
// Returns:
//   - ExpectedErrors: The new slice with the errors appended.
func (e ExpectedErrors) WithErrors(errors ...ExpectedError) ExpectedErrors {
	newSlice := append([]ExpectedError{}, e...)
	return append(newSlice, errors...)
}

// WithOrigin makes all errors in the slice have the given origin and returns
// a new slice with the origin set for all errors.
//
// Parameters:
//   - origin: The origin to set for all errors.
//
// Returns:
//   - ExpectedErrors: The new slice with the origin set for all errors.
func (e ExpectedErrors) WithOrigin(origin string) ExpectedErrors {
	expectedErrors := []ExpectedError{}
	for i := range e {
		expectedErrors = append(expectedErrors, e[i].WithOrigin(origin))
	}
	return expectedErrors
}

// GetByID returns the ExpectedError with the given ID, or nil if not found.
//
// Parameters:
//   - id: The ID of the expected error.
//
// Returns:
//   - *ExpectedError: The ExpectedError with the given ID, or nil if not found.
func (e ExpectedErrors) GetByID(id string) *ExpectedError {
	for i := range e {
		if e[i].ID == id {
			return &e[i]
		}
	}
	return nil
}
