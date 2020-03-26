package service

import (
	"context"
	"database/sql"
	"fmt"
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

func processLoginRequest(ctx context.Context, db *sql.DB, req *api.LoginRequest) (sessionToken string, err error) {
	var uc api.UserContext
	// Get the login details of user from DB
	err = db.QueryRow("SELECT id, username, password FROM users where username = ?", req.UserName).Scan(&uc.UserID, &uc.UserName, &uc.Password)
	if err != nil {
		return "", errors.New(fmt.Sprintf("unauthorised access, username %v does not exist", req.UserName))
	}

	// compare the password
	if err = bcrypt.CompareHashAndPassword([]byte(uc.Password), []byte(req.Password)); err != nil {
		return "", errors.New("unauthorised access, password does not match ")
	}
	// user authenticated, remove password from user context
	uc.Password = ""
	// update last logged in date of the user
	_, err = db.Exec("update users set last_logged_in_at=? where id=?", time.Now(), uc.UserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to update the last logged in date in DB")
	}

	sessionToken, err = api.SetSessionContext(uc)
	return sessionToken, nil
}

func processCreateListRequest(ctx context.Context, db *sql.DB, req *api.CreateListRequest) (string, error) {
	var uc api.UserContext
	// create a new list
	resp, err := db.Exec("insert Into list (name, description, owner, created_at, last_modified_at, deadline, status) values (?,?,?,?,?,?,?)",
		req.Name, req.Description, req.Owner, time.Now(), time.Now(), time.Now().AddDate(1, 0, 0), req.Status)
	if err != nil {
		return "", errors.Wrap(err, "failed to insert new list in DB")
	}
	lid, _ := resp.LastInsertId()
	// add the current user as a contributor of the list
	_, err = db.Exec("insert into list_contributer (list, user, access_type, valid_until) values (?,?,?,?)",
		lid, req.Owner, "edit", time.Now().AddDate(1, 0, 0))
	if err != nil {
		_,_ = db.Exec("delete from list where id=?", lid)
		return "", errors.Wrap(err, "failed to insert new list-user pair in DB")
	}

	uc.UserID = req.Owner
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	return sessionToken, nil
}
