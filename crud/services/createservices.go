package services

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CreateInvoke executes the create operation.
//
// Parameters:
//   - ctx: The context.
//   - connFn: The database connection function.
//   - entity: The entity to create.
//   - mutatorRepo: The mutator repository.
//   - txManager: The transaction manager.
//
// Returns:
//   - Entity: The created entity.
//   - error: Any error that occurred during the operation.
func CreateInvoke(
	ctx context.Context,
	connFn repositorytypes.ConnFn,
	entity databasetypes.Mutator,
	mutatorRepo repositorytypes.MutatorRepo,
	txManager repositorytypes.TxManager[databasetypes.Mutator],
) (databasetypes.Mutator, error) {
	return txManager.WithTransaction(
		ctx,
		connFn,
		func(
			ctx context.Context, tx databasetypes.Tx,
		) (databasetypes.Mutator, error) {
			return mutatorRepo.Insert(ctx, tx, entity)
		},
	)
}

// CreateHandler is the handler implementation for the create endpoint.
type CreateHandler struct {
	entityFactoryFn crudtypes.CreateEntityFactoryFn
	createInvokeFn  crudtypes.CreateInvokeFn
	toOutputFn      crudtypes.ToCreateOutputFn
	beforeCallback  crudtypes.BeforeCreateCallback
}

// NewCreateHandler creates a new create handler.
//
// Parameters:
//   - entityFactoryFn: The function that creates a new entity.
//   - createInvokeFn: The function that invokes the create endpoint.
//   - toOutputFn: The function that converts the entity to the endpoint output.
//   - beforeCallback: The optional function that runs before the create
//     operation.
//
// Returns:
//   - *CreateHandler: The new create handler.
func NewCreateHandler(
	entityFactoryFn crudtypes.CreateEntityFactoryFn,
	createInvokeFn crudtypes.CreateInvokeFn,
	toOutputFn crudtypes.ToCreateOutputFn,
	beforeCallback crudtypes.BeforeCreateCallback,
) *CreateHandler {
	return &CreateHandler{
		entityFactoryFn: entityFactoryFn,
		createInvokeFn:  createInvokeFn,
		toOutputFn:      toOutputFn,
		beforeCallback:  beforeCallback,
	}
}

// Handle processes the create endpoint.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - i: The input.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *CreateHandler) Handle(
	w http.ResponseWriter, r *http.Request, i *crudtypes.CreateInputer,
) (any, error) {
	entity, err := h.entityFactoryFn(r.Context(), i)
	if err != nil {
		return nil, err
	}
	if h.beforeCallback != nil {
		entity, err = h.beforeCallback(w, r, &entity, i)
		if err != nil {
			return nil, err
		}
	}
	createdEntity, err := h.createInvokeFn(r.Context(), entity)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(createdEntity)
}

// WithEntityFactoryFn returns a new create handler with the entity factory
// function.
//
// Parameters:
//   - entityFactoryFn: The function that creates a new entity.
//
// Returns:
//   - *CreateHandler: The new create handler.
func (h *CreateHandler) WithEntityFactoryFn(
	entityFactoryFn crudtypes.CreateEntityFactoryFn,
) *CreateHandler {
	new := *h
	new.entityFactoryFn = entityFactoryFn
	return &new
}

// WithCreateInvokeFn returns a new create handler with the create invoke
// function.
//
// Parameters:
//   - createInvokeFn: The function that invokes the create endpoint.
//
// Returns:
//   - *CreateHandler: The new create handler.
func (h *CreateHandler) WithCreateInvokeFn(
	createInvokeFn crudtypes.CreateInvokeFn,
) *CreateHandler {
	new := *h
	new.createInvokeFn = createInvokeFn
	return &new
}

// WithToOutputFn returns a new create handler with the to output function.
//
// Parameters:
//   - toOutputFn: The function that converts the entity to the endpoint output.
//
// Returns:
//   - *CreateHandler: The new create handler.
func (h *CreateHandler) WithToOutputFn(
	toOutputFn crudtypes.ToCreateOutputFn,
) *CreateHandler {
	new := *h
	new.toOutputFn = toOutputFn
	return &new
}

// WithBeforeCallback returns a new create handler with the before callback.
//
// Parameters:
//   - beforeCallback: The function that runs before the create operation.
//
// Returns:
//   - *CreateHandler: The new create handler.
func (h *CreateHandler) WithBeforeCallback(
	beforeCallback crudtypes.BeforeCreateCallback,
) *CreateHandler {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
