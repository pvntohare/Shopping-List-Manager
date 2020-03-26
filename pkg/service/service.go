package service

import (
	"context"
	"database/sql"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"shoppinglist/pkg/api"
	"time"
)

type Config struct {
	DBConn     string `json:dbconn`
	DBPort     string `json:dbport`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

type Info struct {
	ServiceName string `json:"servicename"`
	Version     string `json:"version"`
	BuildInfo   string `json:"buildinfo"`
	BuildTime   string `json:"buildtime"`
	StartTime   string `json:"starttime"`
}

type basicService struct {
	db           *sql.DB
	logger       log.Logger
	ConfigObject *Config
	serviceInfo  *Info
}

type Service interface {
	Ping(ctx context.Context, req api.PingRequest) (resp api.PingResponse)
	Signup(ctx context.Context, req api.SignupRequest) (resp api.SignupResponse)
	Login(ctx context.Context, req api.LoginRequest) (resp api.LoginResponse)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(db *sql.DB, logger log.Logger, configObject *Config, serviceInfo *Info /*other middlewares here*/) Service {
	var svc Service
	{
		svc = basicService{db, logger, configObject, serviceInfo}
		svc = LoggingMiddleware(logger)(svc)
		/*chain other middleware here*/
	}
	return svc
}

func (s basicService) Ping(ctx context.Context, req api.PingRequest) (resp api.PingResponse) {
	return api.PingResponse{}
}

func (s basicService) Signup(ctx context.Context, req api.SignupRequest) (resp api.SignupResponse) {
	logger := log.With(s.logger, "method", "SingupService")
	err := validateSignupRequest(&req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "request validation failed for signup service")
		return
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	// TBD use nil equivalent date
	req.LastLoggedInAt = time.Now()
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), 8)
	if err != nil {
		resp.Err = errors.Wrapf(err, "failed to encrypt the password")
		return
	}
	req.Password = string(encryptedPass)
	//store the user in DB
	err = processSingupRequest(ctx, s.db, &req)
	if err != nil {
		resp.Err = errors.Wrap(err, "signup service failed")
		return
	}
	logger.Log("successfully_create_user :", req.UserName)
	return
}

func (s basicService) Login(ctx context.Context, req api.LoginRequest) (resp api.LoginResponse) {
	logger := log.With(s.logger, "method", "LoginService")
	err := validateLoginRequest(&req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "request validation failed for login service")
		return
	}
	st, err := processLoginRequest(ctx, s.db, &req)
	if err != nil {
		resp.Err = errors.Wrap(err, "login service failed")
		return
	}
	resp.SessionToke = st
	logger.Log("successfully_logged_in_for_user :", req.UserName)
	return
}
