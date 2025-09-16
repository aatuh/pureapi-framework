package db

import (
	"testing"

	"github.com/aatuh/pureapi-core/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DummyEntity is a test struct that satisfies database.CRUDEntity.
// It uses db tags for mapping.
type DummyEntity struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

// TableName returns the table name for DummyEntity.
func (d *DummyEntity) TableName() string {
	return "dummy"
}

func (d *DummyEntity) InsertedValues() (columns []string, values []any) {
	columns = []string{"id", "name", "age"}
	values = []any{d.ID, d.Name, d.Age}
	return
}

func (d *DummyEntity) ScanRow(row database.Row) error {
	err := row.Scan(&d.ID, &d.Name, &d.Age)
	if err != nil {
		return err
	}
	return nil
}

// dummyEntityFn returns a new DummyEntity.
func dummyEntityFn() *DummyEntity {
	return &DummyEntity{}
}

// dummyOptionEntityFn applies options to a new DummyEntity.
func dummyOptionEntityFn(options ...EntityOption[*DummyEntity]) *DummyEntity {
	entity := &DummyEntity{}
	for _, opt := range options {
		opt(entity)
	}
	return entity
}

// EntityQueryOptionsTestSuite is a test suite for EntityQueryOptions.
type EntityQueryOptionsTestSuite struct {
	suite.Suite
	opts *EntityQueryOptions[*DummyEntity]
}

// TestEntityQueryOptionsTestSuite runs the test suite.
func TestEntityQueryOptionsTestSuite(t *testing.T) {
	suite.Run(t, new(EntityQueryOptionsTestSuite))
}

// SetupTest sets up the test suite.
func (s *EntityQueryOptionsTestSuite) SetupTest() {
	s.opts = NewEntityQueryOptions("dummy", dummyEntityFn, dummyOptionEntityFn)
}

// TestNewEntityQueryOptions verifies the initial state of the
// EntityQueryOptions.
func (s *EntityQueryOptionsTestSuite) TestNewEntityQueryOptions() {
	// Verify initial state.
	assert.Equal(s.T(), "dummy", s.opts.TableName)
	assert.NotNil(s.T(), s.opts.EntityFn)
	assert.NotNil(s.T(), s.opts.OptionEntityFn)
	assert.Empty(s.T(), s.opts.SelectorList, "Selectors should be empty")
	assert.Empty(s.T(), s.opts.UpdateList, "Updates should be empty")
	assert.Empty(s.T(), s.opts.Options, "Options should be empty")
}

// TestAddSelector_Success verifies adding a valid selector.
func (s *EntityQueryOptionsTestSuite) TestAddSelector_Success() {
	// Add a valid selector for field "name".
	s.opts.AddSelector("name", Equal, "Alice")
	selectors := s.opts.Selectors()
	require.Len(s.T(), selectors, 1)
	selector := selectors[0]
	assert.Equal(s.T(), "dummy", selector.Table)
	assert.Equal(s.T(), "name", selector.Column)
	assert.Equal(s.T(), Equal, selector.Predicate)
	assert.Equal(s.T(), "Alice", selector.Value)
}

// TestAddSelector_TypeMismatch verifies that adding a selector with a
// type mismatch panics.
func (s *EntityQueryOptionsTestSuite) TestAddSelector_TypeMismatch() {
	// Field "age" is an int, so passing a string should panic.
	require.Panics(s.T(), func() {
		s.opts.AddSelector("age", GreaterOrEqual, "not-an-int")
	}, "Expected panic when value type mismatches")
}

// TestAddUpdate_Success verifies adding a valid update.
func (s *EntityQueryOptionsTestSuite) TestAddUpdate_Success() {
	// Add a valid update for field "age".
	s.opts.AddUpdate("age", 30)
	updates := s.opts.Updates()
	require.Len(s.T(), updates, 1)
	update := updates[0]
	assert.Equal(s.T(), "age", update.Field)
	assert.Equal(s.T(), 30, update.Value)
}

// TestAddUpdate_TypeMismatch verifies that adding an update with a type
// mismatch panics.
func (s *EntityQueryOptionsTestSuite) TestAddUpdate_TypeMismatch() {
	// Field "age" is an int, so passing a string should panic.
	require.Panics(s.T(), func() {
		s.opts.AddUpdate("age", "thirty")
	}, "Expected panic when update value type mismatches")
}

// TestAddOptionAndEntity_Success verifies that adding an option and
// calling Entity() applies the option.
func (s *EntityQueryOptionsTestSuite) TestAddOptionAndEntity_Success() {
	// Add an option to set the "name" field.
	s.opts.AddOption("name", "Bob")
	// When Entity() is called, the OptionEntityFn should apply the option.
	entity := s.opts.Entity()
	assert.NotNil(s.T(), entity)
	// Since WithOption uses reflection, the "Name" field should be set.
	assert.Equal(s.T(), "Bob", entity.Name)
}

// TestAddOption_InvalidField verifies that adding an option for a
// non-existent field panics.
func (s *EntityQueryOptionsTestSuite) TestAddOption_InvalidField() {
	// Adding an option for a non-existent field should panic when applying it.
	s.opts.AddOption("nonexistent", "value")
	require.Panics(s.T(), func() {
		_ = s.opts.Entity()
	}, "Expected panic for invalid field in WithOption when creating entity")
}

// TestEntityWithoutOptions verifies that Entity() returns the default
func (s *EntityQueryOptionsTestSuite) TestEntityWithoutOptions() {
	// When no options are added, Entity() should return the default entity.
	entity := s.opts.Entity()
	assert.NotNil(s.T(), entity)
	// Default values for DummyEntity fields (zero values).
	assert.Equal(s.T(), 0, entity.ID)
	assert.Equal(s.T(), "", entity.Name)
	assert.Equal(s.T(), 0, entity.Age)
}
