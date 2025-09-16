package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/aatuh/pureapi-core/database"
	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
	"github.com/stretchr/testify/suite"
)

// --- SQLite Query Builder Implementation ---

// sqliteQueryBuilder implements DataReaderQuery and DataMutatorQuery for SQLite.
type sqliteQueryBuilder struct{}

func (qb sqliteQueryBuilder) Get(table string, opts *GetOptions) (string, []any) {
	// For simplicity, ignore filtering options.
	query := fmt.Sprintf("SELECT id, name, age FROM %s", table)
	return query, nil
}

func (qb sqliteQueryBuilder) Count(table string, opts *CountOptions) (string, []any) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	return query, nil
}

func (qb sqliteQueryBuilder) Insert(table string, insertedValuesFunc InsertedValuesFn) (string, []any) {
	// Our DummyEntity: id is auto-generated.
	query := fmt.Sprintf("INSERT INTO %s (name, age) VALUES (?, ?)", table)
	_, vals := insertedValuesFunc()
	// Expect vals: [id, name, age]; ignore id.
	if len(vals) >= 3 {
		return query, []any{vals[1], vals[2]}
	}
	return query, nil
}

func (qb sqliteQueryBuilder) UpdateQuery(table string, updates []Update, selectors []Selector) (string, []any) {
	// For simplicity, assume updates for "name" and "age" and a selector on "id".
	query := fmt.Sprintf("UPDATE %s SET name = ?, age = ? WHERE id = ?", table)
	var name interface{}
	var age interface{}
	for _, u := range updates {
		if u.Field == "name" {
			name = u.Value
		} else if u.Field == "age" {
			age = u.Value
		}
	}
	var id interface{}
	for _, sel := range selectors {
		if sel.Column == "id" {
			id = sel.Value
			break
		}
	}
	return query, []any{name, age, id}
}

func (qb sqliteQueryBuilder) Delete(table string, selectors []Selector, opts *DeleteOptions) (string, []any) {
	// Expect a selector on "id".
	var id interface{}
	for _, sel := range selectors {
		if sel.Column == "id" {
			id = sel.Value
			break
		}
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	return query, []any{id}
}

// --- Integration Test Suite ---

type integrationRepoSuite struct {
	suite.Suite
	db         database.DB // Our database adapter (implements DB).
	underlying *sql.DB     // Underlying *sql.DB.
	qb         sqliteQueryBuilder
	errChecker dummyErrorChecker
	ctx        context.Context
}

func TestIntegrationRepoSuite(t *testing.T) {
	suite.Run(t, new(integrationRepoSuite))
}

func (s *integrationRepoSuite) SetupSuite() {
	var err error
	// Open an in-memory SQLite3 database using NewSQLDBAdapter.
	s.db, err = database.NewSQLDBAdapter("sqlite3", ":memory:")
	s.Require().NoError(err)
	s.underlying = s.db.UnderlyingDB()
	s.ctx = context.Background()
	s.qb = sqliteQueryBuilder{}
	s.errChecker = dummyErrorChecker{}

	// Create table "dummy".
	createTable := `
	CREATE TABLE dummy (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		age INTEGER
	);`
	_, err = s.underlying.Exec(createTable)
	s.Require().NoError(err)
}

func (s *integrationRepoSuite) TearDownSuite() {
	s.db.Close()
}

func (s *integrationRepoSuite) resetTable() {
	_, err := s.underlying.Exec("DELETE FROM dummy;")
	s.Require().NoError(err)
}

// TestInsertAndGetOne tests a happy-path insert and retrieval using GetOne.
func (s *integrationRepoSuite) TestInsertAndGetOne() {
	s.resetTable()
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	readRepo := NewReaderRepo[*DummyEntity](s.qb, s.errChecker)

	entity := &DummyEntity{
		Name: "Alice",
		Age:  25,
	}
	inserted, err := mutRepo.Insert(s.ctx, s.db, entity)
	s.Require().NoError(err)

	// Retrieve the last inserted id.
	var lastID int64
	err = s.underlying.QueryRow("SELECT last_insert_rowid()").Scan(&lastID)
	s.Require().NoError(err)
	inserted.ID = int(lastID)

	getOpts := &GetOptions{
		Selectors: NewSelectors(*NewSelector("id", Equal, inserted.ID)),
	}
	got, err := readRepo.GetOne(s.ctx, s.db, dummyEntityFn, getOpts)
	s.Require().NoError(err)
	s.Equal("Alice", got.Name)
	s.Equal(25, got.Age)
	s.Equal(inserted.ID, got.ID)
}

// TestGetMany tests retrieving multiple rows.
func (s *integrationRepoSuite) TestGetMany() {
	s.resetTable()
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	readRepo := NewReaderRepo[*DummyEntity](s.qb, s.errChecker)

	entities := []*DummyEntity{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 35},
	}
	for _, e := range entities {
		_, err := mutRepo.Insert(s.ctx, s.db, e)
		s.Require().NoError(err)
	}
	getOpts := &GetOptions{}
	got, err := readRepo.GetMany(s.ctx, s.db, dummyEntityFn, getOpts)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(got), 3)
}

// TestCount tests the Count method.
func (s *integrationRepoSuite) TestCount() {
	s.resetTable()
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	readRepo := NewReaderRepo[*DummyEntity](s.qb, s.errChecker)

	entities := []*DummyEntity{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
	}
	for _, e := range entities {
		_, err := mutRepo.Insert(s.ctx, s.db, e)
		s.Require().NoError(err)
	}
	count, err := readRepo.Count(s.ctx, s.db, nil, nil, dummyEntityFn)
	s.Require().NoError(err)
	s.Equal(2, count)
}

// TestUpdate tests updating an entity.
func (s *integrationRepoSuite) TestUpdate() {
	s.resetTable()
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	readRepo := NewReaderRepo[*DummyEntity](s.qb, s.errChecker)

	entity := &DummyEntity{Name: "Alice", Age: 25}
	inserted, err := mutRepo.Insert(s.ctx, s.db, entity)
	s.Require().NoError(err)
	var lastID int64
	err = s.underlying.QueryRow("SELECT last_insert_rowid()").Scan(&lastID)
	s.Require().NoError(err)
	inserted.ID = int(lastID)

	// Update the entity's name and age.
	updates := NewUpdates(NewUpdate("name", "Alice Updated"), NewUpdate("age", 26))
	selectors := NewSelectors(*NewSelector("id", Equal, inserted.ID))
	rowsAffected, err := mutRepo.Update(s.ctx, s.db, inserted, selectors, updates)
	s.Require().NoError(err)
	s.Equal(int64(1), rowsAffected)

	getOpts := &GetOptions{
		Selectors: NewSelectors(*NewSelector("id", Equal, inserted.ID)),
	}
	updated, err := readRepo.GetOne(s.ctx, s.db, dummyEntityFn, getOpts)
	s.Require().NoError(err)
	s.Equal("Alice Updated", updated.Name)
	s.Equal(26, updated.Age)
}

// TestDelete tests deleting an entity.
func (s *integrationRepoSuite) TestDelete() {
	s.resetTable()
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	readRepo := NewReaderRepo[*DummyEntity](s.qb, s.errChecker)

	entity := &DummyEntity{Name: "Alice", Age: 25}
	inserted, err := mutRepo.Insert(s.ctx, s.db, entity)
	s.Require().NoError(err)
	var lastID int64
	err = s.underlying.QueryRow("SELECT last_insert_rowid()").Scan(&lastID)
	s.Require().NoError(err)
	inserted.ID = int(lastID)

	selectors := NewSelectors(*NewSelector("id", Equal, inserted.ID))
	rowsAffected, err := mutRepo.Delete(s.ctx, s.db, inserted, selectors, &DeleteOptions{})
	s.Require().NoError(err)
	s.Equal(int64(1), rowsAffected)

	count, err := readRepo.Count(s.ctx, s.db, nil, nil, dummyEntityFn)
	s.Require().NoError(err)
	s.Equal(0, count)
}

// TestCustomRepo_FullEntity tests QueryCustom when the factory returns a full
// entity.
func (s *integrationRepoSuite) TestCustomRepo_FullEntity() {
	s.resetTable()
	// Insert a row using the mutator repo.
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	entity := &DummyEntity{Name: "Alice", Age: 25}
	_, err := mutRepo.Insert(s.ctx, s.db, entity)
	s.Require().NoError(err)

	// Use CustomRepo with a factory returning *DummyEntity (which implements Getter).
	customRepo := NewCustomRepo[*DummyEntity](s.errChecker)
	results, err := customRepo.QueryCustom(
		s.ctx,
		s.db,
		"SELECT id, name, age FROM dummy",
		nil,
		func() *DummyEntity { return dummyEntityFn() },
	)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(results), 1)
	s.Equal("Alice", results[0].Name)
	s.Equal(25, results[0].Age)
}

// TestCustomRepo_Scalar tests QueryCustom when the query returns a scalar.
func (s *integrationRepoSuite) TestCustomRepo_Scalar() {
	s.resetTable()
	// Insert a row using the mutator repo.
	mutRepo := NewMutatorRepo[*DummyEntity](s.qb, s.errChecker)
	entity := &DummyEntity{Name: "Alice", Age: 25}
	_, err := mutRepo.Insert(s.ctx, s.db, entity)
	s.Require().NoError(err)

	// Use CustomRepo with a factory that returns a pointer to a string.
	customRepo := NewCustomRepo[*string](s.errChecker)
	results, err := customRepo.QueryCustom(
		s.ctx,
		s.db,
		"SELECT name FROM dummy",
		nil,
		func() *string {
			var str string
			return &str
		},
	)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(results), 1)
	s.Equal("Alice", *results[0])
}

// TestRawQueryer tests raw query execution.
func (s *integrationRepoSuite) TestRawQueryer() {
	s.resetTable()
	raw := NewRawQueryer()
	// Exec an INSERT.
	_, err := raw.Exec(s.ctx, s.db, "INSERT INTO dummy (name, age) VALUES (?, ?)", []any{"RawUser", 99})
	s.Require().NoError(err)

	// Query using raw.
	rows, err := raw.Query(s.ctx, s.db, "SELECT name, age FROM dummy", nil)
	s.Require().NoError(err)
	defer rows.Close()
	var name string
	var age int
	found := false
	for rows.Next() {
		err = rows.Scan(&name, &age)
		s.Require().NoError(err)
		found = true
	}
	s.True(found, "Expected at least one row")
}

// TestTxManager tests transaction management.
func (s *integrationRepoSuite) TestTxManager() {
	s.resetTable()
	txManager := NewTxManager[*DummyEntity]()
	entity, err := txManager.WithTransaction(s.ctx, func() (database.DB, error) {
		return s.db, nil
	}, func(ctx context.Context, tx database.Tx) (*DummyEntity, error) {
		res, err := tx.Exec("INSERT INTO dummy (name, age) VALUES (?, ?)", "TxUser", 40)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		return &DummyEntity{ID: int(id), Name: "TxUser", Age: 40}, nil
	})
	s.Require().NoError(err)
	s.Equal("TxUser", entity.Name)

	var count int
	err = s.underlying.QueryRow("SELECT COUNT(*) FROM dummy WHERE name = ?", "TxUser").Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}
