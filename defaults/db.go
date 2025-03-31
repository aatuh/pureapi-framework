package defaults

import (
	databasetypes "github.com/pureapi/pureapi-core/database/types"
	"github.com/pureapi/pureapi-framework/repository"
	repositorytypes "github.com/pureapi/pureapi-framework/repository/types"
	"github.com/pureapi/pureapi-sqlite/errorchecker"
	"github.com/pureapi/pureapi-sqlite/query"
)

// QueryBuilder returns a new QueryBuilder.
//
// Returns:
//   - QueryBuilder: A new QueryBuilder.
func QueryBuilder() repositorytypes.QueryBuilder {
	return &query.Query{}
}

// QueryErrorChecker returns a new QueryErrorChecker.
//
// Returns:
//   - QueryErrorChecker: A new QueryErrorChecker.
func QueryErrorChecker() databasetypes.ErrorChecker {
	return errorchecker.NewErrorChecker()
}

// MutatorRepo returns a new MutatorRepo.
//
// Returns:
//   - MutatorRepo: A new MutatorRepo.
func MutatorRepo[
	Entity databasetypes.Mutator,
]() repositorytypes.MutatorRepo[Entity] {
	return repository.NewMutatorRepo[Entity](
		QueryBuilder(), QueryErrorChecker(),
	)
}

// ReaderRepo returns a new ReaderRepo.
//
// Returns:
//   - ReaderRepo: A new ReaderRepo.
func ReaderRepo[
	Entity databasetypes.Getter,
]() repositorytypes.ReaderRepo[Entity] {
	return repository.NewReaderRepo[Entity](
		QueryBuilder(), QueryErrorChecker(),
	)
}

// RawQueryer returns a new RawQueryer.
//
// Returns:
//   - RawQueryer: A new RawQueryer.
func RawQueryer() repositorytypes.RawQueryer {
	return repository.NewRawQueryer()
}

// TxManager returns a new TxManager.
//
// Returns:
//   - TxManager: A new TxManager.
func TxManager[T any]() repositorytypes.TxManager[T] {
	return repository.NewTxManager[T]()
}
