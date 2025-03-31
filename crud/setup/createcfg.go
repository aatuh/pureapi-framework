package setup

import (
	"context"
	"errors"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/crud/services"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/defaults"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
	"github.com/pureapi/pureapi-framework/util/apimapper"
)

// CreateConfig holds the configuration for the create endpoint.
type CreateConfig[Entity databasetypes.Mutator] struct {
	// Default config for the create input handler.
	DefaultInputHandlerConfig *DefaultCreateInputHandlerConfig[Entity]
	// Override config for the create handler logic.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.CreateInputer[Entity]]

	// Default config for the create handler logic.
	DefaultHandlerLogicConfig *DefaultCreateHandlerLogicConfig[Entity]
	// Override for the create handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.CreateInputer[Entity]]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// Validate validates and sets defaults for the create config.
// It returns a new config with the defaults set.
func (cfg *CreateConfig[Entity]) Validate(
	systemID string,
	emitterLogger utiltypes.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn repositorytypes.ConnFn,
) (*CreateConfig[Entity], error) {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig =
				&DefaultCreateInputHandlerConfig[Entity]{}
		}
		subCfg, err := newCfg.DefaultInputHandlerConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultInputHandlerConfig = subCfg
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.CreateInputer[Entity]] {
			return apimapper.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *crudtypes.CreateInputer[Entity] {
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
				&DefaultCreateHandlerLogicConfig[Entity]{}
		}
		subCfg, err := newCfg.DefaultHandlerLogicConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultHandlerLogicConfig = subCfg
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.CreateInputer[Entity]] {
			return DefaultCreateHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.MutatorRepo,
				newCfg.DefaultHandlerLogicConfig.TxManager,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
				newCfg.DefaultHandlerLogicConfig.AfterCallback,  // Can be nil.
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(systemID, CreateErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(systemID, emitterLogger),
	)

	return &newCfg, nil
}

// DefaultCreateHandler sets up an endpoint handler for the create operation.
func DefaultCreateHandler[Entity databasetypes.Mutator](
	connFn repositorytypes.ConnFn,
	mutatorRepo repositorytypes.MutatorRepo[Entity],
	txManager repositorytypes.TxManager[Entity],
	outputFactoryFn func() crudtypes.CreateOutputer[Entity],
	beforeCallback crudtypes.BeforeCreateCallback[Entity],
	afterCreateFn services.AfterCreate[Entity],
) crudtypes.CreateHandler[Entity] {
	return services.NewCreateHandler(
		func(
			ctx context.Context, input *crudtypes.CreateInputer[Entity],
		) (Entity, error) {
			return (*input).GetEntity(), nil
		},
		func(ctx context.Context, entity Entity) (Entity, error) {
			return services.CreateInvoke(
				ctx, connFn, entity, mutatorRepo, txManager, afterCreateFn,
			)
		},
		func(entity Entity) (crudtypes.CreateOutputer[Entity], error) {
			output := outputFactoryFn()
			output.SetEntities([]Entity{entity})
			return output, nil
		},
		beforeCallback,
	)
}

// DefaultCreateInputHandlerConfig holds the default configuration for the
// create input handler.
type DefaultCreateInputHandlerConfig[Entity databasetypes.Mutator] struct {
	APIFields      apimapper.APIFields
	InputFactoryFn func() crudtypes.CreateInputer[Entity]
}

// Validate validates and sets defaults for the create input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultCreateInputHandlerConfig[Entity]) Validate() (*DefaultCreateInputHandlerConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		return nil, errors.New(
			"Create DefaultInputHandler InputFactoryFn is required",
		)
	}
	return &newCfg, nil
}

// DefaultCreateHandlerLogicConfig holds the default configuration for the
// create handler logic.
type DefaultCreateHandlerLogicConfig[Entity databasetypes.Mutator] struct {
	OutputFactoryFn func() crudtypes.CreateOutputer[Entity]
	BeforeCallback  crudtypes.BeforeCreateCallback[Entity]
	AfterCallback   services.AfterCreate[Entity]
	TxManager       repositorytypes.TxManager[Entity]
	MutatorRepo     repositorytypes.MutatorRepo[Entity]
}

// Validate validates and sets defaults for the create handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultCreateHandlerLogicConfig[Entity]) Validate() (*DefaultCreateHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		return nil, errors.New(
			"Create DefaultHandlerLogic OutputFactoryFn is required",
		)
	}
	if newCfg.TxManager == nil {
		newCfg.TxManager = defaults.TxManager[Entity]()
	}
	if newCfg.MutatorRepo == nil {
		newCfg.MutatorRepo = defaults.MutatorRepo[Entity]()
	}
	return &newCfg, nil
}
