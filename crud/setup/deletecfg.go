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
	"github.com/pureapi/pureapi-framework/db/input"
	querytypes "github.com/pureapi/pureapi-framework/db/query/types"
	"github.com/pureapi/pureapi-framework/defaults"
	"github.com/pureapi/pureapi-framework/repository"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
	"github.com/pureapi/pureapi-framework/util/apimapper"
	apimappertypes "github.com/pureapi/pureapi-framework/util/apimapper/types"
)

type DefaultDeleteInput struct {
	Selectors input.Selectors `json:"selectors"`
}

func NewDeleteInput() *DefaultDeleteInput {
	return &DefaultDeleteInput{}
}

func (d *DefaultDeleteInput) GetSelectors() input.Selectors { return d.Selectors }

type DefaultDeleteOutput struct {
	Count int64 `json:"count"`
}

func NewDefaultDeleteOutput() *DefaultDeleteOutput {
	return &DefaultDeleteOutput{}
}

func (d *DefaultDeleteOutput) SetCount(count int64) { d.Count = count }

// DeleteConfig holds the configuration for the delete endpoint.
type DeleteConfig[Entity databasetypes.Mutator] struct {
	// Default config for the delete input handler.
	DefaultInputHandlerConfig *DefaultDeleteInputHandlerConfig
	// Override for the delete input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.DeleteInputer]

	// Default config for the delete handler logic.
	DefaultHandlerLogicConfig *DefaultDeleteHandlerLogicConfig[Entity]
	// Override for the delete handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.DeleteInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// Validate validates and sets defaults for the delete config.
// It returns a new config with the defaults set.
func (cfg *DeleteConfig[Entity]) Validate(
	systemID string,
	emitterLogger utiltypes.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn repositorytypes.ConnFn,
	apiToDBFields apimappertypes.APIToDBFields,
) (*DeleteConfig[Entity], error) {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig = &DefaultDeleteInputHandlerConfig{}
		}
		subCfg, err := newCfg.DefaultInputHandlerConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultInputHandlerConfig = subCfg
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.DeleteInputer] {
			return apimapper.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *crudtypes.DeleteInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	// If override is not set, validate and use default handler logic.
	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig = &DefaultDeleteHandlerLogicConfig[Entity]{}
		}
		subCfg, err := newCfg.DefaultHandlerLogicConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultHandlerLogicConfig = subCfg
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.DeleteInputer] {
			return DefaultDeleteHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				apiToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback,
				repository.NewMutatorRepo[Entity](
					defaults.QueryBuilder(), defaults.QueryErrorChecker(),
				),
				defaults.TxManager[*int64](),
				newCfg.DefaultHandlerLogicConfig.AfterCallback,
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(systemID, DeleteErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(systemID, emitterLogger),
	)

	return &newCfg, nil
}

// DefaultDeleteHandler sets up an endpoint handler for the delete operation.
func DefaultDeleteHandler[Entity databasetypes.Mutator](
	connFn repositorytypes.ConnFn,
	entityFn func(opts ...querytypes.EntityOption[Entity]) Entity,
	apiToDBFields apimappertypes.APIToDBFields,
	outputFactoryFn func() crudtypes.DeleteOutputer,
	beforeCallback crudtypes.BeforeDeleteCallback[Entity],
	mutatorRepo repositorytypes.MutatorRepo[Entity],
	txManager repositorytypes.TxManager[*int64],
	afterDeleteFn services.AfterDelete,
) crudtypes.DeleteHandler {
	return services.NewDeleteHandler(
		func(
			input *crudtypes.DeleteInputer,
		) (*crudtypes.ParsedDeleteEndpointInput, error) {
			i := *input
			return services.ParseDeleteInput(
				apiToDBFields,
				i.GetSelectors(),
				nil,
				0,
			)
		},
		func(
			ctx context.Context,
			parsedInput *crudtypes.ParsedDeleteEndpointInput,
			entity Entity,
		) (int64, error) {
			return services.DeleteInvoke(
				ctx,
				parsedInput,
				connFn,
				entity,
				mutatorRepo,
				txManager,
				afterDeleteFn,
			)
		},
		func(count int64) (crudtypes.DeleteOutputer, error) {
			output := outputFactoryFn()
			output.SetCount(count)
			return output, nil
		},
		func() Entity { return entityFn() },
		beforeCallback,
	)
}

// DefaultDeleteInputHandlerConfig holds the default configuration for the
// delete input handler.
type DefaultDeleteInputHandlerConfig struct {
	APIFields      apimapper.APIFields
	InputFactoryFn func() crudtypes.DeleteInputer
}

// Validate validates and sets defaults for the delete input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultDeleteInputHandlerConfig) Validate() (*DefaultDeleteInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() crudtypes.DeleteInputer {
			return NewDeleteInput()
		}
	}
	return &newCfg, nil
}

// DefaultDeleteHandlerLogicConfig holds the default configuration for the
// delete handler logic.
type DefaultDeleteHandlerLogicConfig[Entity databasetypes.Mutator] struct {
	OutputFactoryFn func() crudtypes.DeleteOutputer
	BeforeCallback  crudtypes.BeforeDeleteCallback[Entity]
	AfterCallback   services.AfterDelete
	EntityFn        func(...querytypes.EntityOption[Entity]) Entity
}

// Validate validates and sets defaults for the delete handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultDeleteHandlerLogicConfig[Entity]) Validate() (*DefaultDeleteHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() crudtypes.DeleteOutputer {
			return NewDefaultDeleteOutput()
		}
	}
	if newCfg.EntityFn == nil {
		return nil, errors.New(
			"Validate: EntityFn is required in DefaultHandlerLogicConfig",
		)
	}
	return &newCfg, nil
}
