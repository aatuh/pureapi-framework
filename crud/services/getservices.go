package services

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/dbinput"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// ParseGetInput translates API parameters to DB parameters.
//
// Parameters:
//   - apiToDBFields: A map translating API field names to their corresponding
//     database field definitions.
//   - selectors: A slice of API-level selectors.
//   - orders: A slice of API-level orders.
//   - inputPage: A pointer to the input page.
//   - maxPage: The maximum page size.
//   - count: A boolean indicating whether to return the count.
//
// Returns:
//   - *ParsedGetEndpointInput: A pointer to the parsed get endpoint input.
//   - error: An error if the input is invalid.
func ParseGetInput(
	apiToDBFields crudtypes.APIToDBFields,
	selectors dbinput.Selectors,
	orders dbinput.Orders,
	inputPage *dbinput.Page,
	maxPage int,
	count bool,
) (*crudtypes.ParsedGetEndpointInput, error) {
	dbOrders, err := orders.TranslateToDBOrders(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if inputPage == nil {
		inputPage = &dbinput.Page{Offset: 0, Limit: maxPage}
	}
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	return &crudtypes.ParsedGetEndpointInput{
		Orders:    dbOrders,
		Selectors: dbSelectors,
		Page:      inputPage.ToDBPage(),
		Count:     count,
	}, nil
}

// GetInvoke executes the get operation.
//
// Parameters:
//   - ctx: The context.
//   - connFn: The database connection function.
//   - entityFactoryFn: The entity factory function.
//   - readerRepo: The reader repository.
//   - txManager: The transaction manager.
//
// Returns:
//   - []Entity: The entities.
//   - error: Any error that occurred during the operation.
func GetInvoke[Getter databasetypes.Getter](
	ctx context.Context,
	parsedInput *crudtypes.ParsedGetEndpointInput,
	connFn repositorytypes.ConnFn,
	entityFactoryFn repositorytypes.GetterFactoryFn[Getter],
	readerRepo repositorytypes.ReaderRepo[Getter],
	_ repositorytypes.TxManager[Getter],
) ([]Getter, int, error) {
	conn, err := connFn()
	if err != nil {
		return nil, 0, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	if parsedInput.Count {
		count, err := readerRepo.Count(
			ctx, tx, parsedInput.Selectors, parsedInput.Page, entityFactoryFn,
		)
		if err != nil {
			return nil, 0, err
		}
		return nil, count, nil
	}
	entities, err := readerRepo.GetMany(
		ctx,
		tx,
		entityFactoryFn,
		&repositorytypes.GetOptions{
			Selectors: parsedInput.Selectors,
			Orders:    parsedInput.Orders,
			Page:      parsedInput.Page,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return entities, len(entities), nil
}

// GetHandler is the handler for the get endpoint.
type GetHandler[Entity databasetypes.Getter, Input any, Output any] struct {
	parseInputFn    func(input *Input) (*crudtypes.ParsedGetEndpointInput, error)
	getInvokeFn     crudtypes.GetInvokeFn[Entity]
	toOutputFn      crudtypes.ToGetOutputFn[Entity, Output]
	entityFactoryFn repositorytypes.GetterFactoryFn[Entity]
	beforeCallback  crudtypes.BeforeGetCallback[Input, Entity]
}

// NewGetHandler creates a new get handler.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//   - getInvokeFn: The function that invokes the get endpoint.
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//   - entityFactoryFn: The function that creates a new entity.
//   - beforeCallback: The optional function that runs before the get operation.
//
// Returns:
//   - *GetHandler: The new get handler.
func NewGetHandler[Entity databasetypes.Getter, Input any, Output any](
	parseInputFn func(input *Input) (*crudtypes.ParsedGetEndpointInput, error),
	getInvokeFn crudtypes.GetInvokeFn[Entity],
	toOutputFn crudtypes.ToGetOutputFn[Entity, Output],
	entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
	beforeCallback crudtypes.BeforeGetCallback[Input, Entity],
) *GetHandler[Entity, Input, Output] {
	return &GetHandler[Entity, Input, Output]{
		parseInputFn:    parseInputFn,
		getInvokeFn:     getInvokeFn,
		toOutputFn:      toOutputFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
	}
}

// Handle processes the get endpoint.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - i: The input.
//
// Returns:
//   - any: The endpoint output.
//   - error: An error if the request fails.
func (h *GetHandler[Entity, Input, Output]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	parsedInput, err := h.parseInputFn(i)
	if err != nil {
		return nil, err
	}
	if h.beforeCallback != nil {
		parsedInput, err = h.beforeCallback(w, r, parsedInput, i)
		if err != nil {
			return nil, err
		}
	}
	entities, count, err := h.getInvokeFn(
		r.Context(), parsedInput, h.entityFactoryFn,
	)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(entities, count)
}

// WithParseInputFn returns a new get handler with the parse input function.
//
// Parameters:
//   - parseInputFn: The function that parses the input.
//
// Returns:
//   - *GetHandler: The new get handler.
func (h *GetHandler[Entity, Input, Output]) WithParseInputFn(
	parseInputFn func(input *Input) (*crudtypes.ParsedGetEndpointInput, error),
) *GetHandler[Entity, Input, Output] {
	new := *h
	new.parseInputFn = parseInputFn
	return &new
}

// WithGetInvokeFn returns a new get handler with the get invoke function.
//
// Parameters:
//   - getInvokeFn: The function that invokes the get endpoint.
//
// Returns:
//   - *GetHandler: The new get handler.
func (h *GetHandler[Entity, Input, Output]) WithGetInvokeFn(
	getInvokeFn crudtypes.GetInvokeFn[Entity],
) *GetHandler[Entity, Input, Output] {
	new := *h
	new.getInvokeFn = getInvokeFn
	return &new
}

// WithToOutputFn returns a new get handler with the to output function.
//
// Parameters:
//   - toOutputFn: The function that converts the entities to the endpoint
//     output.
//
// Returns:
//   - *GetHandler: The new get handler.
func (h *GetHandler[Entity, Input, Output]) WithToOutputFn(
	toOutputFn crudtypes.ToGetOutputFn[Entity, Output],
) *GetHandler[Entity, Input, Output] {
	new := *h
	new.toOutputFn = toOutputFn
	return &new
}

// WithEntityFactoryFn returns a new get handler with the entity factory
// function.
//
// Parameters:
//   - entityFactoryFn: The function that creates a new entity.
//
// Returns:
//   - *GetHandler: The new get handler.
func (h *GetHandler[Entity, Input, Output]) WithEntityFactoryFn(
	entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
) *GetHandler[Entity, Input, Output] {
	new := *h
	new.entityFactoryFn = entityFactoryFn
	return &new
}

// WithBeforeCallback returns a new get handler with the before callback.
//
// Parameters:
//   - beforeCallback: The function that runs before the get operation.
//
// Returns:
//   - *GetHandler: The new get handler.
func (h *GetHandler[Entity, Input, Output]) WithBeforeCallback(
	beforeCallback crudtypes.BeforeGetCallback[Input, Entity],
) *GetHandler[Entity, Input, Output] {
	new := *h
	new.beforeCallback = beforeCallback
	return &new
}
