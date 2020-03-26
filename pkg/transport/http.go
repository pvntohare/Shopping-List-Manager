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
)

// Api resource locators
const (
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
	// Logs in a already signed up user
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

	// swagger:operation GET /healthz healthz HealthzReq
	//
	// Api for checking status of the service
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/HealthzRes"
	//   "500":
	//     description: StatusInternalServerError
	HealthzURL = "/healthz"
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

	return r
}

// decodeHTTPSignupRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded index request from the HTTP request body. Primarily useful in a
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
// JSON-encoded index request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req api.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func getErrorInfo(err error) (int, string, string) {
	httpStatus := http.StatusInternalServerError
	msg := (errors.Cause(err)).Error()
	trace := fmt.Sprintf("%+v", err)
	/*	if t, ok := cause.(stackTracer); ok {
			fmt.Sprintf("error = %v", cause)
			st := t.StackTrace()
			trace = fmt.Sprintf("%+v", st)
		}
	*/
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
	return json.NewEncoder(w).Encode(response)
}
