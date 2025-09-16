package input

import "fmt"

// APIField holds the core mapping information.
type APIField struct {
	APIName  string
	Alias    string
	DBColumn string
	Required bool
	Default  any
	Source   string
	Validate []string
	Nested   APIFields
	Type     string
}

// WithAPIName returns a copy of APIField with APIName set.
//
// Parameters:
//   - apiName: The API name.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithAPIName(apiName string) APIField {
	field.APIName = apiName
	return field
}

// WithAlias returns a copy of APIField with Alias set.
//
// Parameters:
//   - alias: The alias.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithAlias(alias string) APIField {
	field.Alias = alias
	return field
}

// WithDBColumn returns a copy of APIField with DBColumn set.
//
// Parameters:
//   - dbColumn: The DB column.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithDBColumn(dbColumn string) APIField {
	field.DBColumn = dbColumn
	return field
}

// WithRequired returns a copy of APIField with Required set.
//
// Parameters:
//   - required: The required flag.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithRequired(required bool) APIField {
	field.Required = required
	return field
}

// WithDefault returns a copy of APIField with Default set.
//
// Parameters:
//   - defaultValue: The default value.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithDefault(defaultValue any) APIField {
	field.Default = defaultValue
	return field
}

// WithSource returns a copy of APIField with Source set.
//
// Parameters:
//   - source: The source.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithSource(source string) APIField {
	field.Source = source
	return field
}

// WithValidate returns a copy of APIField with Validate set.
//
// Parameters:
//   - validate: The validation rules.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithValidate(validate []string) APIField {
	field.Validate = validate
	return field
}

// Nested returns a copy of APIField with Nested set.
//
// Parameters:
//   - nested: The nested APIFields.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithNested(nested APIFields) APIField {
	field.Nested = nested
	return field
}

// WithType returns a copy of APIField with Type set.
//
// Parameters:
//   - typ: The type.
//
// Returns:
//   - APIField: The new APIField.
func (field APIField) WithType(typ string) APIField {
	field.Type = typ
	return field
}

// APIFields is a slice of APIField.
type APIFields []APIField

// NewAPIFields returns a new APIFields.
//
// Parameters:
//   - fields: The fields.
//
// Returns:
//   - APIFields: The new APIFields.
func NewAPIFields(fields ...APIField) APIFields {
	return fields
}

// WithFields returns a copy of APIFields with fields appended.
//
// Parameters:
//   - fields: The fields to append.
//
// Returns:
//   - APIFields: The new APIFields.
func (a APIFields) WithFields(fields ...APIField) APIFields {
	return append(a, fields...)
}

// GetAPIField looks up a single field definition by its API name.
//
// Parameters:
//   - field: The field name.
//
// Returns:
//   - APIField: The field definition.
//   - error: An error if the field is not found.
func (a APIFields) GetAPIField(field string) (APIField, error) {
	for _, def := range a {
		if def.APIName == field {
			return def, nil
		}
	}
	return APIField{}, fmt.Errorf(
		"MustGetAPIField: unknown API field %q", field,
	)
}

// MustGetAPIField looks up a single field definition by its API name.
// It panics if the field is not found.
//
// Parameters:
//   - field: The field name.
//
// Returns:
//   - APIField: The field definition.
func (a APIFields) MustGetAPIField(field string) APIField {
	def, err := a.GetAPIField(field)
	if err != nil {
		panic(err)
	}
	return def
}

// GetAPIFields looks up multiple field definitions by their API names.
//
// Parameters:
//   - fields: The field names.
//
// Returns:
//   - APIFields: The field definitions.
//   - error: An error if a field is not found.
func (a APIFields) GetAPIFields(fields []string) (APIFields, error) {
	var defs APIFields
	for _, field := range fields {
		def, err := a.GetAPIField(field)
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}
	return defs, nil
}
