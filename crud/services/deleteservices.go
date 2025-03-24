package services

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/dbinput"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// ParseDeleteInput translates API delete input into DB delete input.
//
// Parameters:
//   - apiToDBFields: A map translating API field names to their corresponding
//     database field definitions.
//   - selectors: A slice of API-level selectors.
//   - orders: A slice of API-level orders.
//   - limit: The maximum number of entities to delete.
//
// Returns:
//   - *ParsedDeleteEndpointInput: A pointer to the parsed delete endpoint
//     input.
//   - error: An error if the input is invalid.
func ParseDeleteInput(
	apiToDBFields crudtypes.APIToDBFields,
	selectors dbinput.Selectors,
	orders dbinput.Orders,
	limit int,
) (*crudtypes.ParsedDeleteEndpointInput, error) {
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbSelectors) == 0 {
		return nil, ErrNeedAtLeastOneSelector
	}
	dbOrders, err := orders.TranslateToDBOrders(apiToDBFields)
	if err != nil {
		return nil, err
	}
	return &crudtypes.ParsedDeleteEndpointInput{
		Selectors: dbSelectors,
		DeleteOpts: &repositorytypes.DeleteOptions{
			Limit:  limit,
			Orders: dbOrders,
		},
	}, nil
}

// DeleteInvoke executes the delete operation.
//
// Parameters:
//   - ctx: The context.
//   - connFn: The database connection function.
//   - entity: The entity to delete.
//   - mutatorRepo: The mutator repository.
//   - txManager: The transaction manager.
//
// Returns:
//   - int64: The number of entities deleted.
//   - error: Any error that occurred during the operation.
func DeleteInvoke[Entity databasetypes.Mutator](
	ctx context.Context,
	parsedInput *crudtypes.ParsedDeleteEndpointInput,
	connFn repositorytypes.ConnFn,
	entity Entity,
	mutatorRepo repositorytypes.MutatorRepo[databasetypes.Mutator],
	txManager repositorytypes.TxManager[*int64],
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx databasetypes.Tx) (*int64, error) {
			c, err := mutatorRepo.Delete(
				ctx, tx, entity, parsedInput.Selectors, parsedInput.DeleteOpts,
			)
			return &c, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// DeleteHandler is the handler implementation for the delete endpoint.
type DeleteHandler[Input any] struct {
	parseInputFn    func(input *Input) (*crudtypes.ParsedDeleteEndpointInput, error)
	deleteInvokeFn  crudtypes.DeleteInvokeFn
	toOutputFn      crudtypes.ToDeleteOutputFn
	entityFactoryFn crudtypes.DeleteEntityFactoryFn
	beforeCallback  crudtypes.BeforeDeleteCallback[Input]
}

// NewDeleteHandler creates a new delete handler.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//   - deleteInvokeFn: The function that invokes the delete endpoint.
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//   - entityFactoryFn: The function that creates a new entity.
//   - beforeCallback: The optional function that runs before the delete
//     operation.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func NewDeleteHandler[Input any](
	parseInputFn func(input *Input) (*crudtypes.ParsedDeleteEndpointInput, error),
	deleteInvokeFn crudtypes.DeleteInvokeFn,
	toOutputFn crudtypes.ToDeleteOutputFn,
	entityFactoryFn crudtypes.DeleteEntityFactoryFn,
	beforeCallback crudtypes.BeforeDeleteCallback[Input],
) *DeleteHandler[Input] {
	return &DeleteHandler[Input]{
		parseInputFn:    parseInputFn,
		deleteInvokeFn:  deleteInvokeFn,
		toOutputFn:      toOutputFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
	}
}

// Handle processes the delete endpoint.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - i: The input.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *DeleteHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	parsedInput, err := h.parseInputFn(i)
	if err != nil {
		return nil, err
	}
	entity := h.entityFactoryFn()
	if h.beforeCallback != nil {
		parsedInput, entity, err = h.beforeCallback(
			w, r, parsedInput, entity, i,
		)
		if err != nil {
			return nil, err
		}
	}
	count, err := h.deleteInvokeFn(r.Context(), parsedInput, entity)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(count)
}

// WithParseInputFn returns a new delete handler with the parse input function.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Input]) WithParseInputFn(
	parseInputFn func(input *Input) (*crudtypes.ParsedDeleteEndpointInput, error),
) *DeleteHandler[Input] {
	new := *h
	new.parseInputFn = parseInputFn
	return &new
}

// WithDeleteInvokeFn returns a new delete handler with the delete invoke
// function.
//
// Parameters:
//   - deleteInvokeFn: The function that invokes the delete endpoint.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Input]) WithDeleteInvokeFn(
	deleteInvokeFn crudtypes.DeleteInvokeFn,
) *DeleteHandler[Input] {
	new := *h
	new.deleteInvokeFn = deleteInvokeFn
	return &new
}

// WithToOutputFn returns a new delete handler with the to output function.
//
// Parameters:
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Input]) WithToOutputFn(
	toOutputFn crudtypes.ToDeleteOutputFn,
) *DeleteHandler[Input] {
	new := *h
	new.toOutputFn = toOutputFn
	return &new
}

// WithEntityFactoryFn returns a new delete handler with the entity factory
// function.
//
// Parameters:
//   - entityFactoryFn: The function that creates a new entity.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Input]) WithEntityFactoryFn(
	entityFactoryFn crudtypes.DeleteEntityFactoryFn,
) *DeleteHandler[Input] {
	new := *h
	new.entityFactoryFn = entityFactoryFn
	return &new
}

// WithBeforeDeleteCallback returns a new delete handler with the before delete
// callback.
//
// Parameters:
//   - beforeCallback: The function that runs before the delete operation.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Input]) WithBeforeDeleteCallback(
	beforeCallback crudtypes.BeforeDeleteCallback[Input],
) *DeleteHandler[Input] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
