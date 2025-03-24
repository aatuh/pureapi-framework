package services

import (
	"github.com/pureapi/pureapi-core/util"
)

// Common errors.
var (
	ErrNeedAtLeastOneSelector = util.NewAPIError("NEED_AT_LEAST_ONE_SELECTOR")
	ErrNeedAtLeastOneUpdate   = util.NewAPIError("NEED_AT_LEAST_ONE_UPDATE")
)
