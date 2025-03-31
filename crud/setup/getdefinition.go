package setup

import (
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
)

// GetDefinition creates a definition for the get endpoint.
func GetDefinition[Entity databasetypes.Getter](
	cfg *GetConfig[Entity],
	url string,
	stack endpointtypes.Stack,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := NewGetEndpointHandler(cfg, emitterLogger)
	return newDefinition(url, http.MethodGet, stack, handler)
}

// GetHandler creates a handler for the get endpoint.
func NewGetEndpointHandler[Entity databasetypes.Getter](
	cfg *GetConfig[Entity],
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.GetInputer] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}
