package api

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

var Cache redis.Conn

// PingRequest api is used for checking health of the service
// swagger:model
type PingRequest struct {
	//none
}

// PingResponse is the response of pingAPI
// swagger:response PingResponse
type PingResponse struct {
	Err error `json:"error,omitempty"`
}

// SignupRequest is request schema for signup request
// It adds a new user under given username with given user details
// swagger:model
type SignupRequest struct {
	UserName       string    `json:"user_name"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLoggedInAt time.Time `json:"last_logged_in_at"`
	Status         string    `json:"status"`
}

// SignupResponse represents the response struct returned by singupAPI
// swagger:response SignupResponse
type SignupResponse struct {
	Err error `json:"error,omitempty"`
}

// LoginRequest will authorize a user with given username and password
// The user with given username should have already been registered
// swagger:model
type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// LoginResponse represents the response struct returned by loginAPI
// swagger:response LoginResponse
type LoginResponse struct {
	SessionToke string
	Err         error `json:"error,omitempty"`
}

// CreateListRequest is request schema for creating new list
// It will create a shopping list for current user
// swagger:model
type CreateListRequest struct {
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Owner            int       `json:"owner"`
	Created_at       time.Time `json:"created_at"`
	Last_modified_at time.Time `json:"last_modified_at"`
	Deadline         time.Time `json:"deadline"`
	Status           string    `json:"status"`
}

// CreateListResponse represents the response struct returned by POST listAPI
// swagger:response CreateListResponse
type CreateListResponse struct {
	Err error `json:"error,omitempty"`
}
