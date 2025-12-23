package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// 可观测性配置项
// LogEnabled: 是否开启日志
// LogExporter: 日志上报方式
// LogLoadInternal: 日志上报的时间间隔
// LogLoadMaxLog: 日志上报的最大日志量
// HttpLogFeedIngesterUrl: 日志上报的HTTP服务地址
type LogSetting struct {
	LogEnabled             bool   `mapstructure:"logEnabled"`
	LogExporter            string `mapstructure:"logExporter"`
	LogLoadInternal        int    `mapstructure:"logLoadInternal"`
	LogLoadMaxLog          int    `mapstructure:"logLoadMaxLog"`
	HttpLogFeedIngesterUrl string `mapstructure:"httpLogFeedIngesterUrl"`
}

// 可观测性配置项
// TraceEnabled: 是否开启trace
// TraceProvider: trace上报方式
// TraceMaxQueueSize: trace上报的队列最大大小
// HttpTraceFeedIngesterUrl: trace上报的HTTP服务地址
// GrpcTraceFeedIngesterUrl: trace上报的GRPC服务地址
// GrpcTraceJobId: trace上报的GRPC任务ID
type TraceSetting struct {
	TraceEnabled             bool   `mapstructure:"traceEnabled"`
	TraceProvider            string `mapstructure:"traceProvider"`
	TraceMaxQueueSize        int    `mapstructure:"traceMaxQueueSize"`
	HttpTraceFeedIngesterUrl string `mapstructure:"httpTraceFeedIngesterUrl"`
	GrpcTraceFeedIngesterUrl string `mapstructure:"grpcTraceFeedIngesterUrl"`
	GrpcTraceJobId           string `mapstructure:"grpcTraceJobId"`
}

// 可观测性配置项
// MetricEnabled: 是否开启metric
// MetricProvider: metric上报方式
// HttpMetricFeedIngesterUrl: metric上报的HTTP服务地址
// MetricIntervalSecond: metric上报的时间间隔
type MetricSetting struct {
	MetricEnabled             bool   `mapstructure:"metricEnabled"`
	MetricProvider            string `mapstructure:"metricProvider"`
	HttpMetricFeedIngesterUrl string `mapstructure:"httpMetricFeedIngesterUrl"`
	MetricIntervalSecond      int    `mapstructure:"metricIntervalSecond"`
}

type ObservabilitySetting struct {
	LogSetting    `mapstructure:",squash"`
	TraceSetting  `mapstructure:",squash"`
	MetricSetting `mapstructure:",squash"`
}

type ServerInfo struct {
	ServerName    string
	ServerVersion string
	Language      string
	GoVersion     string
	GoArch        string
}

func Init(serverInfo ServerInfo, setting ObservabilitySetting) {
	if setting.LogEnabled {
		InitLogExporter(serverInfo, setting.LogSetting)
	}
	if setting.TraceEnabled {
		InitTraceProvider(serverInfo, setting.TraceSetting)
	}
	if setting.MetricEnabled {
		InitMeterProvider(serverInfo, setting.MetricSetting)
	}
}

// 将 Trace 上下文注入到 HTTP Header 中
func InjectTraceHeader(ctx context.Context, header http.Header) {
	if header == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

// 从 HTTP Header 中提取 Trace 上下文
func ExtractTraceHeader(ctx context.Context, header http.Header) context.Context {
	if header == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}
