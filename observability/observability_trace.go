package observability

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/public"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/version"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkResource "go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	HTTP_METHOD    = "http.method"
	HTTP_ROUTE     = "http.route"
	HTTP_CLIENT_IP = "http.client_ip"
	FUNC_PATH      = "func.path"
	DB_QUERY       = "db.query"
	TABLE_NAME     = "table.name"
	DB_SQL         = "db.sql"
	DB_Values      = "db.values"
)

func InitTraceProvider(serverInfo ServerInfo, setting TraceSetting) {
	switch setting.TraceProvider {
	case "http":
		httpTraceProvider(serverInfo, setting)
	case "grpc":
		grpcTraceProvider(serverInfo, setting)
	case "file":
		fileTraceProvider(serverInfo, setting)
	}
}

// 初始化 trace 输出到文件的 tracerProvider
func fileTraceProvider(serverInfo ServerInfo, setting TraceSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	tracerClient := public.NewFileClient("./o11y_trace.json")

	tracerExporter := ar_trace.NewExporter(tracerClient)

	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(tracerExporter,
			sdkTrace.WithBlocking(),
			sdkTrace.WithMaxExportBatchSize(1000)),
		sdkTrace.WithResource(addServerResource(ar_trace.TraceResource(), serverInfo)),
	)

	otel.SetTracerProvider(tracerProvider)
}

// 初始化 trace 采用 http 上报到 AR 的 tracerProvider
func httpTraceProvider(serverInfo ServerInfo, setting TraceSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	tracerClient := public.NewHTTPClient(
		public.WithAnyRobotURL(setting.HttpTraceFeedIngesterUrl),
	)

	tracerExporter := ar_trace.NewExporter(tracerClient)

	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(
			tracerExporter,
			sdkTrace.WithMaxQueueSize(setting.TraceMaxQueueSize),
			sdkTrace.WithBlocking(),
			sdkTrace.WithMaxExportBatchSize(1000),
		),
		sdkTrace.WithResource(addServerResource(ar_trace.TraceResource(), serverInfo)))

	otel.SetTracerProvider(tracerProvider)
}

// 初始化 trace 采用 grpc 上报到 AR 的 tracerProvider
func grpcTraceProvider(serverInfo ServerInfo, setting TraceSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	traceExporter, _ := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(setting.GrpcTraceFeedIngesterUrl))

	attrs := []attribute.KeyValue{
		attribute.String("job_id", setting.GrpcTraceJobId),
	}
	jobResource := sdkResource.NewWithAttributes("", attrs...)

	tempResource, err := sdkResource.Merge(
		jobResource,
		addServerResource(ar_trace.TraceResource(), serverInfo))
	if err == nil {
		jobResource = tempResource
	}

	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(
			traceExporter,
			sdkTrace.WithMaxQueueSize(setting.TraceMaxQueueSize),
			sdkTrace.WithBlocking(),
			sdkTrace.WithMaxExportBatchSize(1000)),
		sdkTrace.WithResource(jobResource),
	)

	otel.SetTracerProvider(tracerProvider)
}

// 服务内函数调用创建 span
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {

		newCtx, span := GlobalTracer().Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindInternal))
		return newCtx, span
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := GlobalTracer().Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))
	return newCtx, span
}

// 跨服务（接口）创建 span
func StartServerSpan(ctx *gin.Context) (context.Context, trace.Span) {
	newCtx := ExtractTraceHeader(ctx.Request.Context(), ctx.Request.Header)
	newCtx, span := GlobalTracer().Start(newCtx, ctx.FullPath(), trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(attribute.String(HTTP_METHOD, ctx.Request.Method))
	span.SetAttributes(attribute.String(HTTP_ROUTE, ctx.FullPath()))
	span.SetAttributes(attribute.String(HTTP_CLIENT_IP, ctx.ClientIP()))
	return newCtx, span
}

// MQ 消费者 创建 span
func StartConsumerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		newCtx, span := GlobalTracer().Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindConsumer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := GlobalTracer().Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// MQ 生产者 创建 span
func StartProducerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		newCtx, span := GlobalTracer().Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindProducer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := GlobalTracer().Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// 关闭span
func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	TelemetrySpanEnd(span, err)
}

// TelemetrySpanEnd 关闭span
func TelemetrySpanEnd(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
	span.End()
}

// 设置attribute
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

// 获取全局 TracerProvider 产生的Tracer
// 代替 ar_trace.Tracer (如果程序中存在多次 SetTracerProvider, 可能导致 ar_trace.Tracer 所属的 TracerProvider 没有更新，导致 trace 埋点失败)
func GlobalTracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(
		version.TraceInstrumentationName,
		trace.WithInstrumentationVersion(version.TelemetrySDKVersion),
		trace.WithSchemaURL(version.TraceInstrumentationURL),
	)
}
