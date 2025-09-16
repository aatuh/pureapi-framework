package errutil

import (
	"net/http"

	"github.com/aatuh/pureapi-framework/api/input"
)

// GenericErrors returns a list of generic API errors.
//
// Returns:
//   - []ExpectedError: A list of generic API errors.
func GenericErrors() ExpectedErrors {
	return []ExpectedError{
		{ID: input.ErrInputDecoding.ID(), MaskedID: input.ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: input.ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrValidation.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}
