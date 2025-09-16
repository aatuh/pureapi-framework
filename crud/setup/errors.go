package setup

import (
	"net/http"

	apidb "github.com/aatuh/pureapi-framework/api/db"
	"github.com/aatuh/pureapi-framework/api/errutil"
	"github.com/aatuh/pureapi-framework/crud/services"
	"github.com/aatuh/pureapi-framework/db"
)

func CreateErrors() errutil.ExpectedErrors {
	return []errutil.ExpectedError{
		{ID: db.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: db.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
	}
}

func GetErrors() errutil.ExpectedErrors {
	return []errutil.ExpectedError{
		{ID: apidb.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrMaxPageLimitExceeded.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: db.ErrNoRows.ID(), Status: http.StatusNotFound, PublicData: true},
	}
}

func UpdateErrors() errutil.ExpectedErrors {
	return []errutil.ExpectedError{
		{ID: apidb.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneUpdate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: db.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: db.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: apidb.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: apidb.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}

func DeleteErrors() errutil.ExpectedErrors {
	return []errutil.ExpectedError{
		{ID: apidb.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: apidb.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: apidb.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}
