package db

import "github.com/aatuh/pureapi-core/apierror"

var (
	// ErrDuplicateEntry is returned when a unique constraint is violated.
	ErrDuplicateEntry = apierror.NewAPIError("DUPLICATE_ENTRY")

	// ErrForeignConstraint is returned when a foreign key constraint is
	// violated.
	ErrForeignConstraint = apierror.NewAPIError("FOREIGN_CONSTRAINT_ERROR")

	// ErrNoRows is returned when no rows are found.
	ErrNoRows = apierror.NewAPIError("NO_ROWS")
)
