package setup

import (
	"errors"
	"fmt"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-framework/api/input"
	"github.com/aatuh/pureapi-framework/crud/services"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/aatuh/pureapi-framework/defaults"
	"github.com/aatuh/pureapi-framework/util/inpututil"
)

// CRUDOperation defines the type of CRUD operations.
type CRUDOperation string

// Supported CRUD operations.
const (
	OpCreate CRUDOperation = "create"
	OpGet    CRUDOperation = "get"
	OpUpdate CRUDOperation = "update"
	OpDelete CRUDOperation = "delete"
)

// CRUDDefinitions holds a collection of enabled endpoint definitions.
type CRUDDefinitions struct {
	Endpoints map[CRUDOperation]endpoint.Definition
}

// CRUDOption represents a functional option for configuring CRUDConfig.
type CRUDOption[Entity database.CRUDEntity] func(*CRUDConfig[Entity])

// CRUDConfig holds all the settings for the CRUD endpoints.
type CRUDConfig[Entity database.CRUDEntity] struct {
	SystemID      string
	URL           string
	Stack         endpoint.Stack
	EmitterLogger event.EmitterLogger

	ConnFn        db.ConnFn
	APIToDBFields inpututil.APIToDBFields

	ConversionRules map[string]func(any) any
	CustomRules     map[string]func(any) error

	Create *CreateConfig[Entity]
	Get    *GetConfig[Entity]
	Update *UpdateConfig[Entity]
	Delete *DeleteConfig[Entity]
}

// MustValidate validates the config and sets defaults and returns a new config.
// It panics if validation fails.
func (cfg *CRUDConfig[Entity]) MustValidate() *CRUDConfig[Entity] {
	cfg, err := cfg.Validate()
	if err != nil {
		panic(err)
	}
	return cfg
}

// Validate validates the config and sets defaults and returns a new config.
func (cfg *CRUDConfig[Entity]) Validate() (*CRUDConfig[Entity], error) {
	cfg, err := cfg.ValidateMainCRUDConfig()
	if err != nil {
		return nil, fmt.Errorf("Validate main config: %w", err)
	}

	createCfg, err := cfg.Create.Validate(
		cfg.SystemID,
		cfg.EmitterLogger,
		cfg.ConversionRules,
		cfg.CustomRules,
		cfg.ConnFn,
	)
	if err != nil {
		return nil, fmt.Errorf("Validate create config: %w", err)
	}

	cfg.Create = createCfg
	getCfg, err := cfg.Get.Validate(
		cfg.SystemID,
		cfg.EmitterLogger,
		cfg.ConversionRules,
		cfg.CustomRules,
		cfg.ConnFn,
		cfg.APIToDBFields,
	)
	if err != nil {
		return nil, fmt.Errorf("Validate get config: %w", err)
	}

	cfg.Get = getCfg
	updateCfg, err := cfg.Update.Validate(
		cfg.SystemID,
		cfg.EmitterLogger,
		cfg.ConversionRules,
		cfg.CustomRules,
		cfg.ConnFn,
		cfg.APIToDBFields,
	)
	if err != nil {
		return nil, fmt.Errorf("Validate update config: %w", err)
	}

	cfg.Update = updateCfg
	deleteCfg, err := cfg.Delete.Validate(
		cfg.SystemID,
		cfg.EmitterLogger,
		cfg.ConversionRules,
		cfg.CustomRules,
		cfg.ConnFn,
		cfg.APIToDBFields,
	)
	if err != nil {
		return nil, fmt.Errorf("Validate delete config: %w", err)
	}
	cfg.Delete = deleteCfg

	return cfg, nil
}

// ValidateMainCRUDConfig validates the main config and sets defaults
// It returns a new config with the defaults set.
func (cfg *CRUDConfig[Entity]) ValidateMainCRUDConfig() (
	*CRUDConfig[Entity], error,
) {
	newCfg := *cfg
	if newCfg.SystemID == "" {
		return nil, errors.New(
			"ValidateMainCRUDConfig: SystemID is required in CRUDConfig",
		)
	}
	if newCfg.URL == "" {
		return nil, errors.New(
			"ValidateMainCRUDConfig: URL is required in CRUDConfig",
		)
	}
	if newCfg.Stack == nil {
		newCfg.Stack = defaults.NewStackBuilder().Build()
	}
	if newCfg.ConnFn == nil {
		return nil, errors.New(
			"ValidateMainCRUDConfig: ConnFn is required in CRUDConfig",
		)
	}
	if newCfg.EmitterLogger == nil {
		newCfg.EmitterLogger = defaults.EmitterLogger()
	}
	if newCfg.ConversionRules == nil {
		newCfg.ConversionRules = defaults.InputConversionRules()
	}
	if newCfg.CustomRules == nil {
		newCfg.CustomRules = defaults.ValidationRules()
	}
	return &newCfg, nil
}

// DefaultCRUDConfig holds the default settings for the CRUD endpoints.
type DefaultCRUDConfig[Entity database.CRUDEntity] struct {
	SystemID              string
	URL                   string
	ConnFn                db.ConnFn
	EntityFn              func(...db.EntityOption[Entity]) Entity
	APIToDBFields         inpututil.APIToDBFields
	AllAPIFields          input.APIFields
	CreateAPIFields       input.APIFields
	CreateInputFactoryFn  func() services.CreateInputer[Entity]
	CreateOutputFactoryFn func() services.CreateOutputer[Entity]
	GetAPIFields          input.APIFields
	GetOutputFactoryFn    func() services.GetOutputer[Entity]
	UpdateAPIFields       input.APIFields
	DeleteAPIFields       input.APIFields
}

// NewDefaultCRUDConfig returns a new CRUDConfig using the default settings.
func NewDefaultCRUDConfig[Entity database.CRUDEntity](
	cfg *DefaultCRUDConfig[Entity],
) *CRUDConfig[Entity] {
	return &CRUDConfig[Entity]{
		SystemID:      cfg.SystemID,
		URL:           cfg.URL,
		ConnFn:        cfg.ConnFn,
		APIToDBFields: cfg.APIToDBFields,
		Create: &CreateConfig[Entity]{
			DefaultInputHandlerConfig: &DefaultCreateInputHandlerConfig[Entity]{
				APIFields:      cfg.CreateAPIFields,
				InputFactoryFn: cfg.CreateInputFactoryFn,
			},
			DefaultHandlerLogicConfig: &DefaultCreateHandlerLogicConfig[Entity]{
				OutputFactoryFn: cfg.CreateOutputFactoryFn,
			},
		},
		Get: &GetConfig[Entity]{
			DefaultInputHandlerConfig: &DefaultGetInputHandlerConfig{
				APIFields: cfg.GetAPIFields,
			},
			DefaultHandlerLogicConfig: &DefaultGetHandlerLogicConfig[Entity]{
				OutputFactoryFn: cfg.GetOutputFactoryFn,
				EntityFn:        cfg.EntityFn,
			},
		},
		Update: &UpdateConfig[Entity]{
			DefaultInputHandlerConfig: &DefaultUpdateInputHandlerConfig{
				APIFields: cfg.UpdateAPIFields,
			},
			DefaultHandlerLogicConfig: &DefaultUpdateHandlerLogicConfig[Entity]{
				EntityFn: cfg.EntityFn,
			},
		},
		Delete: &DeleteConfig[Entity]{
			DefaultInputHandlerConfig: &DefaultDeleteInputHandlerConfig{
				APIFields: cfg.DeleteAPIFields,
			},
			DefaultHandlerLogicConfig: &DefaultDeleteHandlerLogicConfig[Entity]{
				EntityFn: cfg.EntityFn,
			},
		},
	}
}
