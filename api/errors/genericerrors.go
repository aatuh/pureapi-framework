package errors

import (
	"net/http"
)

// GenericErrors returns a list of generic API errors.
//
// Returns:
//   - []ExpectedError: A list of generic API errors.
func GenericErrors() ExpectedErrors {
	return []ExpectedError{
		{ID: ErrInputDecoding.ID(), MaskedID: ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: ErrValidation.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}
