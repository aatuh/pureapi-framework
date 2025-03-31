package apimapper

import (
	"fmt"
	"strings"

	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	"github.com/pureapi/pureapi-framework/db/input"
	"github.com/pureapi/pureapi-framework/defaults"
)

const (
	FieldSelectors = "selectors"
	FieldUpdates   = "updates"
	FieldCount     = "count"
	FieldValue     = "value"
	FieldPredicate = "predicate"
)

func InputHandlerFromAPIFields[Input any](
	apiFields APIFields,
) endpointtypes.InputHandler[Input] {
	return NewMapInputHandler(
		MustMatchStructAPIFields[Input](apiFields),
		defaults.InputConversionRules(),
		defaults.ValidationRules(),
		func() *Input { return new(Input) },
	)
}

// GetAPIFieldToDBColumnMapping builds a mapping from API field names to DB
// fields.
//
// Example:
//
//	apiFields := []APIField{
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
func GetAPIFieldToDBColumnMapping(
	apiFields []APIField,
	tableName string,
) map[string]input.DBField {
	mapping := make(map[string]input.DBField)
	for _, def := range apiFields {
		mapping[def.APIName] = input.DBField{
			Table:  tableName,
			Column: def.DBColumn,
		}
	}
	return mapping
}

// GenericGetAPIFields creates APIFields for a get request.
func GenericGetAPIFields(
	apiFields APIFields,
	predicates map[string]input.Predicates,
	orderable []string,
) APIFields {
	mustMatchPredicates(predicates, apiFields)
	mustMatchFields(orderable, apiFields)
	return APIFields{
		selectorFieldsEntry(apiFields, predicates),
		orderFieldsEntry(apiFields, orderable),
		pageFieldEntry(),
		countFieldEntry(),
	}
}

// // GenericGetOutputAPIFields creates APIFields for a get request.
// func GenericGetOutputAPIFields(
// 	outputKey string, apiFields APIFields,
// ) APIFields {
// 	return APIFields{
// 		{
// 			APIName: outputKey,
// 			Nested:  apiFields,
// 		},
// 		{
// 			APIName: FieldCount,
// 		},
// 	}
// }

// GenericUpdateAPIFields creates APIFields for an update request.
func GenericUpdateAPIFields(
	selectableAPIFields APIFields,
	predicates map[string]input.Predicates,
	updatableAPIFields APIFields,
) APIFields {
	return APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
		updatesFieldEntry(updatableAPIFields),
	}
}

// GenericDeleteAPIFields creates APIFields for a delete request.
func GenericDeleteAPIFields(
	selectableAPIFields APIFields,
	predicates map[string]input.Predicates,
) APIFields {
	mustMatchPredicates(predicates, selectableAPIFields)
	return APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
	}
}

// mustMatchPredicates returns a map of predicates that matches the APIFields.
// It panics if a predicate is not found for a field.
func mustMatchPredicates(
	predicates map[string]input.Predicates,
	apiFields APIFields,
) map[string]input.Predicates {
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

// mustMatchFields returns a slice of fields that matches the APIFields.
// It panics if a field is not found.
func mustMatchFields(fields []string, apiFields APIFields) {
	for _, field := range fields {
		field, err := apiFields.GetAPIField(field)
		if err != nil {
			panic(fmt.Sprintf(
				"mustMatchFields: unknown field %s", field.APIName,
			))
		}
	}
}

// selectorFieldsEntry creates an APIField for selector fields.
func selectorFieldsEntry(
	from APIFields,
	fieldPredicates map[string]input.Predicates,
) APIField {
	return APIField{
		APIName: FieldSelectors,
		Nested:  mustSelectorFields(from, fieldPredicates),
	}
}

// mustSelectorFields creates APIFields for selector fields. It panics if a
// predicate is not found for a field.
func mustSelectorFields(
	from APIFields,
	fieldPredicates map[string]input.Predicates,
) APIFields {
	var fields []APIField
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

// selectorField creates an APIField for a selector field.
func selectorField(
	from APIField, predicates input.Predicates,
) APIField {
	return APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Nested: []APIField{
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

// updatesFieldEntry creates an APIField for updating fields.
func updatesFieldEntry(from APIFields) APIField {
	return APIField{
		APIName: FieldUpdates,
		Nested:  updates(from),
	}
}

// updates creates APIFields for updating fields.
func updates(from APIFields) APIFields {
	var fields []APIField
	for _, field := range from {
		fields = append(fields, update(field))
	}
	return fields

}

// update creates an APIField for updating fields.
func update(from APIField) APIField {
	return APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Validate: from.Validate,
		Type:     from.Type,
	}
}

// orderFieldsEntry creates an APIField for ordering fields.
func orderFieldsEntry(
	from APIFields,
	orderableFields []string,
) APIField {
	return APIField{
		APIName: "orders",
		Nested:  orderFields(from, orderableFields),
	}
}

// orderFields creates APIFields for ordering fields.
func orderFields(from APIFields, orderableFields []string) APIFields {
	fields := []APIField{}
	apiFields, err := from.GetAPIFields(orderableFields)
	if err != nil {
		panic(err)
	}
	for _, field := range apiFields {
		fields = append(fields, APIField{
			APIName: field.APIName,
			Validate: []string{
				"string",
				fmt.Sprintf(
					"oneof=%s %s %s %s",
					input.DirectionAsc,
					input.DirectionDesc,
					input.DirectionAscending,
					input.DirectionDescending,
				),
			},
			Type: "string",
		})
	}
	return fields
}

// pageFieldEntry creates an APIField for pagination.
func pageFieldEntry() APIField {
	return APIField{
		APIName: "page",
		Nested: []APIField{
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

// countFieldEntry creates an APIField for counting fields.
func countFieldEntry() APIField {
	return APIField{
		APIName:  FieldCount,
		Validate: []string{"bool"},
		Type:     "bool",
	}
}
