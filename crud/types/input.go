package types

import "github.com/pureapi/pureapi-framework/dbinput"

type CreateInputer[Entity any] interface {
	GetEntity() Entity
}

type CreateOutputer[Entity any] interface {
	SetEntities(entities []Entity)
}

type GetInputer[Entity any] interface {
	GetSelectors() dbinput.Selectors
	GetOrders() dbinput.Orders
	GetPage() *dbinput.Page
	GetCount() bool
}

type GetOutputer[Entity any] interface {
	SetEntities(entities []Entity)
	SetCount(count int)
}

type UpdateInputer interface {
	GetSelectors() dbinput.Selectors
	GetUpdates() dbinput.Updates
	GetUpsert() bool
}

type UpdateOutputer interface {
	SetCount(count int64)
}

type DeleteInputer interface {
	GetSelectors() dbinput.Selectors
}

type DeleteOutputer interface {
	SetCount(count int64)
}
