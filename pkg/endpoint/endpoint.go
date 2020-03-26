package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"shoppinglist/pkg/api"
	"shoppinglist/pkg/service"
)

type Endpoints struct {
	Ping       endpoint.Endpoint
	Signup     endpoint.Endpoint
	Login      endpoint.Endpoint
	CreateList endpoint.Endpoint
}

func New(s service.Service, logger log.Logger) Endpoints {
	var pingEndpoint endpoint.Endpoint
	{
		pingEndpoint = MakePingEndpoint(s)
	}

	var singupEndpoint endpoint.Endpoint
	{
		singupEndpoint = MakeSignupEndpoint(s)
		singupEndpoint = LoggingMiddleware(log.With(logger, "method", "Signup"))(singupEndpoint)
	}

	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = MakeLoginEndpoint(s)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method", "Login"))(loginEndpoint)
	}

	var createListEndpoint endpoint.Endpoint
	{
		createListEndpoint = MakeCreateListEndpoint(s)
		createListEndpoint = LoggingMiddleware(log.With(logger, "method", "CreateList"))(createListEndpoint)
	}

	return Endpoints{
		Ping:       pingEndpoint,
		Signup:     singupEndpoint,
		Login:      loginEndpoint,
		CreateList: createListEndpoint,
	}
}

func MakePingEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.PingRequest)
		return s.Ping(ctx, req), nil
	}
}

func MakeSignupEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.SignupRequest)
		return s.Signup(ctx, req), nil
	}
}

func MakeLoginEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.LoginRequest)
		return s.Login(ctx, req), nil
	}
}

func MakeCreateListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.CreateListRequest)
		return s.CreateList(ctx, req), nil
	}
}
