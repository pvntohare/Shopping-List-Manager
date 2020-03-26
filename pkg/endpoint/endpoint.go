package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"shoppinglist/pkg/api"
	"shoppinglist/pkg/service"
)

type Endpoints struct {
	Ping   endpoint.Endpoint
	Signup endpoint.Endpoint
	Login  endpoint.Endpoint
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

	return Endpoints{
		Ping:   pingEndpoint,
		Signup: singupEndpoint,
		Login:  loginEndpoint,
	}
}

func MakePingEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.PingRequest)
		response = s.Ping(ctx, req)
		return
	}
}

func MakeSignupEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.SignupRequest)
		response = s.Signup(ctx, req)
		return
	}
}

func MakeLoginEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.LoginRequest)
		return s.Login(ctx, req), nil
	}
}
