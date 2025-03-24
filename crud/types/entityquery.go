package types

import "github.com/pureapi/pureapi-core/database/types"

// EntityOption defines a functional option for configuring an entity.
type EntityOption func(types.CRUDEntity)

// OptionEntityFn is a function that returns an entity with the given options.
type OptionEntityFn func(
	opts ...EntityOption,
) types.CRUDEntity
