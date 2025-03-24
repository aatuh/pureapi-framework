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

// UpdateConfig holds the configuration for the update endpoint.
type UpdateConfig struct {
	// Default config for the update input handler.
	DefaultInputHandlerConfig *DefaultUpdateInputHandlerConfig
	// Override for the update input handler.
	InputHandlerFactoryFn func() endpointtypes.InputHandler[crudtypes.UpdateInputer]

	// Default config for the update handler logic.
	DefaultHandlerLogicConfig *DefaultUpdateHandlerLogicConfig
	// Override for the update handler logic.
	HandlerLogicFnFactoryFn func() endpoint.HandlerLogicFn[crudtypes.UpdateInputer]

	ErrorHandlerFactoryFn  func() endpointtypes.ErrorHandler
	OutputHandlerFactoryFn func() endpointtypes.OutputHandler
}

// MustValidate validates and sets defaults for the update config.
// It returns a new config with the defaults set.
func (cfg *UpdateConfig) MustValidate(
	crudCfg *CRUDConfig,
) *UpdateConfig {
	newCfg := *cfg

	if newCfg.DefaultInputHandlerConfig == nil {
		newCfg.DefaultInputHandlerConfig = &DefaultUpdateInputHandlerConfig{}
	}
	if newCfg.DefaultInputHandlerConfig.InputFactoryFn == nil {
		newCfg.DefaultInputHandlerConfig.InputFactoryFn = func() crudtypes.UpdateInputer {
			return NewUpdateInput()
		}
	}
	newCfg.InputHandlerFactoryFn = withDefaultFactory(
		newCfg.InputHandlerFactoryFn,
		func() endpointtypes.InputHandler[crudtypes.UpdateInputer] {
			return api.NewMapInputHandler(
				newCfg.DefaultInputHandlerConfig.APIFields,
				crudCfg.ConversionRules,
				crudCfg.CustomRules,
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
			newCfg.DefaultHandlerLogicConfig = &DefaultUpdateHandlerLogicConfig{}
		}
		if newCfg.DefaultHandlerLogicConfig.OutputFactoryFn == nil {
			newCfg.DefaultHandlerLogicConfig.OutputFactoryFn = func() crudtypes.UpdateOutputer {
				return NewUpdateOutput()
			}
		}
		if newCfg.DefaultHandlerLogicConfig.EntityFn == nil {
			panic("MustValidate: EntityFn is required in UpdateConfig")
		}
	}
	newCfg.HandlerLogicFnFactoryFn = withDefaultFactory(
		newCfg.HandlerLogicFnFactoryFn,
		func() endpoint.HandlerLogicFn[crudtypes.UpdateInputer] {
			return SetupUpdateHandler(
				crudCfg.ConnFn,
				newCfg.DefaultHandlerLogicConfig.EntityFn,
				crudCfg.APIToDBFields,
				newCfg.DefaultHandlerLogicConfig.OutputFactoryFn,
				newCfg.DefaultHandlerLogicConfig.BeforeCallback, // Can be nil.
				repository.NewMutatorRepo(
					custom.QueryBuilder(), custom.QueryErrorChecker(),
				),
				defaults.DefaultTxManager[*int64](),
			).Handle
		},
	)

	newCfg.ErrorHandlerFactoryFn = withDefaultFactory(
		newCfg.ErrorHandlerFactoryFn,
		defaultErrorHandlerFactory(crudCfg, errors.UpdateErrors()),
	)
	newCfg.OutputHandlerFactoryFn = withDefaultFactory(
		newCfg.OutputHandlerFactoryFn,
		defaultOutputHandlerFactory(crudCfg),
	)

	return &newCfg
}

// UpdateDefinition creates a definition for the update endpoint.
func UpdateDefinition(cfg *CRUDConfig) *endpoint.DefaultDefinition {
	handler := NewUpdateEndpointHandler(cfg.Update, cfg.EmitterLogger)
	return newDefinition(cfg.URL, http.MethodPut, cfg.Stack, handler)
}

// UpdateHandler creates a handler for the update endpoint.
func NewUpdateEndpointHandler(
	cfg *UpdateConfig,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.UpdateInputer] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}

type UpdateInput struct {
	Selectors dbinput.Selectors `json:"selectors"`
	Updates   dbinput.Updates   `json:"updates"`
	Upsert    bool              `json:"upsert"`
}

func NewUpdateInput() *UpdateInput {
	return &UpdateInput{}
}

func (u *UpdateInput) GetSelectors() dbinput.Selectors { return u.Selectors }
func (u *UpdateInput) GetUpdates() dbinput.Updates     { return u.Updates }
func (u *UpdateInput) GetUpsert() bool                 { return u.Upsert }

type UpdateOutput struct {
	Count int64 `json:"count"`
}

func NewUpdateOutput() *UpdateOutput {
	return &UpdateOutput{}
}

func (u *UpdateOutput) SetCount(count int64) { u.Count = count }

// SetupUpdateHandler sets up an endpoint handler for the update operation.
func SetupUpdateHandler(
	connFn repositorytypes.ConnFn,
	entityFn func(opts ...crudtypes.EntityOption) databasetypes.Mutator,
	apiToDBFields crudtypes.APIToDBFields,
	outputFactoryFn func() crudtypes.UpdateOutputer,
	beforeCallback crudtypes.BeforeUpdateCallback,
	mutatorRepo repositorytypes.MutatorRepo,
	txManager repositorytypes.TxManager[*int64],
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
			updater databasetypes.Mutator,
		) (int64, error) {
			return services.UpdateInvoke(
				ctx,
				parsedInput,
				connFn,
				updater,
				mutatorRepo,
				txManager,
			)
		},
		func(count int64) (crudtypes.UpdateOutputer, error) {
			output := outputFactoryFn()
			output.SetCount(count)
			return output, nil
		},
		func() databasetypes.Mutator { return entityFn() },
		beforeCallback,
	)
}
