package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"shoppinglist/pkg/api"
)

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

func processLoginRequest(ctx context.Context, req *api.LoginRequest) error {
	fmt.Println("user logged in")
	return nil
}
