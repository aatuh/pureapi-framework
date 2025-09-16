package setup

import (
	"net/http"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
)

// NewCRUDDefinitions creates endpoint definitions based on the enabled ops.
func NewCRUDDefinitions[Entity database.CRUDEntity](
	cfg *CRUDConfig[Entity],
) *CRUDDefinitions {
	validCfg := cfg.MustValidate()
	defs := &CRUDDefinitions{
		Endpoints: make(map[CRUDOperation]endpoint.Definition),
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

// CreateDefinition creates a definition for the create endpoint.
func CreateDefinition[Entity database.Mutator](
	cfg *CreateConfig[Entity],
	url string,
	stack endpoint.Stack,
	emitterLogger event.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
	return newDefinition(url, http.MethodPost, stack, handler)
}

// GetDefinition creates a definition for the get endpoint.
func GetDefinition[Entity database.Getter](
	cfg *GetConfig[Entity],
	url string,
	stack endpoint.Stack,
	emitterLogger event.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
	return newDefinition(url, http.MethodGet, stack, handler)
}

// UpdateDefinition creates a definition for the update endpoint.
func UpdateDefinition[Entity database.Mutator](
	cfg *UpdateConfig[Entity],
	url string,
	stack endpoint.Stack,
	emitterLogger event.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
	return newDefinition(url, http.MethodPut, stack, handler)
}

// DeleteDefinition creates a definition for the delete endpoint.
func DeleteDefinition[Entity database.Mutator](
	cfg *DeleteConfig[Entity],
	url string,
	stack endpoint.Stack,
	emitterLogger event.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
	return newDefinition(url, http.MethodDelete, stack, handler)
}
