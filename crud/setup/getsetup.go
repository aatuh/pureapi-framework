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
	"github.com/pureapi/pureapi-framework/defaults"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// GetConfig holds the configuration for the get endpoint.
type GetConfig struct {
	// Default config for the get input handler.
	DefaultInputHandlerConfig *DefaultGetInputHandlerConfig
	// Override for the get input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.GetInputer]

	// Default config for the get handler logic.
	DefaultHandlerLogicConfig *DefaultGetHandlerLogicConfig
	// Override for the get handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.GetInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the get config.
// It returns a new config with the defaults set.
func (cfg *GetConfig) MustValidate(
	crudCfg *CRUDConfig,
) *GetConfig {
	newCfg := *cfg

	if newCfg.DefaultInputHandlerConfig == nil {
		newCfg.DefaultInputHandlerConfig =
			&DefaultGetInputHandlerConfig{}
	}
	if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
		newCfg.DefaultInputHandlerConfig.InputFactoryFn = func() crudtypes.GetInputer {
			return NewGetInput()
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.GetInputer] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
				func() *crudtypes.GetInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	// If override is not set, validate and use default handler logic.
	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig =
				&DefaultGetHandlerLogicConfig{}
		}
		if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
			newCfg.DefaultHandlerLogicConfig.OutputFactoryFn = func() crudtypes.GetOutputer {
				return NewGetOutput()
			}
		}
		if newCfg.DefaultHandlerLogicConfig.EntityFn == nil {
			panic("MustValidate: EntityFn is required in GetConfig")
		}
		if newCfg.DefaultHandlerLogicConfig.TxManager == nil {
			newCfg.DefaultHandlerLogicConfig.TxManager = defaults.DefaultTxManager[databasetypes.Getter]()
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.GetInputer] {
			return SetupGetHandler(
				crudCfg.ConnFn,
				crudCfg.ReaderRepo,
				newCfg.DefaultHandlerLogicConfig.TxManager,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
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
func GetDefinition(cfg *CRUDConfig) *endpoint.DefaultDefinition {
	handler := NewGetEndpointHandler(cfg.Get, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodGet, cfg.Stack, handler)
}

// GetHandler creates a handler for the get endpoint.
func NewGetEndpointHandler(
	cfg *GetConfig,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.GetInputer] {
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

type GetOutput struct {
	Entities []databasetypes.Getter `json:"entities"`
	Count    int                    `json:"count"`
}

func NewGetOutput() *GetOutput {
	return &GetOutput{}
}

func (o *GetOutput) SetEntities(entities []databasetypes.Getter) { o.Entities = entities }
func (o *GetOutput) SetCount(count int)                          { o.Count = count }

// SetupGetHandler sets up an endpoint handler for the get operation.
func SetupGetHandler(
	connFn repositorytypes.ConnFn,
	readerRepo repositorytypes.ReaderRepo,
	txManager repositorytypes.TxManager[databasetypes.Getter],
	entityFn func(opts ...crudtypes.EntityOption) databasetypes.Getter,
	apiToDBFields crudtypes.APIToDBFields,
	outputFactoryFn func() crudtypes.GetOutputer,
	beforeCallback crudtypes.BeforeGetCallback,
) crudtypes.GetHandler {
	return services.NewGetHandler(
		func(
			input *crudtypes.GetInputer,
		) (*crudtypes.ParsedGetEndpointInput, error) {
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
			entityFactoryFn repositorytypes.GetterFactoryFn,
		) ([]databasetypes.Getter, int, error) {
			return services.GetInvoke(
				ctx,
				parsedInput,
				connFn,
				entityFactoryFn,
				readerRepo,
				txManager,
			)
		},
		func(
			entities []databasetypes.Getter,
			count int,
		) (crudtypes.GetOutputer, error) {
			output := outputFactoryFn()
			output.SetEntities(entities)
			output.SetCount(count)
			return output, nil
		},
		func() databasetypes.Getter { return entityFn() },
		beforeCallback,
	)
}
