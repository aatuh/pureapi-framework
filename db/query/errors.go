package query

import "github.com/pureapi/pureapi-core/util"

var (
	// ErrDuplicateEntry is returned when a unique constraint is violated.
	ErrDuplicateEntry = util.NewAPIError("DUPLICATE_ENTRY")

	// ErrForeignConstraint is returned when a foreign key constraint is
	// violated.
	ErrForeignConstraint = util.NewAPIError("FOREIGN_CONSTRAINT_ERROR")

	// ErrNoRows is returned when no rows are found.
	ErrNoRows = util.NewAPIError("NO_ROWS")
)
