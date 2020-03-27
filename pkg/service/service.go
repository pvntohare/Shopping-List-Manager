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
	CreateList(ctx context.Context, req api.CreateListRequest) (resp api.CreateListResponse)
	GetLists(ctx context.Context, req api.GetListsRequest) (resp api.GetListsResponse)
	CreateItem(ctx context.Context, req api.CreateItemRequest) (resp api.CreateItemResponse)
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
		resp.Err = errors.Wrap(err, "failed to process signup service")
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
		resp.Err = errors.Wrap(err, "failed to prcoess login service")
		return
	}
	resp.SessionToken = st
	logger.Log("successfully_logged_in_for_user :", req.UserName)
	return
}

func (s basicService) CreateList(ctx context.Context, req api.CreateListRequest) (resp api.CreateListResponse) {
	logger := log.With(s.logger, "method", "CreateListService")
	err := validateCreateListRequest(&req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "request validation failed for create list service")
		return
	}
	st, err := processCreateListRequest(ctx, s.db, &req)
	if err != nil {
		resp.Err = errors.Wrap(err, "failed to process create list service")
	}
	resp.SessionToken = st
	logger.Log("successfully_created_list :", req.List.Name)
	return
}

func (s basicService) GetLists(ctx context.Context, req api.GetListsRequest) (resp api.GetListsResponse) {
	logger := log.With(s.logger, "method", "GetListsService")
	err := validateGetListsRequest(&req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "request validation failed for get lists service")
	}
	lists, st, err := processGetListsRequest(ctx, s.db, &req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "failed to process get lists service")
	}
	resp.Lists = lists
	resp.SessionToken = st
	logger.Log("successfully_got_lists_for_user : ", req.UserID)
	return
}

func (s basicService) CreateItem(ctx context.Context, req api.CreateItemRequest) (resp api.CreateItemResponse) {
	logger := log.With(s.logger, "method", "CreateItemService")
	err := validateCreateItemRequest(&req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "request validation failed for create item service")
	}
	st, err := processCreateItemRequest(ctx, s.db, &req)
	if err != nil {
		resp.Err = errors.Wrapf(err, "failed to process create item service")
	}
	resp.SessionToken = st
	logger.Log("successfully_created_item :", req.Item.Title)
	return
}
