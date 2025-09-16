package db

import (
	"fmt"
	"strings"

	"github.com/aatuh/pureapi-framework/db"
)

// Predicate is a string representation of a filtering predicate.
type Predicate string

// String returns the string representation of the predicate.
//
// Returns:
//   - string: The string representation of the predicate.
func (p Predicate) String() string {
	return string(p)
}

// Predicates is a slice of Predicate values.
type Predicates []Predicate

// String returns a string representation of the predicates.
//
// Returns:
//   - string: A comma-separated string of the predicates.
func (p Predicates) String() string {
	str := make([]string, len(p))
	for i, predicate := range p {
		str[i] = predicate.String()
	}
	return strings.Join(str, ",")
}

// StrSlice returns a slice of strings representing the predicates.
//
// Returns:
//   - []string: A slice of strings representing the predicates.
func (p Predicates) StrSlice() []string {
	str := make([]string, len(p))
	for i, predicate := range p {
		str[i] = predicate.String()
	}
	return str
}

// Predicates for filtering data.
const (
	Greater        Predicate = ">"
	Gt             Predicate = "gt"
	GreaterOrEqual Predicate = ">="
	Ge             Predicate = "ge"
	Equal          Predicate = "="
	Eq             Predicate = "eq"
	NotEqual       Predicate = "!="
	Ne             Predicate = "ne"
	Less           Predicate = "<"
	Lt             Predicate = "LT"
	LessOrEqual    Predicate = "<="
	Le             Predicate = "le"
	In             Predicate = "in"
	NotIn          Predicate = "not_in"
)

// ToDBPredicates maps API-level predicates to database predicates.
var ToDBPredicates = map[Predicate]db.Predicate{
	Greater:        db.Greater,
	Gt:             db.Greater,
	GreaterOrEqual: db.GreaterOrEqual,
	Ge:             db.GreaterOrEqual,
	Equal:          db.Equal,
	Eq:             db.Equal,
	NotEqual:       db.NotEqual,
	Ne:             db.NotEqual,
	Less:           db.Less,
	Lt:             db.Less,
	LessOrEqual:    db.LessOrEqual,
	Le:             db.LessOrEqual,
	In:             db.In,
	NotIn:          db.NotIn,
}

// AllPredicates is a slice of all available predicates.
var AllPredicates = []Predicate{
	Greater, Gt,
	GreaterOrEqual, Ge,
	Equal, Eq,
	NotEqual, Ne,
	Less, Lt,
	LessOrEqual, Le,
	In,
	NotIn,
}

// OnlyEqualPredicates is a slice of predicates that only allow equality.
var OnlyEqualPredicates = []Predicate{Equal, Eq}

// EqualAndNotEqualPredicates is a slice of predicates that allow both equality
// and inequality.
var EqualAndNotEqualPredicates = []Predicate{Equal, Eq, NotEqual, Ne}

// OnlyGreaterPredicates is a slice of predicates that only allow greater
// values.
var OnlyGreaterPredicates = []Predicate{GreaterOrEqual, Ge, Greater, Gt}

// OnlyLessPredicates is a slice of predicates that only allow less values.
var OnlyLessPredicates = []Predicate{LessOrEqual, Le, Less, Lt}

// OnlyInAndNotInPredicates is a slice of predicates that only allow
// IN and NOT_IN.
var OnlyInAndNotInPredicates = []Predicate{In, NotIn}

// APIToDBField is used to translate between API field and database field.
type APIToDBField struct {
	Table  string
	Column string
}

// Page represents a pagination input.
type Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// ToDBPage converts a Page to database Page.
//
// Returns:
//   - *Page: The database Page.
func (p *Page) ToDBPage() *db.Page {
	return &db.Page{
		Offset: p.Offset,
		Limit:  p.Limit,
	}
}

// Direction is used to specify the direction of the order.
type Direction string

// String returns the string representation of the order direction.
//
// Returns:
//   - string: The string representation of the order direction.
func (o Direction) String() string {
	return string(o)
}

// Available order directions.
const (
	DirectionAsc        Direction = "asc"
	DirectionAscending  Direction = "ascending"
	DirectionDesc       Direction = "desc"
	DirectionDescending Direction = "descending"
)

// DirectionsToDB is a map of order directions to database order directions.
var DirectionsToDB = map[Direction]db.Direction{
	DirectionAsc:        db.OrderAsc,
	DirectionAscending:  db.OrderAsc,
	DirectionDesc:       db.OrderDesc,
	DirectionDescending: db.OrderDesc,
}

// Orders is a map of field names to order directions.
type Orders map[string]Direction

// TranslateToDBOrders translates the provided orders into database orders.
// It also returns an error if any of the orders are invalid.
//
// Parameters:
//   - orders: The list of orders to translate.
//   - allowedOrderFields: The list of allowed order fields.
//   - apiToDBFieldMap: The mapping of API field names to database field names.
//
// Returns:
//   - []Order: The list of database orders.
//   - error: An error if any of the orders are invalid.
func (o Orders) TranslateToDBOrders(
	apiToDBFieldMap map[string]APIToDBField,
) ([]db.Order, error) {
	dbOrders, err := o.dedup().ToDBOrders(apiToDBFieldMap)
	if err != nil {
		return nil, fmt.Errorf("TranslateToDBOrders: %w", err)
	}
	return dbOrders, nil
}

// dedup deduplicates the provided orders.
func (o Orders) dedup() Orders {
	dedup := map[string]Direction{}
	existing := make(map[string]bool)
	for field := range o {
		order := o[field]
		if !existing[field] {
			dedup[field] = order
			existing[field] = true
		}
	}
	return dedup
}

// ToDBOrders translates the provided orders into database orders.
// It returns an error if any of the orders are invalid.
//
// Parameters:
//   - apiToDBFieldMap: The mapping of API field names to database field names.
//
// Returns:
//   - []Order: The list of database orders.
//   - error: An error if any of the orders are invalid.
func (o Orders) ToDBOrders(
	apiToDBFieldMap map[string]APIToDBField,
) ([]db.Order, error) {
	dbOrders := []db.Order{}

	for field, direction := range o {
		translatedField := apiToDBFieldMap[field]

		// Translate field.
		dbColumn := translatedField.Column
		if dbColumn == "" {
			return nil, ErrInvalidOrderField.
				WithData(
					ErrInvalidOrderFieldData{Field: field},
				).
				WithMessage(fmt.Sprintf(
					"cannot translate field: %s", field,
				))
		}

		lowerDir := Direction(strings.ToLower(string(direction)))
		dbOrders = append(dbOrders, db.Order{
			Table:     translatedField.Table,
			Field:     dbColumn,
			Direction: DirectionsToDB[lowerDir],
		})
	}

	return dbOrders, nil
}

// ErrInvalidDatabaseSelectorTranslationData is the data for the
// ErrInvalidDatabaseSelectorTranslation error.
type ErrInvalidDatabaseSelectorTranslationData struct {
	Field string `json:"field"`
}

// Selector represents a data selector that specifies criteria for filtering
// data based on fields, predicates, and values.
type Selector struct {
	Predicate Predicate `json:"predicate"` // The predicate to use.
	Value     any       `json:"value"`     // The value to filter on.
}

// NewSelector creates a new API selector with the provided predicate and value.
//
// Parameters:
//   - predicate: The predicate to use.
//   - value: The value to filter on.
//
// Returns:
//   - *APISelector: A new selector.
func NewSelector(predicate Predicate, value any) *Selector {
	return &Selector{
		Predicate: predicate,
		Value:     value,
	}
}

// String returns a string representation of the selector.
// It is useful for debugging and logging purposes.
//
// Returns:
//   - string: A formatted string showing the field, predicate, and value.
func (s Selector) String() string {
	return fmt.Sprintf("%s %v", s.Predicate, s.Value)
}

// APISelectors represents a collection of selectors used for filtering data.
// It is a map where the key is the field name and the value is the selector.
type APISelectors map[string]Selector

// AddSelector adds a new selector to the collection of selectors.
//
// Parameters:
//   - field: The field name.
//   - predicate: The predicate to use.
//   - value: The value to filter on.
//
// Returns:
//   - APISelectors: A new collection of selectors with the new selector added.
func (s APISelectors) AddSelector(
	field string, predicate Predicate, value any,
) APISelectors {
	s[field] = Selector{
		Predicate: predicate,
		Value:     value,
	}
	return s
}

// ToDBSelectors converts a slice of API-level selectors to database selectors.
//
// Parameters:
//   - apiToDBFieldMap: A map translating API field names to their corresponding
//     database field definitions.
//
// Returns:
//   - []types.Selector: A slice of types.Selector, which represents the
//     translated database selectors.
//   - error: An error if any validation fails, such as invalid predicates or
//     unknown fields.
func (s APISelectors) ToDBSelectors(
	apiToDBFieldMap map[string]APIToDBField,
) ([]db.Selector, error) {
	databaseSelectors := []db.Selector{}
	for field := range s {
		// Translate the predicate.
		selector := s[field]
		lowerPredicate := Predicate(strings.ToLower(string(selector.Predicate)))
		dbPredicate, ok := ToDBPredicates[lowerPredicate]
		if !ok {
			return nil, ErrInvalidPredicate.
				WithData(
					ErrInvalidPredicateData{Predicate: selector.Predicate},
				).
				WithMessage(fmt.Sprintf(
					"cannot translate predicate: %s", selector.Predicate,
				))
		}
		// Translate the field.
		dbField, ok := apiToDBFieldMap[field]
		if !ok {
			return nil, ErrInvalidDatabaseSelectorTranslation.
				WithData(
					ErrInvalidDatabaseSelectorTranslationData{Field: field},
				).
				WithMessage(fmt.Sprintf(
					"cannot translate field: %s", field,
				))
		}
		// Create the database selector.
		databaseSelectors = append(databaseSelectors, db.Selector{
			Table:     dbField.Table,
			Column:    dbField.Column,
			Predicate: dbPredicate,
			Value:     selector.Value,
		})
	}

	return databaseSelectors, nil
}

// APIUpdates represents a list of updates to apply to a database entity.
type APIUpdates map[string]any

// ToDBUpdates translates a list of updates to a database update list
// and returns an error if the translation fails.
//
// Parameters:
//   - updates: The list of updates to translate.
//   - apiToDBFieldMap: The mapping of API field names to database field names.
//
// Returns:
//   - []Update: A list of database entity updates.
//   - error: An error if any field translation fails.
func (updates APIUpdates) ToDBUpdates(
	apiToDBFieldMap map[string]APIToDBField,
) ([]db.Update, error) {
	dbUpdates := []db.Update{}

	for field := range updates {
		value := updates[field]

		// Translate the field.
		dbField, ok := apiToDBFieldMap[field]
		if !ok {
			return nil, ErrInvalidDatabaseUpdateTranslation.
				WithData(
					ErrInvalidDatabaseUpdateTranslationData{Field: field},
				).
				WithMessage(fmt.Sprintf(
					"cannot translate field: %s", field,
				))
		}

		dbUpdates = append(dbUpdates, db.Update{
			Field: dbField.Column,
			Value: value,
		})
	}

	return dbUpdates, nil
}
