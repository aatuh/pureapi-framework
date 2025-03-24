package crud

import (
	"context"
	"net/http"

	"github.com/pureapi/pureapi-framework/client"
	"github.com/pureapi/pureapi-framework/crud/setup"
	"github.com/pureapi/pureapi-framework/json"
)

// CRUDClients groups the CRUD clients into a single struct.
type CRUDClients[CreateInput any, CreateOutput any, GetOutput any] struct {
	Create *CreateClient[CreateInput, CreateOutput]
	Get    *GetClient[GetOutput]
	Update *UpdateClient
	Delete *DeleteClient
}

// NewCRUDClients creates a new set of CRUD clients for the given URL.
func NewCRUDClients[CreateInput any, CreateOutput any, GetOutput any](
	url string,
) *CRUDClients[CreateInput, CreateOutput, GetOutput] {
	crudClient := NewCRUDClient(url)
	return &CRUDClients[CreateInput, CreateOutput, GetOutput]{
		Create: NewCreateClient[CreateInput, CreateOutput](crudClient),
		Get:    NewGetClient[GetOutput](crudClient),
		Update: NewUpdateClient(crudClient),
		Delete: NewDeleteClient(crudClient),
	}
}

// CRUDClient is a non‑generic base client with common configuration.
type CRUDClient struct {
	url string
}

// NewCRUDClient creates a new base client.
func NewCRUDClient(url string) *CRUDClient {
	return &CRUDClient{url: url}
}

// CreateClient is a generic wrapper for the Create endpoint.
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
) (*json.Response[json.APIOutput[Output]], error) {
	return client.SendRequest[Input, Output](
		ctx, host, c.url, http.MethodPost, input,
	)
}

// GetClient is a generic wrapper for the Get endpoint.
type GetClient[Output any] struct {
	*CRUDClient
}

// NewGetClient creates a new GetClient given a base client.
func NewGetClient[Output any](base *CRUDClient) *GetClient[Output] {
	return &GetClient[Output]{CRUDClient: base}
}

// Send sends a get request using the strongly typed input and output.
func (c *GetClient[Output]) Send(
	ctx context.Context, host string, input *setup.GetInput,
) (*json.Response[json.APIOutput[Output]], error) {
	return client.SendRequest[setup.GetInput, Output](
		ctx, host, c.url, http.MethodGet, input,
	)
}

// UpdateClient is a generic wrapper for the Update endpoint.
type UpdateClient struct {
	*CRUDClient
}

// NewUpdateClient creates a new UpdateClient given a base client.
func NewUpdateClient(base *CRUDClient) *UpdateClient {
	return &UpdateClient{CRUDClient: base}
}

// TODO: Can work without setup.*Input/Output -> SendRequest remove generic

// Send sends an update request using the strongly typed input and output.
func (c *UpdateClient) Send(
	ctx context.Context, host string, input *setup.UpdateInput,
) (*json.Response[json.APIOutput[setup.UpdateOutput]], error) {
	return client.SendRequest[setup.UpdateInput, setup.UpdateOutput](
		ctx, host, c.url, http.MethodPut, input,
	)
}

// DeleteClient is a generic wrapper for the Delete endpoint.
type DeleteClient struct {
	*CRUDClient
}

// NewDeleteClient creates a new DeleteClient given a base client.
func NewDeleteClient(base *CRUDClient) *DeleteClient {
	return &DeleteClient{CRUDClient: base}
}

// Send sends a delete request using the strongly typed input and output.
func (c *DeleteClient) Send(
	ctx context.Context, host string, input *setup.DeleteInput,
) (*json.Response[json.APIOutput[setup.DeleteOutput]], error) {
	return client.SendRequest[setup.DeleteInput, setup.DeleteOutput](
		ctx, host, c.url, http.MethodDelete, input,
	)
}
