package api

// ErrorBuilder is a builder for ExpectedErrors. It can be used to add errors
// under a specific system ID.
type ErrorBuilder struct {
	systemID string
	errs     []ExpectedError
}

// NewErrorBuilder creates a new ErrorBuilder.
//
// Parameters:
//   - systemID: The system ID to add to the errors.
//
// Returns:
//   - *ErrorBuilder: The new ErrorBuilder.
func NewErrorBuilder(systemID string) *ErrorBuilder {
	return &ErrorBuilder{
		systemID: systemID,
	}
}

// With adds errors to the builder.
//
// Parameters:
//   - errs: The errors to add.
//
// Returns:
//   - *ErrorBuilder: The builder with the errors added.
func (b *ErrorBuilder) With(errs ExpectedErrors) *ErrorBuilder {
	b.errs = append(b.errs, errs...)
	return b
}

// Build returns the ExpectedErrors with the system ID.
//
// Returns:
//   - ExpectedErrors: The ExpectedErrors with the system ID.
func (b *ErrorBuilder) Build() ExpectedErrors {
	return ExpectedErrors(b.errs).WithOrigin(b.systemID)
}
