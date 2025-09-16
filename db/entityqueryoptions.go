package db

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aatuh/pureapi-core/database"
)

// EntityOption defines a functional option for configuring an entity.
type EntityOption[T any] func(T)

// OptionEntityFactoryFn is a function that returns an entity with the given
// options.
type OptionEntityFactoryFn[Entity database.CRUDEntity] func(
	opts ...EntityOption[Entity],
) Entity

// EntityQueryOptions represents a query for an entity.
type EntityQueryOptions[Entity database.CRUDEntity] struct {
	TableName      string
	EntityFn       func() Entity
	OptionEntityFn OptionEntityFactoryFn[Entity]
	SelectorList   Selectors
	UpdateList     Updates
	Options        []EntityOption[Entity]
}

// NewEntityQueryOptions creates a new NewEntityQuery for an entity.
//
// Parameters:
//   - tableName: The name of the table to
//   - entityFn: A function that returns the entity to
//   - optionEntityFn: A function that returns the entity with the given
//     options.
//
// Returns:
//   - *EntityQuery: The new Entity
func NewEntityQueryOptions[Entity database.CRUDEntity](
	tableName string,
	entityFn func() Entity,
	optionEntityFn OptionEntityFactoryFn[Entity],
) *EntityQueryOptions[Entity] {
	return &EntityQueryOptions[Entity]{
		TableName:      tableName,
		EntityFn:       entityFn,
		OptionEntityFn: optionEntityFn,
		SelectorList:   NewSelectors(),
		UpdateList:     NewUpdates(),
		Options:        []EntityOption[Entity]{},
	}
}

// AddSelector appends a selector to the
//
// Parameters:
//   - field: The field to select.
//   - predicate: The predicate to apply.
//   - value: The value to compare.
//
// Returns:
//   - *EntityQuery: The updated Entity
func (q *EntityQueryOptions[Entity]) AddSelector(
	field string, predicate Predicate, value any,
) *EntityQueryOptions[Entity] {
	entity := q.EntityFn()
	selector := MustGetSelector(
		q.TableName, entity, field, predicate, value,
	)
	q.SelectorList = append(q.SelectorList, *selector)
	return q
}

// Selectors returns the current selectors.
//
// Returns:
//   - types.Selectors: The current selectors.
func (q *EntityQueryOptions[Entity]) Selectors() Selectors {
	return q.SelectorList
}

// AddUpdate appends an update clause to the
//
// Parameters:
//   - field: The field to update.
//   - value: The value to set.
//
// Returns:
//   - *EntityQuery: The updated Entity
func (q *EntityQueryOptions[Entity]) AddUpdate(
	field string, value any,
) *EntityQueryOptions[Entity] {
	q.UpdateList = append(
		q.UpdateList,
		*MustGetUpdate(q.EntityFn(), field, value),
	)
	return q
}

// Updates returns the current updates.
//
// Returns:
//   - types.Updates: The current updates.
func (q *EntityQueryOptions[Entity]) Updates() Updates {
	return q.UpdateList
}

// Option creates an entity-specific option.
//
// Parameters:
//   - field: The field to set.
//   - value: The value to set.
//
// Returns:
//   - crud.EntityOption[Entity]: The entity-specific option.
func (q *EntityQueryOptions[Entity]) AddOption(
	field string, value any,
) *EntityQueryOptions[Entity] {
	q.Options = append(
		q.Options, WithOption[Entity](field, value),
	)
	return q
}

// Entity returns the entity that is being queried with the set options.
//
// Returns:
//   - Entity: The entity that is being queried.
func (q *EntityQueryOptions[Entity]) Entity() Entity {
	var entityOpts []EntityOption[Entity]
	for _, opt := range q.Options {
		entityOpts = append(entityOpts, func(e Entity) {
			opt(e)
		})
	}
	return q.OptionEntityFn(entityOpts...)
}

// MustGetSelector creates a new Selector with the given parameters. It panics
// if the provided value's type is not assignable to the expected type.
//
// Parameters:
//   - tableName: The name of the table to select from.
//   - entity: The entity to select from.
//   - column: The column to select.
//   - predicate: The predicate to apply.
//   - value: The value to compare.
//
// Returns:
//   - *types.Selector: The new Selector.
func MustGetSelector(
	tableName string, entity any, column string, predicate Predicate, value any,
) *Selector {
	mustValidateDBMapping(entity, column, value)
	return &Selector{
		Table:     tableName,
		Column:    column,
		Predicate: predicate,
		Value:     value,
	}
}

// MustGetUpdate creates a new Update with the given parameters. It panics
// if the provided value's type is not assignable to the expected type.
//
// Parameters:
//   - entity: The entity to update.
//   - fieldName: The field to update.
//   - value: The value to set.
//
// Returns:
//   - *types.Update: The new Update.
func MustGetUpdate(entity any, fieldName string, value any) *Update {
	mustValidateDBMapping(entity, fieldName, value)
	return &Update{
		Field: fieldName,
		Value: value,
	}
}

// WithOption is a generic functional option that sets a field on an object.
// The field parameter should match the object's struct `db` tag.
// It attempts to automatically handle differences between pointer and
// non-pointer values.
// Only single level pointer differences are handled and passing a multi-level
// pointer field will panic.
//
// Example:
//
//	type SomeEntity struct {
//	    ID   custom.UUID `db:"id"`
//	    Data string    `db:"data"`
//	}
//
//	field = "data"
//	value = "example data"
//
//	Output:
//
//	someEntity.Data = "example data"
//
// Limitations:
//   - Only single-level pointer differences are handled.
//     Multi-level pointers (e.g. **SomeEntity) are not supported.
//   - If the field is not found in the object, or if the provided value's type
//     is not assignable (even after pointer adjustments), the function will
//     panic.
//
// Parameters:
//   - field: The field to set.
//   - value: The value to set.
//
// Returns:
//   - EntityOption[Entity]: The entity-specific option.
func WithOption[T any](
	field string, value any,
) EntityOption[T] {
	return func(t T) {
		v := reflect.ValueOf(t).Elem()
		typ := v.Type()
		// Ensure we're working on the underlying struct.
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			v = v.Elem()
		}
		var found bool
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Tag.Get("db") == field {
				fieldVal := v.Field(i)
				// If the value is zero (or nil for pointers), simply return.
				if reflect.ValueOf(value).IsZero() {
					return
				}
				val := reflect.ValueOf(value)

				// If types directly match, set directly.
				if val.Type().AssignableTo(fieldVal.Type()) {
					fieldVal.Set(val)
					found = true
					break
				}

				// If the value is a pointer and non-nil, but the target is a
				// non-pointer, try setting with the dereferenced value.
				if val.Kind() == reflect.Ptr && !val.IsNil() &&
					val.Elem().Type().AssignableTo(fieldVal.Type()) {
					fieldVal.Set(val.Elem())
					found = true
					break
				}

				// If the target is a pointer but the value is not, allocate a
				// new pointer.
				if fieldVal.Type().Kind() == reflect.Ptr &&
					val.Type().AssignableTo(fieldVal.Type().Elem()) {
					ptrVal := reflect.New(val.Type())
					ptrVal.Elem().Set(val)
					fieldVal.Set(ptrVal)
					found = true
					break
				}

				panic(fmt.Sprintf(
					"WithOption: value for field %q must be assignable to type %v, got %v",
					field,
					fieldVal.Type(),
					reflect.TypeOf(value),
				))
			}
		}
		if !found {
			panic(fmt.Sprintf(
				"WithOption: field %q not found in object struct",
				field,
			))
		}
	}
}

// ScanRow uses reflection to build a slice of pointers to the object fields.+
//
// Parameters:
//   - t: A pointer to the object to be scanned.
//   - row: The row to scan from.
//
// Returns:
//   - error: An error if the scan fails.
func ScanRow[Entity any](t *Entity, row database.Row) error {
	v := reflect.ValueOf(t).Elem()
	typ := v.Type()
	var pointers []any
	for i := 0; i < v.NumField(); i++ {
		if typ.Field(i).Tag.Get("db") == "" {
			continue
		}
		pointers = append(pointers, v.Field(i).Addr().Interface())
	}
	if err := row.Scan(pointers...); err != nil {
		return fmt.Errorf("ScanRow: failed to scan row: %w", err)
	}
	return nil
}

// InsertedValues uses reflection to generate slices of column names and values.
//
// Parameters:
//   - t: A pointer to the object to be scanned.
//
// Returns:
//   - []string: A slice of column names.
//   - []any: A slice of values.
func InsertedValues[Entity any](t *Entity) ([]string, []any) {
	val := reflect.ValueOf(*t)
	typ := reflect.TypeOf(*t)
	var cols []string
	var vals []any
	for i := 0; i < val.NumField(); i++ {
		col := typ.Field(i).Tag.Get("db")
		if col == "" {
			continue
		}
		cols = append(cols, col)
		if typ.Field(i).Type == reflect.TypeOf(time.Time{}) {
			vals = append(vals, val.Field(i).Interface().(time.Time).UnixNano())
		} else {
			vals = append(vals, val.Field(i).Interface())
		}
	}
	return cols, vals
}

// mustValidateDBMapping validates that an object field name and type are valid.
func mustValidateDBMapping(obj any, fieldName string, value any) {
	// Get the mapping of db field names to their types.
	mapping := getDBMapping(obj)
	expectedType, ok := mapping[fieldName]
	if !ok {
		panic(fmt.Errorf("mustValidateDBMapping: unknown field %q", fieldName))
	}
	mustBeAssignable(value, fieldName, expectedType)
}

// getDBMapping builds a mapping from a struct's db tags to their field types.
func getDBMapping(obj any) map[string]reflect.Type {
	mapping := make(map[string]reflect.Type)
	typ := reflect.TypeOf(obj)
	// NumField only supports non-pointer structs.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag != "" {
			mapping[dbTag] = f.Type
		}
	}
	return mapping
}

// mustBeAssignable checks if a provided value is assignable to the expected
// type.
func mustBeAssignable(value any, fieldName string, expectedType reflect.Type) {
	if value == nil {
		panic(fmt.Sprintf(
			"mustBeAssignable: value for field %q is nil; expected type %v",
			fieldName,
			expectedType,
		))
	}
	vType := reflect.TypeOf(value)
	// Check direct assignability.
	if vType.AssignableTo(expectedType) {
		return
	}
	// If value is a pointer and non-nil, check if element type is assignable.
	if vType.Kind() == reflect.Ptr && !reflect.ValueOf(value).IsNil() {
		if vType.Elem().AssignableTo(expectedType) {
			return
		}
	}
	panic(fmt.Sprintf(
		"mustBeAssignable: value for field %q must be assignable to type %v, got %T",
		fieldName,
		expectedType,
		value,
	))
}
