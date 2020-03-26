package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"shoppinglist/pkg/api"
	"time"
)

type userLogin struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func processSingupRequest(ctx context.Context, db *sql.DB, req *api.SignupRequest) error {
	query := fmt.Sprintf("SELECT id, username FROM users where username='%v'", req.UserName)
	res, err := db.Query(query)
	if err != nil {
		return errors.Wrapf(err, "failed to read db for username %v", req.UserName)
	}
	if res.Next() {
		return errors.New(fmt.Sprintf("username %v not available", req.UserName))
	}
	_, err = db.Exec("insert Into users (username, full_name, email, password, created_at, updated_at, last_logged_in_at, status) values (?,?,?,?,?,?,?,?)",
		req.UserName, req.FullName, req.Email, req.Password, req.CreatedAt, req.UpdatedAt, req.LastLoggedInAt, req.Status)
	if err != nil {
		return errors.Wrap(err, "failed to insert user in DB")
	}
	return nil
}

func processLoginRequest(ctx context.Context, db *sql.DB, req *api.LoginRequest) (sesstionToken string, err error) {
	var user userLogin
	// Get the login details of user from DB
	err = db.QueryRow("SELECT id, username, password FROM users where username = ?", req.UserName).Scan(&user.UserID, &user.UserName, &user.Password)
	if err != nil {
		return "", errors.New(fmt.Sprintf("unauthorised access, username %v does not exist", req.UserName))
	}

	// compare the password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("unauthorised access, password does not match ")
	}

	// update last logged in date of the user
	_, err = db.Exec("update users set last_logged_in_at=? where id=?", time.Now(), user.UserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to update the last logged in date in DB")
	}

	// Create a new random session token
	sessionToken := uuid.New().String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = api.Cache.Do("SETEX", sessionToken, "120", req.UserName)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		return "", errors.Wrapf(err, "failed to set the session for username %v", req.UserName)
	}
	return sessionToken, nil
}
