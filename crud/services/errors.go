package services

import "github.com/aatuh/pureapi-core/apierror"

// Common errors.
var (
	ErrNeedAtLeastOneSelector = apierror.NewAPIError("NEED_AT_LEAST_ONE_SELECTOR")
	ErrNeedAtLeastOneUpdate   = apierror.NewAPIError("NEED_AT_LEAST_ONE_UPDATE")
)
