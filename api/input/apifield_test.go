package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APIFieldTestSuite is a suite of tests for APIField.
type APIFieldTestSuite struct {
	suite.Suite
}

// TestAPIFieldTestSuite runs the suite of tests for APIField.
func TestAPIFieldTestSuite(t *testing.T) {
	suite.Run(t, new(APIFieldTestSuite))
}

// TestChainableMethods tests the chainable methods for APIField.
func (s *APIFieldTestSuite) TestChainableMethods() {
	// Create an empty field.
	field := APIField{}

	// Test WithAPIName.
	f1 := field.WithAPIName("name1")
	assert.Equal(s.T(), "name1", f1.APIName)
	assert.Equal(s.T(), field.APIName, "")

	// Test WithAlias.
	f2 := field.WithAlias("alias1")
	assert.Equal(s.T(), "alias1", f2.Alias)

	// Test WithDBColumn.
	f3 := field.WithDBColumn("col1")
	assert.Equal(s.T(), "col1", f3.DBColumn)

	// Test WithRequired.
	f4 := field.WithRequired(true)
	assert.True(s.T(), f4.Required)

	// Test WithDefault.
	f5 := field.WithDefault(42)
	assert.Equal(s.T(), 42, f5.Default)

	// Test WithSource.
	f6 := field.WithSource("body")
	assert.Equal(s.T(), "body", f6.Source)

	// Test WithValidate.
	rules := []string{"rule1", "rule2"}
	f7 := field.WithValidate(rules)
	assert.Equal(s.T(), rules, f7.Validate)

	// Test WithType.
	f8 := field.WithType("string")
	assert.Equal(s.T(), "string", f8.Type)

	// Test WithNested.
	nested := NewAPIFields(APIField{APIName: "nested1"})
	f9 := field.WithNested(nested)
	assert.Equal(s.T(), nested, f9.Nested)
}

// TestNewAPIFieldsAndWithFields tests the NewAPIFields and WithFields methods.
func (s *APIFieldTestSuite) TestNewAPIFieldsAndWithFields() {
	// Use NewAPIFields to create a slice.
	f1 := APIField{APIName: "f1"}
	f2 := APIField{APIName: "f2"}
	fields := NewAPIFields(f1, f2)
	assert.Len(s.T(), fields, 2)
	assert.Equal(s.T(), "f1", fields[0].APIName)
	assert.Equal(s.T(), "f2", fields[1].APIName)

	// Test WithFields to append additional fields.
	f3 := APIField{APIName: "f3"}
	newFields := fields.WithFields(f3)
	assert.Len(s.T(), newFields, 3)
	assert.Equal(s.T(), "f3", newFields[2].APIName)
}

// TestGetAPIField tests the GetAPIField method.
func (s *APIFieldTestSuite) TestGetAPIField() {
	// Create APIFields with multiple fields.
	f1 := APIField{APIName: "field1", Alias: "a1"}
	f2 := APIField{APIName: "field2", Alias: "a2"}
	fields := NewAPIFields(f1, f2)

	// Look up a field that exists.
	res, err := fields.GetAPIField("field1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "a1", res.Alias)

	// Lookup a field that does not exist.
	_, err = fields.GetAPIField("nonexistent")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "unknown API field")
}

// TestMustGetAPIField tests the MustGetAPIField method.
func (s *APIFieldTestSuite) TestMustGetAPIField() {
	f1 := APIField{APIName: "mustField", DBColumn: "col1"}
	fields := NewAPIFields(f1)

	// MustGetAPIField returns the field when found.
	res := fields.MustGetAPIField("mustField")
	assert.Equal(s.T(), "col1", res.DBColumn)

	// MustGetAPIField should panic when the field is not found.
	assert.Panics(s.T(), func() {
		_ = fields.MustGetAPIField("nonexistent")
	})
}

// TestGetAPIFields tests the GetAPIFields method.
func (s *APIFieldTestSuite) TestGetAPIFields() {
	f1 := APIField{APIName: "f1"}
	f2 := APIField{APIName: "f2"}
	fields := NewAPIFields(f1, f2)

	// Get multiple fields.
	res, err := fields.GetAPIFields([]string{"f1", "f2"})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res, 2)

	// Requesting a non-existent field should return an error.
	_, err = fields.GetAPIFields([]string{"f1", "unknown"})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "unknown API field")
}
