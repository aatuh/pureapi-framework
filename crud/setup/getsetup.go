package setup

import (
	"context"
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/api"
	"github.com/pureapi/pureapi-framework/crud/errors"
	"github.com/pureapi/pureapi-framework/crud/services"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/dbinput"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// GetConfig holds the configuration for the get endpoint.
type GetConfig[Entity databasetypes.CRUDEntity] struct {
	// Default config for the get input handler.
	DefaultInputHandlerConfig *DefaultGetInputHandlerConfig[Entity]
	// Override for the get input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.GetInputer[Entity]]

	// Default config for the get handler logic.
	DefaultHandlerLogicConfig *DefaultGetHandlerLogicConfig[Entity]
	// Override for the get handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.GetInputer[Entity]]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the get config.
// It returns a new config with the defaults set.
func (cfg *GetConfig[Entity]) MustValidate(
	crudCfg *CRUDConfig[Entity],
) *GetConfig[Entity] {
	newCfg := *cfg

	if newCfg.DefaultInputHandlerConfig == nil {
		newCfg.DefaultInputHandlerConfig =
			&DefaultGetInputHandlerConfig[Entity]{}
	}
	if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
		newCfg.DefaultInputHandlerConfig.InputFactoryFn = func() crudtypes.GetInputer[Entity] {
			return NewGetInput()
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.GetInputer[Entity]] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
				func() *crudtypes.GetInputer[Entity] {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	if newCfg.DefaultHandlerLogicConfig == nil {
		newCfg.DefaultHandlerLogicConfig =
			&DefaultGetHandlerLogicConfig[Entity]{}
	}
	if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
		newCfg.DefaultHandlerLogicConfig.OutputFactoryFn = func() crudtypes.GetOutputer[Entity] {
			return NewGetOutput[Entity]()
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.GetInputer[Entity]] {
			return SetupGetHandler(
				crudCfg.ConnFn,
				crudCfg.ReaderRepo,
				crudCfg.TxManager,
				crudCfg.EntityFn,
				crudCfg.APIToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(crudCfg, errors.GetErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(crudCfg),
	)

	return &newCfg
}

// GetDefinition creates a definition for the get endpoint.
func GetDefinition[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *endpoint.DefaultDefinition {
	handler := NewGetEndpointHandler(cfg.Get, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodGet, cfg.Stack, handler)
}

// GetHandler creates a handler for the get endpoint.
func NewGetEndpointHandler[Entity databasetypes.CRUDEntity](
	cfg *GetConfig[Entity],
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.GetInputer[Entity]] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}

type GetInput struct {
	Selectors dbinput.Selectors `json:"selectors"`
	Orders    dbinput.Orders    `json:"orders"`
	Page      *dbinput.Page     `json:"page"`
	Count     bool              `json:"count"`
}

func NewGetInput() *GetInput {
	return &GetInput{}
}

func (i *GetInput) GetSelectors() dbinput.Selectors { return i.Selectors }
func (i *GetInput) GetOrders() dbinput.Orders       { return i.Orders }
func (i *GetInput) GetPage() *dbinput.Page          { return i.Page }
func (i *GetInput) GetCount() bool                  { return i.Count }

type GetOutput[Entity any] struct {
	Entities []Entity `json:"entities"`
	Count    int      `json:"count"`
}

func NewGetOutput[Entity any]() *GetOutput[Entity] {
	return &GetOutput[Entity]{}
}

func (o *GetOutput[Entity]) SetEntities(entities []Entity) { o.Entities = entities }
func (o *GetOutput[Entity]) SetCount(count int)            { o.Count = count }

// SetupGetHandler sets up an endpoint handler for the get operation.
func SetupGetHandler[
	Input crudtypes.GetInputer[Entity],
	Output crudtypes.GetOutputer[Entity],
	Entity databasetypes.CRUDEntity,
](
	connFn repositorytypes.ConnFn,
	readerRepo repositorytypes.ReaderRepo[Entity],
	txManager repositorytypes.TxManager[Entity],
	entityFn func(opts ...crudtypes.EntityOption[Entity]) Entity,
	apiToDBFields crudtypes.APIToDBFields,
	outputFactoryFn func() Output,
	beforeCallback crudtypes.BeforeGetCallback[Input, Entity],
) crudtypes.GetHandler[Entity, Input, Output] {
	return services.NewGetHandler(
		func(input *Input) (*crudtypes.ParsedGetEndpointInput, error) {
			i := *input
			return services.ParseGetInput(
				apiToDBFields,
				i.GetSelectors(),
				i.GetOrders(),
				i.GetPage(),
				100,
				i.GetCount(),
			)
		},
		func(
			ctx context.Context,
			parsedInput *crudtypes.ParsedGetEndpointInput,
			entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
		) ([]Entity, int, error) {
			return services.GetInvoke(
				ctx,
				parsedInput,
				connFn,
				entityFactoryFn,
				readerRepo,
				txManager,
			)
		},
		func(entities []Entity, count int) (crudtypes.GetOutputer[Entity], error) {
			output := outputFactoryFn()
			output.SetEntities(entities)
			output.SetCount(count)
			return output, nil
		},
		func() Entity { return entityFn() },
		beforeCallback,
	)
}
