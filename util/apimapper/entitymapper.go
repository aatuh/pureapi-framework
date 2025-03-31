package apimapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pureapi/pureapi-framework/util/apimapper/types"
)

// MustMapInputToEntity is a generic helper function that maps fields from an
// input struct to an entity struct using the provided APIToDBFields.
// It panics if the mapping fails.
//
// Parameters:
//   - input: The input struct to map.
//   - entity: The entity struct to map to.
//   - apiToDBFields: The mapping between API fields and database columns.
//
// Returns:
//   - *Entity: The mapped entity.
func MustMapInputToEntity[Entity any](
	input any, entity *Entity, apiToDBFields types.APIToDBFields,
) *Entity {
	entity, err := MapInputToEntity(input, entity, apiToDBFields)
	if err != nil {
		panic(fmt.Errorf("MapInputToEntity: %w", err))
	}
	return entity
}

// MapInputToEntity is a generic helper function that maps fields from an input
// struct to an entity struct using the provided APIToDBFields.
// It expects the input to be a struct (or pointer to struct) and the entity to
// be a pointer to a struct. It uses the JSON tag of the input and the "db" tag
// of the entity.
//
// Parameters:
//   - input: The input struct to map.
//   - entity: The entity struct to map to.
//   - apiToDBFields: The mapping between API fields and database columns.
//
// Returns:
//   - *Entity: The mapped entity.
//   - error: An error if the mapping fails.
func MapInputToEntity[Entity any](
	input any, entity *Entity, apiToDBFields types.APIToDBFields,
) (*Entity, error) {
	inputVal := reflect.ValueOf(input)
	if inputVal.Kind() == reflect.Ptr {
		inputVal = inputVal.Elem()
	}
	if inputVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("MapInputToEntity: input is not a struct")
	}

	entityVal := reflect.ValueOf(entity)
	if entityVal.Kind() != reflect.Ptr ||
		entityVal.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"MapInputToEntity: entity must be a pointer to a struct",
		)
	}
	entityVal = entityVal.Elem()
	entityType := entityVal.Type()
	inputType := inputVal.Type()

	for apiField, DBColumn := range apiToDBFields {
		var inputFieldVal reflect.Value
		found := false
		// Look for a matching field in the input by checking its JSON tag or
		// field name.
		for i := 0; i < inputType.NumField(); i++ {
			field := inputType.Field(i)
			jsonTag := field.Tag.Get("json")
			// Compare the tag (ignoring extra options) or the field name.
			if jsonTag == apiField || field.Name == apiField {
				inputFieldVal = inputVal.Field(i)
				found = true
				break
			}
		}
		if !found {
			// Skip if no matching input field.
			continue
		}

		// Find the corresponding field in the entity by matching the "db" tag.
		for j := 0; j < entityType.NumField(); j++ {
			eField := entityType.Field(j)
			dbTag := eField.Tag.Get("db")
			if dbTag == DBColumn.Column {
				if entityVal.Field(j).CanSet() {
					// Assume the types are compatible.
					entityVal.Field(j).Set(inputFieldVal)
				}
				break
			}
		}
	}
	return entity, nil
}

// MapEntityToOutput maps fields from an entity struct to an output struct
// using the provided APIToDBFields. It panics if the mapping fails.
//
// Parameters:
//   - entity: The entity to map.
//   - output: The output struct to map to.
//   - apiToDBFields: The mapping between API fields and database columns.
//
// Returns:
//   - output: The mapped output struct.
func MustMapEntityToOutput(
	entity any, output any, apiToDBFields types.APIToDBFields,
) any {
	err := MapEntityToOutput(entity, output, apiToDBFields)
	if err != nil {
		panic(fmt.Errorf("MapEntityToOutput: %w", err))
	}
	return output
}

// MapEntityToOutput maps fields from an entity struct to an output struct
// using the provided APIToDBFields. It expects the entity to be a struct
// (or pointer to struct) and the output to be a pointer to a struct.
// The mapping is done by matching the entity’s "db" tag with the
// colum in the APIToDBFields value, and the output’s JSON tag with the
// APIToDBFields key.
//
// Parameters:
//   - entity: The entity to map.
//   - output: The output struct to map to.
//   - apiToDBFields: The mapping between API fields and database columns.
//
// Returns:
//   - error: An error if the mapping fails.
func MapEntityToOutput(
	entity any, output any, apiToDBFields types.APIToDBFields,
) error {
	entityVal := reflect.ValueOf(entity)
	if entityVal.Kind() == reflect.Ptr {
		entityVal = entityVal.Elem()
	}
	if entityVal.Kind() != reflect.Struct {
		return fmt.Errorf("MapEntityToOutput: entity is not a struct")
	}

	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr ||
		outputVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf(
			"MapEntityToOutput: output must be a pointer to a struct",
		)
	}
	outputVal = outputVal.Elem()
	outputType := outputVal.Type()

	// For each APIField, attempt to find the corresponding field in the output
	// struct.
	for apiField, DBColumn := range apiToDBFields {
		// Look for a field in the output struct whose JSON tag (or field name)
		// equals the APIField.APIName.
		for i := 0; i < outputType.NumField(); i++ {
			outField := outputType.Field(i)
			jsonTag := outField.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = outField.Name
			}
			if jsonTag == apiField || outField.Name == apiField {
				// Found a matching output field; now find the corresponding
				// entity field by matching the "db" tag.
				entityFieldVal, err := findFieldValueByDBTag(
					entityVal, DBColumn.Column,
				)
				if err != nil {
					return fmt.Errorf("MapEntityToOutput: %w", err)
				}
				if outputVal.Field(i).CanSet() {
					outputVal.Field(i).Set(entityFieldVal)
				}
				break
			}
		}
	}
	return nil
}

// findFieldValueByDBTag searches for a struct field in v with a "db" tag
// matching tag.
func findFieldValueByDBTag(v reflect.Value, tag string) (reflect.Value, error) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == tag {
			return v.Field(i), nil
		}
	}
	return reflect.Value{}, fmt.Errorf(
		"findFieldValueByDBTag: no field with db tag %q found", tag,
	)
}

// MustMatchStructAPIFields is a generic function that returns a slice of
// APIFields for any struct type. For each field, if an APIField with a matching
// JSON name exists in allFields, that APIField is used. If the struct field has
// a "validate" tag, its rules override any validations from the APIField.
// Additionally, if a field is tagged with `required:"true"`, then it must
// be found in allFields or else the function panics. It will try to find the
// "required" and "type" tags and these will also override APIField settings.
//
//	Parameters:
//		- Struct: The struct type.
//		- allFields: The APIFields to match against.
//
//	Returns:
//		- APIFields: The matched APIFields.
func MustMatchStructAPIFields[Struct any](allFields APIFields) APIFields {
	apiFields := MatchStructAPIFIelds(
		reflect.TypeOf((*Struct)(nil)).Elem(), allFields, 0,
	)
	err := validateAPIFieldTypes[Struct](apiFields)
	if err != nil {
		panic(fmt.Errorf("MustMatchStructAPIFields: %w", err))
	}
	return apiFields
}

// MatchStructAPIFIelds processes the given type recursively.
func MatchStructAPIFIelds(
	t reflect.Type, allFields APIFields, depth int,
) APIFields {
	// Check if t is empty an type.
	if t.Kind() == reflect.Interface {
		panic("structToAPIFields: type is a null interface")
	}
	// Check if t is not a struct type.
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf(
			"structToAPIFields: type %q is not a struct", t.Name(),
		))
	}

	depth++
	if depth > 100 {
		panic(fmt.Sprintf("structToAPIFields: type %q is too deep", t.Name()))
	}

	// Build a lookup map from APIName to APIField.
	fieldsByName := make(map[string]APIField)
	fieldsByAlias := make(map[string]APIField)
	for _, f := range allFields {
		if f.Alias != "" {
			// Check if the alias is already in the map.
			if _, ok := fieldsByAlias[f.Alias]; ok {
				panic(fmt.Sprintf(
					"structToAPIFields: type %s has duplicate alias %q",
					t.Name(),
					f.Alias,
				))
			}
			fieldsByAlias[f.Alias] = f
		} else {
			// Check if the APIName is already in the map.
			if _, ok := fieldsByName[f.APIName]; ok {
				panic(fmt.Sprintf(
					"structToAPIFields: type %s has duplicate API name %q",
					t.Name(),
					f.APIName,
				))
			}
			fieldsByName[f.APIName] = f
		}
	}

	var inputFields APIFields
	// Iterate over each field of the struct type.
	for i := range t.NumField() {
		field := t.Field(i)
		// Extract the alias tag and use it if present.
		// Alias fields are useful when there are multiple fields with the same
		// JSON name.
		jsonTag := field.Tag.Get("alias")
		useAlias := false
		if jsonTag != "" {
			useAlias = true
		} else {
			// Extract the JSON tag.
			jsonTag = field.Tag.Get("json")
			if jsonTag == "" {
				// Skip fields without JSON tags.
				continue
			}
		}
		// Split tag options (e.g. "my_field,omitempty").
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		// Get the "validate" tag.
		validateTag := field.Tag.Get("validate")
		var validateRules []string
		if validateTag != "" {
			validateRules = strings.Split(validateTag, ",")
		}

		// Get the "required" tag.
		requiredTag := field.Tag.Get("required")
		var required bool
		if requiredTag != "" {
			parsed, err := strconv.ParseBool(strings.TrimSpace(requiredTag))
			if err != nil {
				panic(fmt.Sprintf(
					"structToAPIFields: invalid value for required tag on field %s: %v",
					fieldName,
					err),
				)
			}
			required = parsed
		}

		// Get the "ext" tag
		extTag := field.Tag.Get("ext")

		// Lookup the APIField.
		var apiField APIField
		var exists bool
		if useAlias {
			apiField, exists = fieldsByAlias[fieldName]
			if !exists {
				panic(fmt.Sprintf(
					"structToAPIFields: alias field %s not found in fieldsByAlias",
					fieldName,
				))
			}
		} else {
			apiField, exists = fieldsByName[fieldName]
		}
		if exists {
			// Override validations if the struct provides them.
			if len(validateRules) > 0 {
				apiField.Validate = validateRules
			}
		} else {
			// Field not found in allFields.
			if extTag == "true" {
				panic(fmt.Sprintf(
					"structToAPIFields: APIField not found for external field: %s",
					fieldName,
				))
			}
			apiField = APIField{
				APIName:  fieldName,
				Validate: validateRules,
			}
		}

		// Override the APIField.Required if the "required" tag is present.
		if requiredTag != "" {
			apiField.Required = required
		}

		// Override the APIField.Type if the "type" tag is present.
		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			apiField.Type = typeTag
		} else if apiField.Type == "" {
			// Take type from the struct field and if it is a pointer type,
			// dereference it.
			if field.Type.Kind() == reflect.Ptr {
				field.Type = field.Type.Elem()
				apiField.Type = field.Type.String()
			} else {
				apiField.Type = field.Type.String()
			}
		}

		// Process potential nested input.
		fieldType := field.Type
		// Dereference pointer input.
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if the field is a slice.
		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			// Dereference element pointer if necessary.
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				// Recursively process the slice element type.
				apiField.Nested = MatchStructAPIFIelds(
					elemType, allFields, depth,
				)
			}
		} else if fieldType.Kind() == reflect.Struct {
			// Recursively process the nested struct.
			apiField.Nested = MatchStructAPIFIelds(fieldType, allFields, depth)
		}

		inputFields = append(inputFields, apiField)
	}
	return inputFields
}

// getJSONTag extracts the JSON key from a struct field’s tag.
func getJSONTag(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name // fallback to field name if no json tag exists
	}
	parts := strings.Split(tag, ",")
	return parts[0]
}

// findFieldByJSONTag returns the first field in t whose JSON tag matche
//
//	jsonName.
func findFieldByJSONTag(t reflect.Type, jsonName string) *reflect.StructField {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if getJSONTag(f) == jsonName {
			return &f
		}
	}
	return nil
}

// ValidateAPIFields is a generic function that validates any struct type
// against a slice of APIField definitions using a custom type map.
func validateAPIFieldTypes[Struct any](apiFields []APIField) error {
	var t Struct
	rt := reflect.TypeOf(t)
	// If is a pointer, use its element type.
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("type %s is not a struct", rt.Name())
	}
	return validateAPIFieldTypesForType(rt, apiFields)
}

// validateAPIFieldTypesForType recursively checks that each APIField in
// apiFields matches the corresponding struct field in t.
func validateAPIFieldTypesForType(
	t reflect.Type, apiFields []APIField,
) error {
	// Ensure that t is a struct
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Slice {
		return fmt.Errorf(
			"validateAPIFieldTypesForType: type %s is not a struct or slice",
			t.Name(),
		)
	}

	for _, apiField := range apiFields {
		underlyingType := t
		if t.Kind() == reflect.Slice {
			underlyingType = t.Elem()
		}
		// Find the matching struct field by comparing json tags.
		field := findFieldByJSONTag(underlyingType, apiField.APIName)
		if field == nil {
			// If the struct does not have a matching field, skip validation.
			continue
		}

		// Get the field type and dereference pointer input.
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		// If there are nested APIFields, expect the field to be a struct.
		if len(apiField.Nested) > 0 {
			if ft.Kind() != reflect.Struct && ft.Kind() != reflect.Slice {
				return fmt.Errorf(
					"validateAPIFieldTypesForType: field %s is expected to be a struct or slice for nested API fields, got %s",
					field.Name,
					ft.Kind().String(),
				)
			}
			if err := validateAPIFieldTypesForType(ft, apiField.Nested); err != nil {
				return fmt.Errorf(
					"validateAPIFieldTypesForType: nested type validation error on field %s: %w",
					field.Name,
					err,
				)
			}
			// Skip further type-checking for nested fields.
			continue
		}
	}
	return nil
}
