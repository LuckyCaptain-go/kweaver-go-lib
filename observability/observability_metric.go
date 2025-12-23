package observability

import (
	"time"

	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/ar_metric"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/public"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

func InitMeterProvider(serverInfo ServerInfo, setting MetricSetting) {
	switch setting.MetricProvider {
	case "http":
		httpMeterProvider(serverInfo, setting)
	}
}

// 初始化 metric 采用 http 上报到 AR 的 metricProvider
func httpMeterProvider(serverInfo ServerInfo, setting MetricSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	meterClient := public.NewHTTPClient(
		public.WithAnyRobotURL(setting.HttpMetricFeedIngesterUrl),
		public.WithCompression(1),
		public.WithTimeout(10*time.Second),
		public.WithRetry(true, 5*time.Second, 30*time.Second, 1*time.Minute),
	)

	meterExporter := ar_metric.NewExporter(meterClient)

	meterProvider := sdkMetric.NewMeterProvider(
		sdkMetric.WithReader(
			sdkMetric.NewPeriodicReader(
				meterExporter,
				sdkMetric.WithInterval(time.Duration(setting.MetricIntervalSecond)*time.Second),
				sdkMetric.WithTimeout(10*time.Second),
			)),
		sdkMetric.WithResource(ar_metric.MetricResource()),
	)

	ar_metric.MetricProvider = meterProvider
	ar_metric.Meter = ar_metric.MetricProvider.Meter(
		version.MetricInstrumentationName,
		metric.WithSchemaURL(version.MetricInstrumentationURL),
		metric.WithInstrumentationVersion(version.TelemetrySDKVersion),
	)

	otel.SetMeterProvider(meterProvider)
}
