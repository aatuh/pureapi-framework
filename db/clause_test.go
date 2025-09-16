package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ClauseTestSuite is a suite of tests for Clause.
type ClauseTestSuite struct {
	suite.Suite
}

// TestClauseTestSuite runs the test suite.
func TestClauseTestSuite(t *testing.T) {
	suite.Run(t, new(ClauseTestSuite))
}

// TestSelectorCreation verifies that NewSelector creates a proper selector.
func (s *ClauseTestSuite) TestSelectorCreation() {
	selector := NewSelector("name", Equal, "John")
	assert.Equal(s.T(), "name", selector.Column)
	assert.Equal(s.T(), Equal, selector.Predicate)
	assert.Equal(s.T(), "John", selector.Value)
	assert.Empty(s.T(), selector.Table)
}

// TestSelectorWithMethods checks that WithTable, WithColumn,
// WithPredicate, and WithValue return new selectors with the updated
// values, leaving the original selector unchanged.
func (s *ClauseTestSuite) TestSelectorWithMethods() {
	orig := NewSelector("name", Equal, "John")
	withTable := orig.WithTable("users")
	withColumn := orig.WithColumn("username")
	withPredicate := orig.WithPredicate(NotEqual)
	withValue := orig.WithValue("Jane")

	// Original remains unchanged.
	assert.Empty(s.T(), orig.Table)
	assert.Equal(s.T(), "name", orig.Column)
	assert.Equal(s.T(), Equal, orig.Predicate)
	assert.Equal(s.T(), "John", orig.Value)

	// New selectors have the updated fields.
	assert.Equal(s.T(), "users", withTable.Table)
	assert.Equal(s.T(), "username", withColumn.Column)
	assert.Equal(s.T(), NotEqual, withPredicate.Predicate)
	assert.Equal(s.T(), "Jane", withValue.Value)
}

// TestSelectorsAddAndGet verifies Selectors.Add, GetByField and GetByFields.
func (s *ClauseTestSuite) TestSelectorsAddAndGet() {
	// Start with empty selectors.
	var sels Selectors = NewSelectors()
	sels = sels.Add("id", Equal, 100)
	sels = sels.Add("name", Like, "%Doe%")
	sels = sels.Add("name", NotEqual, "John")

	// GetByField returns the first selector with the field.
	selector := sels.GetByField("name")
	assert.NotNil(s.T(), selector)
	assert.Equal(s.T(), "name", selector.Column)

	// GetByField for a non-existent field returns nil.
	assert.Nil(s.T(), sels.GetByField("age"))

	// GetByFields returns all selectors with matching fields.
	multi := sels.GetByFields("name")
	assert.Len(s.T(), multi, 2)

	// Test with multiple fields.
	multi = sels.GetByFields("id", "name")
	assert.Len(s.T(), multi, 3)
}

// TestUpdateCreation verifies NewUpdate and the WithField/WithValue methods.
func (s *ClauseTestSuite) TestUpdateCreation() {
	update := NewUpdate("age", 30)
	assert.Equal(s.T(), "age", update.Field)
	assert.Equal(s.T(), 30, update.Value)

	// Update with new field and value.
	updatedField := update.WithField("user_age")
	updatedValue := update.WithValue(35)
	assert.Equal(s.T(), "user_age", updatedField.Field)
	assert.Equal(s.T(), 30, updatedField.Value)
	assert.Equal(
		s.T(), "age", update.Field,
		"original update should be unchanged",
	)
	assert.Equal(s.T(), 35, updatedValue.Value)
}

// TestUpdatesAdd verifies that Updates.Add appends new update clauses.
func (s *ClauseTestSuite) TestUpdatesAdd() {
	var ups Updates = NewUpdates()
	ups = ups.Add("name", "Alice")
	ups = ups.Add("age", 28)
	assert.Len(s.T(), ups, 2)
	assert.Equal(s.T(), "name", ups[0].Field)
	assert.Equal(s.T(), "Alice", ups[0].Value)
	assert.Equal(s.T(), "age", ups[1].Field)
	assert.Equal(s.T(), 28, ups[1].Value)
}

// TestJoinCreationAndMethods verifies NewJoin and join modification methods.
func (s *ClauseTestSuite) TestJoinCreationAndMethods() {
	left := ColumnSelector{Table: "users", Column: "id"}
	right := ColumnSelector{Table: "orders", Column: "user_id"}
	join := NewJoin(JoinTypeInner, "orders", left, right)
	assert.Equal(s.T(), JoinTypeInner, join.JoinType)
	assert.Equal(s.T(), "orders", join.Table)
	assert.Equal(s.T(), left, join.OnLeft)
	assert.Equal(s.T(), right, join.OnRight)

	// Change join type.
	join2 := join.WithJoinType(JoinTypeLeft)
	assert.Equal(s.T(), JoinTypeLeft, join2.JoinType)
	// Change table name.
	join3 := join.WithTable("purchases")
	assert.Equal(s.T(), "purchases", join3.Table)
	// Change OnLeft.
	newLeft := ColumnSelector{Table: "clients", Column: "client_id"}
	join4 := join.WithOnLeft(newLeft)
	assert.Equal(s.T(), newLeft, join4.OnLeft)
	// Change OnRight.
	newRight := ColumnSelector{Table: "sales", Column: "client_id"}
	join5 := join.WithOnRight(newRight)
	assert.Equal(s.T(), newRight, join5.OnRight)
}
