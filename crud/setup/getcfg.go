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

// GetHandler is the handler interface for the get endpoint.
type GetHandler interface {
	Handle(
		w http.ResponseWriter, r *http.Request, i *services.GetInputer,
	) (any, error)
}

// DefaultGetInput is the default input for the get endpoint.
type DefaultGetInput struct {
	Selectors apidb.APISelectors `json:"selectors"`
	Orders    apidb.Orders       `json:"orders"`
	Page      *apidb.Page        `json:"page"`
	Count     bool               `json:"count"`
}

// NewGetInput returns a new DefaultGetInput.
func NewGetInput() *DefaultGetInput {
	return &DefaultGetInput{}
}

// GetSelectors returns the selectors for the get db.
func (i *DefaultGetInput) GetSelectors() apidb.APISelectors { return i.Selectors }

// GetOrders returns the orders for the get db.
func (i *DefaultGetInput) GetOrders() apidb.Orders { return i.Orders }

// GetPage returns the page for the get db.
func (i *DefaultGetInput) GetPage() *apidb.Page { return i.Page }

// GetCount returns the count for the get db.
func (i *DefaultGetInput) GetCount() bool { return i.Count }

type DefaultGetOutput[Entity database.Getter] struct {
	Entities []Entity `json:"entities"`
	Count    int      `json:"count"`
}

func NewDefaultGetOutput[Entity database.Getter]() *DefaultGetOutput[Entity] {
	return &DefaultGetOutput[Entity]{}
}

func (o *DefaultGetOutput[Entity]) SetEntities(entities []Entity) { o.Entities = entities }
func (o *DefaultGetOutput[Entity]) SetCount(count int)            { o.Count = count }

// GetConfig holds the configuration for the get endpoint.
type GetConfig[Entity database.Getter] struct {
	// Default config for the get input handler.
	DefaultInputHandlerConfig *DefaultGetInputHandlerConfig
	// Override for the get input handler.
	InputHandlerFactoryFn func() endpoint.InputHandler[services.GetInputer]

	// Default config for the get handler logic.
	DefaultHandlerLogicConfig *DefaultGetHandlerLogicConfig[Entity]
	// Override for the get handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[services.GetInputer]

	ErrorHandlerFactoryFn  func() endpoint.ErrorHandler
	OutputHandlerFactoryFn func() endpoint.OutputHandler
}

// Validate validates and sets defaults for the get config.
// It returns a new config with the defaults set.
func (cfg *GetConfig[Entity]) Validate(
	systemID string,
	emitterLogger event.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn db.ConnFn,
	apiToDBFields inpututil.APIToDBFields,
) (*GetConfig[Entity], error) {
	newCfg := *cfg

	if newCfg.InputHandlerFactoryFn == nil {
		if newCfg.DefaultInputHandlerConfig == nil {
			newCfg.DefaultInputHandlerConfig =
				&DefaultGetInputHandlerConfig{}
		}
		subCfg, err := newCfg.DefaultInputHandlerConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultInputHandlerConfig = subCfg
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpoint.InputHandler[services.GetInputer] {
			return input.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
				func() *services.GetInputer {
					inp := newCfg.DefaultInputHandlerConfig.InputFactoryFn()
					return &inp
				},
			).MustValidateAPIFields()
		},
	)

	// If override is not set, validate and use default handler logic.
	if newCfg.HandlerLogicFnFactoryFn == nil {
		if newCfg.DefaultHandlerLogicConfig == nil {
			newCfg.DefaultHandlerLogicConfig =
				&DefaultGetHandlerLogicConfig[Entity]{}
		}
		subCfg, err := newCfg.DefaultHandlerLogicConfig.Validate()
		if err != nil {
			return nil, err
		}
		newCfg.DefaultHandlerLogicConfig = subCfg
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[services.GetInputer] {
			return DefaultGetHandler(
				connFn,
				newCfg.DefaultHandlerLogicConfig.ReaderRepo,
				newCfg.DefaultHandlerLogicConfig.TxManager,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				apiToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
				newCfg.DefaultHandlerLogicConfig.AfterGetFn,
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(systemID, GetErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(systemID, emitterLogger),
	)

	return &newCfg, nil
}

// DefaultGetHandler sets up an endpoint handler for the get operation.
func DefaultGetHandler[Entity database.Getter](
	connFn db.ConnFn,
	readerRepo db.ReaderRepository[Entity],
	txManager db.TxManager[Entity],
	entityFn func(opts ...db.EntityOption[Entity]) Entity,
	apiToDBFields inpututil.APIToDBFields,
	outputFactoryFn func() services.GetOutputer[Entity],
	beforeCallback services.BeforeGetCallback,
	afterGetFn services.AfterGet[Entity],
) *services.GetHandler[Entity] {
	return services.NewGetHandler(
		func(
			input *services.GetInputer,
		) (*services.ParsedGetEndpointInput, error) {
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
			parsedInput *services.ParsedGetEndpointInput,
			entityFactoryFn db.GetterFactoryFn[Entity],
		) ([]Entity, int, error) {
			return services.GetInvoke(
				ctx,
				parsedInput,
				connFn,
				entityFactoryFn,
				readerRepo,
				txManager,
				afterGetFn,
			)
		},
		func(
			entities []Entity, count int,
		) (services.GetOutputer[Entity], error) {
			output := outputFactoryFn()
			output.SetEntities(entities)
			output.SetCount(count)
			return output, nil
		},
		func() Entity { return entityFn() },
		beforeCallback,
	)
}

// DefaultGetInputHandlerConfig holds the default configuration for the get
// input handler.
type DefaultGetInputHandlerConfig struct {
	APIFields      input.APIFields
	InputFactoryFn func() services.GetInputer
}

// Validate validates and sets defaults for the get input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultGetInputHandlerConfig) Validate() (*DefaultGetInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() services.GetInputer {
			return NewGetInput()
		}
	}
	return &newCfg, nil
}

// DefaultGetHandlerLogicConfig holds the default configuration for the get
// handler logic.
type DefaultGetHandlerLogicConfig[Entity database.Getter] struct {
	OutputFactoryFn func() services.GetOutputer[Entity]
	BeforeCallback  services.BeforeGetCallback
	AfterGetFn      services.AfterGet[Entity]
	EntityFn        func(...db.EntityOption[Entity]) Entity
	TxManager       db.TxManager[Entity]
	ReaderRepo      db.ReaderRepository[Entity]
}

// Validate validates and sets defaults for the get handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultGetHandlerLogicConfig[Entity]) Validate() (*DefaultGetHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() services.GetOutputer[Entity] {
			return NewDefaultGetOutput[Entity]()
		}
	}
	if newCfg.EntityFn == nil {
		return nil, errors.New(
			"Validate: EntityFn is required in GetConfig",
		)
	}
	if newCfg.TxManager == nil {
		newCfg.TxManager = defaults.TxManager[Entity]()
	}
	if newCfg.ReaderRepo == nil {
		newCfg.ReaderRepo = defaults.ReaderRepo[Entity]()
	}
	return &newCfg, nil
}
