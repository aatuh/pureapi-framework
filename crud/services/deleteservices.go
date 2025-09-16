package services

import (
	"context"
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	apidb "github.com/aatuh/pureapi-framework/api/db"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/aatuh/pureapi-framework/util/inpututil"
)

type DeleteInputer interface {
	GetSelectors() apidb.APISelectors
}

type DeleteOutputer interface {
	SetCount(count int64)
}

// DeleteInvokeFn is the function that invokes the delete endpoint.
type ToDeleteOutputFn func(count int64) (DeleteOutputer, error)

// DeleteEntityFactoryFn is the function that creates a new entity.
type DeleteEntityFactoryFn[Entity database.Mutator] func() Entity

// ParsedDeleteEndpointInput represents a parsed delete endpoint db.
type ParsedDeleteEndpointInput struct {
	Selectors  db.Selectors
	DeleteOpts *db.DeleteOptions
}

// DeleteInvokeFn is the function that invokes the delete endpoint.
type DeleteInvokeFn[Entity database.Mutator] func(
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	entity Entity,
) (int64, error)

// BeforeDeleteCallback is the function that runs before the delete operation.
// It can be used to modify the parsed input and entity before they are used.
type BeforeDeleteCallback[Entity database.Mutator] func(
	w http.ResponseWriter,
	r *http.Request,
	parsedInput *ParsedDeleteEndpointInput,
	entity Entity,
	input DeleteInputer,
) (*ParsedDeleteEndpointInput, Entity, error)

// AfterDelete is a function that is called after a delete operation.
type AfterDelete func(
	ctx context.Context, tx database.Tx, count int64,
) (*int64, error)

// ParseDeleteInput translates API delete input into DB delete db.
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
//     db.
//   - error: An error if the input is invalid.
func ParseDeleteInput(
	apiToDBFields inpututil.APIToDBFields,
	selectors apidb.APISelectors,
	orders apidb.Orders,
	limit int,
) (*ParsedDeleteEndpointInput, error) {
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
	return &ParsedDeleteEndpointInput{
		Selectors: dbSelectors,
		DeleteOpts: &db.DeleteOptions{
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
func DeleteInvoke[Entity database.Mutator](
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	connFn db.ConnFn,
	entity Entity,
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[*int64],
	afterDeleteFn AfterDelete,
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (*int64, error) {
			count, err := mutatorRepo.Delete(
				ctx, tx, entity, parsedInput.Selectors, parsedInput.DeleteOpts,
			)
			if err != nil {
				return nil, err
			}
			if afterDeleteFn != nil {
				return afterDeleteFn(ctx, tx, count)
			}
			return &count, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// DeleteHandler is the handler implementation for the delete endpoint.
type DeleteHandler[Entity database.Mutator] struct {
	parseInputFn func(
		input DeleteInputer,
	) (*ParsedDeleteEndpointInput, error)
	deleteInvokeFn  DeleteInvokeFn[Entity]
	toOutputFn      ToDeleteOutputFn
	entityFactoryFn DeleteEntityFactoryFn[Entity]
	beforeCallback  BeforeDeleteCallback[Entity]
}

// NewDeleteHandler creates a new delete handler.
//
// Parameters:
//   - parseInputFn: The function that parses the db.
//   - deleteInvokeFn: The function that invokes the delete endpoint.
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//   - entityFactoryFn: The function that creates a new entity.
//   - beforeCallback: The optional function that runs before the delete
//     operation.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func NewDeleteHandler[Entity database.Mutator](
	parseInputFn func(
		input DeleteInputer,
	) (*ParsedDeleteEndpointInput, error),
	deleteInvokeFn DeleteInvokeFn[Entity],
	toOutputFn ToDeleteOutputFn,
	entityFactoryFn DeleteEntityFactoryFn[Entity],
	beforeCallback BeforeDeleteCallback[Entity],
) *DeleteHandler[Entity] {
	return &DeleteHandler[Entity]{
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
//   - i: The db.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *DeleteHandler[Entity]) Handle(
	w http.ResponseWriter, r *http.Request, i *DeleteInputer,
) (any, error) {
	parsedInput, err := h.parseInputFn(*i)
	if err != nil {
		return nil, err
	}
	entity := h.entityFactoryFn()
	if h.beforeCallback != nil {
		parsedInput, entity, err = h.beforeCallback(
			w, r, parsedInput, entity, *i,
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
//   - parseInputFn: The function that parses the db.
//
// Returns:
//   - *DeleteHandler: The new delete handler.
func (h *DeleteHandler[Entity]) WithParseInputFn(
	parseInputFn func(
		input DeleteInputer,
	) (*ParsedDeleteEndpointInput, error),
) *DeleteHandler[Entity] {
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
func (h *DeleteHandler[Entity]) WithDeleteInvokeFn(
	deleteInvokeFn DeleteInvokeFn[Entity],
) *DeleteHandler[Entity] {
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
func (h *DeleteHandler[Entity]) WithToOutputFn(
	toOutputFn ToDeleteOutputFn,
) *DeleteHandler[Entity] {
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
func (h *DeleteHandler[Entity]) WithEntityFactoryFn(
	entityFactoryFn DeleteEntityFactoryFn[Entity],
) *DeleteHandler[Entity] {
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
func (h *DeleteHandler[Entity]) WithBeforeDeleteCallback(
	beforeCallback BeforeDeleteCallback[Entity],
) *DeleteHandler[Entity] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
