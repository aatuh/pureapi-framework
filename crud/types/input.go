package types

import (
	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-framework/dbinput"
)

type CreateInputer interface {
	GetEntity() databasetypes.Mutator
}

type CreateOutputer interface {
	SetEntities(entities []databasetypes.Mutator)
}

type GetInputer interface {
	GetSelectors() dbinput.Selectors
	GetOrders() dbinput.Orders
	GetPage() *dbinput.Page
	GetCount() bool
}

type GetOutputer interface {
	SetEntities(entities []databasetypes.Getter)
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
