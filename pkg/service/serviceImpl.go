package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"shoppinglist/pkg/api"
	"strings"
	"time"
)

func checkListEditPermission(db *sqlx.DB, query string, ID1 int64, ID2 int64) error {
	var accessType string
	err := db.QueryRow(query, ID1, ID2).Scan(&accessType)
	//rr := db.QueryRow("select access_type from list_contributer where list=? and user=?", params[0], params[1]).Scan(&accessType)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("unauthorised access, user does not have permission to edit the list")
		}
		return errors.Wrapf(err, "error checking list access for user")
	}
	if strings.Compare(accessType, "edit") != 0 {
		return errors.New("unauthorised access, user does not have permission to edit the list")
	}
	return nil
}

func processSingupRequest(ctx context.Context, db *sqlx.DB, req *api.SignupRequest) error {
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

func processLoginRequest(ctx context.Context, db *sqlx.DB, req *api.LoginRequest) (sessionToken string, err error) {
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

func processLogoutRequest(ctx context.Context, _ *sqlx.DB, req *api.LogoutRequest) error {
	err := api.DeleteSessionContext(req.SessionToken)
	if err != nil {
		return errors.Wrapf(err, "failed to delete session from cache while logging out")
	}
	return nil
}

func processCreateListRequest(ctx context.Context, db *sqlx.DB, req *api.CreateListRequest) (string, error) {
	// create a new db transaction
	tx, err := db.Beginx()
	if err != nil {
		return "", errors.Wrapf(err, "failed to begin transaction")
	}
	// create a new list
	resp, err := tx.Exec("insert Into list (name, description, owner, created_at, last_modified_at, deadline, status) values (?,?,?,?,?,?,?)",
		req.List.Name, req.List.Description, req.List.Owner.UserID, time.Now(), time.Now(), time.Now().AddDate(1, 0, 0), req.List.Status)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "failed to insert new list in DB")
	}
	lid, err := resp.LastInsertId()
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "failed to get the id of created list, aborting")
	}
	// add the current user as a contributor of the list
	_, err = tx.Exec("insert into list_contributer (list, user, access_type, valid_until) values (?,?,?,?)",
		lid, req.List.Owner.UserID, api.Edit, time.Now().AddDate(1, 0, 0))
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "failed to insert new list-user pair in DB, aborting")
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "failed to commit db transaction while creating list, aborting")
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

func processGetListsRequest(ctx context.Context, db *sqlx.DB, req *api.GetListsRequest) ([]api.List, string, error) {
	var lists []api.List

	// read lists associated with current user
	query := "select l.id, l.name, l.description, l.owner, l.created_at, l.last_modified_at, l.deadline, " +
		"l.status, lc.access_type, u.username from " +
		"(select id, name, description, owner, created_at, last_modified_at, deadline, status from list) l " +
		"JOIN (select list, access_type from list_contributer where user=?) lc " +
		"JOIN (select id, username from users) u ON l.id=lc.list and u.id=l.owner"
	resp, err := db.Query(query, req.UserID)
	if err != nil {
		return lists, "", errors.Wrapf(err, "failed to query DB for gives user's lists")
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

	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		sessionToken = req.SessionToken
	}
	return lists, sessionToken, nil
}

func processCreateItemRequest(ctx context.Context, db *sqlx.DB, req *api.CreateItemRequest) (string, error) {
	// check user permission to edit the list
	query := "select access_type from list_contributer where list=? and user=?"
	err := checkListEditPermission(db, query, req.Item.ListID, req.Item.CreatedBy.UserID)
	if err != nil {
		return "", err
	}

	// begin a transaction
	tx, err := db.Beginx()
	if err != nil {
		return "", errors.Wrapf(err, "failed to begin transaction")
	}
	//check the list status
	var listStatus string
	err = tx.Get(&listStatus, "select status from list where id=?", req.Item.ListID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return "", errors.New("the mentioned list does not exist")
		}
		return "", errors.Wrapf(err, "error checking list status")
	}
	if strings.Compare(listStatus, api.Todo) != 0 {
		tx.Rollback()
		return "", errors.New(fmt.Sprintf("list status:%v should be %v", listStatus, api.Todo))
	}

	// Check if item category already exists in our DB
	// add it to DB if does not already exist
	if req.Item.Category.ID == 0 {
		// this is a new category, add it to our DB
		resp, err := tx.Exec("insert into category (name, type) values (?,?)", req.Item.Category.Name, req.Item.Category.Type)
		if err != nil {
			tx.Rollback()
			return "", errors.Wrapf(err, "failed to add new category in DB")
		}
		req.Item.Category.ID, _ = resp.LastInsertId()
	}

	// insert the new item
	query = "insert into item (list, title, description, status, category, created_by, last_modified_by, " +
		"created_at, last_modified_at, deadline) values (?,?,?,?,?,?,?,?,?,?)"
	_, err = tx.Exec(query, req.Item.ListID, req.Item.Title, req.Item.Description, api.Todo, req.Item.Category.ID, req.Item.CreatedBy.UserID,
		req.Item.CreatedBy.UserID, time.Now(), time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "failed to add new item")
	}
	err = tx.Commit()
	if err != nil {
		return "", errors.Wrapf(err, "failed to commit transaction for creating item in list")
	}

	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.Item.CreatedBy.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	return sessionToken, nil
}

func processGetListItemsRequest(ctx context.Context, db *sqlx.DB, req *api.GetListItemsRequest) ([]api.Item, string, error) {
	var items []api.Item

	// begin a transaction
	tx, err := db.Beginx()
	if err != nil {
		return items, "", errors.Wrapf(err, "failed to begin a transaction for get list")
	}
	// check if current user have read permission for given list
	var id int64
	err = tx.Get(&id, "select id from list_contributer where user=? and list=?", req.UserID, req.ListID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return items, "", errors.New("current user does not have read access for list")
		}
		return items, "", errors.Wrapf(err, "failed to check list-users connection")
	}

	// read items from give list
	query := "select id, list, title, description, status, category, created_by, last_modified_by, bought_by, created_at," +
		" last_modified_at, bought_at, deadline from item where list=?"
	resp, err := db.Query(query, req.ListID)
	if err != nil {
		tx.Rollback()
		return items, "", errors.Wrapf(err, "failed to read items for given list")
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
		err = tx.Get(&item.CreatedBy.UserName, "select username from users where id=?", item.CreatedBy.UserID)
		if err != nil {
			tx.Rollback()
			return items, "", errors.Wrapf(err, "failed to read username off item creator")
		}
		err = tx.Get(&item.LastModifiedBy.UserName, "select username from users where id=?", item.LastModifiedBy.UserID)
		if err != nil {
			tx.Rollback()
			return items, "", errors.Wrapf(err, "failed to read username off latest item modifier")
		}
		err = tx.Get(&item.Category, "select id, name, type from category where id=?", item.Category.ID)
		if err != nil {
			tx.Rollback()
			return items, "", errors.Wrapf(err, "failed to read category details of item")
		}
		if item.BoughtBy.UserID == 0 {
			items = append(items, item)
			continue
		}
		err = tx.Get(&item.BoughtBy.UserName, "select username from users where id=?", item.BoughtBy.UserID)
		if err != nil {
			tx.Rollback()
			return items, "", errors.Wrapf(err, "failed to read username off item buyer")
		}
		items = append(items, item)
	}
	tx.Commit()
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		sessionToken = req.SessionToken
	}
	return items, sessionToken, nil
}

func processBuyItemRequest(ctx context.Context, db *sqlx.DB, req *api.BuyItemRequest) (string, error) {
	// check if current user had write access to item list
	var (
		listAccessType string
		itemStatus     string
		listID         int64
	)
	err := db.QueryRow("select lc.access_type, i.status, i.list from list_contributer lc, item i "+
		"where i.id=? and lc.user=? and lc.list=i.list", req.ItemID, req.UserID).Scan(&listAccessType, &itemStatus, &listID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("unauthorised access, user does not have permission to edit the list item belongs to")
		}
		return "", errors.Wrapf(err, "failed to read user permission to edit list")
	}
	if strings.Compare(listAccessType, api.Edit) != 0 {
		return "", errors.New("unauthorised access, user have read only permission for list item belongs to")
	}

	// begin a db transaction
	tx, err := db.Beginx()
	if err != nil {
		return "", errors.Wrapf(err, "failed to begin a db transaction for buy item")
	}

	// check list status and item status
	var listStatus string
	err = tx.Get(&listStatus, "select status from list where id=?", listID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return "", errors.New("the mentioned list does not exist")
		}
		return "", errors.Wrapf(err, "failed to read list status")
	}
	if strings.Compare(listStatus, api.Todo) != 0 {
		tx.Rollback()
		return "", errors.New(fmt.Sprintf("list is in %v state, need in todo state", itemStatus))
	}
	if strings.Compare(itemStatus, api.Todo) != 0 {
		tx.Rollback()
		return "", errors.New(fmt.Sprintf("item is in %v state, need in todo state", itemStatus))
	}

	// mark item as bought
	var boughtBy api.User
	boughtBy.UserName = req.UserName
	err = tx.Get(&boughtBy.UserID, "select id from users where username=?", req.UserName)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return "", errors.New(fmt.Sprintf("given buyer username %v is not a registered user", req.UserName))
		}
		return "", errors.Wrapf(err, "failed to read user details for buyer")
	}
	_, err = tx.Exec("update item set status=?, last_modified_by=?, bought_by=?, last_modified_at=?, bought_at=? where id=?",
		api.Bought, req.UserID, boughtBy.UserID, time.Now(), time.Now(), req.ItemID)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "failed to mark item as bought in DB")
	}
	err = tx.Commit()
	if err != nil {
		return "", errors.Wrapf(err, "failed to commit transacton for buy item")
	}
	// TODO update list's last_modified_at field after the item deletion

	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	return sessionToken, nil
}

func processShareListRequest(ctx context.Context, db *sqlx.DB, req *api.ShareListRequest) (string, error) {
	tx, err := db.Beginx()
	if err != nil {
		return "", errors.Wrapf(err, "failed to start db transaction for share list")
	}
	// check if the current user is owner of the list to be shared
	var owner int64
	err = tx.Get(&owner,"select owner from list where id=?", req.ListID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return "", errors.New(fmt.Sprintf("list %v does not exist", req.ListID))
		}
		return "", errors.Wrapf(err, "failed to read list details")
	}
	if owner != req.UserID {
		tx.Rollback()
		return "", errors.New("unauthorised access, only list owner can share the list")
	}

	// share the list
	var uid int64
	err = tx.Get(&uid,"select id from users where username=?", req.UserName)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return "", errors.New(fmt.Sprintf("user %v is not registered", req.UserName))
		}
		return "", errors.Wrapf(err, "failed to read user details")
	}
	_, err = tx.Exec("insert into list_contributer (list, user, access_type, valid_until) values (?,?,?,?)",
		req.ListID, uid, req.AccessType, time.Now().AddDate(1, 0, 0))
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "failed to make an entry in list_contributor table")
	}
	err = tx.Commit()
	if err != nil {
		return "", errors.Wrapf(err, "failed to commit db transaction for share list")
	}

	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}
	return sessionToken, nil
}

func processGetAllCategoriesRequest(ctx context.Context, db *sqlx.DB, req *api.GetAllCategoriesRequest) ([]api.Category, string, error) {
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

func processDeleteListRequest(ctx context.Context, db *sqlx.DB, req *api.DeleteListRequest) (string, error) {
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}

	// check if user has edit permission for list
	query := "select access_type from list_contributer where list=? and user=?"
	err = checkListEditPermission(db, query, req.ListID, req.UserID)
	if err != nil {
		return sessionToken, err
	}

	// mark the list as deleted
	_, err = db.Exec("update list set status=? where id=?", api.Deleted, req.ListID)
	if err != nil {
		return sessionToken, errors.Wrapf(err, "failed to mark list:%v as deleted", req.ListID)
	}

	return sessionToken, nil
}

func processDeleteItemRequest(ctx context.Context, db *sqlx.DB, req *api.DeleteItemRequest) (string, error) {
	// Refresh user session
	var uc api.UserContext
	uc.UserID = req.UserID
	uc.SessionToken = req.SessionToken
	sessionToken, err := api.RefreshSessionContext(uc)
	if err != nil {
		return req.SessionToken, nil
	}

	// check if user has edit permissions for list
	query := "select access_type from list_contributer where user=? and list in (select list from item where id=?)"
	err = checkListEditPermission(db, query, req.UserID, req.ItemID)
	if err != nil {
		return sessionToken, err
	}

	// mark the item as deleted
	_, err = db.Exec("update item set status=? where id=?", api.Deleted, req.ItemID)
	if err != nil {
		return sessionToken, errors.Wrapf(err, "failed to mark item:%v as deleted", req.ItemID)
	}

	return sessionToken, nil
}
