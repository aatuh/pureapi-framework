package errutil

import "github.com/aatuh/pureapi-core/apierror"

// APIErrorFactory is a factory for creating API errors with both the ID and
// origin set.
type APIErrorFactory struct {
	SystemID string
}

// NewErrorFactory creates a new ErrorFactory for the given system ID.
func NewErrorFactory(systemID string) *APIErrorFactory {
	return &APIErrorFactory{
		SystemID: systemID,
	}
}

// APIError creates a new API error with the origin already set.
func (f *APIErrorFactory) APIError(id string) *apierror.DefaultAPIError {
	return apierror.NewAPIError(id).WithOrigin(f.SystemID)
}
