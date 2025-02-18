package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"shoppinglist/pkg/api"
	"shoppinglist/pkg/service"
)

type Endpoints struct {
	Ping             endpoint.Endpoint
	Signup           endpoint.Endpoint
	Login            endpoint.Endpoint
	CreateList       endpoint.Endpoint
	GetLists         endpoint.Endpoint
	CreateItem       endpoint.Endpoint
	GetListItems     endpoint.Endpoint
	BuyItem          endpoint.Endpoint
	ShareList        endpoint.Endpoint
	Logout           endpoint.Endpoint
	GetAllCategories endpoint.Endpoint
	DeleteList       endpoint.Endpoint
	DeleteItem       endpoint.Endpoint
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

	var logoutEndpoint endpoint.Endpoint
	{
		logoutEndpoint = MakeLogoutEndpoint(s)
		logoutEndpoint = LoggingMiddleware(log.With(logger, "method", "Logout"))(logoutEndpoint)
	}

	var createListEndpoint endpoint.Endpoint
	{
		createListEndpoint = MakeCreateListEndpoint(s)
		createListEndpoint = LoggingMiddleware(log.With(logger, "method", "CreateList"))(createListEndpoint)
	}

	var getListsEndpoint endpoint.Endpoint
	{
		getListsEndpoint = MakeGetListsEndpoint(s)
		getListsEndpoint = LoggingMiddleware(log.With(logger, "method", "GetLists"))(getListsEndpoint)
	}

	var createItemEndpoint endpoint.Endpoint
	{
		createItemEndpoint = MakeCreateItemEndpoint(s)
		createItemEndpoint = LoggingMiddleware(log.With(logger, "method", "GetItem"))(createItemEndpoint)
	}

	var getListItemsEndpoint endpoint.Endpoint
	{
		getListItemsEndpoint = MakeGetListItemsEndpoint(s)
		getListItemsEndpoint = LoggingMiddleware(log.With(logger, "method", "GetItem"))(getListItemsEndpoint)
	}

	var buyItemEndpoint endpoint.Endpoint
	{
		buyItemEndpoint = MakeBuyItemEndpoint(s)
		buyItemEndpoint = LoggingMiddleware(log.With(logger, "method", "BuyItem"))(buyItemEndpoint)
	}

	var shareListEndpoint endpoint.Endpoint
	{
		shareListEndpoint = MakeShareListEndpoint(s)
		shareListEndpoint = LoggingMiddleware(log.With(logger, "method", "ShareList"))(shareListEndpoint)
	}

	var getAllCategoriesEndpoint endpoint.Endpoint
	{
		getAllCategoriesEndpoint = MakeGetAllCategoriesEndpoint(s)
		getAllCategoriesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetAllCategories"))(getAllCategoriesEndpoint)
	}

	var deleteListEndpoint endpoint.Endpoint
	{
		deleteListEndpoint = MakeDeleteListEndpoint(s)
		deleteListEndpoint = LoggingMiddleware(log.With(logger, "method", "DeleteList"))(deleteListEndpoint)
	}

	var deleteItemEndpoint endpoint.Endpoint
	{
		deleteItemEndpoint = MakeDeleteItemEndpoint(s)
		deleteItemEndpoint = LoggingMiddleware(log.With(logger, "method", "DeleteItem"))(deleteItemEndpoint)
	}

	return Endpoints{
		Ping:             pingEndpoint,
		Signup:           singupEndpoint,
		Login:            loginEndpoint,
		CreateList:       createListEndpoint,
		GetLists:         getListsEndpoint,
		CreateItem:       createItemEndpoint,
		GetListItems:     getListItemsEndpoint,
		BuyItem:          buyItemEndpoint,
		ShareList:        shareListEndpoint,
		Logout:           logoutEndpoint,
		GetAllCategories: getAllCategoriesEndpoint,
		DeleteList:       deleteListEndpoint,
		DeleteItem:       deleteItemEndpoint,
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

func MakeLogoutEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.LogoutRequest)
		return s.Logout(ctx, req), nil
	}
}

func MakeCreateListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.CreateListRequest)
		return s.CreateList(ctx, req), nil
	}
}

func MakeGetListsEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.GetListsRequest)
		return s.GetLists(ctx, req), nil
	}
}

func MakeCreateItemEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.CreateItemRequest)
		return s.CreateItem(ctx, req), nil
	}
}

func MakeGetListItemsEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.GetListItemsRequest)
		return s.GetListItems(ctx, req), nil
	}
}

func MakeBuyItemEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.BuyItemRequest)
		return s.BuyItem(ctx, req), nil
	}
}

func MakeShareListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.ShareListRequest)
		return s.ShareList(ctx, req), nil
	}
}

func MakeGetAllCategoriesEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.GetAllCategoriesRequest)
		return s.GetAllCategories(ctx, req), nil
	}
}

func MakeDeleteListEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.DeleteListRequest)
		return s.DeleteList(ctx, req), nil
	}
}

func MakeDeleteItemEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(api.DeleteItemRequest)
		return s.DeleteItem(ctx, req), nil
	}
}
