package setup

import (
	databasetypes "github.com/pureapi/pureapi-core/database/types"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/defaults"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
)

// CRUDConfig holds all the settings for the CRUD endpoints.
type CRUDConfig[Entity databasetypes.CRUDEntity] struct {
	SystemID      string
	URL           string
	Stack         endpointtypes.Stack
	EmitterLogger utiltypes.EmitterLogger

	EntityFn      func(...crudtypes.EntityOption[Entity]) Entity
	ConnFn        repositorytypes.ConnFn
	APIToDBFields crudtypes.APIToDBFields

	MutatorRepo repositorytypes.MutatorRepo[Entity]
	ReaderRepo  repositorytypes.ReaderRepo[Entity]
	TxManager   repositorytypes.TxManager[Entity]

	ConversionRules map[string]func(any) any
	CustomRules     map[string]func(any) error

	Create *CreateConfig[Entity]
	Get    *GetConfig[Entity]
	Update *UpdateConfig[Entity]
	Delete *DeleteConfig[Entity]
}

// NewCRUDDefinitions creates endpoint definitions based on the enabled ops.
func NewCRUDDefinitions[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
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
func MustValidateCRUDConfig[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *CRUDConfig[Entity] {
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
func MustValidateMainCRUDConfig[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *CRUDConfig[Entity] {
	newCfg := *cfg

	if newCfg.SystemID == "" {
		panic("SystemID is required")
	}
	if newCfg.URL == "" {
		panic("URL is required")
	}
	if newCfg.Stack == nil {
		newCfg.Stack = defaults.DefaultStackBuilder().Build()
	}
	if newCfg.EntityFn == nil {
		panic("EntityFn is required")
	}
	if newCfg.ConnFn == nil {
		panic("ConnFn is required")
	}
	if newCfg.MutatorRepo == nil {
		newCfg.MutatorRepo = defaults.DefaultMutatorRepo[Entity]()
	}
	if newCfg.ReaderRepo == nil {
		newCfg.ReaderRepo = defaults.DefaultReaderRepo[Entity]()
	}
	if newCfg.TxManager == nil {
		newCfg.TxManager = defaults.DefaultTxManager[Entity]()
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
