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

// DeleteHandler is the handler interface for the delete endpoint.
type DeleteHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i services.DeleteInputer,
	) (any, error)
}

type DefaultDeleteInput struct {
	Selectors apidb.APISelectors `json:"selectors"`
}

func NewDeleteInput() *DefaultDeleteInput {
	return &DefaultDeleteInput{}
}

func (d *DefaultDeleteInput) GetSelectors() apidb.APISelectors { return d.Selectors }

type DefaultDeleteOutput struct {
	Count int64 `json:"count"`
}

func NewDefaultDeleteOutput() *DefaultDeleteOutput {
	return &DefaultDeleteOutput{}
}

func (d *DefaultDeleteOutput) SetCount(count int64) { d.Count = count }

// DeleteConfig holds the configuration for the delete endpoint.
type DeleteConfig[Entity database.Mutator] struct {
	// Default config for the delete input handler.
	DefaultInputHandlerConfig *DefaultDeleteInputHandlerConfig
	// Override for the delete input handler.
	InputHandlerFactoryFn func() endpoint.InputHandler[services.DeleteInputer]

	// Default config for the delete handler logic.
	DefaultHandlerLogicConfig *DefaultDeleteHandlerLogicConfig[Entity]
	// Override for the delete handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[services.DeleteInputer]

	ErrorHandlerFactoryFn  func() endpoint.ErrorHandler
	OutputHandlerFactoryFn func() endpoint.OutputHandler
}

// Validate validates and sets defaults for the delete config.
// It returns a new config with the defaults set.
func (cfg *DeleteConfig[Entity]) Validate(
	systemID string,
	emitterLogger event.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn db.ConnFn,
	apiToDBFields inpututil.APIToDBFields,
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
		func() endpoint.InputHandler[services.DeleteInputer] {
			x := input.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *services.DeleteInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			)
			return x.MustValidateAPIFields()
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
		func() endpoint.HandlerLogicFn[services.DeleteInputer] {
			return DefaultDeleteHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				apiToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback,
				db.NewMutatorRepo[Entity](
					defaults.Query(), defaults.QueryErrorChecker(),
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
func DefaultDeleteHandler[Entity database.Mutator](
	connFn db.ConnFn,
	entityFn func(opts ...db.EntityOption[Entity]) Entity,
	apiToDBFields inpututil.APIToDBFields,
	outputFactoryFn func() services.DeleteOutputer,
	beforeCallback services.BeforeDeleteCallback[Entity],
	mutatorRepo db.MutatorRepository[Entity],
	txManager db.TxManager[*int64],
	afterDeleteFn services.AfterDelete,
) *services.DeleteHandler[Entity] {
	return services.NewDeleteHandler(
		func(
			input services.DeleteInputer,
		) (*services.ParsedDeleteEndpointInput, error) {
			return services.ParseDeleteInput(
				apiToDBFields,
				input.GetSelectors(),
				nil,
				0,
			)
		},
		func(
			ctx context.Context,
			parsedInput *services.ParsedDeleteEndpointInput,
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
		func(count int64) (services.DeleteOutputer, error) {
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
	APIFields      input.APIFields
	InputFactoryFn func() services.DeleteInputer
}

// Validate validates and sets defaults for the delete input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultDeleteInputHandlerConfig) Validate() (*DefaultDeleteInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() services.DeleteInputer {
			return NewDeleteInput()
		}
	}
	return &newCfg, nil
}

// DefaultDeleteHandlerLogicConfig holds the default configuration for the
// delete handler logic.
type DefaultDeleteHandlerLogicConfig[Entity database.Mutator] struct {
	OutputFactoryFn func() services.DeleteOutputer
	BeforeCallback  services.BeforeDeleteCallback[Entity]
	AfterCallback   services.AfterDelete
	EntityFn        func(...db.EntityOption[Entity]) Entity
}

// Validate validates and sets defaults for the delete handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultDeleteHandlerLogicConfig[Entity]) Validate() (*DefaultDeleteHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() services.DeleteOutputer {
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
