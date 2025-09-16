package db_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	apidb "github.com/aatuh/pureapi-framework/api/db"
	coredb "github.com/aatuh/pureapi-framework/db"
)

// InputTestSuite is a suite of tests for Input.
type InputTestSuite struct {
	suite.Suite
	apiToDBFieldMap map[string]apidb.APIToDBField
}

// TestInputTestSuite runs the test suite.
func TestInputTestSuite(t *testing.T) {
	suite.Run(t, new(InputTestSuite))
}

// SetupTest sets up the test suite.
func (s *InputTestSuite) SetupTest() {
	s.apiToDBFieldMap = map[string]apidb.APIToDBField{
		"name": {Table: "users", Column: "name"},
		"age":  {Table: "users", Column: "age"},
	}
}

// Test Predicate.String() returns the underlying string.
func (s *InputTestSuite) TestPredicateString() {
	p := apidb.Equal // "="
	assert.Equal(s.T(), "=", p.String())
}

// Test Predicates.String() returns a comma-separated string.
func (s *InputTestSuite) TestPredicatesString() {
	p := apidb.Predicates{apidb.Equal, apidb.Greater, apidb.Less}
	expected := strings.Join([]string{
		apidb.Equal.String(), apidb.Greater.String(), apidb.Less.String(),
	}, ",")
	assert.Equal(s.T(), expected, p.String())
}

// Test Predicates.StrSlice() returns a slice of predicate strings.
func (s *InputTestSuite) TestPredicatesStrSlice() {
	p := apidb.Predicates{apidb.Equal, apidb.Greater, apidb.Less}
	expected := []string{
		apidb.Equal.String(), apidb.Greater.String(), apidb.Less.String(),
	}
	assert.Equal(s.T(), expected, p.StrSlice())
}

// Test Page.ToDBPage converts the API Page to a core DB Page.
func (s *InputTestSuite) TestPageToDBPage() {
	// Create an API Page.
	inputPage := &apidb.Page{
		Offset: 5,
		Limit:  50,
	}
	// Convert to a core DB Page.
	converted := inputPage.ToDBPage()
	assert.Equal(s.T(), inputPage.Offset, converted.Offset)
	assert.Equal(s.T(), inputPage.Limit, converted.Limit)
}

// Test Direction.String() returns the underlying string.
func (s *InputTestSuite) TestDirectionString() {
	d := apidb.DirectionAsc
	assert.Equal(s.T(), "asc", d.String())
}

// TestTranslateToDBOrders tests the TranslateToDBOrders function.
func (s *InputTestSuite) TestTranslateToDBOrders() {
	tests := []struct {
		name        string
		orders      apidb.Orders
		mapping     map[string]apidb.APIToDBField
		expected    []coredb.Order
		expectError bool
		errContains string
	}{
		{
			name:   "valid order ascending",
			orders: apidb.Orders{"name": apidb.DirectionAsc},
			mapping: map[string]apidb.APIToDBField{
				"name": {Table: "users", Column: "name"},
			},
			expected: []coredb.Order{{
				Table:     "users",
				Field:     "name",
				Direction: coredb.OrderAsc,
			}},
		},
		{
			name:   "valid order descending with different casing",
			orders: apidb.Orders{"age": apidb.DirectionDescending},
			mapping: map[string]apidb.APIToDBField{
				"age": {Table: "users", Column: "age"},
			},
			expected: []coredb.Order{{
				Table:     "users",
				Field:     "age",
				Direction: coredb.OrderDesc,
			}},
		},
		{
			name:        "invalid field (missing mapping)",
			orders:      apidb.Orders{"invalid": apidb.DirectionAsc},
			mapping:     map[string]apidb.APIToDBField{},
			expectError: true,
			errContains: "cannot translate field: invalid",
		},
		{
			name:   "empty column in mapping",
			orders: apidb.Orders{"name": apidb.DirectionAsc},
			mapping: map[string]apidb.APIToDBField{
				"name": {Table: "users", Column: ""},
			},
			expectError: true,
			errContains: "cannot translate field: name",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			orders, err := tc.orders.TranslateToDBOrders(tc.mapping)
			if tc.expectError {
				require.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tc.errContains)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expected, orders)
			}
		})
	}
}

// TestSelectorString tests Selector.String() and NewSelector.
func (s *InputTestSuite) TestSelectorString() {
	sel := apidb.NewSelector(apidb.Equal, 100)
	expected := fmt.Sprintf("%s %v", apidb.Equal, 100)
	assert.Equal(s.T(), expected, sel.String())
}

// TestAPISelectorsToDBSelectors tests AddSelector and ToDBSelectors.
func (s *InputTestSuite) TestAPISelectorsToDBSelectors() {
	tests := []struct {
		name        string
		selectors   apidb.APISelectors
		mapping     map[string]apidb.APIToDBField
		expected    []coredb.Selector
		expectError bool
		errContains string
	}{
		{
			name: "valid selector",
			selectors: apidb.APISelectors{
				"name": {Predicate: apidb.Equal, Value: "john"},
			},
			mapping: map[string]apidb.APIToDBField{
				"name": {Table: "users", Column: "name"},
			},
			expected: []coredb.Selector{{
				Table:     "users",
				Column:    "name",
				Predicate: coredb.Equal,
				Value:     "john",
			}},
		},
		{
			name: "invalid predicate",
			selectors: apidb.APISelectors{
				"name": {Predicate: "invalid", Value: "john"},
			},
			mapping: map[string]apidb.APIToDBField{
				"name": {Table: "users", Column: "name"},
			},
			expectError: true,
			errContains: "cannot translate predicate: invalid",
		},
		{
			name: "invalid field (missing mapping)",
			selectors: apidb.APISelectors{
				"unknown": {Predicate: apidb.Equal, Value: "john"},
			},
			mapping:     map[string]apidb.APIToDBField{},
			expectError: true,
			errContains: "cannot translate field: unknown",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			sels, err := tc.selectors.ToDBSelectors(tc.mapping)
			if tc.expectError {
				require.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tc.errContains)
			} else {
				require.NoError(s.T(), err)
				// Compare without assuming order.
				assert.Len(s.T(), sels, len(tc.expected))
				for _, exp := range tc.expected {
					found := false
					for _, sel := range sels {
						if sel.Table == exp.Table &&
							sel.Column == exp.Column &&
							sel.Predicate == exp.Predicate &&
							fmt.Sprintf("%v", sel.Value) ==
								fmt.Sprintf("%v", exp.Value) {
							found = true
							break
						}
					}
					assert.True(s.T(), found,
						"expected selector %+v not found",
						exp,
					)
				}
			}
		})
	}
}

// TestAPIUpdatesToDBUpdates tests APIUpdates ToDBUpdates.
func (s *InputTestSuite) TestAPIUpdatesToDBUpdates() {
	tests := []struct {
		name        string
		updates     apidb.APIUpdates
		mapping     map[string]apidb.APIToDBField
		expected    []coredb.Update
		expectError bool
		errContains string
	}{
		{
			name:    "valid update",
			updates: apidb.APIUpdates{"name": "john"},
			mapping: map[string]apidb.APIToDBField{
				"name": {Table: "users", Column: "name"},
			},
			expected: []coredb.Update{{
				Field: "name",
				Value: "john",
			}},
		},
		{
			name:        "invalid field",
			updates:     apidb.APIUpdates{"unknown": "value"},
			mapping:     map[string]apidb.APIToDBField{},
			expectError: true,
			errContains: "cannot translate field: unknown",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			upds, err := tc.updates.ToDBUpdates(tc.mapping)
			if tc.expectError {
				require.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tc.errContains)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expected, upds)
			}
		})
	}
}

// TestAPISelectorsAddSelector tests AddSelector adds a selector.
func (s *InputTestSuite) TestAPISelectorsAddSelector() {
	selectors := apidb.APISelectors{}
	selectors = selectors.AddSelector("name", apidb.Equal, "john")
	sel, ok := selectors["name"]
	require.True(s.T(), ok)
	assert.Equal(s.T(), apidb.Equal, sel.Predicate)
	assert.Equal(s.T(), "john", sel.Value)
}
