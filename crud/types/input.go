package types

import (
	"github.com/pureapi/pureapi-framework/db/input"
)

type CreateInputer[Entity any] interface {
	GetEntity() Entity
}

type CreateOutputer[Entity any] interface {
	SetEntities(entities []Entity)
}

type GetInputer interface {
	GetSelectors() input.Selectors
	GetOrders() input.Orders
	GetPage() *input.Page
	GetCount() bool
}

type GetOutputer[Entity any] interface {
	SetEntities(entities []Entity)
	SetCount(count int)
}

type UpdateInputer interface {
	GetSelectors() input.Selectors
	GetUpdates() input.Updates
	GetUpsert() bool
}

type UpdateOutputer interface {
	SetCount(count int64)
}

type DeleteInputer interface {
	GetSelectors() input.Selectors
}

type DeleteOutputer interface {
	SetCount(count int64)
}
