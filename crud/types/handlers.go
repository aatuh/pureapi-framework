package types

import (
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
)

// CreateHandler is the handler interface for the create endpoint.
type CreateHandler[Entity databasetypes.Mutator] interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *CreateInputer[Entity],
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
