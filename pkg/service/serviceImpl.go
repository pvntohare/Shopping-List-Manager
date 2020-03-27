package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"shoppinglist/pkg/api"
	"strings"
	"time"
)

func processSingupRequest(ctx context.Context, db *sql.DB, req *api.SignupRequest) error {
	query := fmt.Sprintf("SELECT id, username FROM users where username='%v'", req.UserName)
	res, err := db.Query(query)
	if err != nil {
		return errors.Wrapf(err, "failed to read db for username %v", req.UserName)
	}
	defer res.Close()
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
	// create a new list
	resp, err := db.Exec("insert Into list (name, description, owner, created_at, last_modified_at, deadline, status) values (?,?,?,?,?,?,?)",
		req.List.Name, req.List.Description, req.List.Owner.UserID, time.Now(), time.Now(), time.Now().AddDate(1, 0, 0), req.List.Status)
	if err != nil {
		return "", errors.Wrap(err, "failed to insert new list in DB")
	}
	lid, _ := resp.LastInsertId()
	// add the current user as a contributor of the list
	_, err = db.Exec("insert into list_contributer (list, user, access_type, valid_until) values (?,?,?,?)",
		lid, req.List.Owner.UserID, "edit", time.Now().AddDate(1, 0, 0))
	if err != nil {
		_, _ = db.Exec("delete from list where id=?", lid)
		return "", errors.Wrap(err, "failed to insert new list-user pair in DB")
	}
	var uc api.UserContext
	uc.UserID = req.List.Owner.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	return sessionToken, nil
}

func processGetListsRequest(ctx context.Context, db *sql.DB, req *api.GetListsRequest) (lists []api.List, st string, err error) {
	query := "select l.id, l.name, l.description, l.owner, l.created_at, l.last_modified_at, l.deadline, " +
		"l.status, lc.access_type, u.username from " +
		"(select id, name, description, owner, created_at, last_modified_at, deadline, status from list) l " +
		"JOIN (select list, access_type from list_contributer where user=?) lc " +
		"JOIN (select id, username from users) u ON l.id=lc.list and u.id=l.owner"
	resp, err := db.Query(query, req.UserID)
	if err != nil {
		return lists, req.SessionToken, errors.Wrapf(err, "failed to query DB for gives user's lists")
	}
	defer resp.Close()
	for resp.Next() {
		var list api.List
		resp.Scan(&list.ID, &list.Name, &list.Description, &list.Owner.UserID, &list.CreatedAt, &list.LastModifiedAt,
			&list.Deadline, &list.Status, &list.AccessType, &list.Owner.UserName)
		list.CreatedByMe = false
		if list.Owner.UserID == req.UserID {
			list.CreatedByMe = true
		}
		lists = append(lists, list)
	}
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	st, err = api.RefreshSessionContext(uc)
	if err != nil {
		return lists, req.SessionToken, nil
	}
	return
}

func processCreateItemRequest(ctx context.Context, db *sql.DB, req *api.CreateItemRequest) (string, error) {
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.Item.CreatedBy.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}

	// check user permission to edit the list
	var accessType string
	err = db.QueryRow("select access_type from list_contributer where list=? and user=?", req.Item.ListID, req.Item.CreatedBy.UserID).Scan(&accessType)
	if err != nil {
		return sessionToken, errors.Wrapf(err, "error checking list access for user")
	}
	if strings.Compare(accessType, "edit") != 0 {
		return sessionToken, errors.New("unauthorised access, user does not have permission to edit the list")
	}

	// Check if item category already exists in our DB
	// add it to DB if does not already exist
	if req.Item.Category.ID == 0 {
		// this is a new category, add it to our DB
		resp, err := db.Exec("insert into category (name, type) values (?,?)", req.Item.Category.Name, req.Item.Category.Type)
		if err != nil {
			return sessionToken, errors.Wrapf(err, "failed to add new category in DB")
		}
		req.Item.Category.ID, _ = resp.LastInsertId()
	}

	// insert the new item
	query := "insert into item (list, title, description, status, category, created_by, last_modified_by, " +
		"created_at, last_modified_at, deadline) values (?,?,?,?,?,?,?,?,?,?)"
	_, err = db.Exec(query, req.Item.ListID, req.Item.Title, req.Item.Description, "todo", req.Item.Category.ID, req.Item.CreatedBy.UserID,
		req.Item.CreatedBy.UserID, time.Now(), time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil {
		return sessionToken, errors.Wrapf(err, "failed to add new item")
	}

	return sessionToken, nil
}
