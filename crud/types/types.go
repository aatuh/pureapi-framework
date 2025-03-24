package types

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CreateInvokeFn is the function invokes the create endpoint.
type CreateInvokeFn func(
	ctx context.Context, entity databasetypes.Mutator,
) (databasetypes.Mutator, error)

// CreateEntityFactoryFn is the function that creates a new entity.
type CreateEntityFactoryFn func(
	ctx context.Context, input *CreateInputer,
) (databasetypes.Mutator, error)

// ToCreateOutputFn is the function that converts the entity to the endpoint
// output.
type ToCreateOutputFn func(entity databasetypes.Mutator) (CreateOutputer, error)

// BeforeCreateCallback is the function that runs before the create operation.
// It can be used to modify the entity before it is created.
type BeforeCreateCallback func(
	w http.ResponseWriter,
	r *http.Request,
	entity *databasetypes.Mutator,
	input *CreateInputer,
) (databasetypes.Mutator, error)

// GetInvokeFn is the function that invokes the get endpoint.
type GetInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedGetEndpointInput,
	entityFactoryFn repositorytypes.GetterFactoryFn,
) ([]databasetypes.Getter, int, error)

// ToGetOutputFn is the function that converts the entities to the endpoint
// output.
type ToGetOutputFn func(
	entities []databasetypes.Getter, count int,
) (GetOutputer, error)

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
type UpdateEntityFactoryFn func() databasetypes.Mutator

// UpdateInvokeFn is the function that invokes the update endpoint.
type UpdateInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	updater databasetypes.Mutator,
) (int64, error)

// BeforeUpdateCallback is the function that runs before the update operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeUpdateCallback func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedUpdateEndpointInput,
	entity databasetypes.Mutator,
	input *UpdateInputer,
) (*ParsedUpdateEndpointInput, databasetypes.Mutator, error)

// DeleteInvokeFn is the function that invokes the delete endpoint.
type ToDeleteOutputFn func(count int64) (DeleteOutputer, error)

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
type BeforeDeleteCallback func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedDeleteEndpointInput,
	entity databasetypes.Mutator,
	input *DeleteInputer,
) (*ParsedDeleteEndpointInput, databasetypes.Mutator, error)
