package dbquery

import "github.com/pureapi/pureapi-core/util"

// Commmon database errors.
var (
	ErrDuplicateEntry    = util.NewAPIError("DUPLICATE_ENTRY")
	ErrForeignConstraint = util.NewAPIError("FOREIGN_CONSTRAINT_ERROR")
	ErrNoRows            = util.NewAPIError("NO_ROWS")
)
