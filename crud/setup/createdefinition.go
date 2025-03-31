package setup

import (
	"net/http"

	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
)

// CreateDefinition creates a definition for the create endpoint.
func CreateDefinition[Entity databasetypes.Mutator](
	cfg *CreateConfig[Entity],
	url string,
	stack endpointtypes.Stack,
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultDefinition {
	handler := NewCreateEndpointHandler(cfg, emitterLogger)
	return newDefinition(url, http.MethodPost, stack, handler)
}

// CreateHandler creates a handler for the create endpoint.
func NewCreateEndpointHandler[Entity databasetypes.Mutator](
	cfg *CreateConfig[Entity],
	emitterLogger utiltypes.EmitterLogger,
) *endpoint.DefaultHandler[crudtypes.CreateInputer[Entity]] {
	return newHandler(
		cfg.InputHandlerFactoryFn(),
		cfg.HandlerLogicFnFactoryFn(),
		cfg.ErrorHandlerFactoryFn(),
		cfg.OutputHandlerFactoryFn(),
		emitterLogger,
	)
}
