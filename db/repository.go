package db

import (
	"context"

	"github.com/aatuh/pureapi-core/database"
)

// ConnFn returns a database connection.
type ConnFn func() (database.DB, error)

// GetterFactoryFn returns a Getter factory function.
type GetterFactoryFn[Entity database.Getter] func() Entity

// ReaderRepository defines retrieval-related operations.
type ReaderRepository[Entity database.Getter] interface {
	// GetOne retrieves a single record from the DB.
	GetOne(
		ctx context.Context,
		preparer database.Preparer,
		factoryFn GetterFactoryFn[Entity],
		getOptions *GetOptions,
	) (Entity, error)

	// GetMany retrieves multiple records from the DB.
	GetMany(
		ctx context.Context,
		preparer database.Preparer,
		factoryFn GetterFactoryFn[Entity],
		getOptions *GetOptions,
	) ([]Entity, error)

	// Count returns a record count.
	Count(
		ctx context.Context,
		preparer database.Preparer,
		selectors Selectors,
		page *Page,
		factoryFn GetterFactoryFn[Entity],
	) (int, error)
}

// MutatorRepository defines mutation-related operations.
type MutatorRepository[Entity database.Mutator] interface {
	// Insert builds an insert query and executes it.
	Insert(
		ctx context.Context, preparer database.Preparer, entity Entity,
	) (Entity, error)

	// Update builds an update query and executes it.
	Update(
		ctx context.Context,
		preparer database.Preparer,
		entity Entity,
		selectors Selectors,
		updates Updates,
	) (int64, error)

	// Delete builds a delete query and executes it.
	Delete(
		ctx context.Context,
		preparer database.Preparer,
		entity Entity,
		selectors Selectors,
		deleteOpts *DeleteOptions,
	) (int64, error)
}

// CustomRepository defines methods for executing custom SQL queries and mapping
// the results into custom entities.
type CustomRepository[Entity any] interface {
	// QueryCustom executes a custom SQL  It returns a slice of T or an
	// error if the query or scan fails.
	QueryCustom(
		ctx context.Context,
		preparer database.Preparer,
		query string,
		parameters []any,
		factoryFn func() Entity,
	) ([]Entity, error)
}

// RawQueryer defines generic methods for executing raw queries and commands.
type RawQueryer interface {
	// Exec executes a query using a prepared statement that does not return
	// rows.
	Exec(
		ctx context.Context,
		preparer database.Preparer,
		query string,
		parameters []any,
	) (database.Result, error)

	// ExecRaw executes a query directly on the DB without explicit preparation.
	ExecRaw(
		ctx context.Context,
		db database.DB,
		query string,
		parameters []any,
	) (database.Result, error)

	// Query prepares and executes a query that returns rows. Returns rows.
	// The caller is responsible for closing the returned rows.
	Query(
		ctx context.Context,
		preparer database.Preparer,
		query string,
		parameters []any,
	) (database.Rows, error)

	// QueryRaw executes a query directly on the DB without preparation and
	// returns rows. The caller is responsible for closing the returned rows.
	QueryRaw(ctx context.Context, db database.DB, query string, parameters []any,
	) (database.Rows, error)
}

// TxManager is an interface for transaction management.
type TxManager[T any] interface {
	// WithTransaction wraps a function call in a DB transaction.
	WithTransaction(
		ctx context.Context, connFn ConnFn, callback database.TxFn[T],
	) (T, error)
}
