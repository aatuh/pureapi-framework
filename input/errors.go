package input

import "github.com/pureapi/pureapi-core/util"

// TODO: Move elsewhere

// Commmon input errors.
var (
	ErrValidation    = util.NewAPIError("VALIDATION_ERROR")
	ErrInputDecoding = util.NewAPIError("ERROR_DECODING_INPUT")
	ErrInvalidInput  = util.NewAPIError("INVALID_INPUT")
)
