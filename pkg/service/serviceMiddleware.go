package service

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"shoppinglist/pkg/api"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Ping(ctx context.Context, req api.PingRequest) (resp api.PingResponse) {
	defer func() {
		if resp.Err != nil {
			err1 := errors.Wrap(resp.Err, "failure in ping request")
			mw.logger.Log("ping_failed", err1)
		}
	}()
	return mw.next.Ping(ctx, req)
}

func (mw loggingMiddleware) Signup(ctx context.Context, req api.SignupRequest) (resp api.SignupResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "Signup", "req", req.UserName, "resp", resp)
		} else {
			mw.logger.Log("failed for input signup req :", req.UserName, "error : ", resp.Err)
		}
	}()
	return mw.next.Signup(ctx, req)
}

func (mw loggingMiddleware) Login(ctx context.Context, req api.LoginRequest) (resp api.LoginResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "Login", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input login req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.Login(ctx, req)
}
