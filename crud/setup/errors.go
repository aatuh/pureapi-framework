package setup

import (
	"net/http"

	"github.com/pureapi/pureapi-framework/api/errors"
	"github.com/pureapi/pureapi-framework/crud/services"
	"github.com/pureapi/pureapi-framework/db/input"
	"github.com/pureapi/pureapi-framework/db/query"
)

func CreateErrors() errors.ExpectedErrors {
	return []errors.ExpectedError{
		{ID: query.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: query.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
	}
}

func GetErrors() errors.ExpectedErrors {
	return []errors.ExpectedError{
		{ID: input.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrMaxPageLimitExceeded.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: query.ErrNoRows.ID(), Status: http.StatusNotFound, PublicData: true},
	}
}

func UpdateErrors() errors.ExpectedErrors {
	return []errors.ExpectedError{
		{ID: input.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneUpdate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: query.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: query.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: input.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: input.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}

func DeleteErrors() errors.ExpectedErrors {
	return []errors.ExpectedError{
		{ID: input.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: input.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}
