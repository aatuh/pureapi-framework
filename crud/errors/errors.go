package errors

import (
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-framework/api"
	"github.com/pureapi/pureapi-framework/crud/services"
	"github.com/pureapi/pureapi-framework/dbinput"
	"github.com/pureapi/pureapi-framework/dbquery"
	"github.com/pureapi/pureapi-framework/input"
)

// MaskedErrorMapping returns a map of masked error IDs to ExpectedErrors.
//
// Example:
//
//	expectedErrors := ExpectedErrors{
//	    {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
//
//	mappedErrors := map[string]ExpectedError{
//	    "masked_id": {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
func MaskedErrorMapping(
	expectedErrors api.ExpectedErrors,
) map[string]api.ExpectedError {
	mappedErrors := map[string]api.ExpectedError{}
	for i := range expectedErrors {
		mappedErrors[expectedErrors[i].MaskedID] = expectedErrors[i]
	}
	return mappedErrors
}

// MustApplyErrorMapping applies a mapping of error IDs to new IDs by modifying
// the masked error IDs in the default errors. If an error ID is not found in
// the mapping, it panics.
//
// Example:
//
//	defaultErrors := ExpectedErrors{
//	    {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
//
//	mappedErrors := map[string]ExpectedError{
//	    "masked_id": {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
//
//	newExpectedErrors := MustApplyErrorMapping(defaultErrors, mappedErrors)
//
// Contents of newExpectedErrors:
//
//	{
//	    {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
func MustApplyErrorMapping(
	expectedErrors api.ExpectedErrors,
	maskedErrorMap map[string]api.ExpectedError,
) api.ExpectedErrors {
	newExpectedErrors := expectedErrors
	for i := range maskedErrorMap {
		maskedError := maskedErrorMap[i]
		expectedError := expectedErrors.GetByID(maskedError.ID)
		if expectedError == nil {
			panic(fmt.Sprintf(
				"mustApplyErrorMapping: mapped error ID %q not found in default errors",
				maskedError.ID,
			))
		}
		// Replace
		newExpectedErrors = newExpectedErrors.With(maskedError)
	}
	return newExpectedErrors
}

func GenericErrors() api.ExpectedErrors {
	return []api.ExpectedError{
		{ID: input.ErrInputDecoding.ID(), MaskedID: input.ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: input.ErrInvalidInput.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: input.ErrValidation.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}

func CreateErrors() api.ExpectedErrors {
	return []api.ExpectedError{
		{ID: dbquery.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: dbquery.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
	}
}

func GetErrors() api.ExpectedErrors {
	return []api.ExpectedError{
		{ID: dbinput.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrMaxPageLimitExceeded.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbquery.ErrNoRows.ID(), Status: http.StatusNotFound, PublicData: true},
	}
}

func UpdateErrors() api.ExpectedErrors {
	return []api.ExpectedError{
		{ID: dbinput.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneUpdate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidOrderField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbquery.ErrDuplicateEntry.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: dbquery.ErrForeignConstraint.ID(), Status: http.StatusBadRequest, PublicData: false},
		{ID: dbinput.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: dbinput.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}

func DeleteErrors() api.ExpectedErrors {
	return []api.ExpectedError{
		{ID: dbinput.ErrInvalidPredicate.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrPredicateNotAllowed.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: services.ErrNeedAtLeastOneSelector.ID(), Status: http.StatusBadRequest, PublicData: true},
		{ID: dbinput.ErrInvalidDatabaseSelectorTranslation.ID(), MaskedID: dbinput.ErrInvalidSelectorField.ID(), Status: http.StatusBadRequest, PublicData: true},
	}
}
