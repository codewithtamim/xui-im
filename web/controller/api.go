package controller

import (
	"net/http"

	"github.com/codewithtamim/xui-im/v2/web/service"
	"github.com/codewithtamim/xui-im/v2/web/session"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// APIController handles the main API routes for the xui-im panel, including inbounds and server management.
type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
	apiKeyService     service.ApiKeyService
	userService       service.UserService
	swaggerEnabled    bool
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(g *gin.RouterGroup, customGeo *service.CustomGeoService, swaggerEnabled bool) *APIController {
	a := &APIController{swaggerEnabled: swaggerEnabled}
	a.initRouter(g, customGeo)
	return a
}

// checkAPIAuth is a middleware that authenticates requests via session cookie
// or X-API-Key header, returning 404 for unauthenticated requests to hide
// the existence of API endpoints from unauthorized users.
// When API key auth succeeds, the first user is set in context so handlers
// that use session.GetLoginUser(c) work correctly.
func (a *APIController) checkAPIAuth(c *gin.Context) {
	if session.IsLogin(c) {
		c.Next()
		return
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" && a.apiKeyService.ValidateApiKey(apiKey) {
		user, err := a.userService.GetFirstUser()
		if err == nil && user != nil {
			c.Set(session.APIUserKey, user)
		}
		c.Next()
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup, customGeo *service.CustomGeoService) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	NewCustomGeoController(api.Group("/custom-geo"), customGeo)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)

	// API docs - always registered; returns 404 when disabled, serves Swagger when enabled
	api.GET("/docs/*any", a.serveDocs)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}

// serveDocs serves Swagger UI when enabled, or returns 404 HTML when disabled.
func (a *APIController) serveDocs(c *gin.Context) {
	if !a.swaggerEnabled {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusNotFound, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>404 Not Found</title></head><body style="font-family:system-ui,sans-serif;max-width:600px;margin:4rem auto;padding:2rem;text-align:center"><h1>404 Not Found</h1><p>API Documentation is disabled. Enable it in Panel Settings → API Keys.</p></body></html>`)
		return
	}
	ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
}
