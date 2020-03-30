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
		if err == sql.ErrNoRows {
			return "", errors.New(fmt.Sprintf("unauthorised access, username %v does not exist", req.UserName))
		}
		return "", errors.Wrapf(err, "failed to query DB for given user")
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

func processLogoutRequest(ctx context.Context, _ *sql.DB, req *api.LogoutRequest) error {
	err := api.DeleteSessionContext(req.SessionToken)
	if err != nil {
		return errors.Wrapf(err, "failed to delete session from cache while logging out")
	}
	return nil
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

func processGetListsRequest(ctx context.Context, db *sql.DB, req *api.GetListsRequest) ([]api.List, string, error) {
	var lists []api.List
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		sessionToken = req.SessionToken
	}

	// read lists associated with current user
	query := "select l.id, l.name, l.description, l.owner, l.created_at, l.last_modified_at, l.deadline, " +
		"l.status, lc.access_type, u.username from " +
		"(select id, name, description, owner, created_at, last_modified_at, deadline, status from list) l " +
		"JOIN (select list, access_type from list_contributer where user=?) lc " +
		"JOIN (select id, username from users) u ON l.id=lc.list and u.id=l.owner"
	resp, err := db.Query(query, req.UserID)
	if err != nil {
		return lists, sessionToken, errors.Wrapf(err, "failed to query DB for gives user's lists")
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
	return lists, sessionToken, nil
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
		if err == sql.ErrNoRows {
			return sessionToken, errors.New("unauthorised access, user does not have permission to edit the list")
		}
		return sessionToken, errors.Wrapf(err, "error checking list access for user")
	}
	if strings.Compare(accessType, "edit") != 0 {
		return sessionToken, errors.New("unauthorised access, user does not have permission to edit the list")
	}

	//check if the list status
	var listStatus string
	err = db.QueryRow("select status from list where id=?", req.Item.ListID).Scan(&listStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return sessionToken, errors.New("the mentioned list does not exist")
		}
		return sessionToken, errors.Wrapf(err, "error checking list status")
	}
	if strings.Compare(listStatus, "todo") != 0 {
		return sessionToken, errors.New(fmt.Sprintf("list status:%v should be todo", listStatus))
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

func processGetListItemsRequest(ctx context.Context, db *sql.DB, req *api.GetListItemsRequest) ([]api.Item, string, error) {
	var items []api.Item
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		sessionToken = req.SessionToken
	}

	// check if current user have read permission for given list
	resp, err := db.Query("select id from list_contributer where user=? and list=?", req.UserID, req.ListID)
	if err != nil {
		return items, sessionToken, errors.Wrapf(err, "failed to check list-users connection")
	}
	defer resp.Close()
	if !resp.Next() {
		// we did not find any entry for list-user connection
		return items, sessionToken, errors.New("current user does not have read access for list")
	}

	// read items from give list
	query := "select id, list, title, description, status, category, created_by, last_modified_by, bought_by, created_at," +
		" last_modified_at, bought_at, deadline from item where list=?"
	resp, err = db.Query(query, req.ListID)
	if err != nil {
		return items, sessionToken, errors.Wrapf(err, "failed to read items for given list")
	}
	defer resp.Close()
	for resp.Next() {
		var item api.Item
		var boughtBy, boughtAt interface{}
		resp.Scan(&item.ID, &item.ListID, &item.Title, &item.Description, &item.Status, &item.Category.ID, &item.CreatedBy.UserID, &item.LastModifiedBy.UserID,
			&boughtBy, &item.CreatedAt, &item.LastModifiedAt, &boughtAt, &item.Deadline)
		if boughtBy != nil {
			item.BoughtBy.UserID = boughtBy.(int64)
		}
		if boughtAt != nil {
			item.BoughtAt = boughtAt.(time.Time)
		}
		err = db.QueryRow("select username from users where id=?", item.CreatedBy.UserID).Scan(&item.CreatedBy.UserName)
		if err != nil {
			return items, sessionToken, errors.Wrapf(err, "failed to read username off item creator")
		}
		err = db.QueryRow("select username from users where id=?", item.LastModifiedBy.UserID).Scan(&item.LastModifiedBy.UserName)
		if err != nil {
			return items, sessionToken, errors.Wrapf(err, "failed to read username off latest item modifier")
		}
		err = db.QueryRow("select name, type from category where id=?", item.Category.ID).Scan(&item.Category.Name, &item.Category.Type)
		if err != nil {
			return items, sessionToken, errors.Wrapf(err, "failed to read category details of item")
		}
		if item.BoughtBy.UserID == 0 {
			items = append(items, item)
			continue
		}
		err = db.QueryRow("select username from users where id=?", item.BoughtBy.UserID).Scan(&item.BoughtBy.UserName)
		if err != nil {
			return items, sessionToken, errors.Wrapf(err, "failed to read username off item buyer")
		}
		items = append(items, item)
	}

	return items, sessionToken, nil
}

func processBuyItemRequest(ctx context.Context, db *sql.DB, req *api.BuyItemRequest) (string, error) {
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	// check if current user had write access to item list
	var (
		listAccessType string
		itemStatus     string
		listID         int64
	)
	err = db.QueryRow("select lc.access_type, i.status, i.list from list_contributer lc, item i "+
		"where i.id=? and lc.user=? and lc.list=i.list", req.ItemID, req.UserID).Scan(&listAccessType, &itemStatus, &listID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sessionToken, errors.New("unauthorised access, user does not have permission to edit the list item belongs to")
		}
		return sessionToken, errors.Wrapf(err, "failed to read user permission to edit list")
	}
	if strings.Compare(listAccessType, "edit") != 0 {
		return sessionToken, errors.New("unauthorised access, user have read only permission for list item belongs to")
	}
	if strings.Compare(itemStatus, "todo") != 0 {
		return sessionToken, errors.New(fmt.Sprintf("item is in %v state, need in todo state", itemStatus))
	}

	// mark item as bought
	var boughtBy api.User
	boughtBy.UserName = req.UserName
	err = db.QueryRow("select id from users where username=?", req.UserName).Scan(&boughtBy.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return sessionToken, errors.New(fmt.Sprintf("given buyer username %v is not a registered user", req.UserName))
		}
		return sessionToken, errors.Wrapf(err, "failed to read user details for buyer")
	}
	_, err = db.Exec("update item set status=?, last_modified_by=?, bought_by=?, last_modified_at=?, bought_at=? where id=?",
		"bought", req.UserID, boughtBy.UserID, time.Now(), time.Now(), req.ItemID)
	if err != nil {
		return sessionToken, errors.Wrapf(err, "failed to mark item as bought in DB")
	}
	// TODO update list's last_modified_at field after the item deletion

	return sessionToken, nil
}

func processShareListRequest(ctx context.Context, db *sql.DB, req *api.ShareListRequest) (string, error) {
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}

	// check if the current user is owner of the list to be shared
	var owner int64
	err = db.QueryRow("select owner from list where id=?", req.ListID).Scan(&owner)
	if err != nil {
		if err == sql.ErrNoRows {
			return sessionToken, errors.New(fmt.Sprintf("list %v does not exist", req.ListID))
		}
		return sessionToken, errors.Wrapf(err, "failed to read list details")
	}
	if owner != req.UserID {
		return sessionToken, errors.New("unauthorised access, only list owner can share the list")
	}

	// share the list
	var uid int64
	err = db.QueryRow("select id from users where username=?", req.UserName).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return sessionToken, errors.New(fmt.Sprintf("user %v is not registered", req.UserName))
		}
		return sessionToken, errors.Wrapf(err, "failed to read user details")
	}
	_, err = db.Exec("insert into list_contributer (list, user, access_type, valid_until) values (?,?,?,?)",
		req.ListID, uid, req.AccessType, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return sessionToken, errors.Wrapf(err, "failed to make an entry in list_contributor table")
	}

	return sessionToken, nil
}

func processGetAllCategoriesRequest(ctx context.Context, db *sql.DB, req *api.GetAllCategoriesRequest) ([]api.Category, string, error) {
	var categories []api.Category
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return categories, req.SessionToken, nil
	}

	resp, err := db.Query("select id, name, type from category")
	if err != nil {
		return categories, sessionToken, errors.Wrapf(err, "failed to read categories from system")
	}
	defer resp.Close()
	for resp.Next() {
		var category api.Category
		resp.Scan(&category.ID, &category.Name, &category.Type)
		categories = append(categories, category)
	}
	return categories, sessionToken, nil
}
