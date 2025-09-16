package defaults

import (
	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-framework/db"
	"github.com/pureapi/pureapi-sqlite/errorchecker"
	"github.com/pureapi/pureapi-sqlite/query"
)

// Query returns a new Query.
//
// Returns:
//   - Query: A new Query.
func Query() db.Query {
	return &query.Query{}
}

// QueryErrorChecker returns a new QueryErrorChecker.
//
// Returns:
//   - QueryErrorChecker: A new QueryErrorChecker.
func QueryErrorChecker() database.ErrorChecker {
	return errorchecker.NewErrorChecker()
}

// MutatorRepo returns a new MutatorRepo.
//
// Returns:
//   - MutatorRepo: A new MutatorRepo.
func MutatorRepo[
	Entity database.Mutator,
]() db.MutatorRepository[Entity] {
	return db.NewMutatorRepo[Entity](
		Query(), QueryErrorChecker(),
	)
}

// ReaderRepo returns a new ReaderRepo.
//
// Returns:
//   - ReaderRepo: A new ReaderRepo.
func ReaderRepo[
	Entity database.Getter,
]() db.ReaderRepository[Entity] {
	return db.NewReaderRepo[Entity](
		Query(), QueryErrorChecker(),
	)
}

// RawQueryer returns a new RawQueryer.
//
// Returns:
//   - RawQueryer: A new RawQueryer.
func RawQueryer() db.RawQueryer {
	return db.NewRawQueryer()
}

// TxManager returns a new TxManager.
//
// Returns:
//   - TxManager: A new TxManager.
func TxManager[T any]() db.TxManager[T] {
	return db.NewTxManager[T]()
}
