package services

import (
	"context"
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-framework/db"
)

// CreateInputer is the input interface for the create endpoint.
type CreateInputer[Entity any] interface {
	GetEntity() Entity
}

// CreateOutputer is the output interface for the create endpoint.
type CreateOutputer[Entity any] interface {
	SetEntities(entities []Entity)
}

// CreateInvokeFn is the function invokes the create endpoint.
type CreateInvokeFn[Entity database.Mutator] func(
	ctx context.Context, entity Entity,
) (Entity, error)

// CreateEntityFactoryFn is the function that creates a new entity.
type CreateEntityFactoryFn[Entity database.Mutator] func(
	ctx context.Context, input *CreateInputer[Entity],
) (Entity, error)

// ToCreateOutputFn is the function that converts the entity to the endpoint
// output.
type ToCreateOutputFn[Entity database.Mutator] func(
	entity Entity,
) (CreateOutputer[Entity], error)

// BeforeCreateCallback is the function that runs before the create operation.
// It can be used to modify the entity before it is created.
type BeforeCreateCallback[Entity database.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	entity Entity,
	input CreateInputer[Entity],
) (Entity, error)

// AfterCreate is a function that is called after a create operation.
type AfterCreate[Entity database.Mutator] func(
	ctx context.Context, tx database.Tx, entity Entity,
) (Entity, error)

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
func CreateInvoke[Entity database.Mutator](
	ctx context.Context,
	connFn db.ConnFn,
	entity Entity,
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[Entity],
	afterCreateFn AfterCreate[Entity],
) (Entity, error) {
	return txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (Entity, error) {
			entity, err := mutatorRepo.Insert(ctx, tx, entity)
			if err != nil {
				return entity, err
			}
			if afterCreateFn != nil {
				return afterCreateFn(ctx, tx, entity)
			}
			return entity, nil
		},
	)
}

// CreateHandler is the handler implementation for the create endpoint.
type CreateHandler[Entity database.Mutator] struct {
	entityFactoryFn CreateEntityFactoryFn[Entity]
	createInvokeFn  CreateInvokeFn[Entity]
	toOutputFn      ToCreateOutputFn[Entity]
	beforeCallback  BeforeCreateCallback[Entity]
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
func NewCreateHandler[Entity database.Mutator](
	entityFactoryFn CreateEntityFactoryFn[Entity],
	createInvokeFn CreateInvokeFn[Entity],
	toOutputFn ToCreateOutputFn[Entity],
	beforeCallback BeforeCreateCallback[Entity],
) *CreateHandler[Entity] {
	return &CreateHandler[Entity]{
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
func (h *CreateHandler[Entity]) Handle(
	w http.ResponseWriter, r *http.Request, i *CreateInputer[Entity],
) (any, error) {
	entity, err := h.entityFactoryFn(r.Context(), i)
	if err != nil {
		return nil, err
	}
	if h.beforeCallback != nil {
		entity, err = h.beforeCallback(w, r, entity, *i)
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
func (h *CreateHandler[Entity]) WithEntityFactoryFn(
	entityFactoryFn CreateEntityFactoryFn[Entity],
) *CreateHandler[Entity] {
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
func (h *CreateHandler[Entity]) WithCreateInvokeFn(
	createInvokeFn CreateInvokeFn[Entity],
) *CreateHandler[Entity] {
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
func (h *CreateHandler[Entity]) WithToOutputFn(
	toOutputFn ToCreateOutputFn[Entity],
) *CreateHandler[Entity] {
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
func (h *CreateHandler[Entity]) WithBeforeCallback(
	beforeCallback BeforeCreateCallback[Entity],
) *CreateHandler[Entity] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
