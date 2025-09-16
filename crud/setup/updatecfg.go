package setup

import (
	"context"
	"errors"
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	apidb "github.com/aatuh/pureapi-framework/api/db"
	"github.com/aatuh/pureapi-framework/api/input"
	"github.com/aatuh/pureapi-framework/crud/services"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/aatuh/pureapi-framework/defaults"
	"github.com/aatuh/pureapi-framework/util/inpututil"
)

// UpdateHandler is the handler interface for the update endpoint.
type UpdateHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *services.UpdateInputer,
	) (any, error)
}

type DefaultUpdateInput struct {
	Selectors apidb.APISelectors `json:"selectors"`
	Updates   apidb.APIUpdates   `json:"updates"`
	Upsert    bool               `json:"upsert"`
}

func NewDefaultUpdateInput() *DefaultUpdateInput {
	return &DefaultUpdateInput{}
}

func (u *DefaultUpdateInput) GetSelectors() apidb.APISelectors { return u.Selectors }
func (u *DefaultUpdateInput) GetUpdates() apidb.APIUpdates     { return u.Updates }
func (u *DefaultUpdateInput) GetUpsert() bool                  { return u.Upsert }

type DefaultUpdateOutput struct {
	Count int64 `json:"count"`
}

func NewDefaultUpdateOutput() *DefaultUpdateOutput {
	return &DefaultUpdateOutput{}
}

func (u *DefaultUpdateOutput) SetCount(count int64) { u.Count = count }

// UpdateConfig holds the configuration for the update endpoint.
type UpdateConfig[Entity database.Mutator] struct {
	// Default config for the update input handler.
	DefaultInputHandlerConfig *DefaultUpdateInputHandlerConfig
	// Override for the update input handler.
	InputHandlerFactoryFn func() endpoint.InputHandler[services.UpdateInputer]

	// Default config for the update handler logic.
	DefaultHandlerLogicConfig *DefaultUpdateHandlerLogicConfig[Entity]
	// Override for the update handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[services.UpdateInputer]

	ErrorHandlerFactoryFn  func() endpoint.ErrorHandler
	OutputHandlerFactoryFn func() endpoint.OutputHandler
}

// Validate validates and sets defaults for the update config.
// It returns a new config with the defaults set.
func (cfg *UpdateConfig[Entity]) Validate(
	systemID string,
	emitterLogger event.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn db.ConnFn,
	apiToDBFields inpututil.APIToDBFields,
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
		func() endpoint.InputHandler[services.UpdateInputer] {
			return input.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *services.UpdateInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			).MustValidateAPIFields()
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
		func() endpoint.HandlerLogicFn[services.UpdateInputer] {
			return DefaultUpdateHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				apiToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
				db.NewMutatorRepo[Entity](
					defaults.Query(), defaults.QueryErrorChecker(),
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
func DefaultUpdateHandler[Entity database.Mutator](
	connFn db.ConnFn,
	entityFn func(opts ...db.EntityOption[Entity]) Entity,
	apiToDBFields inpututil.APIToDBFields,
	outputFactoryFn func() services.UpdateOutputer,
	beforeCallback services.BeforeUpdateCallback[Entity],
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[*int64],
	afterUpdateFn services.AfterUpdate[Entity],
) *services.UpdateHandler[Entity] {
	return services.NewUpdateHandler(
		func(
			input *services.UpdateInputer,
		) (*services.ParsedUpdateEndpointInput, error) {
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
			parsedInput *services.ParsedUpdateEndpointInput,
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
		func(count int64) (services.UpdateOutputer, error) {
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
	APIFields      input.APIFields
	InputFactoryFn func() services.UpdateInputer
}

// Validate validates and sets defaults for the update input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultUpdateInputHandlerConfig) Validate() (*DefaultUpdateInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() services.UpdateInputer {
			return NewDefaultUpdateInput()
		}
	}
	return &newCfg, nil
}

// DefaultUpdateHandlerLogicConfig holds the default configuration for the
// update handler logic.
type DefaultUpdateHandlerLogicConfig[Entity database.Mutator] struct {
	OutputFactoryFn func() services.UpdateOutputer
	BeforeCallback  services.BeforeUpdateCallback[Entity]
	AfterUpdateFn   services.AfterUpdate[Entity]
	EntityFn        func(...db.EntityOption[Entity]) Entity
}

// Validate validates and sets defaults for the update handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultUpdateHandlerLogicConfig[Entity]) Validate() (*DefaultUpdateHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() services.UpdateOutputer {
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
