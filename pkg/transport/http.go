package transport

import (
	"context"
	"encoding/json"
	"fmt"
	goendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
	"shoppinglist/pkg/api"
	"shoppinglist/pkg/endpoint"
	"strconv"
	"strings"
	"time"
)

// Api resource locators
const (
	// swagger:operation GET /ping ping PingReq
	//
	// Api for checking status of the service
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/PingRes"
	//   "500":
	//     description: StatusInternalServerError
	PingURL = "/ping"

	// swagger:operation POST /signup singup SingupRequest
	//
	// Enrolls a new user in the system
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: SignupRequest
	//   in: body
	//   description: request Parameters for signup
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/SignupRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/SignupResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	SignupURL = "/signup"

	// swagger:operation POST /login login LoginRequest
	//
	// Logs in a registered user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: LoginRequest
	//   in: body
	//   description: request Parameters for login
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/LoginRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/LoginResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	LoginURL = "/login"

	// swagger:operation POST /logout logout LogoutRequest
	//
	// Logs out a logged in user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: LogoutRequest
	//   in: body
	//   description: request Parameters for logout
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/LogoutRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/LogoutResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	LogoutURL = "/logout"

	// swagger:operation POST /list list CreateListRequest
	//
	// Creates a new shopping list for logged in user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: CreateListRequest
	//   in: body
	//   description: request Parameters for create list
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/CreateListRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/CreateListResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	CreateListURL = "/list"

	// swagger:operation GET /list list GetListsRequest
	//
	// Returns all list associated with logged in user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: GetListsRequest
	//   in: body
	//   description: request Parameters fetching lists
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/GetListsRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/GetListsResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	GetListsURL = "/list"

	// swagger:operation POST /item item CreateItemRequest
	//
	// Creates an item in given shopping list
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: CreateItemRequest
	//   in: body
	//   description: request parameters for create item
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/CreateItemRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/CreateItemResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	CreateItemURL = "/item"

	// swagger:operation GET /item item GetListItemsRequest
	//
	// Returns all items of a list associated with logged in user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: GetListItemsRequest
	//   in: body
	//   description: request Parameters fetching items of list
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/GetListItemsRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/GetListItemsResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	GetListItemsURL = "/item"

	// swagger:operation POST /buy buy BuyItemRequest
	//
	// Mark an item as bought by given user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: BuyItemRequest
	//   in: body
	//   description: mark item as bought
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/BuyItemRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/buyItemResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	BuyItemURL = "/buy"

	// swagger:operation POST /share share ShareListRequest
	//
	// Share a list with given user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: ShareListRequest
	//   in: body
	//   description: share a list with user
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/ShareListRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ShareListResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	ShareListURL = "/share"

	// swagger:operation GET /categories categories GetAllCategoriesRequest
	//
	// Get a list of all registered categories
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: GetAllCategoriesRequest
	//   in: body
	//   description: get list of all categories
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/GetAllCategoriesRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/GetAllCategoriesResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	CategoriesURL = "/categories"

	// swagger:operation POST /delete/list/{lid} delete_list DeleteListRequest
	//
	// Mark given list as deleted
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: DeleteListRequest
	//   in: body
	//   description: mark given list as deleted
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/DeleteListRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/DeleteListResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	DeleteListURL = "/delete/list/{lid}"

	// swagger:operation POST /delete/item/{iid} delete_item DeleteItemRequest
	//
	// Mark given item as deleted
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: DeleteItemRequest
	//   in: body
	//   description: mark given item as deleted
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/DeleteItemRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/DeleteItemResponse"
	//   "400":
	//     "$ref": "#/responses/ServiceError"
	//   "500":
	//     "$ref": "#/responses/ServiceError"
	DeleteItemURL = "/delete/item/{iid}"
)

func commonHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// NewHTTPHandler returns an HTTP handler that makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {

	r := mux.NewRouter()
	r.Use(commonHTTPMiddleware)

	r.Methods("GET").Path(PingURL).Handler(httptransport.NewServer(
		endpoints.Ping,
		decodeHTTPPingRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(SignupURL).Handler(httptransport.NewServer(
		endpoints.Signup,
		decodeHTTPSignupRequest,
		encodeResponse,
	))

	r.Methods("GET").Path(LoginURL).Handler(httptransport.NewServer(
		endpoints.Login,
		decodeHTTPLoginRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(LogoutURL).Handler(httptransport.NewServer(
		endpoints.Logout,
		decodeHTTPLogoutRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(CreateListURL).Handler(httptransport.NewServer(
		endpoints.CreateList,
		decodeHTTPCreateListRequest,
		encodeResponse,
	))

	r.Methods("GET").Path(GetListsURL).Handler(httptransport.NewServer(
		endpoints.GetLists,
		decodeHTTPGetListsRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(CreateItemURL).Handler(httptransport.NewServer(
		endpoints.CreateItem,
		decodeHTTPCreateItemRequest,
		encodeResponse,
	))

	r.Methods("GET").Path(GetListItemsURL).Handler(httptransport.NewServer(
		endpoints.GetListItems,
		decodeHTTPGetListItemsRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(BuyItemURL).Handler(httptransport.NewServer(
		endpoints.BuyItem,
		decodeHTTPBuyItemRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(ShareListURL).Handler(httptransport.NewServer(
		endpoints.ShareList,
		decodeHTTPShareListRequest,
		encodeResponse,
	))

	r.Methods("GET").Path(CategoriesURL).Handler(httptransport.NewServer(
		endpoints.GetAllCategories,
		decodeHTTPGetAllCategoriesRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(DeleteListURL).Handler(httptransport.NewServer(
		endpoints.DeleteList,
		decodeHTTPDeleteListRequest,
		encodeResponse,
	))

	r.Methods("POST").Path(DeleteItemURL).Handler(httptransport.NewServer(
		endpoints.DeleteItem,
		decodeHTTPDeleteItemRequest,
		encodeResponse,
	))

	return r
}

func decodeHTTPPingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req api.PingRequest
	//err := json.NewDecoder(r.Body).Decode(&req)
	return req, nil
}

// decodeHTTPSignupRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded signup request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPSignupRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.SignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// decodeHTTPLoginRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded login request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// decodeHTTPLogoutRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded login request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPLogoutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.LogoutRequest
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPCreateListRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded create list request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPCreateListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.CreateListRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.List.Owner.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPGetListsRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded get lists request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPGetListsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.GetListsRequest
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPCreateItemRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded create item request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPCreateItemRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.CreateItemRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.Item.CreatedBy.UserID = uc.UserID
	req.Item.LastModifiedBy.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPGetListItemsRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded get list items request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPGetListItemsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.GetListItemsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPBuyItemRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded buy item request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPBuyItemRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.BuyItemRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPShareListRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded share list request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPShareListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.ShareListRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPGetAllCategoriesRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded share list request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPGetAllCategoriesRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.GetAllCategoriesRequest
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	return req, nil
}

// decodeHTTPDeleteListRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded share list request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPDeleteListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.DeleteListRequest
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	params := mux.Vars(r)
	lid, err := strconv.ParseInt(params["lid"], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid list id in url")
	}
	req.ListID = lid
	return req, nil
}

// decodeHTTPDeleteItemRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded share list request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPDeleteItemRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.DeleteItemRequest
	uc, err := api.GetUserContextFromSession(r)
	if err != nil {
		return nil, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	req.UserID = uc.UserID
	req.SessionToken = uc.SessionToken
	params := mux.Vars(r)
	iid, err := strconv.ParseInt(params["iid"], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid item id in url")
	}
	req.ItemID = iid
	return req, nil
}

func getErrorInfo(err error) (int, string, string) {
	httpStatus := http.StatusInternalServerError
	if strings.Contains(err.Error(), "unauthorised access") {
		httpStatus = http.StatusUnauthorized
	}
	msg := (errors.Cause(err)).Error()
	trace := fmt.Sprintf("%+v", err)
	return httpStatus, msg, trace

}

// swagger:response ServiceError
type ServiceError struct {
	// HTTP Error Codes
	ErrCode int `json:"errcode"`
	// Very Detailed Error Msg describing the stack trace of error
	ErrMsg string `json:"errmsg"`
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	httpStatus, msg, _ := getErrorInfo(err)
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(ServiceError{ErrCode: httpStatus, ErrMsg: msg})
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(goendpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch response.(type) {
	case api.PingResponse:
		p := []byte("pong")
		_, err := w.Write(p)
		return err
	case api.LoginResponse:
		resp := response.(api.LoginResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		return json.NewEncoder(w).Encode(struct{}{})
	case api.CreateListResponse:
		resp := response.(api.CreateListResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		return json.NewEncoder(w).Encode(struct{}{})
	case api.GetListsResponse:
		resp := response.(api.GetListsResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.CreateItemResponse:
		resp := response.(api.CreateItemResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.GetListItemsResponse:
		resp := response.(api.GetListItemsResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.BuyItemResponse:
		resp := response.(api.BuyItemResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.ShareListResponse:
		resp := response.(api.ShareListResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.GetAllCategoriesResponse:
		resp := response.(api.GetAllCategoriesResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.DeleteListResponse:
		resp := response.(api.DeleteListResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	case api.DeleteItemResponse:
		resp := response.(api.DeleteItemResponse)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   resp.SessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})
		resp.SessionToken = ""
		return json.NewEncoder(w).Encode(resp)
	default:
		return json.NewEncoder(w).Encode(response)
	}
}
