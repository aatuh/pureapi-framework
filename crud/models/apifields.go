package models

import (
	"fmt"
	"strings"

	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	"github.com/pureapi/pureapi-framework/api"
	"github.com/pureapi/pureapi-framework/crud/util"
	"github.com/pureapi/pureapi-framework/dbinput"
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
	apiFields api.APIFields,
) endpointtypes.InputHandler[Input] {
	return api.NewMapInputHandler(
		util.MustMatchStructAPIFields[Input](apiFields),
		defaults.DefaultConversionRules(),
		defaults.DefaultCustomRules(),
		func() *Input { return new(Input) },
	)
}

// GetAPIFieldToDBColumnMapping builds a mapping from API field names to DB
// fields.
//
// Example:
//
//	apiFields := []api.APIField{
//	    {
//	        APIName:  "name",
//	        DBColumn: "name",
//	    },
//	}
//	tableName := "users"
//
// Output:
//
//	mapping := map[string]api.DBField{
//	    "name": {
//	        Table:  "users",
//	        Column: "name",
//	    },
//	}
func GetAPIFieldToDBColumnMapping(
	apiFields []api.APIField,
	tableName string,
) map[string]dbinput.DBField {
	mapping := make(map[string]dbinput.DBField)
	for _, def := range apiFields {
		mapping[def.APIName] = dbinput.DBField{
			Table:  tableName,
			Column: def.DBColumn,
		}
	}
	return mapping
}

// GenericGetAPIFields creates api.APIFields for a get request.
func GenericGetAPIFields(
	apiFields api.APIFields,
	predicates map[string]dbinput.Predicates,
	orderable []string,
) api.APIFields {
	mustMatchPredicates(predicates, apiFields)
	mustMatchFields(orderable, apiFields)
	return api.APIFields{
		selectorFieldsEntry(apiFields, predicates),
		orderFieldsEntry(apiFields, orderable),
		pageFieldEntry(),
		countFieldEntry(),
	}
}

// // GenericGetOutputAPIFields creates api.APIFields for a get request.
// func GenericGetOutputAPIFields(
// 	outputKey string, apiFields api.APIFields,
// ) api.APIFields {
// 	return api.APIFields{
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
	selectableAPIFields api.APIFields,
	predicates map[string]dbinput.Predicates,
	updatableAPIFields api.APIFields,
) api.APIFields {
	return api.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
		updatesFieldEntry(updatableAPIFields),
	}
}

// GenericDeleteAPIFields creates APIFields for a delete request.
func GenericDeleteAPIFields(
	selectableAPIFields api.APIFields,
	predicates map[string]dbinput.Predicates,
) api.APIFields {
	mustMatchPredicates(predicates, selectableAPIFields)
	return api.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
	}
}

// mustMatchPredicates returns a map of predicates that matches the APIFields.
// It panics if a predicate is not found for a field.
func mustMatchPredicates(
	predicates map[string]dbinput.Predicates,
	apiFields api.APIFields,
) map[string]dbinput.Predicates {
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
func mustMatchFields(fields []string, apiFields api.APIFields) {
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
	from api.APIFields,
	fieldPredicates map[string]dbinput.Predicates,
) api.APIField {
	return api.APIField{
		APIName: FieldSelectors,
		Nested:  mustSelectorFields(from, fieldPredicates),
	}
}

// mustSelectorFields creates APIFields for selector fields. It panics if a
// predicate is not found for a field.
func mustSelectorFields(
	from api.APIFields,
	fieldPredicates map[string]dbinput.Predicates,
) api.APIFields {
	var fields []api.APIField
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
	from api.APIField, predicates dbinput.Predicates,
) api.APIField {
	return api.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Nested: []api.APIField{
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
func updatesFieldEntry(from api.APIFields) api.APIField {
	return api.APIField{
		APIName: FieldUpdates,
		Nested:  updates(from),
	}
}

// updates creates APIFields for updating fields.
func updates(from api.APIFields) api.APIFields {
	var fields []api.APIField
	for _, field := range from {
		fields = append(fields, update(field))
	}
	return fields

}

// update creates an APIField for updating fields.
func update(from api.APIField) api.APIField {
	return api.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Validate: from.Validate,
		Type:     from.Type,
	}
}

// orderFieldsEntry creates an APIField for ordering fields.
func orderFieldsEntry(
	from api.APIFields,
	orderableFields []string,
) api.APIField {
	return api.APIField{
		APIName: "orders",
		Nested:  orderFields(from, orderableFields),
	}
}

// orderFields creates APIFields for ordering fields.
func orderFields(from api.APIFields, orderableFields []string) api.APIFields {
	fields := []api.APIField{}
	apiFields, err := from.GetAPIFields(orderableFields)
	if err != nil {
		panic(err)
	}
	for _, field := range apiFields {
		fields = append(fields, api.APIField{
			APIName: field.APIName,
			Validate: []string{
				"string",
				fmt.Sprintf(
					"oneof=%s %s %s %s",
					dbinput.DirectionAsc,
					dbinput.DirectionDesc,
					dbinput.DirectionAscending,
					dbinput.DirectionDescending,
				),
			},
			Type: "string",
		})
	}
	return fields
}

// pageFieldEntry creates an APIField for pagination.
func pageFieldEntry() api.APIField {
	return api.APIField{
		APIName: "page",
		Nested: []api.APIField{
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
func countFieldEntry() api.APIField {
	return api.APIField{
		APIName:  FieldCount,
		Validate: []string{"bool"},
		Type:     "bool",
	}
}
