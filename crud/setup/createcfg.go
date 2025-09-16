package setup

import (
	"context"
	"errors"
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-framework/api/input"
	"github.com/aatuh/pureapi-framework/crud/services"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/aatuh/pureapi-framework/defaults"
)

// CreateHandler is the handler interface for the create endpoint.
type CreateHandler[Entity database.Mutator] interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *services.CreateInputer[Entity],
	) (any, error)
}

// CreateConfig holds the configuration for the create endpoint.
type CreateConfig[Entity database.Mutator] struct {
	// Default config for the create input handler.
	DefaultInputHandlerConfig *DefaultCreateInputHandlerConfig[Entity]
	// Override config for the create handler logic.
	InputHandlerFactoryFn func() endpoint.InputHandler[services.CreateInputer[Entity]]

	// Default config for the create handler logic.
	DefaultHandlerLogicConfig *DefaultCreateHandlerLogicConfig[Entity]
	// Override for the create handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[services.CreateInputer[Entity]]

	ErrorHandlerFactoryFn  func() endpoint.ErrorHandler
	OutputHandlerFactoryFn func() endpoint.OutputHandler
}

// Validate validates and sets defaults for the create config.
// It returns a new config with the defaults set.
func (cfg *CreateConfig[Entity]) Validate(
	systemID string,
	emitterLogger event.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn db.ConnFn,
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
		func() endpoint.InputHandler[services.CreateInputer[Entity]] {
			return input.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *services.CreateInputer[Entity] {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			).MustValidateAPIFields()
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
		func() endpoint.HandlerLogicFn[services.CreateInputer[Entity]] {
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
func DefaultCreateHandler[Entity database.Mutator](
	connFn db.ConnFn,
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[Entity],
	outputFactoryFn func() services.CreateOutputer[Entity],
	beforeCallback services.BeforeCreateCallback[Entity],
	afterCreateFn services.AfterCreate[Entity],
) *services.CreateHandler[Entity] {
	return services.NewCreateHandler(
		func(
			ctx context.Context, input *services.CreateInputer[Entity],
		) (Entity, error) {
			return (*input).GetEntity(), nil
		},
		func(ctx context.Context, entity Entity) (Entity, error) {
			return services.CreateInvoke(
				ctx, connFn, entity, mutatorRepo, txManager, afterCreateFn,
			)
		},
		func(entity Entity) (services.CreateOutputer[Entity], error) {
			output := outputFactoryFn()
			output.SetEntities([]Entity{entity})
			return output, nil
		},
		beforeCallback,
	)
}

// DefaultCreateInputHandlerConfig holds the default configuration for the
// create input handler.
type DefaultCreateInputHandlerConfig[Entity database.Mutator] struct {
	APIFields      input.APIFields
	InputFactoryFn func() services.CreateInputer[Entity]
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
type DefaultCreateHandlerLogicConfig[Entity database.Mutator] struct {
	OutputFactoryFn func() services.CreateOutputer[Entity]
	BeforeCallback  services.BeforeCreateCallback[Entity]
	AfterCallback   services.AfterCreate[Entity]
	TxManager       db.TxManager[Entity]
	MutatorRepo     db.MutatorRepository[Entity]
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
