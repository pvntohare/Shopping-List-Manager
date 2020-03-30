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

func (mw loggingMiddleware) Logout(ctx context.Context, req api.LogoutRequest) (resp api.LogoutResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "Logout", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input logout req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.Logout(ctx, req)
}

func (mw loggingMiddleware) CreateList(ctx context.Context, req api.CreateListRequest) (resp api.CreateListResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "CreateList", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input CreateList req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.CreateList(ctx, req)
}

func (mw loggingMiddleware) GetLists(ctx context.Context, req api.GetListsRequest) (resp api.GetListsResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "GetLists", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input GetLists req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.GetLists(ctx, req)
}

func (mw loggingMiddleware) CreateItem(ctx context.Context, req api.CreateItemRequest) (resp api.CreateItemResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "CreateItem", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input CreateItem req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.CreateItem(ctx, req)
}

func (mw loggingMiddleware) GetListItems(ctx context.Context, req api.GetListItemsRequest) (resp api.GetListItemsResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "GetListItems", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input GetListItem req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.GetListItems(ctx, req)
}

func (mw loggingMiddleware) BuyItem(ctx context.Context, req api.BuyItemRequest) (resp api.BuyItemResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "BuyItem", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input BuyItem req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.BuyItem(ctx, req)
}

func (mw loggingMiddleware) ShareList(ctx context.Context, req api.ShareListRequest) (resp api.ShareListResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "ShareList", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed for input ShareList req :", req, "error : ", resp.Err)
		}
	}()
	return mw.next.ShareList(ctx, req)
}

func (mw loggingMiddleware) GetAllCategories(ctx context.Context, req api.GetAllCategoriesRequest) (resp api.GetAllCategoriesResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "GetAllCategories", "req", req, "resp", resp)
		} else {
			mw.logger.Log("failed to get categories err: ", resp.Err)
		}
	}()
	return mw.next.GetAllCategories(ctx, req)
}

func (mw loggingMiddleware) DeleteList(ctx context.Context, req api.DeleteListRequest) (resp api.DeleteListResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "DeleteList", "list_id", req.ListID, "resp", resp)
		} else {
			mw.logger.Log("failed for input DeleteList list_id :", req.ListID, "error : ", resp.Err)
		}
	}()
	return mw.next.DeleteList(ctx, req)
}

func (mw loggingMiddleware) DeleteItem(ctx context.Context, req api.DeleteItemRequest) (resp api.DeleteItemResponse) {
	defer func() {
		if resp.Err == nil {
			mw.logger.Log("method", "DeleteItem", "item_id", req.ItemID, "resp", resp)
		} else {
			mw.logger.Log("failed for input DeleteItem item_id :", req.ItemID, "error : ", resp.Err)
		}
	}()
	return mw.next.DeleteItem(ctx, req)
}
