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
type CreateEntityFactoryFn[Input any, Entity databasetypes.Mutator] func(
	ctx context.Context, input *Input,
) (Entity, error)

// ToCreateOutputFn is the function that converts the entity to the endpoint
// output.
type ToCreateOutputFn[Entity any] func(entity Entity) (any, error)

// BeforeCreateCallback is the function that runs before the create operation.
// It can be used to modify the entity before it is created.
type BeforeCreateCallback[Input any, Entity databasetypes.Mutator] func(
	w http.ResponseWriter, r *http.Request, entity *Entity, input *Input,
) (Entity, error)

// GetInvokeFn is the function that invokes the get endpoint.
type GetInvokeFn[Entity databasetypes.Getter] func(
	ctx context.Context,
	parsedInput *ParsedGetEndpointInput,
	entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
) ([]Entity, int, error)

// ToGetOutputFn is the function that converts the entities to the endpoint
// output.
type ToGetOutputFn[Entity any, Output any] func(
	entities []Entity, count int,
) (Output, error)

// BeforeGetCallback is the function that runs before the get operation.
// It can be used to modify the parsed input before it is used.
type BeforeGetCallback[Input any, Entity databasetypes.Getter] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedGetEndpointInput,
	input *Input,
) (*ParsedGetEndpointInput, error)

// UpdateInvokeFn is the function that invokes the update endpoint.
type ToUpdateOutputFn func(count int64) (any, error)

// UpdateEntityFactoryFn is the function that creates a new entity.
type UpdateEntityFactoryFn func() databasetypes.Mutator

// UpdateInvokeFn is the function that invokes the update endpoint.
type UpdateInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	updater databasetypes.Mutator,
) (int64, error)

// BeforeUpdateCallback is the function that runs before the update operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeUpdateCallback[Input any] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedUpdateEndpointInput,
	entity databasetypes.Mutator,
	input *Input,
) (*ParsedUpdateEndpointInput, databasetypes.Mutator, error)

// DeleteInvokeFn is the function that invokes the delete endpoint.
type ToDeleteOutputFn func(count int64) (any, error)

// DeleteEntityFactoryFn is the function that creates a new entity.
type DeleteEntityFactoryFn func() databasetypes.Mutator

// DeleteInvokeFn is the function that invokes the delete endpoint.
type DeleteInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	entity databasetypes.Mutator,
) (int64, error)

// BeforeDeleteCallback is the function that runs before the delete operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeDeleteCallback[Input any] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedDeleteEndpointInput,
	entity databasetypes.Mutator,
	input *Input,
) (*ParsedDeleteEndpointInput, databasetypes.Mutator, error)
