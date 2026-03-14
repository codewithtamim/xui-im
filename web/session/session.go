// Package session provides session management utilities for the 3x-ui web panel.
// It handles user authentication state, login sessions, and session storage using Gin sessions.
package session

import (
	"encoding/gob"
	"net/http"

	"github.com/mhsanaei/3x-ui/v2/database/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	loginUserKey = "LOGIN_USER"
	defaultPath  = "/"

	// APIUserKey is the gin context key for the user when authenticated via X-API-Key.
	// Used when there is no session (e.g. API key auth) so handlers can still get the user.
	APIUserKey = "API_USER"
)

func init() {
	gob.Register(model.User{})
}

// SetLoginUser stores the authenticated user in the session.
// The user object is serialized and stored for subsequent requests.
func SetLoginUser(c *gin.Context, user *model.User) {
	if user == nil {
		return
	}
	s := sessions.Default(c)
	s.Set(loginUserKey, *user)
}

// SetMaxAge configures the session cookie maximum age in seconds.
// This controls how long the session remains valid before requiring re-authentication.
func SetMaxAge(c *gin.Context, maxAge int) {
	s := sessions.Default(c)
	s.Options(sessions.Options{
		Path:     defaultPath,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetLoginUser retrieves the authenticated user from the session or from API key context.
// Returns nil if no user is logged in or if the session/context data is invalid.
func GetLoginUser(c *gin.Context) *model.User {
	s := sessions.Default(c)
	obj := s.Get(loginUserKey)
	if obj != nil {
		user, ok := obj.(model.User)
		if ok {
			return &user
		}
		s.Delete(loginUserKey)
	}
	// Fallback: user set by API key auth middleware (no session)
	if u, exists := c.Get(APIUserKey); exists && u != nil {
		if user, ok := u.(*model.User); ok {
			return user
		}
	}
	return nil
}

// IsLogin checks if a user is currently authenticated in the session.
// Returns true if a valid user session exists, false otherwise.
func IsLogin(c *gin.Context) bool {
	return GetLoginUser(c) != nil
}

// ClearSession removes all session data and invalidates the session.
// This effectively logs out the user and clears any stored session information.
func ClearSession(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Options(sessions.Options{
		Path:     defaultPath,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
