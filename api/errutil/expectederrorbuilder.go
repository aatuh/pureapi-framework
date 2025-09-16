package errutil

// ExpectedErrorBuilder is a builder for ExpectedErrors. It can be used to get
// a list of ExpectedErrors with a system ID.
type ExpectedErrorBuilder struct {
	systemID string
	errs     []ExpectedError
}

// NewExpectedErrorBuilder creates a new ExpectedErrorBuilder.
//
// Parameters:
//   - systemID: The system ID to add to the errors.
//
// Returns:
//   - *ExpectedErrorBuilder: The new ExpectedErrorBuilder.
func NewExpectedErrorBuilder(systemID string) *ExpectedErrorBuilder {
	return &ExpectedErrorBuilder{
		systemID: systemID,
	}
}

// With adds errors to the builder.
//
// Parameters:
//   - errors: The errors to add.
//
// Returns:
//   - *ErrorBuilder: The builder with the errors added.
func (b *ExpectedErrorBuilder) WithErrors(
	errors ExpectedErrors,
) *ExpectedErrorBuilder {
	b.errs = append(b.errs, errors...)
	return b
}

// Build returns the ExpectedErrors with the system ID.
//
// Returns:
//   - ExpectedErrors: The ExpectedErrors with the system ID.
func (b *ExpectedErrorBuilder) Build() ExpectedErrors {
	return ExpectedErrors(b.errs).WithOrigin(b.systemID)
}
