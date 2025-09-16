package crud

import (
	"context"
	"net/http"

	"github.com/aatuh/pureapi-framework/api/json"
	"github.com/aatuh/pureapi-framework/crud/setup"
	"github.com/aatuh/pureapi-framework/defaults"
)

// DefaultClients groups the CRUD clients into a single struct.
type DefaultClients[CreateInput any, CreateOutput any, GetOutput any] struct {
	Create *CreateClient[CreateInput, json.APIOutput[CreateOutput]]
	Get    *GetClient[setup.DefaultGetInput, json.APIOutput[GetOutput]]
	Update *UpdateClient[setup.DefaultUpdateInput, json.APIOutput[setup.DefaultUpdateOutput]]
	Delete *DeleteClient[setup.DefaultDeleteInput, json.APIOutput[setup.DefaultDeleteOutput]]
}

// NewDefaultClients creates a new set of CRUD clients for the given URL.
func NewDefaultClients[CreateInput any, CreateOutput any, GetOutput any](
	url string,
) *DefaultClients[CreateInput, CreateOutput, GetOutput] {
	crudClient := NewDefaultClient(url)
	return &DefaultClients[CreateInput, CreateOutput, GetOutput]{
		Create: NewCreateClient[CreateInput, json.APIOutput[CreateOutput]](crudClient),
		Get:    NewGetClient[setup.DefaultGetInput, json.APIOutput[GetOutput]](crudClient),
		Update: NewUpdateClient[setup.DefaultUpdateInput, json.APIOutput[setup.DefaultUpdateOutput]](crudClient),
		Delete: NewDeleteClient[setup.DefaultDeleteInput, json.APIOutput[setup.DefaultDeleteOutput]](crudClient),
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
) (*json.Response[Output], error) {
	return json.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodPost, input, defaults.CtxLogger,
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
) (*json.Response[Output], error) {
	return json.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodGet, input, defaults.CtxLogger,
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
) (*json.Response[Output], error) {
	return json.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodPut, input, defaults.CtxLogger,
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
) (*json.Response[Output], error) {
	return json.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodDelete, input, defaults.CtxLogger,
	)
}
