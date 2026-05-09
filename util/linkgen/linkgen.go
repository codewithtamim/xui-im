package linkgen

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/codewithtamim/xui-im/v2/database/model"
	"github.com/codewithtamim/xui-im/v2/util/random"
)

// GenLink generates a subscription link URI for a single inbound client.
func GenLink(inbound *model.Inbound, client model.Client, host string) string {
	switch inbound.Protocol {
	case "vmess":
		return genVmessLink(inbound, client, host)
	case "vless":
		return genVlessLink(inbound, client, host)
	case "trojan":
		return genTrojanLink(inbound, client, host)
	case "shadowsocks":
		return genShadowsocksLink(inbound, client, host)
	case "hysteria", "hysteria2":
		return genHysteriaLink(inbound, client, host)
	}
	return ""
}

func resolveAddress(inbound *model.Inbound, host string) string {
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		return host
	}
	return inbound.Listen
}

func parseStream(stream string) map[string]any {
	var s map[string]any
	json.Unmarshal([]byte(stream), &s)
	return s
}

func searchKey(data any, key string) (any, bool) {
	switch val := data.(type) {
	case map[string]any:
		for k, v := range val {
			if k == key {
				return v, true
			}
			if result, ok := searchKey(v, key); ok {
				return result, true
			}
		}
	case []any:
		for _, v := range val {
			if result, ok := searchKey(v, key); ok {
				return result, true
			}
		}
	}
	return nil, false
}

func genRemark(inbound *model.Inbound, email string) string {
	if inbound.Remark != "" && email != "" {
		return inbound.Remark + " - " + email
	}
	if inbound.Remark != "" {
		return inbound.Remark
	}
	return email
}

func searchHost(headers any) string {
	data, _ := headers.(map[string]any)
	for k, v := range data {
		if strings.EqualFold(k, "host") {
			switch val := v.(type) {
			case []any:
				if len(val) > 0 {
					return val[0].(string)
				}
				return ""
			case string:
				return val
			}
		}
	}
	return ""
}

func applyXhttpPaddingParams(xhttp map[string]any, params url.Values) {
	if xhttp == nil {
		return
	}
	if xpb, ok := xhttp["xPaddingBytes"].(string); ok && len(xpb) > 0 {
		params.Set("x_padding_bytes", xpb)
	}
	extra := map[string]any{}
	if xpb, ok := xhttp["xPaddingBytes"].(string); ok && len(xpb) > 0 {
		extra["xPaddingBytes"] = xpb
	}
	if obfs, ok := xhttp["xPaddingObfsMode"].(bool); ok && obfs {
		extra["xPaddingObfsMode"] = true
		for _, field := range []string{"xPaddingKey", "xPaddingHeader", "xPaddingPlacement", "xPaddingMethod"} {
			if v, ok := xhttp[field].(string); ok && len(v) > 0 {
				extra[field] = v
			}
		}
	}
	if len(extra) > 0 {
		if b, err := json.Marshal(extra); err == nil {
			params.Set("extra", string(b))
		}
	}
}

func applyXhttpPaddingToObj(xhttp map[string]any, obj map[string]any) {
	if xhttp == nil {
		return
	}
	if xpb, ok := xhttp["xPaddingBytes"].(string); ok && len(xpb) > 0 {
		obj["x_padding_bytes"] = xpb
	}
	if obfs, ok := xhttp["xPaddingObfsMode"].(bool); ok && obfs {
		obj["xPaddingObfsMode"] = true
		for _, field := range []string{"xPaddingKey", "xPaddingHeader", "xPaddingPlacement", "xPaddingMethod"} {
			if v, ok := xhttp[field].(string); ok && len(v) > 0 {
				obj[field] = v
			}
		}
	}
}

func genVmessLink(inbound *model.Inbound, client model.Client, host string) string {
	address := resolveAddress(inbound, host)
	obj := map[string]any{
		"v":    "2",
		"add":  address,
		"port": inbound.Port,
		"type": "none",
	}
	stream := parseStream(inbound.StreamSettings)
	network, _ := stream["network"].(string)
	obj["net"] = network

	switch network {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		obj["type"] = typeStr
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			if len(requestPath) > 0 {
				obj["path"] = requestPath[0].(string)
			}
			headers, _ := request["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		obj["type"], _ = header["type"].(string)
		obj["path"], _ = kcp["seed"].(string)
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		obj["path"] = ws["path"].(string)
		if h, ok := ws["host"].(string); ok && len(h) > 0 {
			obj["host"] = h
		} else {
			headers, _ := ws["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		obj["path"] = grpc["serviceName"].(string)
		obj["authority"] = grpc["authority"].(string)
		if grpc["multiMode"].(bool) {
			obj["type"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		obj["path"] = httpupgrade["path"].(string)
		if h, ok := httpupgrade["host"].(string); ok && len(h) > 0 {
			obj["host"] = h
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		obj["path"] = xhttp["path"].(string)
		if h, ok := xhttp["host"].(string); ok && len(h) > 0 {
			obj["host"] = h
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
		obj["mode"], _ = xhttp["mode"].(string)
		applyXhttpPaddingToObj(xhttp, obj)
	}

	security, _ := stream["security"].(string)
	obj["tls"] = security
	if security == "tls" {
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		if len(alpns) > 0 {
			var alpn []string
			for _, a := range alpns {
				alpn = append(alpn, a.(string))
			}
			obj["alpn"] = strings.Join(alpn, ",")
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			obj["sni"], _ = sniValue.(string)
		}
		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				obj["fp"], _ = fpValue.(string)
			}
		}
	}

	obj["id"] = client.ID
	obj["scy"] = client.Security
	obj["ps"] = genRemark(inbound, client.Email)

	jsonStr, _ := json.MarshalIndent(obj, "", "  ")
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
}

func genVlessLink(inbound *model.Inbound, client model.Client, host string) string {
	address := resolveAddress(inbound, host)
	stream := parseStream(inbound.StreamSettings)

	var inboundSettings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &inboundSettings)
	encryption, _ := inboundSettings["encryption"].(string)

	params := url.Values{}
	network, _ := stream["network"].(string)
	params.Set("type", network)
	params.Set("encryption", encryption)

	switch network {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			if len(requestPath) > 0 {
				params.Set("path", requestPath[0].(string))
			}
			headers, _ := request["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
			params.Set("headerType", "http")
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		params.Set("headerType", header["type"].(string))
		params.Set("seed", kcp["seed"].(string))
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		params.Set("path", ws["path"].(string))
		if h, ok := ws["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := ws["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		params.Set("serviceName", grpc["serviceName"].(string))
		params.Set("authority", grpc["authority"].(string))
		if grpc["multiMode"].(bool) {
			params.Set("mode", "multi")
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		params.Set("path", httpupgrade["path"].(string))
		if h, ok := httpupgrade["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		params.Set("path", xhttp["path"].(string))
		if h, ok := xhttp["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
		if mode, ok := xhttp["mode"].(string); ok {
			params.Set("mode", mode)
		}
		applyXhttpPaddingParams(xhttp, params)
	}

	security, _ := stream["security"].(string)
	if security == "tls" {
		params.Set("security", "tls")
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			params.Set("sni", sniValue.(string))
		}
		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				params.Set("fp", fpValue.(string))
			}
		}
		if network == "tcp" && len(client.Flow) > 0 {
			params.Set("flow", client.Flow)
		}
	}

	if security == "reality" {
		params.Set("security", "reality")
		realitySetting, _ := stream["realitySettings"].(map[string]any)
		realitySettings, _ := searchKey(realitySetting, "settings")
		if realitySetting != nil {
			if sniValue, ok := searchKey(realitySetting, "serverNames"); ok {
				sNames, _ := sniValue.([]any)
				params.Set("sni", sNames[random.Num(len(sNames))].(string))
			}
			if pbkValue, ok := searchKey(realitySettings, "publicKey"); ok {
				params.Set("pbk", pbkValue.(string))
			}
			if sidValue, ok := searchKey(realitySetting, "shortIds"); ok {
				shortIds, _ := sidValue.([]any)
				params.Set("sid", shortIds[random.Num(len(shortIds))].(string))
			}
			if fpValue, ok := searchKey(realitySettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok && len(fp) > 0 {
					params.Set("fp", fp)
				}
			}
			if pqvValue, ok := searchKey(realitySettings, "mldsa65Verify"); ok {
				if pqv, ok := pqvValue.(string); ok && len(pqv) > 0 {
					params.Set("pqv", pqv)
				}
			}
			params.Set("spx", "/"+random.Seq(15))
		}
		if network == "tcp" && len(client.Flow) > 0 {
			params.Set("flow", client.Flow)
		}
	}

	if security != "tls" && security != "reality" {
		params.Set("security", "none")
	}

	link := fmt.Sprintf("vless://%s@%s:%d", client.ID, address, inbound.Port)
	u, _ := url.Parse(link)
	u.RawQuery = params.Encode()
	u.Fragment = genRemark(inbound, client.Email)
	return u.String()
}

func genTrojanLink(inbound *model.Inbound, client model.Client, host string) string {
	address := resolveAddress(inbound, host)
	stream := parseStream(inbound.StreamSettings)
	network, _ := stream["network"].(string)
	params := url.Values{}
	params.Set("type", network)

	switch network {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			if len(requestPath) > 0 {
				params.Set("path", requestPath[0].(string))
			}
			headers, _ := request["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
			params.Set("headerType", "http")
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		params.Set("headerType", header["type"].(string))
		params.Set("seed", kcp["seed"].(string))
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		params.Set("path", ws["path"].(string))
		if h, ok := ws["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := ws["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		params.Set("serviceName", grpc["serviceName"].(string))
		params.Set("authority", grpc["authority"].(string))
		if grpc["multiMode"].(bool) {
			params.Set("mode", "multi")
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		params.Set("path", httpupgrade["path"].(string))
		if h, ok := httpupgrade["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		params.Set("path", xhttp["path"].(string))
		if h, ok := xhttp["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
		if mode, ok := xhttp["mode"].(string); ok {
			params.Set("mode", mode)
		}
		applyXhttpPaddingParams(xhttp, params)
	}

	security, _ := stream["security"].(string)
	if security == "tls" {
		params.Set("security", "tls")
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			params.Set("sni", sniValue.(string))
		}
		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				params.Set("fp", fpValue.(string))
			}
		}
	}

	if security == "reality" {
		params.Set("security", "reality")
		realitySetting, _ := stream["realitySettings"].(map[string]any)
		realitySettings, _ := searchKey(realitySetting, "settings")
		if realitySetting != nil {
			if sniValue, ok := searchKey(realitySetting, "serverNames"); ok {
				sNames, _ := sniValue.([]any)
				params.Set("sni", sNames[random.Num(len(sNames))].(string))
			}
			if pbkValue, ok := searchKey(realitySettings, "publicKey"); ok {
				params.Set("pbk", pbkValue.(string))
			}
			if sidValue, ok := searchKey(realitySetting, "shortIds"); ok {
				shortIds, _ := sidValue.([]any)
				params.Set("sid", shortIds[random.Num(len(shortIds))].(string))
			}
			if fpValue, ok := searchKey(realitySettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok && len(fp) > 0 {
					params.Set("fp", fp)
				}
			}
			if pqvValue, ok := searchKey(realitySettings, "mldsa65Verify"); ok {
				if pqv, ok := pqvValue.(string); ok && len(pqv) > 0 {
					params.Set("pqv", pqv)
				}
			}
			params.Set("spx", "/"+random.Seq(15))
		}
		if network == "tcp" && len(client.Flow) > 0 {
			params.Set("flow", client.Flow)
		}
	}

	if security != "tls" && security != "reality" {
		params.Set("security", "none")
	}

	link := fmt.Sprintf("trojan://%s@%s:%d", client.Password, address, inbound.Port)
	u, _ := url.Parse(link)
	u.RawQuery = params.Encode()
	u.Fragment = genRemark(inbound, client.Email)
	return u.String()
}

func genShadowsocksLink(inbound *model.Inbound, client model.Client, host string) string {
	address := resolveAddress(inbound, host)
	stream := parseStream(inbound.StreamSettings)
	network, _ := stream["network"].(string)
	params := url.Values{}
	params.Set("type", network)

	switch network {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			if len(requestPath) > 0 {
				params.Set("path", requestPath[0].(string))
			}
			headers, _ := request["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
			params.Set("headerType", "http")
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		params.Set("headerType", header["type"].(string))
		params.Set("seed", kcp["seed"].(string))
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		params.Set("path", ws["path"].(string))
		if h, ok := ws["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := ws["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		params.Set("serviceName", grpc["serviceName"].(string))
		params.Set("authority", grpc["authority"].(string))
		if grpc["multiMode"].(bool) {
			params.Set("mode", "multi")
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		params.Set("path", httpupgrade["path"].(string))
		if h, ok := httpupgrade["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		params.Set("path", xhttp["path"].(string))
		if h, ok := xhttp["host"].(string); ok && len(h) > 0 {
			params.Set("host", h)
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			h := searchHost(headers)
			if h != "" {
				params.Set("host", h)
			}
		}
		if mode, ok := xhttp["mode"].(string); ok {
			params.Set("mode", mode)
		}
		applyXhttpPaddingParams(xhttp, params)
	}

	security, _ := stream["security"].(string)
	if security == "tls" {
		params.Set("security", "tls")
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			params.Set("sni", sniValue.(string))
		}
		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				params.Set("fp", fpValue.(string))
			}
		}
	}

	if security == "reality" {
		params.Set("security", "reality")
		realitySetting, _ := stream["realitySettings"].(map[string]any)
		realitySettings, _ := searchKey(realitySetting, "settings")
		if realitySetting != nil {
			if sniValue, ok := searchKey(realitySetting, "serverNames"); ok {
				sNames, _ := sniValue.([]any)
				params.Set("sni", sNames[random.Num(len(sNames))].(string))
			}
			if pbkValue, ok := searchKey(realitySettings, "publicKey"); ok {
				params.Set("pbk", pbkValue.(string))
			}
			if sidValue, ok := searchKey(realitySetting, "shortIds"); ok {
				shortIds, _ := sidValue.([]any)
				params.Set("sid", shortIds[random.Num(len(shortIds))].(string))
			}
			if fpValue, ok := searchKey(realitySettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok && len(fp) > 0 {
					params.Set("fp", fp)
				}
			}
			if pqvValue, ok := searchKey(realitySettings, "mldsa65Verify"); ok {
				if pqv, ok := pqvValue.(string); ok && len(pqv) > 0 {
					params.Set("pqv", pqv)
				}
			}
			params.Set("spx", "/"+random.Seq(15))
		}
	}

	if security != "tls" && security != "reality" {
		params.Set("security", "none")
	}

	var inboundSettings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &inboundSettings)
	method, _ := inboundSettings["method"].(string)
	encPart := fmt.Sprintf("%s:%s", method, client.Password)
	if strings.HasPrefix(method, "2022") {
		if serverPassword, ok := inboundSettings["password"].(string); ok {
			encPart = fmt.Sprintf("%s:%s:%s", method, serverPassword, client.Password)
		}
	}

	link := fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(encPart)), address, inbound.Port)
	u, _ := url.Parse(link)
	u.RawQuery = params.Encode()
	u.Fragment = genRemark(inbound, client.Email)
	return u.String()
}

func genHysteriaLink(inbound *model.Inbound, client model.Client, host string) string {
	if !model.IsHysteria(inbound.Protocol) {
		return ""
	}
	stream := parseStream(inbound.StreamSettings)
	auth := client.Auth
	params := url.Values{}
	params.Set("security", "tls")

	tlsSetting, _ := stream["tlsSettings"].(map[string]any)
	alpns, _ := tlsSetting["alpn"].([]any)
	var alpn []string
	for _, a := range alpns {
		alpn = append(alpn, a.(string))
	}
	if len(alpn) > 0 {
		params.Set("alpn", strings.Join(alpn, ","))
	}
	if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
		params.Set("sni", sniValue.(string))
	}

	tlsSettings, _ := searchKey(tlsSetting, "settings")
	if tlsSettings != nil {
		if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
			params.Set("fp", fpValue.(string))
		}
		if insecure, ok := searchKey(tlsSettings, "allowInsecure"); ok {
			if insecure.(bool) {
				params.Set("insecure", "1")
			}
		}
	}

	if finalmask, ok := stream["finalmask"].(map[string]any); ok {
		if udpMasks, ok := finalmask["udp"].([]any); ok {
			for _, m := range udpMasks {
				mask, _ := m.(map[string]any)
				if mask == nil || mask["type"] != "salamander" {
					continue
				}
				settings, _ := mask["settings"].(map[string]any)
				if pw, ok := settings["password"].(string); ok && pw != "" {
					params.Set("obfs", "salamander")
					params.Set("obfs-password", pw)
					break
				}
			}
		}
	}

	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	version, _ := settings["version"].(float64)
	protocol := "hysteria2"
	if int(version) == 1 {
		protocol = "hysteria"
	}

	address := resolveAddress(inbound, host)
	link := fmt.Sprintf("%s://%s@%s:%d", protocol, auth, address, inbound.Port)
	u, _ := url.Parse(link)
	u.RawQuery = params.Encode()
	u.Fragment = genRemark(inbound, client.Email)
	return u.String()
}
