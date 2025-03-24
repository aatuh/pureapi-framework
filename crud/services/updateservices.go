package services

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/dbinput"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// ParseUpdateInput translates API update input into DB update input.
//
// Parameters:
//   - apiToDBFields: A map translating API field names to their corresponding
//     database field definitions.
//   - selectors: A slice of API-level selectors.
//   - updates: A map of API-level updates.
//   - upsert: A boolean indicating whether to upsert.
//
// Returns:
//   - *ParsedUpdateEndpointInput: A pointer to the parsed update endpoint
//     input.
//   - error: An error if the input is invalid.
func ParseUpdateInput(
	apiToDBFields crudtypes.APIToDBFields,
	selectors dbinput.Selectors,
	updates dbinput.Updates,
	upsert bool,
) (*crudtypes.ParsedUpdateEndpointInput, error) {
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbSelectors) == 0 {
		return nil, ErrNeedAtLeastOneSelector
	}
	dbUpdates, err := updates.ToDBUpdates(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbUpdates) == 0 {
		return nil, ErrNeedAtLeastOneUpdate
	}
	return &crudtypes.ParsedUpdateEndpointInput{
		Selectors: dbSelectors,
		Updates:   dbUpdates,
		Upsert:    upsert,
	}, nil
}

// UpdateInvoke executes the update operation.
//
// Parameters:
//   - ctx: The context.
//   - connFn: The database connection function.
//   - entity: The entity to update.
//   - mutatorRepo: The mutator repository.
//   - txManager: The transaction manager.
//
// Returns:
//   - int64: The number of entities updated.
//   - error: Any error that occurred during the operation.
func UpdateInvoke(
	ctx context.Context,
	parsedInput *crudtypes.ParsedUpdateEndpointInput,
	connFn repositorytypes.ConnFn,
	entity databasetypes.Mutator,
	mutatorRepo repositorytypes.MutatorRepo[databasetypes.Mutator],
	txManager repositorytypes.TxManager[*int64],
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx databasetypes.Tx) (*int64, error) {
			c, err := mutatorRepo.Update(
				ctx, tx, entity, parsedInput.Selectors, parsedInput.Updates,
			)
			return &c, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// UpdateHandler is the handler implementation for the update endpoint.
type UpdateHandler[Input any] struct {
	parseInputFn    func(input *Input) (*crudtypes.ParsedUpdateEndpointInput, error)
	updateInvokeFn  crudtypes.UpdateInvokeFn
	toOutputFn      crudtypes.ToUpdateOutputFn
	entityFactoryFn crudtypes.UpdateEntityFactoryFn
	beforeCallback  crudtypes.BeforeUpdateCallback[Input]
}

// NewUpdateHandler creates a new update handler.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//   - updateInvokeFn: The function that invokes the update endpoint.
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//   - entityFactoryFn: The function that creates a new entity.
//   - beforeCallback: The optional function that runs before the update
//     operation.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func NewUpdateHandler[Input any](
	parseInputFn func(input *Input) (*crudtypes.ParsedUpdateEndpointInput, error),
	updateInvokeFn crudtypes.UpdateInvokeFn,
	toOutputFn crudtypes.ToUpdateOutputFn,
	entityFactoryFn crudtypes.UpdateEntityFactoryFn,
	beforeCallback crudtypes.BeforeUpdateCallback[Input],
) *UpdateHandler[Input] {
	return &UpdateHandler[Input]{
		parseInputFn:    parseInputFn,
		updateInvokeFn:  updateInvokeFn,
		toOutputFn:      toOutputFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
	}
}

// Handle processes the update endpoint.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - i: The input.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *UpdateHandler[Input]) Handle(
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
	count, err := h.updateInvokeFn(r.Context(), parsedInput, entity)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(count)
}

// WithParseInputFn returns a new update handler with the parse input function.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Input]) WithParseInputFn(
	parseInputFn func(input *Input) (*crudtypes.ParsedUpdateEndpointInput, error),
) *UpdateHandler[Input] {
	new := *h
	new.parseInputFn = parseInputFn
	return &new
}

// WithUpdateInvokeFn returns a new update handler with the update invoke
// function.
//
// Parameters:
//   - updateInvokeFn: The function that invokes the update endpoint.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Input]) WithUpdateInvokeFn(
	updateInvokeFn crudtypes.UpdateInvokeFn,
) *UpdateHandler[Input] {
	new := *h
	new.updateInvokeFn = updateInvokeFn
	return &new
}

// WithToOutputFn returns a new update handler with the to output function.
//
// Parameters:
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Input]) WithToOutputFn(
	toOutputFn crudtypes.ToUpdateOutputFn,
) *UpdateHandler[Input] {
	new := *h
	new.toOutputFn = toOutputFn
	return &new
}

// WithEntityFactoryFn returns a new update handler with the entity factory
// function.
//
// Parameters:
//   - entityFactoryFn: The function that creates a new entity.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Input]) WithEntityFactoryFn(
	entityFactoryFn crudtypes.UpdateEntityFactoryFn,
) *UpdateHandler[Input] {
	new := *h
	new.entityFactoryFn = entityFactoryFn
	return &new
}

// WithBeforeUpdateCallback returns a new update handler with the before update
// callback.
//
// Parameters:
//   - beforeCallback: The function that runs before the update operation.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Input]) WithBeforeUpdateCallback(
	beforeCallback crudtypes.BeforeUpdateCallback[Input],
) *UpdateHandler[Input] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
