package setup

import (
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
)

// DeleteDefinition creates a definition for the delete endpoint.
func DeleteDefinition[Entity databasetypes.Mutator](
	cfg *DeleteConfig[Entity],
	url string,
	stack endpointtypes.Stack,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := NewDeleteEndpointHandler(cfg, emitterLogger)
	return newDefinition(url, http.MethodDelete, stack, handler)
}

// DeleteHandler creates a handler for the delete endpoint.
func NewDeleteEndpointHandler[Entity databasetypes.Mutator](
	cfg *DeleteConfig[Entity],
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
