package setup

import (
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/defaults"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CRUDConfig holds all the settings for the CRUD endpoints.
type CRUDConfig struct {
	SystemID      string
	URL           string
	Stack         endpointtypes.Stack
	EmitterLogger utiltypes.EmitterLogger

	ConnFn        repositorytypes.ConnFn
	APIToDBFields crudtypes.APIToDBFields

	MutatorRepo repositorytypes.MutatorRepo
	ReaderRepo  repositorytypes.ReaderRepo

	ConversionRules map[string]func(any) any
	CustomRules     map[string]func(any) error

	Create *CreateConfig
	Get    *GetConfig
	Update *UpdateConfig
	Delete *DeleteConfig
}

// NewCRUDDefinitions creates endpoint definitions based on the enabled ops.
func NewCRUDDefinitions(
	cfg *CRUDConfig,
) *CRUDDefinitions {
	validCfg := MustValidateCRUDConfig(cfg)
	defs := &CRUDDefinitions{
		Endpoints: make(map[CRUDOperation]endpointtypes.Definition),
	}
	if validCfg.Create != nil {
		defs.Endpoints[OpCreate] = CreateDefinition(validCfg)
	}
	if validCfg.Get != nil {
		defs.Endpoints[OpGet] = GetDefinition(validCfg)
	}
	if validCfg.Update != nil {
		defs.Endpoints[OpUpdate] = UpdateDefinition(validCfg)
	}
	if validCfg.Delete != nil {
		defs.Endpoints[OpDelete] = DeleteDefinition(validCfg)
	}
	return defs
}

// MustValidateCRUDConfig validates the config and sets defaults.
func MustValidateCRUDConfig(
	cfg *CRUDConfig,
) *CRUDConfig {
	if cfg == nil {
		panic("CRUDConfig is required")
	}

	cfg = MustValidateMainCRUDConfig(cfg)
	cfg.Create = cfg.Create.MustValidate(cfg)
	cfg.Get = cfg.Get.MustValidate(cfg)
	cfg.Update = cfg.Update.MustValidate(cfg)
	cfg.Delete = cfg.Delete.MustValidate(cfg)

	return cfg
}

// MustValidateMainCRUDConfig validates the main config and sets defaults
// It returns a new config with the defaults set.
func MustValidateMainCRUDConfig(
	cfg *CRUDConfig,
) *CRUDConfig {
	newCfg := *cfg

	if newCfg.SystemID == "" {
		panic("MustValidateMainCRUDConfig: SystemID is required in CRUDConfig")
	}
	if newCfg.URL == "" {
		panic("MustValidateMainCRUDConfig: URL is required in CRUDConfig")
	}
	if newCfg.Stack == nil {
		newCfg.Stack = defaults.DefaultStackBuilder().Build()
	}
	if newCfg.ConnFn == nil {
		panic("MustValidateMainCRUDConfig: ConnFn is required in CRUDConfig")
	}
	if newCfg.MutatorRepo == nil {
		newCfg.MutatorRepo = defaults.DefaultMutatorRepo()
	}
	if newCfg.ReaderRepo == nil {
		newCfg.ReaderRepo = defaults.DefaultReaderRepo()
	}
	if newCfg.EmitterLogger == nil {
		newCfg.EmitterLogger = custom.EmitterLogger()
	}
	if newCfg.ConversionRules == nil {
		newCfg.ConversionRules = defaults.DefaultConversionRules()
	}
	if newCfg.CustomRules == nil {
		newCfg.CustomRules = defaults.DefaultCustomRules()
	}

	return &newCfg
}
