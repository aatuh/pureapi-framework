package crud

import (
	"context"
	"net/http"

	"github.com/pureapi/pureapi-framework/crud/setup"
	"github.com/pureapi/pureapi-framework/jsonapi"
)

// DefaultClients groups the CRUD clients into a single struct.
type DefaultClients[CreateInput any, CreateOutput any, GetOutput any] struct {
	Create *CreateClient[CreateInput, jsonapi.APIOutput[CreateOutput]]
	Get    *GetClient[setup.DefaultGetInput, jsonapi.APIOutput[GetOutput]]
	Update *UpdateClient[setup.DefaultUpdateInput, jsonapi.APIOutput[setup.DefaultUpdateOutput]]
	Delete *DeleteClient[setup.DefaultDeleteInput, jsonapi.APIOutput[setup.DefaultDeleteOutput]]
}

// NewDefaultClients creates a new set of CRUD clients for the given URL.
func NewDefaultClients[CreateInput any, CreateOutput any, GetOutput any](
	url string,
) *DefaultClients[CreateInput, CreateOutput, GetOutput] {
	crudClient := NewDefaultClient(url)
	return &DefaultClients[CreateInput, CreateOutput, GetOutput]{
		Create: NewCreateClient[CreateInput, jsonapi.APIOutput[CreateOutput]](crudClient),
		Get:    NewGetClient[setup.DefaultGetInput, jsonapi.APIOutput[GetOutput]](crudClient),
		Update: NewUpdateClient[setup.DefaultUpdateInput, jsonapi.APIOutput[setup.DefaultUpdateOutput]](crudClient),
		Delete: NewDeleteClient[setup.DefaultDeleteInput, jsonapi.APIOutput[setup.DefaultDeleteOutput]](crudClient),
	}
}

// CRUDClient is a base client with common configuration.
type CRUDClient struct {
	url string
}

// NewDefaultClient creates a new base client.
func NewDefaultClient(url string) *CRUDClient {
	return &CRUDClient{url: url}
}

// CreateClient is a generic wrapper for the create endpoint.
type CreateClient[Input, Output any] struct {
	*CRUDClient
}

// NewCreateClient creates a new CreateClient given a base client.
func NewCreateClient[Input, Output any](
	base *CRUDClient,
) *CreateClient[Input, Output] {
	return &CreateClient[Input, Output]{CRUDClient: base}
}

// Send sends a create request using the strongly typed input and output.
func (c *CreateClient[Input, Output]) Send(
	ctx context.Context, host string, input *Input,
) (*jsonapi.Response[Output], error) {
	return jsonapi.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodPost, input,
	)
}

// GetClient is a generic wrapper for the get endpoint.
type GetClient[Input any, Output any] struct {
	*CRUDClient
}

// NewGetClient creates a new GetClient given a base client.
func NewGetClient[Input any, Output any](
	base *CRUDClient,
) *GetClient[Input, Output] {
	return &GetClient[Input, Output]{CRUDClient: base}
}

// Send sends a get request using the strongly typed input and output.
func (c *GetClient[Input, Output]) Send(
	ctx context.Context, host string, input *Input,
) (*jsonapi.Response[Output], error) {
	return jsonapi.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodGet, input,
	)
}

// UpdateClient is a generic wrapper for the update endpoint.
type UpdateClient[Input any, Output any] struct {
	*CRUDClient
}

// NewUpdateClient creates a new UpdateClient given a base client.
func NewUpdateClient[Input any, Output any](
	base *CRUDClient,
) *UpdateClient[Input, Output] {
	return &UpdateClient[Input, Output]{CRUDClient: base}
}

// Send sends an update request using the strongly typed input and output.
func (c *UpdateClient[Input, Output]) Send(
	ctx context.Context, host string, input *Input,
) (*jsonapi.Response[Output], error) {
	return jsonapi.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodPut, input,
	)
}

// DeleteClient is a generic wrapper for the delete endpoint.
type DeleteClient[Input any, Output any] struct {
	*CRUDClient
}

// NewDeleteClient creates a new DeleteClient given a base client.
func NewDeleteClient[Input any, Output any](
	base *CRUDClient,
) *DeleteClient[Input, Output] {
	return &DeleteClient[Input, Output]{CRUDClient: base}
}

// Send sends a delete request using the strongly typed input and output.
func (c *DeleteClient[Input, Output]) Send(
	ctx context.Context, host string, input *Input,
) (*jsonapi.Response[Output], error) {
	return jsonapi.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodDelete, input,
	)
}
