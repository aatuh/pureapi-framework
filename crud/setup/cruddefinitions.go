package setup

import (
	databasetypes "github.com/pureapi/pureapi-core/database/types"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
)

// NewCRUDDefinitions creates endpoint definitions based on the enabled ops.
func NewCRUDDefinitions[Entity databasetypes.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *CRUDDefinitions {
	validCfg := cfg.MustValidate()
	defs := &CRUDDefinitions{
		Endpoints: make(map[CRUDOperation]endpointtypes.Definition),
	}
	if validCfg.Create != nil {
		defs.Endpoints[OpCreate] = CreateDefinition(
			validCfg.Create,
			validCfg.URL,
			validCfg.Stack,
			validCfg.EmitterLogger,
		)
	}
	if validCfg.Get != nil {
		defs.Endpoints[OpGet] = GetDefinition(
			validCfg.Get,
			validCfg.URL,
			validCfg.Stack,
			validCfg.EmitterLogger,
		)
	}
	if validCfg.Update != nil {
		defs.Endpoints[OpUpdate] = UpdateDefinition(
			validCfg.Update,
			validCfg.URL,
			validCfg.Stack,
			validCfg.EmitterLogger,
		)
	}
	if validCfg.Delete != nil {
		defs.Endpoints[OpDelete] = DeleteDefinition(
			validCfg.Delete,
			validCfg.URL,
			validCfg.Stack,
			validCfg.EmitterLogger,
		)
	}
	return defs
}
