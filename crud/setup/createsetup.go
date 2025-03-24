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
	"github.com/pureapi/pureapi-framework/defaults"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CreateConfig holds the configuration for the create endpoint.
type CreateConfig struct {
	// Default config for the create input handler.
	DefaultInputHandlerConfig *DefaultCreateInputHandlerConfig
	// Default config for the create handler logic.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.CreateInputer]

	// Default config for the create handler logic.
	DefaultHandlerLogicConfig *DefaultCreateHandlerLogicConfig
	// Override for the create handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.CreateInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the create config.
// It returns a new config with the defaults set.
func (cfg *CreateConfig) MustValidate(
	crudCfg *CRUDConfig,
) *CreateConfig {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig =
				&DefaultCreateInputHandlerConfig{}
		}
		if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
			panic("Create DefaultInputHandler InputFactoryFn is required")
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.CreateInputer] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
				func() *crudtypes.CreateInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	// If no override is provided, validate and use default handler logic.
	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig =
				&DefaultCreateHandlerLogicConfig{}
		}
		if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
			panic("Create DefaultHandlerLogic OutputFactoryFn is required")
		}
		if newCfg.DefaultHandlerLogicConfig.TxManager == nil {
			newCfg.DefaultHandlerLogicConfig.TxManager = defaults.DefaultTxManager[databasetypes.Mutator]()
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.CreateInputer] {
			return SetupCreateHandler(
				crudCfg.ConnFn,
				crudCfg.MutatorRepo,
				newCfg.DefaultHandlerLogicConfig.TxManager,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(crudCfg, errors.CreateErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(crudCfg),
	)

	return &newCfg
}

// CreateDefinition creates a definition for the create endpoint.
func CreateDefinition(cfg *CRUDConfig) *endpoint.DefaultDefinition {
	handler := NewCreateEndpointHandler(cfg.Create, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodPost, cfg.Stack, handler)
}

// CreateHandler creates a handler for the create endpoint.
func NewCreateEndpointHandler(
	cfg *CreateConfig,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.CreateInputer] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}

// SetupCreateHandler sets up an endpoint handler for the create operation.
func SetupCreateHandler(
	connFn repositorytypes.ConnFn,
	mutatorRepo repositorytypes.MutatorRepo,
	txManager repositorytypes.TxManager[databasetypes.Mutator],
	outputFactoryFn func() crudtypes.CreateOutputer,
	beforeCallback crudtypes.BeforeCreateCallback,
) crudtypes.CreateHandler {
	return services.NewCreateHandler(
		func(
			ctx context.Context,
			input *crudtypes.CreateInputer,
		) (databasetypes.Mutator, error) {
			return (*input).GetEntity(), nil
		},
		func(
			ctx context.Context,
			entity databasetypes.Mutator,
		) (databasetypes.Mutator, error) {
			return services.CreateInvoke(
				ctx,
				connFn,
				entity,
				mutatorRepo,
				txManager,
			)
		},
		func(entity databasetypes.Mutator) (crudtypes.CreateOutputer, error) {
			output := outputFactoryFn()
			output.SetEntities([]databasetypes.Mutator{entity})
			return output, nil
		},
		beforeCallback,
	)
}
