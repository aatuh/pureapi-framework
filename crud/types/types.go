package types

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CreateInvokeFn is the function invokes the create endpoint.
type CreateInvokeFn[Entity databasetypes.Mutator] func(
	ctx context.Context, entity Entity,
) (Entity, error)

// CreateEntityFactoryFn is the function that creates a new entity.
type CreateEntityFactoryFn[Entity databasetypes.Mutator] func(
	ctx context.Context, input *CreateInputer[Entity],
) (Entity, error)

// ToCreateOutputFn is the function that converts the entity to the endpoint
// output.
type ToCreateOutputFn[Entity databasetypes.Mutator] func(
	entity Entity,
) (CreateOutputer[Entity], error)

// BeforeCreateCallback is the function that runs before the create operation.
// It can be used to modify the entity before it is created.
type BeforeCreateCallback[Entity databasetypes.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	entity *Entity,
	input *CreateInputer[Entity],
) (Entity, error)

// GetInvokeFn is the function that invokes the get endpoint.
type GetInvokeFn[Entity databasetypes.Getter] func(
	ctx context.Context,
	parsedInput *ParsedGetEndpointInput,
	entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
) ([]Entity, int, error)

// ToGetOutputFn is the function that converts the entities to the endpoint
// output.
type ToGetOutputFn[Entity databasetypes.Getter] func(
	entities []Entity, count int,
) (GetOutputer[Entity], error)

// BeforeGetCallback is the function that runs before the get operation.
// It can be used to modify the parsed input before it is used.
type BeforeGetCallback func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedGetEndpointInput,
	input *GetInputer,
) (*ParsedGetEndpointInput, error)

// UpdateInvokeFn is the function that invokes the update endpoint.
type ToUpdateOutputFn func(count int64) (UpdateOutputer, error)

// UpdateEntityFactoryFn is the function that creates a new entity.
type UpdateEntityFactoryFn[Entity databasetypes.Mutator] func() Entity

// UpdateInvokeFn is the function that invokes the update endpoint.
type UpdateInvokeFn[Entity databasetypes.Mutator] func(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	updater Entity,
) (int64, error)

// BeforeUpdateCallback is the function that runs before the update operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeUpdateCallback[Entity databasetypes.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedUpdateEndpointInput,
	entity Entity,
	input *UpdateInputer,
) (*ParsedUpdateEndpointInput, Entity, error)

// DeleteInvokeFn is the function that invokes the delete endpoint.
type ToDeleteOutputFn func(count int64) (DeleteOutputer, error)

// DeleteEntityFactoryFn is the function that creates a new entity.
type DeleteEntityFactoryFn[Entity databasetypes.Mutator] func() Entity

// DeleteInvokeFn is the function that invokes the delete endpoint.
type DeleteInvokeFn[Entity databasetypes.Mutator] func(
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	entity Entity,
) (int64, error)

// BeforeDeleteCallback is the function that runs before the delete operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeDeleteCallback[Entity databasetypes.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedDeleteEndpointInput,
	entity Entity,
	input *DeleteInputer,
) (*ParsedDeleteEndpointInput, Entity, error)
