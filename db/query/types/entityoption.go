package types

import "github.com/pureapi/pureapi-core/database/types"

// EntityOption defines a functional option for configuring an entity.
type EntityOption[T any] func(T)

// OptionEntityFactoryFn is a function that returns an entity with the given
// options.
type OptionEntityFactoryFn[Entity types.CRUDEntity] func(
	opts ...EntityOption[Entity],
) Entity
