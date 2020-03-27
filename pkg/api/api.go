package api

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

var Cache redis.Conn

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
type User struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}
type List struct {
	ID             int64     `json:"list_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Owner          User      `json:"owner"`
	CreatedAt      time.Time `json:"created_at"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	Deadline       time.Time `json:"deadline"`
	Status         string    `json:"status"`
	AccessType     string    `json:"access_type,omitempty"`
	CreatedByMe    bool      `json:"created_by_me,omitempty"`
}

type Item struct {
	ID             int64     `json:"id"`
	ListID         int64     `json:"list_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Category       Category  `json:"category"`
	CreatedBy      User      `json:"created_by"`
	LastModifiedBy User      `json:"last_modified_by"`
	BoughtBy       User      `json:"bought_by"`
	CreatedAt      time.Time `json:"created_at"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	BoughtAt       time.Time `json:"bought_at"`
	Deadline       time.Time `json:"deadline"`
}

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
	SessionToken string
	Err          error `json:"error,omitempty"`
}

// CreateListRequest is request schema for creating new list
// It will create a shopping list for current user
// swagger:model
type CreateListRequest struct {
	SessionToken string
	List         List `json:"list"`
}

// CreateListResponse represents the response struct returned by POST listAPI
// swagger:response CreateListResponse
type CreateListResponse struct {
	SessionToken string
	Err          error `json:"error,omitempty"`
}

// GetListsRequest is request schema for reading the lists
// It will read all the lists for a user
// swagger:model
type GetListsRequest struct {
	SessionToken string
	UserID       int64
}

// GetListsResponse represents the response struct returned by GET listAPI
// swagger:response GetListsResponse
type GetListsResponse struct {
	SessionToken string
	Lists        []List `json:"lists"`
	Err          error  `json:"error,omitempty"`
}

// CreateItemRequest is request schema for creating new item
// It will create an item for given shopping list
// swagger:model
type CreateItemRequest struct {
	SessionToken string
	Item         Item `json:"item"`
}

// CreateItemResponse represents the response struct returned by POST itemAPI
// swagger:response CreateItemResponse
type CreateItemResponse struct {
	SessionToken string
	Err          error `json:"error,omitempty"`
}

// GetListItemsRequest is request schema for reading items
// It will read all the items from a lists for a user
// swagger:model
type GetListItemsRequest struct {
	SessionToken string
	UserID       int64
	ListID       int64 `json:"list_id"`
}

// GetListsResponse represents the response struct returned by GET itemAPI
// swagger:response GetListItemsResponse
type GetListItemsResponse struct {
	SessionToken string
	Items        []Item `json:"items"`
	Err          error  `json:"error,omitempty"`
}
