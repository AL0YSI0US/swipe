// Code generated by Swipe. DO NOT EDIT.

//go:generate swipe
package rest

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	prometheus2 "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/swipe-io/swipe/fixtures/service"
	"github.com/swipe-io/swipe/fixtures/user"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	http2 "net/http"
	"net/url"
	"strconv"
	"time"
)

type createRequestServiceInterface struct {
	Name string `json:"name"`
}

type createResponseServiceInterface struct {
	Err error `json:"-"`
}

func (r createResponseServiceInterface) Failed() (_ error) {
	return r.Err
}

func makeCreateEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createRequestServiceInterface)
		err := s.Create(ctx, req.Name)
		return createResponseServiceInterface{Err: err}, nil
	}
	return w
}

type getRequestServiceInterface struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Fname string  `json:"fname"`
	Price float32 `json:"price"`
	N     int     `json:"n"`
}

type getResponseServiceInterface struct {
	Data user.User `json:"data"`
	Err  error     `json:"-"`
}

func (r getResponseServiceInterface) Failed() (_ error) {
	return r.Err
}

func makeGetEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getRequestServiceInterface)
		data, err := s.Get(ctx, req.Id, req.Name, req.Fname, req.Price, req.N)
		return getResponseServiceInterface{Data: data, Err: err}, nil
	}
	return w
}

type getAllResponseServiceInterface struct {
	Data []user.User `json:"data"`
	Err  error       `json:"-"`
}

func (r getAllResponseServiceInterface) Failed() (_ error) {
	return r.Err
}

func makeGetAllEndpoint(s service.Interface) endpoint.Endpoint {
	w := func(ctx context.Context, request interface{}) (interface{}, error) {
		data, err := s.GetAll(ctx)
		return getAllResponseServiceInterface{Data: data, Err: err}, nil
	}
	return w
}

type loggingMiddlewareServiceInterface struct {
	next   service.Interface
	logger log.Logger
}

func (s *loggingMiddlewareServiceInterface) Create(ctx context.Context, name string) (err error) {
	defer func(now time.Time) {
		s.logger.Log("method", "Create", "took", time.Since(now), "name", name, "err", err)
	}(time.Now())
	return s.next.Create(ctx, name)
}

func (s *loggingMiddlewareServiceInterface) Get(ctx context.Context, id int, name string, fname string, price float32, n int) (data user.User, err error) {
	defer func(now time.Time) {
		s.logger.Log("method", "Get", "took", time.Since(now), "id", id, "name", name, "fname", fname, "price", price, "n", n, "err", err)
	}(time.Now())
	return s.next.Get(ctx, id, name, fname, price, n)
}

func (s *loggingMiddlewareServiceInterface) GetAll(ctx context.Context) (data []user.User, err error) {
	defer func(now time.Time) {
		s.logger.Log("method", "GetAll", "took", time.Since(now), "data", len(data), "err", err)
	}(time.Now())
	return s.next.GetAll(ctx)
}

type instrumentingMiddlewareServiceInterface struct {
	next           service.Interface
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (s *instrumentingMiddlewareServiceInterface) Create(ctx context.Context, name string) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Create").Add(1)
		s.requestLatency.With("method", "Create").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Create(ctx, name)
}

func (s *instrumentingMiddlewareServiceInterface) Get(ctx context.Context, id int, name string, fname string, price float32, n int) (data user.User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Get").Add(1)
		s.requestLatency.With("method", "Get").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Get(ctx, id, name, fname, price, n)
}

func (s *instrumentingMiddlewareServiceInterface) GetAll(ctx context.Context) (data []user.User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "GetAll").Add(1)
		s.requestLatency.With("method", "GetAll").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.GetAll(ctx)
}

func ErrorDecode(code int) (_ error) {
	switch code {
	default:
		return fmt.Errorf("error code %d", code)
	case 403:
		return new(service.ErrUnauthorized)
	}
}

type clientServiceInterfaceOption func(*clientServiceInterface)

func ServiceInterfaceGenericClientOptions(opt ...http.ClientOption) (_ clientServiceInterfaceOption) {
	return func(c *clientServiceInterface) { c.genericClientOption = opt }
}

func ServiceInterfaceCreateClientOptions(opt ...http.ClientOption) (_ clientServiceInterfaceOption) {
	return func(c *clientServiceInterface) { c.createClientOption = opt }
}

func ServiceInterfaceGetClientOptions(opt ...http.ClientOption) (_ clientServiceInterfaceOption) {
	return func(c *clientServiceInterface) { c.getClientOption = opt }
}

func ServiceInterfaceGetAllClientOptions(opt ...http.ClientOption) (_ clientServiceInterfaceOption) {
	return func(c *clientServiceInterface) { c.getAllClientOption = opt }
}

type clientServiceInterface struct {
	createEndpoint            endpoint.Endpoint
	createClientOption        []http.ClientOption
	createEndpointMiddleware  []endpoint.Middleware
	getEndpoint               endpoint.Endpoint
	getClientOption           []http.ClientOption
	getEndpointMiddleware     []endpoint.Middleware
	getAllEndpoint            endpoint.Endpoint
	getAllClientOption        []http.ClientOption
	getAllEndpointMiddleware  []endpoint.Middleware
	genericClientOption       []http.ClientOption
	genericEndpointMiddleware []endpoint.Middleware
}

func (c *clientServiceInterface) Create(ctx context.Context, name string) (err error) {
	resp, err := c.createEndpoint(ctx, createRequestServiceInterface{Name: name})
	if err != nil {
		return err
	}
	response := resp.(createResponseServiceInterface)
	return response.Err
}

func (c *clientServiceInterface) Get(ctx context.Context, id int, name string, fname string, price float32, n int) (data user.User, err error) {
	resp, err := c.getEndpoint(ctx, getRequestServiceInterface{Id: id, Name: name, Fname: fname, Price: price, N: n})
	if err != nil {
		return data, err
	}
	response := resp.(getResponseServiceInterface)
	return response.Data, response.Err
}

func (c *clientServiceInterface) GetAll(ctx context.Context) (data []user.User, err error) {
	resp, err := c.getAllEndpoint(ctx, nil)
	if err != nil {
		return data, err
	}
	response := resp.(getAllResponseServiceInterface)
	return response.Data, response.Err
}

func NewClientRESTServiceInterface(tgt string, opts ...clientServiceInterfaceOption) (service.Interface, error) {
	c := &clientServiceInterface{}
	for _, o := range opts {
		o(c)
	}
	u, err := url.Parse(tgt)
	if err != nil {
		return nil, err
	}
	c.createEndpoint = http.NewClient(
		fasthttp.MethodPost,
		u,
		func(_ context.Context, r *http2.Request, request interface{}) error {
			req, ok := request.(createRequestServiceInterface)
			if !ok {
				return fmt.Errorf("couldn't assert request as createRequestServiceInterface, got %T", request)
			}
			r.Method = fasthttp.MethodPost
			r.URL.Path = fmt.Sprintf("/users")
			data, err := ffjson.Marshal(req)
			if err != nil {
				return fmt.Errorf("couldn't marshal request %T: %s", req, err)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			return nil
		},
		func(_ context.Context, r *http2.Response) (interface{}, error) {
			if statusCode := r.StatusCode; statusCode != http2.StatusOK {
				return nil, ErrorDecode(statusCode)
			}
			var resp createResponseServiceInterface
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			err = ffjson.Unmarshal(b, &resp)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to createResponseServiceInterface: %s", err)
			}
			return resp, nil
		},
		append(c.genericClientOption, c.createClientOption...)...,
	).Endpoint()
	for _, e := range c.createEndpointMiddleware {
		c.createEndpoint = e(c.createEndpoint)
	}
	c.getEndpoint = http.NewClient(
		fasthttp.MethodGet,
		u,
		func(_ context.Context, r *http2.Request, request interface{}) error {
			req, ok := request.(getRequestServiceInterface)
			if !ok {
				return fmt.Errorf("couldn't assert request as getRequestServiceInterface, got %T", request)
			}
			r.Method = fasthttp.MethodGet
			r.URL.Path = fmt.Sprintf("/users/%s", req.Name)
			q := r.URL.Query()
			q.Add("price", strconv.FormatFloat(float64(req.Price), 'w', 2, 32))
			r.URL.RawQuery = q.Encode()
			r.Header.Add("x-num", strconv.FormatInt(int64(req.N), 10))
			return nil
		},
		func(_ context.Context, r *http2.Response) (interface{}, error) {
			if statusCode := r.StatusCode; statusCode != http2.StatusOK {
				return nil, ErrorDecode(statusCode)
			}
			var resp getResponseServiceInterface
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			err = ffjson.Unmarshal(b, &resp)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to getResponseServiceInterface: %s", err)
			}
			return resp, nil
		},
		append(c.genericClientOption, c.getClientOption...)...,
	).Endpoint()
	for _, e := range c.getEndpointMiddleware {
		c.getEndpoint = e(c.getEndpoint)
	}
	c.getAllEndpoint = http.NewClient(
		fasthttp.MethodGet,
		u,
		func(_ context.Context, r *http2.Request, request interface{}) error {
			r.Method = fasthttp.MethodGet
			r.URL.Path = fmt.Sprintf("/users")
			return nil
		},
		func(_ context.Context, r *http2.Response) (interface{}, error) {
			if statusCode := r.StatusCode; statusCode != http2.StatusOK {
				return nil, ErrorDecode(statusCode)
			}
			var resp getAllResponseServiceInterface
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			err = ffjson.Unmarshal(b, &resp)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to getAllResponseServiceInterface: %s", err)
			}
			return resp, nil
		},
		append(c.genericClientOption, c.getAllClientOption...)...,
	).Endpoint()
	for _, e := range c.getAllEndpointMiddleware {
		c.getAllEndpoint = e(c.getAllEndpoint)
	}
	return c, nil
}

type serverServiceInterfaceOption func(*serverServiceInterfaceOpts)
type serverServiceInterfaceOpts struct {
	genericServerOption []http.ServerOption
	createServerOption  []http.ServerOption
	getServerOption     []http.ServerOption
	getAllServerOption  []http.ServerOption
}

func ServiceInterfaceGenericServerOptions(v ...http.ServerOption) (_ serverServiceInterfaceOption) {
	return func(o *serverServiceInterfaceOpts) { o.genericServerOption = v }
}

func ServiceInterfaceCreateServerOptions(opt ...http.ServerOption) (_ serverServiceInterfaceOption) {
	return func(c *serverServiceInterfaceOpts) { c.createServerOption = opt }
}

func ServiceInterfaceGetServerOptions(opt ...http.ServerOption) (_ serverServiceInterfaceOption) {
	return func(c *serverServiceInterfaceOpts) { c.getServerOption = opt }
}

func ServiceInterfaceGetAllServerOptions(opt ...http.ServerOption) (_ serverServiceInterfaceOption) {
	return func(c *serverServiceInterfaceOpts) { c.getAllServerOption = opt }
}

// HTTP REST Transport
type errorWrapper struct {
	Error string `json:"error"`
}

func encodeResponseHTTPServiceInterface(ctx context.Context, w http2.ResponseWriter, response interface{}) error {
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		data, err := ffjson.Marshal(errorWrapper{Error: e.Failed().Error()})
		if err != nil {
			return err
		}
		w.Write(data)
		return nil
	}
	data, err := ffjson.Marshal(response)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

func MakeHandlerRESTServiceInterface(s service.Interface, logger log.Logger, opts ...serverServiceInterfaceOption) (http2.Handler, error) {
	sopt := &serverServiceInterfaceOpts{}
	for _, o := range opts {
		o(sopt)
	}
	s = &loggingMiddlewareServiceInterface{next: s, logger: logger}
	s = &instrumentingMiddlewareServiceInterface{
		next: s,
		requestCount: prometheus2.NewCounterFrom(prometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		requestLatency: prometheus2.NewSummaryFrom(prometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	}
	r := mux.NewRouter()
	r.Methods(fasthttp.MethodPost).Path("/users").Handler(http.NewServer(
		makeCreateEndpoint(s),
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req createRequestServiceInterface
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, fmt.Errorf("couldn't read body for createRequestServiceInterface: %s", err)
			}
			err = ffjson.Unmarshal(b, &req)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("couldn't unmarshal body to createRequestServiceInterface: %s", err)
				return nil, err
			}
			return req, nil
		},
		encodeResponseHTTPServiceInterface,
		append(sopt.genericServerOption, sopt.createServerOption...)...,
	))
	r.Methods(fasthttp.MethodGet).Path("/users/{name:[a-z]}").Handler(http.NewServer(
		makeGetEndpoint(s),
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			var req getRequestServiceInterface
			vars := mux.Vars(r)
			req.Name = vars["name"]
			q := r.URL.Query()
			priceFloat32, err := strconv.ParseFloat(q.Get("price"), 32)
			if err != nil {
				return nil, fmt.Errorf("convert error: %w", err)
			}
			req.Price = float32(priceFloat32)
			nInt, err := strconv.Atoi(r.Header.Get("x-num"))
			if err != nil {
				return nil, fmt.Errorf("convert error: %w", err)
			}
			req.N = int(nInt)
			return req, nil
		},
		encodeResponseHTTPServiceInterface,
		append(sopt.genericServerOption, sopt.getServerOption...)...,
	))
	r.Methods(fasthttp.MethodGet).Path("/users").Handler(http.NewServer(
		makeGetAllEndpoint(s),
		func(ctx context.Context, r *http2.Request) (interface{}, error) {
			return nil, nil
		},
		encodeResponseHTTPServiceInterface,
		append(sopt.genericServerOption, sopt.getAllServerOption...)...,
	))
	return r, nil
}
