package input

import "github.com/pureapi/pureapi-core/util"

// ErrMaxPageLimitExceededData is the data for the ErrMaxPageLimitExceeded
// error.
type ErrMaxPageLimitExceededData struct {
	MaxLimit int `json:"max_limit"`
}

// ErrMaxPageLimitExceeded is returned when a page limit is exceeded.
var ErrMaxPageLimitExceeded = util.NewAPIError("MAX_PAGE_LIMIT_EXCEEDED")

// ErrInvalidOrderFieldData is the data for the ErrInvalidOrderField error.
type ErrInvalidOrderFieldData struct {
	Field string `json:"field"`
}

// ErrInvalidOrderField is returned when a field is not allowed.
var ErrInvalidOrderField = util.NewAPIError("INVALID_ORDER_FIELD")

// ErrInvalidDatabaseSelectorTranslation indicates that the translation of a
// database selector failed.
var ErrInvalidDatabaseSelectorTranslation = util.NewAPIError(
	"INVALID_DATABASE_SELECTOR_TRANSLATION",
)

// ErrInvalidPredicateData is the data for the ErrInvalidPredicate error.
type ErrInvalidPredicateData struct {
	Predicate Predicate `json:""`
}

// ErrInvalidPredicate is returned when a predicate is not allowed.
var ErrInvalidPredicate = util.NewAPIError("INVALID_PREDICATE")

// ErrInvalidSelectorFieldData is the data for the ErrInvalidSelectorField
// error.
type ErrInvalidSelectorFieldData struct {
	Field string `json:"field"`
}

// ErrInvalidSelectorField is returned when a field is not allowed.
var ErrInvalidSelectorField = util.NewAPIError("INVALID_SELECTOR_FIELD")

// ErrPredicateNotAllowedData is the data for the ErrPredicateNotAllowed
// error.
type ErrPredicateNotAllowedData struct {
	Predicate Predicate `json:"predicate"`
}

// ErrPredicateNotAllowed is returned when a predicate is not allowed.
var ErrPredicateNotAllowed = util.NewAPIError("PREDICATE_NOT_ALLOWED")

// ErrInvalidDatabaseUpdateTranslationData is the data for the
// ErrInvalidDatabaseUpdateTranslation error.
type ErrInvalidDatabaseUpdateTranslationData struct {
	Field string `json:"field"`
}

// ErrInvalidDatabaseUpdateTranslation is used to indicate that the
// translation of a database update failed.
var ErrInvalidDatabaseUpdateTranslation = util.NewAPIError(
	"INVALID_DATABASE_UPDATE_TRANSLATION",
)
