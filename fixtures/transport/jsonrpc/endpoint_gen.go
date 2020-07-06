//+build !swipe

// Code generated by Swipe v1.22.2. DO NOT EDIT.

//go:generate swipe
package jsonrpc

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/swipe-io/swipe/fixtures/service"
	"github.com/swipe-io/swipe/fixtures/user"
	"io"
)

type EndpointSet struct {
	CreateEndpoint      endpoint.Endpoint
	DeleteEndpoint      endpoint.Endpoint
	GetEndpoint         endpoint.Endpoint
	GetAllEndpoint      endpoint.Endpoint
	TestMethodEndpoint  endpoint.Endpoint
	TestMethod2Endpoint endpoint.Endpoint
}

func MakeEndpointSet(s service.Interface) EndpointSet {
	return EndpointSet{
		CreateEndpoint:      makeCreateEndpoint(s),
		DeleteEndpoint:      makeDeleteEndpoint(s),
		GetEndpoint:         makeGetEndpoint(s),
		GetAllEndpoint:      makeGetAllEndpoint(s),
		TestMethodEndpoint:  makeTestMethodEndpoint(s),
		TestMethod2Endpoint: makeTestMethod2Endpoint(s),
	}
}

type createRequestServiceInterface struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}
type createResponseServiceInterface struct {
}

func makeCreateEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createRequestServiceInterface)
		err := s.Create(ctx, req.Name, req.Data)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return w
}

type deleteRequestServiceInterface struct {
	Id uint `json:"id"`
}
type deleteResponseServiceInterface struct {
	A string `json:"a"`
	B string `json:"b"`
}

func makeDeleteEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteRequestServiceInterface)
		a, b, err := s.Delete(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		return deleteResponseServiceInterface{A: a, B: b}, nil
	}
	return w
}

type getRequestServiceInterface struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Fname string  `json:"fname"`
	Price float32 `json:"price"`
	N     int     `json:"n"`
	B     int     `json:"b"`
	C     int     `json:"c"`
}
type getResponseServiceInterface struct {
	Data user.User `json:"data"`
}

func makeGetEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getRequestServiceInterface)
		result, err := s.Get(ctx, req.Id, req.Name, req.Fname, req.Price, req.N, req.B, req.C)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return w
}

func makeGetAllEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return w
}

type testMethodRequestServiceInterface struct {
	Data map[string]interface{} `json:"data"`
	Ss   interface{}            `json:"ss"`
}
type testMethodResponseServiceInterface struct {
	States map[string]map[int][]string `json:"states"`
}

func makeTestMethodEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(testMethodRequestServiceInterface)
		result, err := s.TestMethod(req.Data, req.Ss)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return w
}

type testMethod2RequestServiceInterface struct {
	Ns         string `json:"ns"`
	Utype      string `json:"utype"`
	User       string `json:"user"`
	Restype    string `json:"restype"`
	Resource   string `json:"resource"`
	Permission string `json:"permission"`
}

func makeTestMethod2Endpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(testMethod2RequestServiceInterface)
		err := s.TestMethod2(ctx, req.Ns, req.Utype, req.User, req.Restype, req.Resource, req.Permission)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return w
}

func CreateEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeCreateEndpoint(s), nil, nil

}

func DeleteEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeDeleteEndpoint(s), nil, nil

}

func GetEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeGetEndpoint(s), nil, nil

}

func GetAllEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeGetAllEndpoint(s), nil, nil

}

func TestMethodEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeTestMethodEndpoint(s), nil, nil

}

func TestMethod2EndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	s, err := NewClientJSONRPCServiceInterface(instance)
	if err != nil {
		return nil, nil, err
	}
	return makeTestMethod2Endpoint(s), nil, nil

}
