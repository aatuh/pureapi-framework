package hooks

import "context"

// AuthorizationPolicy validates whether the current request is allowed.
type AuthorizationPolicy interface {
	Authorize(ctx context.Context, input any) error
}

// AuthorizationPolicyFunc lifts a function into an AuthorizationPolicy.
type AuthorizationPolicyFunc func(ctx context.Context, input any) error

// Authorize implements AuthorizationPolicy.
func (f AuthorizationPolicyFunc) Authorize(ctx context.Context, input any) error {
	return f(ctx, input)
}

// AuthorizationError maps to an explicit catalog entry (e.g., unauthorized, forbidden).
type AuthorizationError struct {
	catalogID string
	message   string
}

// Error implements error.
func (e AuthorizationError) Error() string {
	if e.message != "" {
		return e.message
	}
	return e.catalogID
}

// CatalogID allows plugging into the error mapper.
func (e AuthorizationError) CatalogID() string {
	return e.catalogID
}

// WireMessage optionally overrides the wire error message.
func (e AuthorizationError) WireMessage() string {
	return e.message
}

// NewAuthorizationError returns an AuthorizationError with the given catalog ID and message.
func NewAuthorizationError(catalogID, message string) AuthorizationError {
	return AuthorizationError{catalogID: catalogID, message: message}
}

// ErrUnauthorized returns an AuthorizationError mapped to the "unauthorized" catalog entry.
func ErrUnauthorized(message string) AuthorizationError {
	return NewAuthorizationError("unauthorized", message)
}

// ErrForbidden returns an AuthorizationError mapped to the "forbidden" catalog entry.
func ErrForbidden(message string) AuthorizationError {
	return NewAuthorizationError("forbidden", message)
}
