package api

import "time"

type SignupRequest struct {
	UserID         string    `json:"userid"`
	UserName       string    `json:"user_name"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLoggedInAt time.Time `json:"last_logged_in_at"`
	Status         string    `json:"status"`
}

type SignupResponse struct {
	Err error `json:"error,omitempty"`
}

type LoginRequest struct {
	UserName string `json:"user_name"`
}

type LoginResponse struct {
	Err error `json:"error,omitempty"`
}
