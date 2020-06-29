//+build !swipe

// Code generated by Swipe v1.20.0. DO NOT EDIT.

//go:generate swipe
package jsonrpc

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/l-vitaly/go-kit/transport/http/jsonrpc"
	"github.com/swipe-io/swipe/fixtures/user"
)

type ServiceInterfaceClientOption func(*clientServiceInterface)

func ServiceInterfaceGenericClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.genericClientOption = opt }
}

func ServiceInterfaceGenericClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.genericEndpointMiddleware = opt }
}

func ServiceInterfaceGetAllClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.getAllClientOption = opt }
}

func ServiceInterfaceGetAllClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.getAllEndpointMiddleware = opt }
}

func ServiceInterfaceTestMethodClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.testMethodClientOption = opt }
}

func ServiceInterfaceTestMethodClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.testMethodEndpointMiddleware = opt }
}

func ServiceInterfaceCreateClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.createClientOption = opt }
}

func ServiceInterfaceCreateClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.createEndpointMiddleware = opt }
}

func ServiceInterfaceDeleteClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.deleteClientOption = opt }
}

func ServiceInterfaceDeleteClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.deleteEndpointMiddleware = opt }
}

func ServiceInterfaceGetClientOptions(opt ...jsonrpc.ClientOption) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.getClientOption = opt }
}

func ServiceInterfaceGetClientEndpointMiddlewares(opt ...endpoint.Middleware) (_ ServiceInterfaceClientOption) {
	return func(c *clientServiceInterface) { c.getEndpointMiddleware = opt }
}

type clientServiceInterface struct {
	createEndpoint               endpoint.Endpoint
	createClientOption           []jsonrpc.ClientOption
	createEndpointMiddleware     []endpoint.Middleware
	deleteEndpoint               endpoint.Endpoint
	deleteClientOption           []jsonrpc.ClientOption
	deleteEndpointMiddleware     []endpoint.Middleware
	getEndpoint                  endpoint.Endpoint
	getClientOption              []jsonrpc.ClientOption
	getEndpointMiddleware        []endpoint.Middleware
	getAllEndpoint               endpoint.Endpoint
	getAllClientOption           []jsonrpc.ClientOption
	getAllEndpointMiddleware     []endpoint.Middleware
	testMethodEndpoint           endpoint.Endpoint
	testMethodClientOption       []jsonrpc.ClientOption
	testMethodEndpointMiddleware []endpoint.Middleware
	genericClientOption          []jsonrpc.ClientOption
	genericEndpointMiddleware    []endpoint.Middleware
}

func (c *clientServiceInterface) TestMethod(data map[string]interface{}, ss interface{}) (_ map[string]map[int][]string, _ error) {
	resp, err := c.testMethodEndpoint(context.Background(), testMethodRequestServiceInterface{Data: data, Ss: ss})
	if err != nil {
		return nil, err
	}
	response := resp.(testMethodResponseServiceInterface)
	return response.States, nil
}

func (c *clientServiceInterface) Create(ctx context.Context, name string, data []byte) (_ error) {
	_, err := c.createEndpoint(ctx, createRequestServiceInterface{Name: name, Data: data})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientServiceInterface) Delete(ctx context.Context, id uint) (_ string, _ string, _ error) {
	resp, err := c.deleteEndpoint(ctx, deleteRequestServiceInterface{Id: id})
	if err != nil {
		return "", "", err
	}
	response := resp.(deleteResponseServiceInterface)
	return response.A, response.B, nil
}

func (c *clientServiceInterface) Get(ctx context.Context, id int, name string, fname string, price float32, n int) (_ user.User, _ error) {
	resp, err := c.getEndpoint(ctx, getRequestServiceInterface{Id: id, Name: name, Fname: fname, Price: price, N: n})
	if err != nil {
		return user.User{}, err
	}
	response := resp.(getResponseServiceInterface)
	return response.Data, nil
}

func (c *clientServiceInterface) GetAll(ctx context.Context) (_ []*user.User, _ error) {
	resp, err := c.getAllEndpoint(ctx, nil)
	if err != nil {
		return nil, err
	}
	response := resp.([]*user.User)
	return response, nil
}