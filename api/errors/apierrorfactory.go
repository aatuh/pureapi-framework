package errors

import "github.com/pureapi/pureapi-core/util"

// APIErrorFactory is a factory for creating API errors with the ID and origin
// set.
type APIErrorFactory struct {
	SystemID string
}

// NewErrorFactory creates a new ErrorFactory for the given system ID.
func NewErrorFactory(systemID string) *APIErrorFactory {
	return &APIErrorFactory{
		SystemID: systemID,
	}
}

// New creates a new API error with the origin already set.
func (f *APIErrorFactory) New(id string) *util.DefaultAPIError {
	return util.NewAPIError(id).WithOrigin(f.SystemID)
}
