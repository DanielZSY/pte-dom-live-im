package appid

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	HeaderAppID = "AppId"
	HeaderToken = "Token"
	HeaderExtend = "Extend"
	HeaderUserID = "UserId"

	MetadataKey = "appid"
)

var httpHeaderKeys = []string{"AppId", "appId", "appid", "SystemId", "systemId"}

var metadataKeys = []string{"appid", "AppId", "app-id", "systemid", "SystemId"}

// FromHTTP 解析 HTTP Header 中的 appId（兼容旧 SystemId）
func FromHTTP(r *http.Request) string {
	for _, key := range httpHeaderKeys {
		if v := strings.TrimSpace(r.Header.Get(key)); v != "" {
			return v
		}
	}
	return appIDFromQuery(r)
}

func appIDFromQuery(r *http.Request) string {
	for _, key := range []string{"appId", "app_id", "AppID", "AppId"} {
		if v := strings.TrimSpace(r.FormValue(key)); v != "" {
			return v
		}
	}
	return ""
}

// TokenFromHTTP WebSocket 握手 token（仅 IM WS；HTTP API 仍用 authori-zation）。
// 优先级：Header authori-zation: Bearer → Header Token → Query token=
func TokenFromHTTP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("authori-zation")); v != "" {
		return StripBearer(v)
	}
	if v := strings.TrimSpace(r.Header.Get(HeaderToken)); v != "" {
		return StripBearer(v)
	}
	if v := strings.TrimSpace(r.FormValue("token")); v != "" {
		return StripBearer(v)
	}
	return ""
}

// ExtendFromHTTP 扩展 JSON（H5 WebSocket 走 Query extend=）
func ExtendFromHTTP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get(HeaderExtend)); v != "" {
		return v
	}
	return strings.TrimSpace(r.FormValue("extend"))
}

// FromContext gRPC metadata 中的 appId
func FromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range metadataKeys {
		if vals := md.Get(key); len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

func StripBearer(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 7 && strings.EqualFold(v[:7], "Bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return v
}
