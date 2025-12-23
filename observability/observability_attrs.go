package observability

import (
	"fmt"
	"os"
	"strings"

	"github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/field"
	"github.com/gin-gonic/gin"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkResource "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	"github.com/AISHU-Technology/kweaver-go-lib/rest"
)

const (
	// trace 属性 key
	KEY_HTTP_URL                    = "http.url"
	KEY_HTTP_METHOD                 = "http.method"
	KEY_HTTP_HEADER_METHOD_OVERRIDE = "http.header.X-Http-Method-Override"
	KEY_HTTP_HEADER_X_LANGUAGE      = "http.header.X-Language"
	KEY_HTTP_HEADER_CONTENT_TYPE    = "http.header.Content-Type"
	KEY_HTTP_HEADER_USER_AGENT      = "http.header.User-Agent"
	KEY_HTTP_HEADER_USERID          = "http.header.Userid"
	KEY_HTTP_STATUS                 = "http.status"
	KEY_HTTP_ERROR_CODE             = "http.error_code"
	KEY_HTTP_ROUTE                  = "http.route"
	KEY_HTTP_CLIENT_IP              = "http.client_ip"

	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_FORWARDED_FOR   = "X-Forwarded-For"
	HTTP_HEADER_METHOD_OVERRIDE = "X-Http-Method-Override"
	HTTP_HEADER_X_LANGUAGE      = "X-Language"
	HTTP_HEADER_USER_AGENT      = "User-Agent"
	HTTP_HEADER_USERID          = "Userid"
)

var (
	POD_NAME string = os.Getenv("POD_NAME")
)

type TraceAttrs struct {
	HttpUrl            string
	HttpMethod         string
	HttpMethodOverride string
	HttpXLanguage      string
	HttpContentType    string
	HttpUserAgent      string
	HttpUserID         string
	HttpRoute          string
	HttpClientIP       string
}

func GetAttrsByGinCtx(c *gin.Context) TraceAttrs {
	attr := TraceAttrs{
		HttpUrl:            fmt.Sprintf("http://%s%s", c.Request.Host, c.Request.RequestURI),
		HttpMethod:         c.Request.Method,
		HttpContentType:    c.GetHeader(CONTENT_TYPE_NAME),
		HttpMethodOverride: c.GetHeader(HTTP_HEADER_METHOD_OVERRIDE),
		HttpXLanguage:      c.GetHeader(HTTP_HEADER_X_LANGUAGE),
		HttpUserAgent:      c.GetHeader(HTTP_HEADER_USER_AGENT),
		HttpUserID:         c.GetHeader(HTTP_HEADER_USERID),
		HttpRoute:          c.FullPath(),
		HttpClientIP:       serverClientIP(c.GetHeader(HTTP_HEADER_FORWARDED_FOR)),
	}
	return attr
}

// 附加 server 的 resource 信息,只适用于 trace 和 metric
func addServerResource(resource *sdkResource.Resource, serverInfo ServerInfo) *sdkResource.Resource {
	// 服务使用的语言，语言版本
	serverResource := sdkResource.NewWithAttributes("",
		attr.Key("service.language").String(serverInfo.Language),
		attr.Key("service.languageVersion").String(serverInfo.GoVersion),
		attr.Key("service.arch").String(serverInfo.GoArch))
	mergeResource, err := sdkResource.Merge(resource, serverResource)
	if err != nil {
		// return original resource
		return resource
	}
	return mergeResource
}

// 附加 server 的 resource 信息，只适用于 log
func addServerResource4Log(resource field.Field, serverInfo ServerInfo) field.Field {
	switch resource.Type() {
	case field.MapType:
		langMap := make(map[string]interface{}, 2)
		langMap["name"] = serverInfo.Language
		langMap["version"] = serverInfo.GoVersion

		service := resource.(field.MapField)["service"].(map[string]interface{})
		service["language"] = langMap
		service["arch"] = serverInfo.GoArch

		resource.(field.MapField).Append("service", field.MapField(service))
	}
	return resource
}

// 设置 trace 的相关 api 的属性
func AddHttpAttrs4API(span trace.Span, attrs TraceAttrs) {
	// 设置 trace 属性
	span.SetAttributes(
		attr.Key(KEY_HTTP_URL).String(attrs.HttpUrl),
		attr.Key(KEY_HTTP_METHOD).String(attrs.HttpMethod),
		attr.Key(KEY_HTTP_HEADER_CONTENT_TYPE).String(attrs.HttpContentType),
		attr.Key(KEY_HTTP_HEADER_X_LANGUAGE).String(attrs.HttpXLanguage),
		attr.Key(KEY_HTTP_HEADER_USER_AGENT).String(attrs.HttpUserAgent),
		attr.Key(KEY_HTTP_HEADER_USERID).String(attrs.HttpUserID),
		attr.Key(KEY_HTTP_ROUTE).String(attrs.HttpRoute),
		attr.Key(KEY_HTTP_CLIENT_IP).String(attrs.HttpClientIP),
	)
	// 兼容没有 method override 的 api
	if attrs.HttpMethodOverride != "" {
		span.SetAttributes(
			attr.Key(KEY_HTTP_HEADER_METHOD_OVERRIDE).String(attrs.HttpMethodOverride),
		)
	}
}

// 设置 trace 的错误信息的 attributes
func AddHttpAttrs4Error(span trace.Span, status int, errorCode string, statusDescription string) {
	// 设置 trace 的 attributes
	span.SetAttributes(
		attr.Key(KEY_HTTP_STATUS).Int(status),
		attr.Key(KEY_HTTP_ERROR_CODE).String(errorCode),
	)
	span.SetStatus(codes.Error, statusDescription)
}

// 设置 trace 的HTTP错误的 attributes
func AddHttpAttrs4HttpError(span trace.Span, err *rest.HTTPError) {
	// 设置 trace 的 attributes
	span.SetAttributes(
		attr.Key(KEY_HTTP_STATUS).Int(err.HTTPCode),
		attr.Key(KEY_HTTP_ERROR_CODE).String(err.BaseError.ErrorCode),
	)
	span.SetStatus(codes.Error, fmt.Sprintf("%v", err.BaseError.ErrorDetails))
}

// 设置 trace 的成功信息的 attributes
func AddHttpAttrs4Ok(span trace.Span, status int) {
	span.SetAttributes(attr.Key(KEY_HTTP_STATUS).Int(status))
	span.SetStatus(codes.Ok, "")
}

// 设置 trace 的相关内部 http 请求的 attributes
func AddAttrs4InternalHttp(span trace.Span, attrs TraceAttrs) {
	// 设置 trace 属性
	span.SetAttributes(
		attr.Key(KEY_HTTP_URL).String(attrs.HttpUrl),
		attr.Key(KEY_HTTP_METHOD).String(attrs.HttpMethod),
		attr.Key(KEY_HTTP_HEADER_CONTENT_TYPE).String(attrs.HttpContentType),
	)
	// 兼容没有 method override 的 api
	if attrs.HttpMethodOverride != "" {
		span.SetAttributes(attr.Key(KEY_HTTP_HEADER_METHOD_OVERRIDE).String(attrs.HttpMethodOverride))
	}
}

// 生成clientIP
func serverClientIP(xForwardedFor string) string {
	if idx := strings.Index(xForwardedFor, ","); idx >= 0 {
		xForwardedFor = xForwardedFor[:idx]
	}
	return xForwardedFor
}
