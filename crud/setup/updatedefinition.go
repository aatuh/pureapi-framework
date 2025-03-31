package setup

import (
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
)

// UpdateDefinition creates a definition for the update endpoint.
func UpdateDefinition[Entity databasetypes.Mutator](
	cfg *UpdateConfig[Entity],
	url string,
	stack endpointtypes.Stack,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := NewUpdateEndpointHandler(cfg, emitterLogger)
	return newDefinition(url, http.MethodPut, stack, handler)
}

// UpdateHandler creates a handler for the update endpoint.
func NewUpdateEndpointHandler[Entity databasetypes.Mutator](
	cfg *UpdateConfig[Entity],
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
