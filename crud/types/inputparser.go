package types

import (
	"github.com/pureapi/pureapi-framework/db/query"
	"github.com/pureapi/pureapi-framework/repository/types"
)

// ParsedGetEndpointInput represents a parsed get endpoint input.
type ParsedGetEndpointInput struct {
	Selectors query.Selectors
	Orders    []query.Order
	Page      *query.Page
	Count     bool
}

// ParsedUpdateEndpointInput represents a parsed update endpoint input.
type ParsedUpdateEndpointInput struct {
	Selectors query.Selectors
	Updates   []query.Update
	Upsert    bool
}

// ParsedDeleteEndpointInput represents a parsed delete endpoint input.
type ParsedDeleteEndpointInput struct {
	Selectors  query.Selectors
	DeleteOpts *types.DeleteOptions
}
