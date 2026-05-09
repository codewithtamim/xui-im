package controller

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codewithtamim/xui-im/v2/database/model"
	"github.com/codewithtamim/xui-im/v2/util/json_util"
	"github.com/codewithtamim/xui-im/v2/util/random"
	"github.com/codewithtamim/xui-im/v2/web/service"
	"github.com/codewithtamim/xui-im/v2/web/session"
	"github.com/codewithtamim/xui-im/v2/web/websocket"

	"github.com/gin-gonic/gin"
)

// InboundController handles HTTP requests related to Xray inbounds management.
type InboundController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
}

// NewInboundController creates a new InboundController and sets up its routes.
func NewInboundController(g *gin.RouterGroup) *InboundController {
	a := &InboundController{}
	a.initRouter(g)
	return a
}

// initRouter initializes the routes for inbound-related operations.
func (a *InboundController) initRouter(g *gin.RouterGroup) {

	g.GET("/list", a.getInbounds)
	g.GET("/get/:id", a.getInbound)
	g.GET("/getClientTraffics/:email", a.getClientTraffics)
	g.GET("/getClientTrafficsById/:id", a.getClientTrafficsById)

	g.POST("/add", a.addInbound)
	g.POST("/del/:id", a.delInbound)
	g.POST("/update/:id", a.updateInbound)
	g.POST("/clientIps/:email", a.getClientIps)
	g.POST("/clearClientIps/:email", a.clearClientIps)
	g.POST("/addClient", a.addInboundClient)
	g.POST("/addClientWithConfig", a.addInboundClientWithConfig)
	g.POST("/:id/copyClients", a.copyInboundClients)
	g.POST("/:id/delClient/:clientId", a.delInboundClient)
	g.POST("/updateClient/:clientId", a.updateInboundClient)
	g.POST("/:id/resetClientTraffic/:email", a.resetClientTraffic)
	g.POST("/resetAllTraffics", a.resetAllTraffics)
	g.POST("/resetAllClientTraffics/:id", a.resetAllClientTraffics)
	g.POST("/delDepletedClients/:id", a.delDepletedClients)
	g.POST("/import", a.importInbound)
	g.POST("/onlines", a.onlines)
	g.POST("/lastOnline", a.lastOnline)
	g.POST("/updateClientTraffic/:email", a.updateClientTraffic)
	g.POST("/:id/delClientByEmail/:email", a.delInboundClientByEmail)
}

type CopyInboundClientsRequest struct {
	SourceInboundID int      `form:"sourceInboundId" json:"sourceInboundId"`
	ClientEmails    []string `form:"clientEmails" json:"clientEmails"`
	Flow            string   `form:"flow" json:"flow"`
}

// getInbounds retrieves the list of inbounds for the logged-in user.
// @Summary List all inbounds
// @Tags Inbound
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/list [get]
func (a *InboundController) getInbounds(c *gin.Context) {
	user := session.GetLoginUser(c)
	inbounds, err := a.inboundService.GetInbounds(user.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbounds, nil)
}

// getInbound retrieves a specific inbound by its ID.
// @Summary Get inbound by ID
// @Tags Inbound
// @Produce json
// @Param id path int true "Inbound ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/get/{id} [get]
func (a *InboundController) getInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "get"), err)
		return
	}
	inbound, err := a.inboundService.GetInbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbound, nil)
}

// getClientTraffics retrieves client traffic information by email.
// @Summary Get client traffics by email
// @Tags Inbound
// @Produce json
// @Param email path string true "Client email"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/getClientTraffics/{email} [get]
func (a *InboundController) getClientTraffics(c *gin.Context) {
	email := c.Param("email")
	clientTraffics, err := a.inboundService.GetClientTrafficByEmail(email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.trafficGetError"), err)
		return
	}
	jsonObj(c, clientTraffics, nil)
}

// getClientTrafficsById retrieves client traffic information by inbound ID.
// @Summary Get client traffics by ID
// @Tags Inbound
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/getClientTrafficsById/{id} [get]
func (a *InboundController) getClientTrafficsById(c *gin.Context) {
	id := c.Param("id")
	clientTraffics, err := a.inboundService.GetClientTrafficByID(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.trafficGetError"), err)
		return
	}
	jsonObj(c, clientTraffics, nil)
}

// addInbound creates a new inbound configuration.
// @Summary Add a new inbound
// @Tags Inbound
// @Accept json
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/add [post]
func (a *InboundController) addInbound(c *gin.Context) {
	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), err)
		return
	}
	user := session.GetLoginUser(c)
	inbound.UserId = user.Id
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	} else {
		inbound.Tag = fmt.Sprintf("inbound-%v:%v", inbound.Listen, inbound.Port)
	}

	inbound, needRestart, err := a.inboundService.AddInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), inbound, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
	// Broadcast inbounds update via WebSocket
	inbounds, _ := a.inboundService.GetInbounds(user.Id)
	websocket.BroadcastInbounds(inbounds)
}

// delInbound deletes an inbound configuration by its ID.
// @Summary Delete an inbound
// @Tags Inbound
// @Produce json
// @Param id path int true "Inbound ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/del/{id} [post]
func (a *InboundController) delInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundDeleteSuccess"), err)
		return
	}
	needRestart, err := a.inboundService.DelInbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundDeleteSuccess"), id, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
	// Broadcast inbounds update via WebSocket
	user := session.GetLoginUser(c)
	inbounds, _ := a.inboundService.GetInbounds(user.Id)
	websocket.BroadcastInbounds(inbounds)
}

// updateInbound updates an existing inbound configuration.
// @Summary Update an inbound
// @Tags Inbound
// @Accept json
// @Produce json
// @Param id path int true "Inbound ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/update/{id} [post]
func (a *InboundController) updateInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	inbound := &model.Inbound{
		Id: id,
	}
	err = c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	inbound, needRestart, err := a.inboundService.UpdateInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), inbound, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
	// Broadcast inbounds update via WebSocket
	user := session.GetLoginUser(c)
	inbounds, _ := a.inboundService.GetInbounds(user.Id)
	websocket.BroadcastInbounds(inbounds)
}

// getClientIps retrieves the IP addresses associated with a client by email.
// @Summary Get client IPs
// @Tags Inbound
// @Produce json
// @Param email path string true "Client email"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/clientIps/{email} [post]
func (a *InboundController) getClientIps(c *gin.Context) {
	email := c.Param("email")

	ips, err := a.inboundService.GetInboundClientIps(email)
	if err != nil || ips == "" {
		jsonObj(c, "No IP Record", nil)
		return
	}

	// Prefer returning a normalized string list for consistent UI rendering
	type ipWithTimestamp struct {
		IP        string `json:"ip"`
		Timestamp int64  `json:"timestamp"`
	}

	var ipsWithTime []ipWithTimestamp
	if err := json.Unmarshal([]byte(ips), &ipsWithTime); err == nil && len(ipsWithTime) > 0 {
		formatted := make([]string, 0, len(ipsWithTime))
		for _, item := range ipsWithTime {
			if item.IP == "" {
				continue
			}
			if item.Timestamp > 0 {
				ts := time.Unix(item.Timestamp, 0).Local().Format("2006-01-02 15:04:05")
				formatted = append(formatted, fmt.Sprintf("%s (%s)", item.IP, ts))
				continue
			}
			formatted = append(formatted, item.IP)
		}
		jsonObj(c, formatted, nil)
		return
	}

	var oldIps []string
	if err := json.Unmarshal([]byte(ips), &oldIps); err == nil && len(oldIps) > 0 {
		jsonObj(c, oldIps, nil)
		return
	}

	// If parsing fails, return as string
	jsonObj(c, ips, nil)
}

// clearClientIps clears the IP addresses for a client by email.
func (a *InboundController) clearClientIps(c *gin.Context) {
	email := c.Param("email")

	err := a.inboundService.ClearClientIps(email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.updateSuccess"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.logCleanSuccess"), nil)
}

// addInboundClient adds a new client to an existing inbound.
// @Summary Add client to inbound
// @Tags Inbound
// @Accept json
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/addClient [post]
func (a *InboundController) addInboundClient(c *gin.Context) {
	data := &model.Inbound{}
	err := c.ShouldBind(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientAddSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// addInboundClientWithConfig adds a new client to an existing inbound and returns the generated Xray client config.
// @Summary Add client to inbound and return config
// @Tags Inbound
// @Accept json
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/addClientWithConfig [post]
func (a *InboundController) addInboundClientWithConfig(c *gin.Context) {
	data := &model.Inbound{}
	err := c.ShouldBind(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	// Extract the new client(s) before adding
	newClients, err := a.inboundService.GetClients(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	if len(newClients) == 0 {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("no client provided"))
		return
	}

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	// Get updated inbound
	inbound, err := a.inboundService.GetInbound(data.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	// Determine host
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.GetHeader("X-Real-IP")
	}
	if host == "" {
		var hErr error
		host, _, hErr = net.SplitHostPort(c.Request.Host)
		if hErr != nil {
			host = c.Request.Host
		}
	}

	// Generate config for the first newly added client
	client := newClients[0]
	configObj := generateClientConfig(inbound, client, host)

	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientAddSuccess"), configObj, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// generateClientConfig builds a full Xray client JSON config for a single inbound+client.
func generateClientConfig(inbound *model.Inbound, client model.Client, host string) map[string]any {
	// Base client config template (same as sub/default.json)
	baseConfig := map[string]any{}
	json.Unmarshal([]byte(defaultClientConfig), &baseConfig)

	stream := processStreamSettings(inbound.StreamSettings)

	// Use request host if inbound listen is empty or wildcard
	address := inbound.Listen
	if address == "" || address == "0.0.0.0" || address == "::" || address == "::0" {
		address = host
	}

	var proxyOutbound json_util.RawMessage
	switch inbound.Protocol {
	case "vmess":
		proxyOutbound = genVmessOutbound(address, inbound.Port, stream, client)
	case "vless":
		proxyOutbound = genVlessOutbound(address, inbound.Port, inbound.Settings, stream, client)
	case "trojan", "shadowsocks":
		proxyOutbound = genServerOutbound(address, inbound.Port, inbound.Protocol, inbound.Settings, stream, client)
	case "hysteria", "hysteria2":
		proxyOutbound = genHysteriaOutbound(address, inbound.Port, inbound.Protocol, inbound.Settings, inbound.StreamSettings, stream, client)
	}

	defaultOutbounds := []json_util.RawMessage{
		json_util.RawMessage(`{"tag":"direct","protocol":"freedom","settings":{"domainStrategy":"AsIs","redirect":""}}`),
		json_util.RawMessage(`{"tag":"block","protocol":"blackhole","settings":{"response":{"type":"http"}}}`),
	}

	outbounds := []json_util.RawMessage{proxyOutbound}
	outbounds = append(outbounds, defaultOutbounds...)

	baseConfig["outbounds"] = outbounds
	baseConfig["remarks"] = inbound.Remark + " - " + client.Email

	return baseConfig
}

func processStreamSettings(stream string) map[string]any {
	var streamSettings map[string]any
	json.Unmarshal([]byte(stream), &streamSettings)
	security, _ := streamSettings["security"].(string)
	switch security {
	case "tls":
		streamSettings["tlsSettings"] = processTlsData(streamSettings["tlsSettings"])
	case "reality":
		streamSettings["realitySettings"] = processRealityData(streamSettings["realitySettings"])
	}
	delete(streamSettings, "sockopt")

	network, _ := streamSettings["network"].(string)
	switch network {
	case "tcp":
		streamSettings["tcpSettings"] = removeAcceptProxy(streamSettings["tcpSettings"])
	case "ws":
		streamSettings["wsSettings"] = removeAcceptProxy(streamSettings["wsSettings"])
	case "httpupgrade":
		streamSettings["httpupgradeSettings"] = removeAcceptProxy(streamSettings["httpupgradeSettings"])
	}
	return streamSettings
}

func removeAcceptProxy(setting any) map[string]any {
	netSettings, ok := setting.(map[string]any)
	if ok {
		delete(netSettings, "acceptProxyProtocol")
	}
	return netSettings
}

func processTlsData(tData any) map[string]any {
	tlsMap, _ := tData.(map[string]any)
	if tlsMap == nil {
		return map[string]any{}
	}
	result := make(map[string]any)
	result["serverName"] = tlsMap["serverName"]
	result["alpn"] = tlsMap["alpn"]
	if settings, ok := tlsMap["settings"].(map[string]any); ok {
		if fp, ok := settings["fingerprint"].(string); ok {
			result["fingerprint"] = fp
		}
	}
	return result
}

func processRealityData(rData any) map[string]any {
	rMap, _ := rData.(map[string]any)
	if rMap == nil {
		return map[string]any{}
	}
	result := make(map[string]any)
	result["show"] = false
	if settings, ok := rMap["settings"].(map[string]any); ok {
		result["publicKey"] = settings["publicKey"]
		result["fingerprint"] = settings["fingerprint"]
		result["mldsa65Verify"] = settings["mldsa65Verify"]
	}
	result["spiderX"] = "/" + random.Seq(15)
	if shortIds, ok := rMap["shortIds"].([]any); ok && len(shortIds) > 0 {
		result["shortId"] = shortIds[random.Num(len(shortIds))].(string)
	} else {
		result["shortId"] = ""
	}
	if serverNames, ok := rMap["serverNames"].([]any); ok && len(serverNames) > 0 {
		result["serverName"] = serverNames[random.Num(len(serverNames))].(string)
	} else {
		result["serverName"] = ""
	}
	return result
}

func genVmessOutbound(address string, port int, stream map[string]any, client model.Client) json_util.RawMessage {
	outbound := map[string]any{
		"protocol": "vmess",
		"tag":      "proxy",
	}
	streamBytes, _ := json.MarshalIndent(stream, "", "  ")
	outbound["streamSettings"] = json.RawMessage(streamBytes)

	users := []map[string]any{
		{"id": client.ID, "email": client.Email, "security": client.Security},
	}
	outbound["settings"] = map[string]any{
		"vnext": []map[string]any{
			{"address": address, "port": port, "users": users},
		},
	}
	result, _ := json.MarshalIndent(outbound, "", "  ")
	return result
}

func genVlessOutbound(address string, port int, inboundSettings string, stream map[string]any, client model.Client) json_util.RawMessage {
	outbound := map[string]any{
		"protocol": "vless",
		"tag":      "proxy",
	}
	streamBytes, _ := json.MarshalIndent(stream, "", "  ")
	outbound["streamSettings"] = json.RawMessage(streamBytes)

	var settings map[string]any
	json.Unmarshal([]byte(inboundSettings), &settings)
	encryption, _ := settings["encryption"].(string)

	user := map[string]any{
		"id":         client.ID,
		"level":      8,
		"encryption": encryption,
	}
	if client.Flow != "" {
		user["flow"] = client.Flow
	}

	outbound["settings"] = map[string]any{
		"vnext": []map[string]any{
			{"address": address, "port": port, "users": []map[string]any{user}},
		},
	}
	result, _ := json.MarshalIndent(outbound, "", "  ")
	return result
}

func genServerOutbound(address string, port int, protocol model.Protocol, inboundSettings string, stream map[string]any, client model.Client) json_util.RawMessage {
	outbound := map[string]any{
		"protocol": string(protocol),
		"tag":      "proxy",
	}
	streamBytes, _ := json.MarshalIndent(stream, "", "  ")
	outbound["streamSettings"] = json.RawMessage(streamBytes)

	server := map[string]any{
		"address":  address,
		"port":     port,
		"level":    8,
		"password": client.Password,
	}

	if protocol == model.Shadowsocks {
		var settings map[string]any
		json.Unmarshal([]byte(inboundSettings), &settings)
		method, _ := settings["method"].(string)
		server["method"] = method
		if strings.HasPrefix(method, "2022") {
			if serverPassword, ok := settings["password"].(string); ok {
				server["password"] = fmt.Sprintf("%s:%s", serverPassword, client.Password)
			}
		}
	}

	outbound["settings"] = map[string]any{
		"servers": []map[string]any{server},
	}
	result, _ := json.MarshalIndent(outbound, "", "  ")
	return result
}

func genHysteriaOutbound(address string, port int, protocol model.Protocol, inboundSettings string, inboundStream string, stream map[string]any, client model.Client) json_util.RawMessage {
	outbound := map[string]any{
		"protocol": string(protocol),
		"tag":      "proxy",
	}

	var settings map[string]any
	json.Unmarshal([]byte(inboundSettings), &settings)
	version, _ := settings["version"].(float64)

	outbound["settings"] = map[string]any{
		"version": int(version),
		"address": address,
		"port":    port,
	}

	var s map[string]any
	json.Unmarshal([]byte(inboundStream), &s)
	hyStream, _ := s["hysteriaSettings"].(map[string]any)
	outHyStream := map[string]any{
		"version": int(version),
		"auth":    client.Auth,
	}
	if hyStream != nil {
		if udpIdleTimeout, ok := hyStream["udpIdleTimeout"].(float64); ok {
			outHyStream["udpIdleTimeout"] = int(udpIdleTimeout)
		}
		if finalmask, ok := hyStream["finalmask"].(map[string]any); ok {
			stream["finalmask"] = finalmask
		}
	}
	stream["hysteriaSettings"] = outHyStream
	stream["network"] = "hysteria"
	stream["security"] = "tls"

	streamBytes, _ := json.MarshalIndent(stream, "", "  ")
	outbound["streamSettings"] = json.RawMessage(streamBytes)

	result, _ := json.MarshalIndent(outbound, "", "  ")
	return result
}

// defaultClientConfig is the standard Xray client template (same as sub/default.json).
const defaultClientConfig = `{
  "remarks": "",
  "dns": {
    "tag": "dns_out",
    "queryStrategy": "UseIP",
    "servers": [
      {
        "address": "8.8.8.8",
        "skipFallback": false
      }
    ]
  },
  "inbounds": [
    {
      "port": 10808,
      "protocol": "mixed",
      "settings": {
        "auth": "noauth",
        "udp": true,
        "userLevel": 8
      },
      "sniffing": {
        "destOverride": [
          "http",
          "tls",
          "quic",
          "fakedns"
        ],
        "enabled": true
      },
      "tag": "mixed"
    },
    {
      "port": 10809,
      "protocol": "http",
      "settings": {
        "userLevel": 8
      },
      "tag": "http"
    }
  ],
  "log": {
    "loglevel": "warning"
  },
  "outbounds": [],
  "policy": {
    "levels": {
      "8": {
        "connIdle": 300,
        "downlinkOnly": 1,
        "handshake": 4,
        "uplinkOnly": 1
      }
    },
    "system": {
      "statsOutboundUplink": true,
      "statsOutboundDownlink": true
    }
  },
  "routing": {
    "domainStrategy": "AsIs",
    "rules": [
      {
        "type": "field",
        "network": "tcp,udp",
        "outboundTag": "proxy"
      }
    ]
  },
  "stats": {}
}`

// copyInboundClients copies clients from source inbound to target inbound.
func (a *InboundController) copyInboundClients(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	req := &CopyInboundClientsRequest{}
	err = c.ShouldBind(req)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	if req.SourceInboundID <= 0 {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid source inbound id"))
		return
	}

	result, needRestart, err := a.inboundService.CopyInboundClients(targetID, req.SourceInboundID, req.ClientEmails, req.Flow)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonObj(c, result, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delInboundClient deletes a client from an inbound by inbound ID and client ID.
// @Summary Delete client from inbound
// @Tags Inbound
// @Produce json
// @Param id path int true "Inbound ID"
// @Param clientId path string true "Client ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/{id}/delClient/{clientId} [post]
func (a *InboundController) delInboundClient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	clientId := c.Param("clientId")

	needRestart, err := a.inboundService.DelInboundClient(id, clientId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientDeleteSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// updateInboundClient updates a client's configuration in an inbound.
// @Summary Update inbound client
// @Tags Inbound
// @Accept json
// @Produce json
// @Param clientId path string true "Client ID"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/updateClient/{clientId} [post]
func (a *InboundController) updateInboundClient(c *gin.Context) {
	clientId := c.Param("clientId")

	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	needRestart, err := a.inboundService.UpdateInboundClient(inbound, clientId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientUpdateSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// resetClientTraffic resets the traffic counter for a specific client in an inbound.
// @Summary Reset client traffic
// @Tags Inbound
// @Produce json
// @Param id path int true "Inbound ID"
// @Param email path string true "Client email"
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/{id}/resetClientTraffic/{email} [post]
func (a *InboundController) resetClientTraffic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	email := c.Param("email")

	needRestart, err := a.inboundService.ResetClientTraffic(id, email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetInboundClientTrafficSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// resetAllTraffics resets all traffic counters across all inbounds.
// @Summary Reset all traffics
// @Tags Inbound
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/resetAllTraffics [post]
func (a *InboundController) resetAllTraffics(c *gin.Context) {
	err := a.inboundService.ResetAllTraffics()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	} else {
		a.xrayService.SetToNeedRestart()
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetAllTrafficSuccess"), nil)
}

// resetAllClientTraffics resets traffic counters for all clients in a specific inbound.
func (a *InboundController) resetAllClientTraffics(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	err = a.inboundService.ResetAllClientTraffics(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	} else {
		a.xrayService.SetToNeedRestart()
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetAllClientTrafficSuccess"), nil)
}

// importInbound imports an inbound configuration from provided data.
func (a *InboundController) importInbound(c *gin.Context) {
	inbound := &model.Inbound{}
	err := json.Unmarshal([]byte(c.PostForm("data")), inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	user := session.GetLoginUser(c)
	inbound.Id = 0
	inbound.UserId = user.Id
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	} else {
		inbound.Tag = fmt.Sprintf("inbound-%v:%v", inbound.Listen, inbound.Port)
	}

	for index := range inbound.ClientStats {
		inbound.ClientStats[index].Id = 0
		inbound.ClientStats[index].Enable = true
	}

	needRestart := false
	inbound, needRestart, err = a.inboundService.AddInbound(inbound)
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), inbound, err)
	if err == nil && needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delDepletedClients deletes clients in an inbound who have exhausted their traffic limits.
func (a *InboundController) delDepletedClients(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	err = a.inboundService.DelDepletedClients(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.delDepletedClientsSuccess"), nil)
}

// onlines retrieves the list of currently online clients.
// @Summary Get online clients
// @Tags Inbound
// @Produce json
// @Success 200 {object} entity.Msg
// @Security ApiKeyAuth
// @Router /panel/api/inbounds/onlines [post]
func (a *InboundController) onlines(c *gin.Context) {
	jsonObj(c, a.inboundService.GetOnlineClients(), nil)
}

// lastOnline retrieves the last online timestamps for clients.
func (a *InboundController) lastOnline(c *gin.Context) {
	data, err := a.inboundService.GetClientsLastOnline()
	jsonObj(c, data, err)
}

// updateClientTraffic updates the traffic statistics for a client by email.
func (a *InboundController) updateClientTraffic(c *gin.Context) {
	email := c.Param("email")

	// Define the request structure for traffic update
	type TrafficUpdateRequest struct {
		Upload   int64 `json:"upload"`
		Download int64 `json:"download"`
	}

	var request TrafficUpdateRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	err = a.inboundService.UpdateClientTrafficByEmail(email, request.Upload, request.Download)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientUpdateSuccess"), nil)
}

// delInboundClientByEmail deletes a client from an inbound by email address.
func (a *InboundController) delInboundClientByEmail(c *gin.Context) {
	inboundId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Invalid inbound ID", err)
		return
	}

	email := c.Param("email")
	needRestart, err := a.inboundService.DelInboundClientByEmail(inboundId, email)
	if err != nil {
		jsonMsg(c, "Failed to delete client by email", err)
		return
	}

	jsonMsg(c, "Client deleted successfully", nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}
