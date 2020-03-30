package api

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type UserContext struct {
	UserID       int64
	UserName     string
	Password     string
	SessionToken string
}

const (
	Todo     = "todo"
	Deleted  = "deleted"
	Bought   = "bought"
	Edit     = "edit"
	ReadOnly = "read_only"
)

func GetUserContextFromSession(r *http.Request) (uc UserContext, err error) {
	// obtain the session token from the requests cookies
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			return uc, errors.New("unauthorised access")
		}
		// For any other type of error, return a bad request status
		return uc, errors.New("internal server error")
	}
	sessionToken := c.Value
	uc.SessionToken = sessionToken
	// get the user id from cache
	response, err := Cache.Do("GET", sessionToken)
	if err != nil {
		return uc, errors.New("failed to read user id from cache")
	}
	if response == nil {
		return uc, errors.New("unauthorised access")
	}
	uid := string(response.([]byte))
	uc.UserID, err = strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return uc, errors.Wrap(err, "unauthorised access,could not read userid from cache")
	}
	return uc, nil
}

func SetSessionContext(uc UserContext) (sessionToken string, err error) {
	// Create a new random session token
	sessionToken = uuid.New().String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = Cache.Do("SETEX", sessionToken, "120", uc.UserID)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		return "", errors.Wrapf(err, "failed to set the session for username %v", uc.UserName)
	}
	return
}

func DeleteSessionContext(sessionToken string) error {
	// Delete the older session token
	_, err := Cache.Do("DEL", sessionToken)
	if err != nil {
		return errors.Wrap(err, "failed to delete old session")
	}
	return nil
}

func RefreshSessionContext(uc UserContext) (string, error) {
	newSessionToken, err := SetSessionContext(uc)
	if err != nil {
		return "", errors.Wrap(err, "failed to refresh user session")
	}

	// Delete the older session token
	err = DeleteSessionContext(uc.SessionToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to delete old session while refreshing user session")
	}
	return newSessionToken, nil
}
