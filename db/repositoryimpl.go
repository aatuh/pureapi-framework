package db

import (
	"context"
	"fmt"

	"github.com/aatuh/pureapi-core/database"
)

// DefaultReaderRepo implements read operations.
type DefaultReaderRepo[Entity database.Getter] struct {
	ReaderQuery  ReaderQuery
	ErrorChecker database.ErrorChecker
}

// DefaultReaderRepo implements ReaderRepo.
var _ ReaderRepository[database.Getter] = (*DefaultReaderRepo[database.Getter])(nil)

// NewReaderRepo creates a new readerRepo.
//
// Parameters:
//   - ctx: Context to use.
//   - readerQuery: The query builder to use for building queries.
//   - errorChecker: The ErrorChecker to use for checking errors.
//
// Returns:
//   - *readerRepo: A new readerRepo.
func NewReaderRepo[Entity database.Getter](
	readerQuery ReaderQuery,
	errorChecker database.ErrorChecker,
) *DefaultReaderRepo[Entity] {
	return &DefaultReaderRepo[Entity]{
		ReaderQuery:  readerQuery,
		ErrorChecker: errorChecker,
	}
}

// GetOne retrieves a single record from the DB by delegating to dbOps.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - factoryFn: A function that returns a new instance of T.
//   - getOpts: The GetOptions to use for the
//
// Returns:
//   - T: The entity scanned from the
//   - error: An error if the query fails.
func (r *DefaultReaderRepo[Entity]) GetOne(
	ctx context.Context,
	preparer database.Preparer,
	factoryFn GetterFactoryFn[Entity],
	getOpts *GetOptions,
) (Entity, error) {
	tableName := factoryFn().TableName()
	query, params := r.ReaderQuery.Get(tableName, getOpts)
	return database.QuerySingleEntity(
		ctx, preparer, query, params, r.ErrorChecker, factoryFn,
	)
}

// GetMany retrieves multiple records from the DB.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - factoryFn: A function that returns a new instance of T.
//   - getOpts: The GetOptions to use for the
//
// Returns:
//   - []T: A slice of entities scanned from the
//   - error: An error if the query fails.
func (r *DefaultReaderRepo[Entity]) GetMany(
	ctx context.Context,
	preparer database.Preparer,
	factoryFn GetterFactoryFn[Entity],
	getOpts *GetOptions,
) ([]Entity, error) {
	tableName := factoryFn().TableName()
	query, params := r.ReaderQuery.Get(tableName, getOpts)
	return database.QueryEntities(
		ctx, preparer, query, params, r.ErrorChecker, factoryFn,
	)
}

// Count returns the count of matching records.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - selectors: The Selectors to use for the
//   - page: The Page to use for the
//   - factoryFn: A function that returns a new instance of T.
//
// Returns:
//   - int: The count of matching records.
//   - error: An error if the query fails.
func (r *DefaultReaderRepo[Entity]) Count(
	ctx context.Context,
	preparer database.Preparer,
	selectors Selectors,
	page *Page,
	factoryFn GetterFactoryFn[Entity],
) (int, error) {
	tableName := factoryFn().TableName()
	countOpts := &CountOptions{
		Selectors: selectors,
		Page:      page,
	}
	query, params := r.ReaderQuery.Count(tableName, countOpts)
	result, err := database.QuerySingleValue(
		ctx,
		preparer,
		query,
		params,
		r.ErrorChecker, func() *int {
			return new(int)
		},
	)
	if err != nil {
		return 0, err
	}
	return *result, nil
}

// Query executes a custom SQL query that is already built.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//   - factoryFn: A function that returns a new instance of T.
//
// Returns:
//   - []T: A slice of entities scanned from the
//   - error: An error if the query fails.
func (r *DefaultReaderRepo[Entity]) Query(
	ctx context.Context,
	preparer database.Preparer,
	query string,
	parameters []any,
	factoryFn GetterFactoryFn[Entity],
) ([]Entity, error) {
	return database.QueryEntities(
		ctx, preparer, query, parameters, r.ErrorChecker, factoryFn,
	)
}

// DefaultMutatorRepo implements mutation operations.
type DefaultMutatorRepo[Entity database.Mutator] struct {
	MutatorQuery MutatorQuery
	ErrorChecker database.ErrorChecker
}

// DefaultMutatorRepo implements MutatorRepo.
var _ MutatorRepository[database.Mutator] = (*DefaultMutatorRepo[database.Mutator])(nil)

// NewMutatorRepo creates a new mutatorRepo.
//
// Parameters:
//   - ctx: Context to use.
//   - mutatorQuery: The query builder to use for the repository.
//   - errorChecker: The error checker to use for the repository.
//
// Returns:
//   - *mutatorRepo: A new mutatorRepo.
func NewMutatorRepo[Entity database.Mutator](
	mutatorQuery MutatorQuery,
	errorChecker database.ErrorChecker,
) *DefaultMutatorRepo[Entity] {
	return &DefaultMutatorRepo[Entity]{
		MutatorQuery: mutatorQuery,
		ErrorChecker: errorChecker,
	}
}

// Insert builds an insert query and executes it.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - mutator: The entity to insert.
//
// Returns:
//   - T: The inserted entity.
//   - error: An error if the query fails.
func (r *DefaultMutatorRepo[Entity]) Insert(
	ctx context.Context, preparer database.Preparer, mutator Entity,
) (Entity, error) {
	query, params := r.MutatorQuery.Insert(
		mutator.TableName(), mutator.InsertedValues,
	)
	result, err := database.Exec(ctx, preparer, query, params, r.ErrorChecker)
	if err != nil {
		return mutator, err
	}
	_, err = result.LastInsertId()
	if err != nil && r.ErrorChecker != nil {
		return mutator, r.ErrorChecker.Check(err)
	}
	return mutator, err
}

// Update builds an update query and executes it.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - updater: The entity to update.
//   - selectors: The selectors to use for the update.
//   - updates: The updates to apply to the entity.
//
// Returns:
//   - int64: The number of rows affected by the update.
//   - error: An error if the query fails.
func (r *DefaultMutatorRepo[Entity]) Update(
	ctx context.Context,
	preparer database.Preparer,
	updater Entity,
	selectors Selectors,
	updates Updates,
) (int64, error) {
	query, params := r.MutatorQuery.UpdateQuery(
		updater.TableName(), updates, selectors,
	)
	result, err := database.Exec(ctx, preparer, query, params, r.ErrorChecker)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if r.ErrorChecker != nil {
			return 0, r.ErrorChecker.Check(err)
		}
		return 0, err
	}
	return rowsAffected, nil
}

// Delete builds a delete query and executes it.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - deleter: The entity to delete.
//   - selectors: The selectors to use for the delete.
//   - deleteOpts: The delete options.
//
// Returns:
//   - int64: The number of rows affected by the delete.
//   - error: An error if the query fails.
func (r *DefaultMutatorRepo[Entity]) Delete(
	ctx context.Context,
	preparer database.Preparer,
	deleter Entity,
	selectors Selectors,
	deleteOpts *DeleteOptions,
) (int64, error) {
	query, params := r.MutatorQuery.Delete(
		deleter.TableName(), selectors, deleteOpts,
	)
	result, err := database.Exec(ctx, preparer, query, params, r.ErrorChecker)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if r.ErrorChecker != nil {
			return 0, r.ErrorChecker.Check(err)
		}
		return 0, err
	}
	return rowsAffected, nil
}

// DefaultCustomRepo implements the CustomRepo interface.
// It can be used to execute custom SQL queries and map the results into custom
// entities. It supports both object and scalar values.
type DefaultCustomRepo[T any] struct {
	ErrorChecker database.ErrorChecker
}

// customRepo implements the CustomRepo interface.
var _ CustomRepository[any] = (*DefaultCustomRepo[any])(nil)

// NewCustomRepo creates a new customRepo.
func NewCustomRepo[T any](
	errorChecker database.ErrorChecker,
) CustomRepository[T] {
	return &DefaultCustomRepo[T]{ErrorChecker: errorChecker}
}

// QueryCustom executes a custom SQL query and maps the results into a slice of
// custom entities. It detects is the custom entity is a database.Getter and
// uses the getter to populate the entity.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the query.
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//   - factoryFn: A function that creates a custom entity from a database row.
//
// Returns:
//   - []T: A slice of custom entities scanned from the query.
//   - error: An error if the query fails.
func (r *DefaultCustomRepo[T]) QueryCustom(
	ctx context.Context,
	preparer database.Preparer,
	query string,
	parameters []any,
	factoryFn func() T,
) ([]T, error) {
	if preparer == nil {
		return nil, fmt.Errorf("QueryCustom: preparer is nil")
	}
	rows, stmt, err := database.Query(
		ctx, preparer, query, parameters, r.ErrorChecker,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If sample implements Getter, perform full-entity scanning manually.
	sample := factoryFn()
	if _, ok := any(sample).(database.Getter); ok {
		var results []T
		for rows.Next() {
			e := factoryFn()
			getter, ok := any(e).(database.Getter)
			if !ok {
				return nil, fmt.Errorf("QueryCustom: type assertion failed")
			}
			if err := getter.ScanRow(rows); err != nil {
				return nil, err
			}
			results = append(results, e)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return results, nil
	}

	// Otherwise, assume a scalar query.
	return database.RowsToAnyScalars(ctx, rows, factoryFn)
}

// DefaultRawQueryer provides direct query execution.
type DefaultRawQueryer struct{}

// DefaultRawQueryer implements RawQueryer.
var _ RawQueryer = (*DefaultRawQueryer)(nil)

// NewRawQueryer creates a new rawQueryer.
//
// Returns:
//   - *rawQueryer: A new rawQueryer.
func NewRawQueryer() *DefaultRawQueryer {
	return &DefaultRawQueryer{}
}

// rawQueryer implements RawQueryer.
var _ RawQueryer = (*DefaultRawQueryer)(nil)

// Exec executes a query using a prepared statement.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Result: The Result of the
//   - error: An error if the query fails.
func (rq *DefaultRawQueryer) Exec(
	ctx context.Context,
	preparer database.Preparer,
	query string,
	parameters []any,
) (database.Result, error) {
	return database.Exec(ctx, preparer, query, parameters, nil)
}

// ExecRaw executes a query directly on the DB.
//
// Parameters:
//   - ctx: Context to use.
//   - db: The DB to execute the query on.
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Result: The Result of the
//   - error: An error if the query fails.
func (rq *DefaultRawQueryer) ExecRaw(
	ctx context.Context, db database.DB, query string, parameters []any,
) (database.Result, error) {
	return database.ExecRaw(ctx, db, query, parameters, nil)
}

// Query executes a query that returns rows. The caller is responsible for
// closing the rows.
//
// Parameters:
//   - ctx: Context to use.
//   - preparer: The preparer to use for the
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Rows: The rows of the
//   - error: An error if the query fails.
func (rq *DefaultRawQueryer) Query(
	ctx context.Context,
	preparer database.Preparer,
	query string,
	parameters []any,
) (database.Rows, error) {
	rows, stmt, err := database.Query(ctx, preparer, query, parameters, nil)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return rows, nil
}

// QueryRaw executes a query directly on the DB without preparation.
//
// Parameters:
//   - ctx: Context to use.
//   - db: The DB to execute the query on.
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Rows: The rows of the
//   - error: An error if the query fails.
//
//nolint:ireturn
func (rq *DefaultRawQueryer) QueryRaw(
	ctx context.Context, db database.DB, query string, parameters []any,
) (database.Rows, error) {
	return database.QueryRaw(ctx, db, query, parameters, nil)
}

// DefaultTxManager is the default transaction manager.
type DefaultTxManager[Entity any] struct{}

// DefaultTxManager implements TxManager.
var _ TxManager[any] = (*DefaultTxManager[any])(nil)

// NewTxManager returns a new txManager.
//
// Returns:
//   - *txManager[Entity]: The new txManager.
func NewTxManager[Entity any]() *DefaultTxManager[Entity] {
	return &DefaultTxManager[Entity]{}
}

// WithTransaction wraps a function call in a DB transaction.
//
// Parameters:
//   - ctx: Context to use.
//   - ctx: The context to use for the transaction.
//   - connFn: The function to get a DB connection.
//   - callback: The function to call in the transaction.
//
// Returns:
//   - Entity: The result of the callback.
//   - error: An error if the transaction fails.
func (t *DefaultTxManager[Entity]) WithTransaction(
	ctx context.Context,
	connFn ConnFn,
	callback database.TxFn[Entity],
) (Entity, error) {
	conn, err := connFn()
	if err != nil {
		var zero Entity
		return zero, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		var zero Entity
		return zero, err
	}
	return database.Transaction(ctx, tx, callback)
}
