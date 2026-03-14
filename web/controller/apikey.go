package controller

import (
	"strconv"

	"github.com/mhsanaei/3x-ui/v2/web/service"

	"github.com/gin-gonic/gin"
)

type ApiKeyController struct {
	apiKeyService service.ApiKeyService
}

func NewApiKeyController(g *gin.RouterGroup) *ApiKeyController {
	a := &ApiKeyController{}
	a.initRouter(g)
	return a
}

func (a *ApiKeyController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/apikeys")

	g.POST("/list", a.list)
	g.POST("/create", a.create)
	g.POST("/delete/:id", a.delete)
	g.POST("/toggle/:id", a.toggle)
}

// list godoc
// @Summary List all API keys
// @Tags ApiKey
// @Produce json
// @Success 200 {object} entity.Msg
// @Router /panel/setting/apikeys/list [post]
func (a *ApiKeyController) list(c *gin.Context) {
	keys, err := a.apiKeyService.ListApiKeys()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, keys, nil)
}

type createApiKeyForm struct {
	Description string `json:"description" form:"description" binding:"required"`
	ExpiryTime  int64  `json:"expiryTime" form:"expiryTime"`
}

// create godoc
// @Summary Create a new API key
// @Tags ApiKey
// @Accept json
// @Produce json
// @Param body body createApiKeyForm true "API key details"
// @Success 200 {object} entity.Msg
// @Router /panel/setting/apikeys/create [post]
func (a *ApiKeyController) create(c *gin.Context) {
	form := &createApiKeyForm{}
	if err := c.ShouldBind(form); err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.descriptionRequired"), err)
		return
	}
	key, err := a.apiKeyService.CreateApiKey(form.Description, form.ExpiryTime)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.createSuccess"), err)
		return
	}
	jsonObj(c, key, nil)
}

// delete godoc
// @Summary Delete an API key
// @Tags ApiKey
// @Produce json
// @Param id path int true "API Key ID"
// @Success 200 {object} entity.Msg
// @Router /panel/setting/apikeys/delete/{id} [post]
func (a *ApiKeyController) delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.deleteSuccess"), err)
		return
	}
	err = a.apiKeyService.DeleteApiKey(id)
	jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.deleteSuccess"), err)
}

// toggle godoc
// @Summary Toggle an API key enabled/disabled
// @Tags ApiKey
// @Produce json
// @Param id path int true "API Key ID"
// @Success 200 {object} entity.Msg
// @Router /panel/setting/apikeys/toggle/{id} [post]
func (a *ApiKeyController) toggle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.toggleSuccess"), err)
		return
	}
	err = a.apiKeyService.ToggleApiKey(id)
	jsonMsg(c, I18nWeb(c, "pages.settings.apiKeys.toggleSuccess"), err)
}
