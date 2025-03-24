package setup

import (
	databasetypes "github.com/pureapi/pureapi-core/database/types"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	"github.com/pureapi/pureapi-framework/api"
	crudtypes "github.com/pureapi/pureapi-framework/crud/types"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
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
	Endpoints map[CRUDOperation]endpointtypes.Definition
}

// CRUDOption represents a functional option for configuring CRUDConfig.
type CRUDOption func(*CRUDConfig)

// DefaultCreateInputHandlerConfig holds the default configuration for the
// create input handler.
type DefaultCreateInputHandlerConfig struct {
	APIFields      api.APIFields
	InputFactoryFn func() crudtypes.CreateInputer
}

// DefaultCreateHandlerLogicConfig holds the default configuration for the
// create handler logic.
type DefaultCreateHandlerLogicConfig struct {
	OutputFactoryFn func() crudtypes.CreateOutputer
	BeforeCallback  crudtypes.BeforeCreateCallback
	TxManager       repositorytypes.TxManager[databasetypes.Mutator]
}

// DefaultGetInputHandlerConfig holds the default configuration for the get
// input handler.
type DefaultGetInputHandlerConfig struct {
	APIFields      api.APIFields
	InputFactoryFn func() crudtypes.GetInputer
}

// DefaultGetHandlerLogicConfig holds the default configuration for the get
// handler logic.
type DefaultGetHandlerLogicConfig struct {
	OutputFactoryFn func() crudtypes.GetOutputer
	BeforeCallback  crudtypes.BeforeGetCallback
	EntityFn        func(...crudtypes.EntityOption) databasetypes.Getter
	TxManager       repositorytypes.TxManager[databasetypes.Getter]
}

// DefaultUpdateInputHandlerConfig holds the default configuration for the
// update input handler.
type DefaultUpdateInputHandlerConfig struct {
	APIFields      api.APIFields
	InputFactoryFn func() crudtypes.UpdateInputer
}

// DefaultUpdateHandlerLogicConfig holds the default configuration for the
// update handler logic.
type DefaultUpdateHandlerLogicConfig struct {
	OutputFactoryFn func() crudtypes.UpdateOutputer
	BeforeCallback  crudtypes.BeforeUpdateCallback
	EntityFn        func(...crudtypes.EntityOption) databasetypes.Mutator
}

// DefaultDeleteInputHandlerConfig holds the default configuration for the
// delete input handler.
type DefaultDeleteInputHandlerConfig struct {
	APIFields      api.APIFields
	InputFactoryFn func() crudtypes.DeleteInputer
}

// DefaultDeleteHandlerLogicConfig holds the default configuration for the
// delete handler logic.
type DefaultDeleteHandlerLogicConfig struct {
	OutputFactoryFn func() crudtypes.DeleteOutputer
	BeforeCallback  crudtypes.BeforeDeleteCallback
	EntityFn        func(...crudtypes.EntityOption) databasetypes.Mutator
}
