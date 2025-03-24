package types

import (
	"github.com/pureapi/pureapi-framework/dbinput"
	"github.com/pureapi/pureapi-framework/dbquery"
	"github.com/pureapi/pureapi-framework/repository/types"
)

// APIToDBFields maps API fields to database fields.
type APIToDBFields map[string]dbinput.DBField

// ParsedGetEndpointInput represents a parsed get endpoint input.
type ParsedGetEndpointInput struct {
	Selectors dbquery.Selectors
	Orders    []dbquery.Order
	Page      *dbquery.Page
	Count     bool
}

// ParsedUpdateEndpointInput represents a parsed update endpoint input.
type ParsedUpdateEndpointInput struct {
	Selectors dbquery.Selectors
	Updates   []dbquery.Update
	Upsert    bool
}

// ParsedDeleteEndpointInput represents a parsed delete endpoint input.
type ParsedDeleteEndpointInput struct {
	Selectors  dbquery.Selectors
	DeleteOpts *types.DeleteOptions
}
