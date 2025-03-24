package types

import (
	"net/http"
)

// CreateHandler is the handler interface for the create endpoint.
type CreateHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *CreateInputer,
	) (any, error)
}

// GetHandler is the handler interface for the get endpoint.
type GetHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *GetInputer,
	) (any, error)
}

// UpdateHandler is the handler interface for the update endpoint.
type UpdateHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, i *UpdateInputer) (any, error)
}

// DeleteHandler is the handler interface for the delete endpoint.
type DeleteHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *DeleteInputer,
	) (any, error)
}
