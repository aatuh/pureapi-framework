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
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CreateConfig holds the configuration for the create endpoint.
type CreateConfig[Entity databasetypes.CRUDEntity] struct {
	// Default config for the create input handler.
	DefaultInputHandlerConfig *DefaultCreateInputHandlerConfig[Entity]
	// Default config for the create handler logic.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.CreateInputer[Entity]]

	// Default config for the create handler logic.
	DefaultHandlerLogicConfig *DefaultCreateHandlerLogicConfig[Entity]
	// Override for the create handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.CreateInputer[Entity]]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the create config.
// It returns a new config with the defaults set.
func (cfg *CreateConfig[Entity]) MustValidate(
	crudCfg *CRUDConfig[Entity],
) *CreateConfig[Entity] {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig =
				&DefaultCreateInputHandlerConfig[Entity]{}
		}
		if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
			panic("Create DefaultInputHandler InputFactoryFn is required")
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.CreateInputer[Entity]] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
				func() *crudtypes.CreateInputer[Entity] {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig =
				&DefaultCreateHandlerLogicConfig[Entity]{}
		}
		if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
			panic("Create DefaultHandlerLogic OutputFactoryFn is required")
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.CreateInputer[Entity]] {
			return SetupCreateHandler(
				crudCfg.ConnFn,
				crudCfg.MutatorRepo,
				crudCfg.TxManager,
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
func CreateDefinition[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *endpoint.DefaultDefinition {
	handler := NewCreateEndpointHandler(cfg.Create, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodPost, cfg.Stack, handler)
}

// CreateHandler creates a handler for the create endpoint.
func NewCreateEndpointHandler[Entity databasetypes.CRUDEntity](
	cfg *CreateConfig[Entity],
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.CreateInputer[Entity]] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}

// SetupCreateHandler sets up an endpoint handler for the create operation.
func SetupCreateHandler[
	Input crudtypes.CreateInputer[Entity],
	Output crudtypes.CreateOutputer[Entity],
	Entity databasetypes.CRUDEntity,
](
	connFn repositorytypes.ConnFn,
	mutatorRepo repositorytypes.MutatorRepo[Entity],
	txManager repositorytypes.TxManager[Entity],
	outputFactoryFn func() Output,
	beforeCallback crudtypes.BeforeCreateCallback[Input, Entity],
) crudtypes.CreateHandler[Entity, Input] {
	return services.NewCreateHandler(
		func(ctx context.Context, input *Input) (Entity, error) {
			return (*input).GetEntity(), nil
		},
		func(ctx context.Context, entity Entity) (Entity, error) {
			return services.CreateInvoke(
				ctx,
				connFn,
				entity,
				mutatorRepo,
				txManager,
			)
		},
		func(entity Entity) (any, error) {
			output := outputFactoryFn()
			output.SetEntities([]Entity{entity})
			return output, nil
		},
		beforeCallback,
	)
}
