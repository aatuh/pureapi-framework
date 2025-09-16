package db

import (
	"context"
	"errors"
	"testing"

	"github.com/aatuh/pureapi-core/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var errDummy = errors.New("dummy error")

// dummyPreparer implements the assumed database.Preparer interface.
type dummyPreparer struct{}

func (d *dummyPreparer) Prepare(query string) (database.Stmt, error) {
	return nil, errDummy
}
func (dp *dummyPreparer) Exec(
	query string, args ...any,
) (database.Result, error) {
	return nil, errDummy
}
func (dp *dummyPreparer) Query(
	query string, args ...any,
) (database.Rows, error) {
	return nil, errDummy
}
func (dp *dummyPreparer) BeginTx(
	ctx context.Context, opts any,
) (database.Tx, error) {
	return nil, errDummy
}

// dummyErrorChecker implements the assumed database.ErrorChecker interface.
type dummyErrorChecker struct{}

func (d dummyErrorChecker) Check(err error) error {
	return err
}

// dummyQuery implements DataReaderQuery and DataMutatorQuery.
type dummyQuery struct{}

func (d dummyQuery) Get(table string, opts *GetOptions) (string, []any) {
	return "SELECT * FROM " + table, []any{table}
}
func (d dummyQuery) Count(
	table string, opts *CountOptions,
) (string, []any) {
	return "SELECT COUNT(*) FROM " + table, []any{table}
}
func (d dummyQuery) Insert(
	table string, insertedValuesFunc InsertedValuesFn,
) (string, []any) {
	return "INSERT INTO " + table, []any{table}
}
func (d dummyQuery) UpdateQuery(
	table string, updates []Update, selectors []Selector,
) (string, []any) {
	return "UPDATE " + table, []any{table}
}
func (d dummyQuery) Delete(
	table string, selectors []Selector, opts *DeleteOptions,
) (string, []any) {
	return "DELETE FROM " + table, []any{table}
}

// dummyGetterFactory returns a new DummyEntity.
func dummyGetterFactory() *DummyEntity {
	return dummyEntityFn()
}

// dummyConnFn simulates obtaining a database connection and returns an error.
func dummyConnFn() (database.DB, error) {
	return nil, errDummy
}

// RepositoryImplTestSuite tests the RepositoryImpl.
type RepositoryImplTestSuite struct {
	suite.Suite
	preparer   database.Preparer
	query      dummyQuery
	errChecker dummyErrorChecker
	ctx        context.Context
}

// TestRepositoryImplTestSuite runs the test suite.
func TestRepositoryImplTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryImplTestSuite))
}

// SetupTest sets up the test suite.
func (s *RepositoryImplTestSuite) SetupTest() {
	s.preparer = &dummyPreparer{}
	s.query = dummyQuery{}
	s.errChecker = dummyErrorChecker{}
	s.ctx = context.Background()
}

// TestReaderRepo_GetOne_PreparerFails tests the GetOne method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestReaderRepo_GetOne_PreparerFails() {
	repo := NewReaderRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.GetOne(s.ctx, s.preparer, dummyGetterFactory, &GetOptions{})
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestReaderRepo_GetMany_PreparerFails tests the GetMany method with a
// preparer that fails.
func (s *RepositoryImplTestSuite) TestReaderRepo_GetMany_PreparerFails() {
	repo := NewReaderRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.GetMany(s.ctx, s.preparer, dummyGetterFactory, &GetOptions{})
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestReaderRepo_Count_PreparerFails tests the Count method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestReaderRepo_Count_PreparerFails() {
	repo := NewReaderRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.Count(s.ctx, s.preparer, nil, nil, dummyGetterFactory)
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestReaderRepo_Query_PreparerFails tests the Query method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestReaderRepo_Query_PreparerFails() {
	repo := NewReaderRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.Query(s.ctx, s.preparer, "SELECT 1", nil, dummyGetterFactory)
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestMutatorRepo_Insert_PreparerFails tests the Insert method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestMutatorRepo_Insert_PreparerFails() {
	mutator := dummyEntityFn()
	repo := NewMutatorRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.Insert(s.ctx, s.preparer, mutator)
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestMutatorRepo_Update_PreparerFails tests the Update method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestMutatorRepo_Update_PreparerFails() {
	updater := dummyEntityFn()
	repo := NewMutatorRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.Update(s.ctx, s.preparer, updater, nil, nil)
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestMutatorRepo_Delete_PreparerFails tests the Delete method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestMutatorRepo_Delete_PreparerFails() {
	deleter := dummyEntityFn()
	repo := NewMutatorRepo[*DummyEntity](s.query, s.errChecker)
	_, err := repo.Delete(s.ctx, s.preparer, deleter, nil, &DeleteOptions{})
	assert.Error(s.T(), err, "Expected error when preparer fails")
}

// TestCustomRepo_QueryCustom_PreparerNil tests the QueryCustom method with a
// nil preparer.
func (s *RepositoryImplTestSuite) TestCustomRepo_QueryCustom_PreparerNil() {
	repo := NewCustomRepo[*DummyEntity](s.errChecker)
	_, err := repo.QueryCustom(
		s.ctx,
		nil,
		"SELECT 1",
		nil,
		func() *DummyEntity { return dummyEntityFn() },
	)
	assert.Error(s.T(), err, "Expected error when preparer is nil")
	assert.Contains(s.T(), err.Error(), "preparer is nil")
}

// TestRawQueryer_Exec_PreparerFails tests the Exec method with a preparer that
// fails.
func (s *RepositoryImplTestSuite) TestRawQueryer_Exec_PreparerFails() {
	raw := NewRawQueryer()
	_, err := raw.Exec(s.ctx, s.preparer, "DUMMY", nil)
	assert.Error(s.T(), err, "Expected error from Exec")
}

// TestRawQueryer_ExecRaw_PreparerFails tests the ExecRaw method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestRawQueryer_ExecRaw_PreparerFails() {
	raw := NewRawQueryer()
	_, err := raw.ExecRaw(s.ctx, nil, "DUMMY", nil)
	assert.Error(s.T(), err, "Expected error from ExecRaw")
}

// TestRawQueryer_Query_PreparerFails tests the Query method with a preparer
// that fails.
func (s *RepositoryImplTestSuite) TestRawQueryer_Query_PreparerFails() {
	raw := NewRawQueryer()
	_, err := raw.Query(s.ctx, s.preparer, "DUMMY", nil)
	assert.Error(s.T(), err, "Expected error from Query")
}

// TestRawQueryer_QueryRaw_PreparerFails tests the QueryRaw method with a
// preparer that fails.
func (s *RepositoryImplTestSuite) TestRawQueryer_QueryRaw_PreparerFails() {
	raw := NewRawQueryer()
	_, err := raw.QueryRaw(s.ctx, nil, "DUMMY", nil)
	assert.Error(s.T(), err, "Expected error from QueryRaw")
}

// TestTxManager_WithTransaction_BeginFails tests the WithTransaction method
// with a transaction that fails to begin.
func (s *RepositoryImplTestSuite) TestTxManager_WithTransaction_BeginFails() {
	txManager := NewTxManager[*DummyEntity]()
	_, err := txManager.WithTransaction(
		s.ctx,
		dummyConnFn,
		func(ctx context.Context, tx database.Tx) (*DummyEntity, error) {
			return dummyEntityFn(), nil
		},
	)
	assert.Error(s.T(), err, "Expected error when transaction begins fail")
}
