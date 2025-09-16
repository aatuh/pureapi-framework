package db

// Predicates for filtering data.
const (
	Greater        Predicate = ">"
	GreaterOrEqual Predicate = ">="
	Equal          Predicate = "="
	NotEqual       Predicate = "!="
	Less           Predicate = "<"
	LessOrEqual    Predicate = "<="
	In             Predicate = "IN"
	NotIn          Predicate = "NOT IN"
	Like           Predicate = "LIKE"
	NotLike        Predicate = "NOT LIKE"
)

// Order directions.
const (
	OrderAsc  Direction = "ASC"
	OrderDesc Direction = "DESC"
)

// Join types.
const (
	JoinTypeInner JoinType = "INNER"
	JoinTypeLeft  JoinType = "LEFT"
	JoinTypeRight JoinType = "RIGHT"
	JoinTypeFull  JoinType = "FULL"
)

// Predicate represents the predicate of a database selector.
type Predicate string

// Direction is used to specify the order of the result set.
type Direction string

// Order is used to specify the order of the result set.
type Order struct {
	Table     string
	Field     string
	Direction Direction
}

// Orders is a list of orders.
type Orders []Order

// ColumnSelector represents a column selector.
type ColumnSelector struct {
	Table  string
	Column string
}

// Projection represents a projected column in a query.
type Projection struct {
	Table  string
	Column string
	Alias  string
}

// Projections is a list of projections.
type Projections []Projection

// Selector represents a database selector.
type Selector struct {
	Table     string
	Column    string
	Predicate Predicate
	Value     any
}

// NewSelector creates a new selector with the given parameters.
//
// Parameters:
//   - column: the column name.
//   - predicate: the predicate.
//   - value: the value.
//
// Returns:
//   - *Selector: The new selector.
func NewSelector(column string, predicate Predicate, value any) *Selector {
	return &Selector{
		Column:    column,
		Predicate: predicate,
		Value:     value,
	}
}

// WithTable returns a new selector with the provided table name.
//
// Parameters:
//   - table: the table name.
//
// Returns:
//   - *Selector: The new selector.
func (s *Selector) WithTable(table string) *Selector {
	newSelector := *s
	newSelector.Table = table
	return &newSelector
}

// WithColumn returns a new selector with the provided column name.
//
// Parameters:
//   - column: the column name.
//
// Returns:
//   - *Selector: The new selector.
func (s *Selector) WithColumn(column string) *Selector {
	newSelector := *s
	newSelector.Column = column
	return &newSelector
}

// WithPredicate returns a new selector with the provided predicate.
//
// Parameters:
//   - predicate: the predicate.
//
// Returns:
//   - *Selector: The new selector.
func (s *Selector) WithPredicate(predicate Predicate) *Selector {
	newSelector := *s
	newSelector.Predicate = predicate
	return &newSelector
}

// WithValue returns a new selector with the provided value.
//
// Parameters:
//   - value: the value.
//
// Returns:
//   - *Selector: The new selector.
func (s *Selector) WithValue(value any) *Selector {
	newSelector := *s
	newSelector.Value = value
	return &newSelector
}

// Selectors represents a list of database selectors.
type Selectors []Selector

// NewSelectors returns a new list of selectors.
//
// Parameters:
//   - selectors: The selectors.
//
// Returns:
//   - Selectors: The new list of selectors.
func NewSelectors(selectors ...Selector) Selectors {
	return selectors
}

// Add adds a new selector to the list.
//
// Parameters:
//   - column: The column name.
//   - predicate: The predicate.
//   - value: The value.
//
// Returns:
//   - Selectors: The new list of selectors.
func (s Selectors) Add(
	column string, predicate Predicate, value any,
) Selectors {
	return append(s, *NewSelector(column, predicate, value))
}

// GetByField returns selector with the given field.
//
// Parameters:
//   - field: the field to search for.
//
// Returns:
//   - *Selector: The selector.
func (s Selectors) GetByField(field string) *Selector {
	for j := range s {
		if s[j].Column == field {
			return &s[j]
		}
	}
	return nil
}

// GetByFields returns selectors with the given fields.
//
// Parameters:
//   - fields: the fields to search for.
//
// Returns:
//   - []Selector: A list of selectors.
func (s Selectors) GetByFields(fields ...string) []Selector {
	var result []Selector
	for _, field := range fields {
		for i := range s {
			if s[i].Column == field {
				result = append(result, s[i])
			}
		}
	}
	return result
}

// Update is the options struct used for update queries.
type Update struct {
	Field string
	Value any
}

// NewUpdate creates a new update field.
//
// Parameters:
//   - field: The field.
//   - value: The value
//
// Returns:
//   - Update: The new update field.
func NewUpdate(field string, value any) Update {
	return Update{
		Field: field,
		Value: value,
	}
}

// WithField returns a new update field with the provided field name.
//
// Parameters:
//   - field: The field.
//
// Returns:
//   - Update: The new update field.
func (u Update) WithField(field string) Update {
	newUpdate := u
	newUpdate.Field = field
	return newUpdate
}

// WithValue returns a new update field with the provided value.
//
// Parameters:
//   - value: The value.
//
// Returns:
//   - Update: The new update field.
func (u Update) WithValue(value any) Update {
	newUpdate := u
	newUpdate.Value = value
	return newUpdate
}

// Updates is a list of update fields
type Updates []Update

// NewUpdates creates a new list of updates.
//
// Parameters:
//   - updates: The updates.
//
// Returns:
//   - Updates: The new list of updates.
func NewUpdates(updates ...Update) Updates {
	return updates
}

// Add adds a new update field to the list.
//
// Parameters:
//   - field: The field.
//   - value: The value.
//
// Returns:
//   - Updates: The new list of updates.
func (u Updates) Add(field string, value any) Updates {
	return append(u, Update{Field: field, Value: value})
}

// Page is used to specify the page of the result set.
type Page struct {
	Offset int
	Limit  int
}

// JoinType represents the type of join.
type JoinType string

// Join represents a database join clause.
type Join struct {
	JoinType JoinType
	Table    string
	OnLeft   ColumnSelector
	OnRight  ColumnSelector
}

// NewJoin creates a new join clause.
//
// Parameters:
//   - joinType: The type of join.
//   - table: The table name.
//   - onLeft: The left column selector.
//   - onRight: The right column selector.
//
// Returns:
//   - Join: The new join clause.
func NewJoin(
	joinType JoinType, table string, onLeft, onRight ColumnSelector,
) Join {
	return Join{
		JoinType: joinType,
		Table:    table,
		OnLeft:   onLeft,
		OnRight:  onRight,
	}
}

// WithJoinType returns a new join with the provided join type.
//
// Parameters:
//   - joinType: The join type.
//
// Returns:
//   - Join: The new join.
func (j Join) WithJoinType(joinType JoinType) Join {
	newJoin := j
	newJoin.JoinType = joinType
	return newJoin
}

// WithTable returns a new join with the provided table name.
//
// Parameters:
//   - table: The table name.
//
// Returns:
//   - Join: The new join.
func (j Join) WithTable(table string) Join {
	newJoin := j
	newJoin.Table = table
	return newJoin
}

// WithOnLeft returns a new join with the provided left column selector.
//
// Parameters:
//   - onLeft: The left column selector.
//
// Returns:
//   - Join: The new join.
func (j Join) WithOnLeft(onLeft ColumnSelector) Join {
	newJoin := j
	newJoin.OnLeft = onLeft
	return newJoin
}

// WithOnRight returns a new join with the provided right column selector.
//
// Parameters:
//   - onRight: The right column selector.
//
// Returns:
//   - Join: The new join.
func (j Join) WithOnRight(onRight ColumnSelector) Join {
	newJoin := j
	newJoin.OnRight = onRight
	return newJoin
}

// Joins is a list of joins.
type Joins []Join
