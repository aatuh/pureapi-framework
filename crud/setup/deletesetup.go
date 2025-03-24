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
	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/dbinput"
	"github.com/pureapi/pureapi-framework/defaults"
	"github.com/pureapi/pureapi-framework/repository"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// DeleteConfig holds the configuration for the delete endpoint.
type DeleteConfig struct {
	// Default config for the delete input handler.
	DefaultInputHandlerConfig *DefaultDeleteInputHandlerConfig
	// Override for the delete input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.DeleteInputer]

	// Default config for the delete handler logic.
	DefaultHandlerLogicConfig *DefaultDeleteHandlerLogicConfig
	// Override for the delete handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.DeleteInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the delete config.
// It returns a new config with the defaults set.
func (cfg *DeleteConfig) MustValidate(crudCfg *CRUDConfig) *DeleteConfig {
	newCfg := *cfg

	if newCfg.DefaultInputHandlerConfig == nil {
		newCfg.DefaultInputHandlerConfig = &DefaultDeleteInputHandlerConfig{}
	}
	if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
		newCfg.DefaultInputHandlerConfig.InputFactoryFn = func() crudtypes.DeleteInputer {
			return NewDeleteInput()
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.DeleteInputer] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
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
			newCfg.DefaultHandlerLogicConfig = &DefaultDeleteHandlerLogicConfig{}
		}
		if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
			newCfg.DefaultHandlerLogicConfig.OutputFactoryFn = func() crudtypes.DeleteOutputer {
				return NewDeleteOutput()
			}
		}
		if newCfg.DefaultHandlerLogicConfig.EntityFn == nil {
			panic("MustValidate: EntityFn is required in DefaultHandlerLogicConfig")
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.DeleteInputer] {
			return SetupDeleteHandler(
				crudCfg.ConnFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				crudCfg.APIToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback,
				repository.NewMutatorRepo(
					custom.QueryBuilder(), custom.QueryErrorChecker(),
				),
				defaults.DefaultTxManager[*int64](),
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(crudCfg, errors.DeleteErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(crudCfg),
	)

	return &newCfg
}

// DeleteDefinition creates a definition for the delete endpoint.
func DeleteDefinition(
	cfg *CRUDConfig,
) *endpoint.DefaultDefinition {
	handler := NewDeleteEndpointHandler(cfg.Delete, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodDelete, cfg.Stack, handler)
}

// DeleteHandler creates a handler for the delete endpoint.
func NewDeleteEndpointHandler(
	cfg *DeleteConfig,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.DeleteInputer] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}

type DeleteInput struct {
	Selectors dbinput.Selectors `json:"selectors"`
}

func NewDeleteInput() *DeleteInput {
	return &DeleteInput{}
}

func (d *DeleteInput) GetSelectors() dbinput.Selectors { return d.Selectors }

type DeleteOutput struct {
	Count int64 `json:"count"`
}

func NewDeleteOutput() *DeleteOutput {
	return &DeleteOutput{}
}

func (d *DeleteOutput) SetCount(count int64) { d.Count = count }

// SetupDeleteHandler sets up an endpoint handler for the delete operation.
func SetupDeleteHandler(
	connFn repositorytypes.ConnFn,
	entityFn func(opts ...crudtypes.EntityOption) databasetypes.Mutator,
	apiToDBFields crudtypes.APIToDBFields,
	outputFactoryFn func() crudtypes.DeleteOutputer,
	beforeCallback crudtypes.BeforeDeleteCallback,
	mutatorRepo repositorytypes.MutatorRepo,
	txManager repositorytypes.TxManager[*int64],
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
			entity databasetypes.Mutator,
		) (int64, error) {
			return services.DeleteInvoke(
				ctx, parsedInput, connFn, entity, mutatorRepo, txManager,
			)
		},
		func(count int64) (crudtypes.DeleteOutputer, error) {
			output := outputFactoryFn()
			output.SetCount(count)
			return output, nil
		},
		func() databasetypes.Mutator { return entityFn() },
		beforeCallback,
	)
}
