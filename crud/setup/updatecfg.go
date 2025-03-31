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

type DefaultUpdateInput struct {
	Selectors input.Selectors `json:"selectors"`
	Updates   input.Updates   `json:"updates"`
	Upsert    bool            `json:"upsert"`
}

func NewDefaultUpdateInput() *DefaultUpdateInput {
	return &DefaultUpdateInput{}
}

func (u *DefaultUpdateInput) GetSelectors() input.Selectors { return u.Selectors }
func (u *DefaultUpdateInput) GetUpdates() input.Updates     { return u.Updates }
func (u *DefaultUpdateInput) GetUpsert() bool               { return u.Upsert }

type DefaultUpdateOutput struct {
	Count int64 `json:"count"`
}

func NewDefaultUpdateOutput() *DefaultUpdateOutput {
	return &DefaultUpdateOutput{}
}

func (u *DefaultUpdateOutput) SetCount(count int64) { u.Count = count }

// UpdateConfig holds the configuration for the update endpoint.
type UpdateConfig[Entity databasetypes.Mutator] struct {
	// Default config for the update input handler.
	DefaultInputHandlerConfig *DefaultUpdateInputHandlerConfig
	// Override for the update input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.UpdateInputer]

	// Default config for the update handler logic.
	DefaultHandlerLogicConfig *DefaultUpdateHandlerLogicConfig[Entity]
	// Override for the update handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.UpdateInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// Validate validates and sets defaults for the update config.
// It returns a new config with the defaults set.
func (cfg *UpdateConfig[Entity]) Validate(
	systemID string,
	emitterLogger utiltypes.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn repositorytypes.ConnFn,
	apiToDBFields apimappertypes.APIToDBFields,
) (*UpdateConfig[Entity], error) {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig = &DefaultUpdateInputHandlerConfig{}
		}
		subCfg, err := newCfg.DefaultInputHandlerConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultInputHandlerConfig = subCfg
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.UpdateInputer] {
			return apimapper.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *crudtypes.UpdateInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
		},
	)

	// If override is not set, validate and use default handler logic.
	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig = &DefaultUpdateHandlerLogicConfig[Entity]{}
		}
		subCfg, err := newCfg.DefaultHandlerLogicConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultHandlerLogicConfig = subCfg
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.UpdateInputer] {
			return DefaultUpdateHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				apiToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
				repository.NewMutatorRepo[Entity](
					defaults.QueryBuilder(), defaults.QueryErrorChecker(),
				),
				defaults.TxManager[*int64](),
				newCfg.DefaultHandlerLogicConfig.AfterUpdateFn,
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(systemID, UpdateErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(systemID, emitterLogger),
	)

	return &newCfg, nil
}

// DefaultUpdateHandler sets up an endpoint handler for the update operation.
func DefaultUpdateHandler[Entity databasetypes.Mutator](
	connFn repositorytypes.ConnFn,
	entityFn func(opts ...querytypes.EntityOption[Entity]) Entity,
	apiToDBFields apimappertypes.APIToDBFields,
	outputFactoryFn func() crudtypes.UpdateOutputer,
	beforeCallback crudtypes.BeforeUpdateCallback[Entity],
	mutatorRepo repositorytypes.MutatorRepo[Entity],
	txManager repositorytypes.TxManager[*int64],
	afterUpdateFn services.AfterUpdate[Entity],
) crudtypes.UpdateHandler {
	return services.NewUpdateHandler(
		func(
			input *crudtypes.UpdateInputer,
		) (*crudtypes.ParsedUpdateEndpointInput, error) {
			i := *input
			return services.ParseUpdateInput(
				apiToDBFields,
				i.GetSelectors(),
				i.GetUpdates(),
				i.GetUpsert(),
			)
		},
		func(
			ctx context.Context,
			parsedInput *crudtypes.ParsedUpdateEndpointInput,
			updater Entity,
		) (int64, error) {
			return services.UpdateInvoke(
				ctx,
				parsedInput,
				connFn,
				updater,
				mutatorRepo,
				txManager,
				afterUpdateFn,
			)
		},
		func(count int64) (crudtypes.UpdateOutputer, error) {
			output := outputFactoryFn()
			output.SetCount(count)
			return output, nil
		},
		func() Entity { return entityFn() },
		beforeCallback,
	)
}

// DefaultUpdateInputHandlerConfig holds the default configuration for the
// update input handler.
type DefaultUpdateInputHandlerConfig struct {
	APIFields      apimapper.APIFields
	InputFactoryFn func() crudtypes.UpdateInputer
}

// Validate validates and sets defaults for the update input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultUpdateInputHandlerConfig) Validate() (*DefaultUpdateInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() crudtypes.UpdateInputer {
			return NewDefaultUpdateInput()
		}
	}
	return &newCfg, nil
}

// DefaultUpdateHandlerLogicConfig holds the default configuration for the
// update handler logic.
type DefaultUpdateHandlerLogicConfig[Entity databasetypes.Mutator] struct {
	OutputFactoryFn func() crudtypes.UpdateOutputer
	BeforeCallback  crudtypes.BeforeUpdateCallback[Entity]
	AfterUpdateFn   services.AfterUpdate[Entity]
	EntityFn        func(...querytypes.EntityOption[Entity]) Entity
}

// Validate validates and sets defaults for the update handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultUpdateHandlerLogicConfig[Entity]) Validate() (*DefaultUpdateHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() crudtypes.UpdateOutputer {
			return NewDefaultUpdateOutput()
		}
	}
	if newCfg.EntityFn == nil {
		return nil, errors.New(
			"Validate: EntityFn is required in UpdateConfig",
		)
	}
	return &newCfg, nil
}
