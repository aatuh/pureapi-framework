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
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
	"github.com/pureapi/pureapi-framework/util/apimapper"
	apimappertypes "github.com/pureapi/pureapi-framework/util/apimapper/types"
)

type DefaultGetInput struct {
	Selectors input.Selectors `json:"selectors"`
	Orders    input.Orders    `json:"orders"`
	Page      *input.Page     `json:"page"`
	Count     bool            `json:"count"`
}

func NewGetInput() *DefaultGetInput {
	return &DefaultGetInput{}
}

func (i *DefaultGetInput) GetSelectors() input.Selectors { return i.Selectors }
func (i *DefaultGetInput) GetOrders() input.Orders       { return i.Orders }
func (i *DefaultGetInput) GetPage() *input.Page          { return i.Page }
func (i *DefaultGetInput) GetCount() bool                { return i.Count }

type DefaultGetOutput[Entity databasetypes.Getter] struct {
	Entities []Entity `json:"entities"`
	Count    int      `json:"count"`
}

func NewDefaultGetOutput[Entity databasetypes.Getter]() *DefaultGetOutput[Entity] {
	return &DefaultGetOutput[Entity]{}
}

func (o *DefaultGetOutput[Entity]) SetEntities(entities []Entity) { o.Entities = entities }
func (o *DefaultGetOutput[Entity]) SetCount(count int)            { o.Count = count }

// GetConfig holds the configuration for the get endpoint.
type GetConfig[Entity databasetypes.Getter] struct {
	// Default config for the get input handler.
	DefaultInputHandlerConfig *DefaultGetInputHandlerConfig
	// Override for the get input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.GetInputer]

	// Default config for the get handler logic.
	DefaultHandlerLogicConfig *DefaultGetHandlerLogicConfig[Entity]
	// Override for the get handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.GetInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// Validate validates and sets defaults for the get config.
// It returns a new config with the defaults set.
func (cfg *GetConfig[Entity]) Validate(
	systemID string,
	emitterLogger utiltypes.EmitterLogger,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
	connFn repositorytypes.ConnFn,
	apiToDBFields apimappertypes.APIToDBFields,
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
		func() endpointtypes.InputHandler[crudtypes.GetInputer] {
			return apimapper.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				conversionRules,
				customRules,
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
		func() endpoint.HandlerLogicFn[crudtypes.GetInputer] {
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
func DefaultGetHandler[Entity databasetypes.Getter](
	connFn repositorytypes.ConnFn,
	readerRepo repositorytypes.ReaderRepo[Entity],
	txManager repositorytypes.TxManager[Entity],
	entityFn func(opts ...querytypes.EntityOption[Entity]) Entity,
	apiToDBFields apimappertypes.APIToDBFields,
	outputFactoryFn func() crudtypes.GetOutputer[Entity],
	beforeCallback crudtypes.BeforeGetCallback,
	afterGetFn services.AfterGet[Entity],
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
			entityFactoryFn repositorytypes.GetterFactoryFn[Entity],
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
		) (crudtypes.GetOutputer[Entity], error) {
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
	APIFields      apimapper.APIFields
	InputFactoryFn func() crudtypes.GetInputer
}

// Validate validates and sets defaults for the get input handler config.
// It returns a new config with the defaults set.
func (cfg *DefaultGetInputHandlerConfig) Validate() (*DefaultGetInputHandlerConfig, error) {
	newCfg := *cfg
	if newCfg.InputFactoryFn == nil {
		newCfg.InputFactoryFn = func() crudtypes.GetInputer {
			return NewGetInput()
		}
	}
	return &newCfg, nil
}

// DefaultGetHandlerLogicConfig holds the default configuration for the get
// handler logic.
type DefaultGetHandlerLogicConfig[Entity databasetypes.Getter] struct {
	OutputFactoryFn func() crudtypes.GetOutputer[Entity]
	BeforeCallback  crudtypes.BeforeGetCallback
	AfterGetFn      services.AfterGet[Entity]
	EntityFn        func(...querytypes.EntityOption[Entity]) Entity
	TxManager       repositorytypes.TxManager[Entity]
	ReaderRepo      repositorytypes.ReaderRepo[Entity]
}

// Validate validates and sets defaults for the get handler logic config.
// It returns a new config with the defaults set.
func (cfg *DefaultGetHandlerLogicConfig[Entity]) Validate() (*DefaultGetHandlerLogicConfig[Entity], error) {
	newCfg := *cfg
	if newCfg.OutputFactoryFn == nil {
		newCfg.OutputFactoryFn = func() crudtypes.GetOutputer[Entity] {
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
