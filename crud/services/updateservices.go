package services

import (
	"context"
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	apidb "github.com/aatuh/pureapi-framework/api/db"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/aatuh/pureapi-framework/util/inpututil"
)

type UpdateInputer interface {
	GetSelectors() apidb.APISelectors
	GetUpdates() apidb.APIUpdates
	GetUpsert() bool
}

type UpdateOutputer interface {
	SetCount(count int64)
}

// UpdateInvokeFn is the function that invokes the update endpoint.
type ToUpdateOutputFn func(count int64) (UpdateOutputer, error)

// UpdateEntityFactoryFn is the function that creates a new entity.
type UpdateEntityFactoryFn[Entity database.Mutator] func() Entity

// ParsedUpdateEndpointInput represents a parsed update endpoint db.
type ParsedUpdateEndpointInput struct {
	Selectors db.Selectors
	Updates   []db.Update
	Upsert    bool
}

// UpdateInvokeFn is the function that invokes the update endpoint.
type UpdateInvokeFn[Entity database.Mutator] func(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	updater Entity,
) (int64, error)

// BeforeUpdateCallback is the function that runs before the update operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeUpdateCallback[Entity database.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedUpdateEndpointInput,
	entity Entity,
	input *UpdateInputer,
) (*ParsedUpdateEndpointInput, Entity, error)

// AfterUpdate is a function that is called after an update operation.
type AfterUpdate[Entity database.Mutator] func(
	ctx context.Context, tx database.Tx, count int64,
) (*int64, error)

// ParseUpdateInput translates API update input into DB update db.
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
//     db.
//   - error: An error if the input is invalid.
func ParseUpdateInput(
	apiToDBFields inpututil.APIToDBFields,
	selectors apidb.APISelectors,
	updates apidb.APIUpdates,
	upsert bool,
) (*ParsedUpdateEndpointInput, error) {
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
	return &ParsedUpdateEndpointInput{
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
func UpdateInvoke[Entity database.Mutator](
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	connFn db.ConnFn,
	entity Entity,
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[*int64],
	afterUpdateFn AfterUpdate[Entity],
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (*int64, error) {
			count, err := mutatorRepo.Update(
				ctx, tx, entity, parsedInput.Selectors, parsedInput.Updates,
			)
			if err != nil {
				return nil, err
			}
			if afterUpdateFn != nil {
				return afterUpdateFn(ctx, tx, count)
			}
			return &count, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// UpdateHandler is the handler implementation for the update endpoint.
type UpdateHandler[Entity database.Mutator] struct {
	parseInputFn func(
		input *UpdateInputer,
	) (*ParsedUpdateEndpointInput, error)
	updateInvokeFn  UpdateInvokeFn[Entity]
	toOutputFn      ToUpdateOutputFn
	entityFactoryFn UpdateEntityFactoryFn[Entity]
	beforeCallback  BeforeUpdateCallback[Entity]
}

// NewUpdateHandler creates a new update handler.
//
// Parameters:
//   - parseInputFn: The function that parses the db.
//   - updateInvokeFn: The function that invokes the update endpoint.
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//   - entityFactoryFn: The function that creates a new entity.
//   - beforeCallback: The optional function that runs before the update
//     operation.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func NewUpdateHandler[Entity database.Mutator](
	parseInputFn func(
		input *UpdateInputer,
	) (*ParsedUpdateEndpointInput, error),
	updateInvokeFn UpdateInvokeFn[Entity],
	toOutputFn ToUpdateOutputFn,
	entityFactoryFn UpdateEntityFactoryFn[Entity],
	beforeCallback BeforeUpdateCallback[Entity],
) *UpdateHandler[Entity] {
	return &UpdateHandler[Entity]{
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
//   - i: The db.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *UpdateHandler[Entity]) Handle(
	w http.ResponseWriter, r *http.Request, i *UpdateInputer,
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
//   - parseInputFn: The function that parses the db.
//
// Returns:
//   - *UpdateHandler: The new update handler.
func (h *UpdateHandler[Entity]) WithParseInputFn(
	parseInputFn func(input *UpdateInputer,
	) (*ParsedUpdateEndpointInput, error),
) *UpdateHandler[Entity] {
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
func (h *UpdateHandler[Entity]) WithUpdateInvokeFn(
	updateInvokeFn UpdateInvokeFn[Entity],
) *UpdateHandler[Entity] {
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
func (h *UpdateHandler[Entity]) WithToOutputFn(
	toOutputFn ToUpdateOutputFn,
) *UpdateHandler[Entity] {
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
func (h *UpdateHandler[Entity]) WithEntityFactoryFn(
	entityFactoryFn UpdateEntityFactoryFn[Entity],
) *UpdateHandler[Entity] {
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
func (h *UpdateHandler[Entity]) WithBeforeUpdateCallback(
	beforeCallback BeforeUpdateCallback[Entity],
) *UpdateHandler[Entity] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
