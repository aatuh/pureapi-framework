package inpututil

import (
	"fmt"
	"strings"

	"github.com/aatuh/pureapi-core/endpoint"
	apidb "github.com/aatuh/pureapi-framework/api/db"
	"github.com/aatuh/pureapi-framework/api/input"
)

const (
	FieldSelectors = "selectors"
	FieldUpdates   = "updates"
	FieldCount     = "count"
	FieldValue     = "value"
	FieldPredicate = "predicate"
)

func InputHandlerFromAPIFields[Input any](
	inputFactoryFn func() *Input,
	apiFields input.APIFields,
	conversionRules map[string]func(any) any,
	customRules map[string]func(any) error,
) endpoint.InputHandler[Input] {
	return input.NewMapInputHandler(
		MustMatchStructAPIFields[Input](apiFields),
		conversionRules,
		customRules,
		inputFactoryFn,
	).MustValidateAPIFields()
}

// GetAPIToDBMap builds a mapping from API field names to DB
// fields.
//
// Example:
//
//	apiFields := []input.APIField{
//	    {
//	        APIName:  "name",
//	        DBColumn: "name",
//	    },
//	}
//	tableName := "users"
//
// Output:
//
//	mapping := map[string]DBField{
//	    "name": {
//	        Table:  "users",
//	        Column: "name",
//	    },
//	}
func GetAPIToDBMap(
	apiFields []input.APIField,
	tableName string,
) map[string]apidb.APIToDBField {
	mapping := make(map[string]apidb.APIToDBField)
	for _, def := range apiFields {
		mapping[def.APIName] = apidb.APIToDBField{
			Table:  tableName,
			Column: def.DBColumn,
		}
	}
	return mapping
}

// GenericCreateAPIFields creates input.APIFields for a create request.
//
// Parameters:
//   - wrapperName: The name of the wrapper field.
//   - nested: The nested input.APIFields.
//
// Returns:
//   - input.APIFields: The input.APIFields.
func GenericCreateAPIFields[T any](
	wrapperName string, nested input.APIFields,
) input.APIFields {
	return MustMatchStructAPIFields[T](input.APIFields{
		{
			APIName: wrapperName,
			Nested:  nested,
		},
	})
}

// GenericGetAPIFields creates input.APIFields for a get request.
func GenericGetAPIFields(
	apiFields input.APIFields,
	predicates map[string]apidb.Predicates,
	orderable []string,
) input.APIFields {
	mustMatchPredicates(predicates, apiFields)
	mustMatchFields(orderable, apiFields)
	return input.APIFields{
		selectorFieldsEntry(apiFields, predicates),
		orderFieldsEntry(apiFields, orderable),
		pageFieldEntry(),
		countFieldEntry(),
	}
}

// // GenericGetOutputAPIFields creates input.APIFields for a get request.
// func GenericGetOutputAPIFields(
// 	outputKey string, apiFields input.APIFields,
// ) input.APIFields {
// 	return input.APIFields{
// 		{
// 			APIName: outputKey,
// 			Nested:  apiFields,
// 		},
// 		{
// 			APIName: FieldCount,
// 		},
// 	}
// }

// GenericUpdateAPIFields creates input.APIFields for an update request.
func GenericUpdateAPIFields(
	selectableAPIFields input.APIFields,
	predicates map[string]apidb.Predicates,
	updatableAPIFields input.APIFields,
) input.APIFields {
	return input.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
		updatesFieldEntry(updatableAPIFields),
	}
}

// GenericDeleteAPIFields creates input.APIFields for a delete request.
func GenericDeleteAPIFields(
	selectableAPIFields input.APIFields,
	predicates map[string]apidb.Predicates,
) input.APIFields {
	mustMatchPredicates(predicates, selectableAPIFields)
	return input.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
	}
}

// mustMatchPredicates returns a map of predicates that matches the input.APIFields.
// It panics if a predicate is not found for a field.
func mustMatchPredicates(
	predicates map[string]apidb.Predicates,
	apiFields input.APIFields,
) map[string]apidb.Predicates {
	for _, field := range apiFields {
		if predicates[field.APIName] == nil {
			panic(fmt.Sprintf(
				"mustMatchPredicates: predicate not found for field: %s",
				field.APIName,
			))
		}
	}
	return predicates
}

// mustMatchFields returns a slice of fields that matches the input.APIFields.
// It panics if a field is not found.
func mustMatchFields(fields []string, apiFields input.APIFields) {
	for _, field := range fields {
		field, err := apiFields.GetAPIField(field)
		if err != nil {
			panic(fmt.Sprintf(
				"mustMatchFields: unknown field %s", field.APIName,
			))
		}
	}
}

// selectorFieldsEntry creates an input.APIField for selector fields.
func selectorFieldsEntry(
	from input.APIFields,
	fieldPredicates map[string]apidb.Predicates,
) input.APIField {
	return input.APIField{
		APIName: FieldSelectors,
		Nested:  mustSelectorFields(from, fieldPredicates),
	}
}

// mustSelectorFields creates input.APIFields for selector fields. It panics if a
// predicate is not found for a field.
func mustSelectorFields(
	from input.APIFields,
	fieldPredicates map[string]apidb.Predicates,
) input.APIFields {
	var fields []input.APIField
	for _, field := range from {
		predicates, ok := fieldPredicates[field.APIName]
		if !ok {
			panic(fmt.Sprintf(
				"mustSelectorFields: no predicates found for field %q",
				field.APIName,
			))
		}
		fields = append(fields, selectorField(field, predicates))
	}
	return fields
}

// selectorField creates an input.APIField for a selector field.
func selectorField(
	from input.APIField, predicates apidb.Predicates,
) input.APIField {
	return input.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Nested: []input.APIField{
			{
				APIName:  FieldValue,
				Validate: from.Validate,
				Type:     from.Type,
				Required: true,
			},
			{
				APIName: FieldPredicate,
				Validate: []string{
					"string",
					fmt.Sprintf(
						"oneof=%s", strings.Join(predicates.StrSlice(), " "),
					),
				},
				Type:     "string",
				Required: true,
			},
		},
	}
}

// updatesFieldEntry creates an input.APIField for updating fields.
func updatesFieldEntry(from input.APIFields) input.APIField {
	return input.APIField{
		APIName: FieldUpdates,
		Nested:  updates(from),
	}
}

// updates creates input.APIFields for updating fields.
func updates(from input.APIFields) input.APIFields {
	var fields []input.APIField
	for _, field := range from {
		fields = append(fields, update(field))
	}
	return fields

}

// update creates an input.APIField for updating fields.
func update(from input.APIField) input.APIField {
	return input.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Validate: from.Validate,
		Type:     from.Type,
	}
}

// orderFieldsEntry creates an input.APIField for ordering fields.
func orderFieldsEntry(
	from input.APIFields,
	orderableFields []string,
) input.APIField {
	return input.APIField{
		APIName: "orders",
		Nested:  orderFields(from, orderableFields),
	}
}

// orderFields creates input.APIFields for ordering fields.
func orderFields(from input.APIFields, orderableFields []string) input.APIFields {
	fields := []input.APIField{}
	apiFields, err := from.GetAPIFields(orderableFields)
	if err != nil {
		panic(err)
	}
	for _, field := range apiFields {
		fields = append(fields, input.APIField{
			APIName: field.APIName,
			Validate: []string{
				"string",
				fmt.Sprintf(
					"oneof=%s %s %s %s",
					apidb.DirectionAsc,
					apidb.DirectionDesc,
					apidb.DirectionAscending,
					apidb.DirectionDescending,
				),
			},
			Type: "string",
		})
	}
	return fields
}

// pageFieldEntry creates an input.APIField for pagination.
func pageFieldEntry() input.APIField {
	return input.APIField{
		APIName: "page",
		Nested: []input.APIField{
			{
				APIName:  "offset",
				Validate: []string{"int64", "min=0"},
				Type:     "int64",
			},
			{
				APIName:  "limit",
				Validate: []string{"int64", "min=1", "max=1000"},
				Default:  int64(1),
				Type:     "int64",
			},
		},
	}
}

// countFieldEntry creates an input.APIField for counting fields.
func countFieldEntry() input.APIField {
	return input.APIField{
		APIName:  FieldCount,
		Validate: []string{"bool"},
		Type:     "bool",
	}
}
