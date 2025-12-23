package observability

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/ar_log"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/public"
	"github.com/AISHU-Technology/TelemetrySDK-Go/exporter/v2/resource"
	"github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/encoder"
	"github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/exporter"
	"github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/field"
	spanLog "github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/log"
	"github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/open_standard"
	sdkRuntime "github.com/AISHU-Technology/TelemetrySDK-Go/span/v2/runtime"
)

// 输出到控制台的程序日志记录器。主要用于异常时在对 wrap 后的错误信息的在最上层打印能包含有堆栈信息
// var SystemLogger4Consol = spanLog.NewSamplerLogger(spanLog.WithSample(1.0), spanLog.WithLevel(spanLog.ErrorLevel))

// 上报到 AR 的程序日志记录器。主要用于把日志信息上报到 AR，结合 trace，便于分析
var SystemLogger = spanLog.NewDefaultSamplerLogger()

// 记录日志: msg 拼接上文件、行号、函数名。用于日志记录时把位置信息带上
func Info(ctx context.Context, msg string) {
	pc, filename, line, _ := runtime.Caller(1)
	SystemLogger.Info(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		field.WithContext(ctx))
}

func InfoWithAttr(ctx context.Context, msg string, options ...field.LogOptionFunc) {
	pc, filename, line, _ := runtime.Caller(1)
	optionsWithCtx := append([]field.LogOptionFunc{field.WithContext(ctx)}, options...)
	SystemLogger.Info(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		optionsWithCtx...)
}

func Error(ctx context.Context, msg string) {
	pc, filename, line, _ := runtime.Caller(1)
	SystemLogger.Error(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		field.WithContext(ctx))
}

func ErrorWithAttr(ctx context.Context, msg string, options ...field.LogOptionFunc) {
	pc, filename, line, _ := runtime.Caller(1)
	optionsWithCtx := append([]field.LogOptionFunc{field.WithContext(ctx)}, options...)
	SystemLogger.Error(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		optionsWithCtx...)
}

func Warn(ctx context.Context, msg string) {
	pc, filename, line, _ := runtime.Caller(1)
	SystemLogger.Warn(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		field.WithContext(ctx))
}

func WarnWithAttr(ctx context.Context, msg string, options ...field.LogOptionFunc) {
	pc, filename, line, _ := runtime.Caller(1)
	optionsWithCtx := append([]field.LogOptionFunc{field.WithContext(ctx)}, options...)
	SystemLogger.Warn(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		optionsWithCtx...)
}

func Debug(ctx context.Context, msg string) {
	pc, filename, line, _ := runtime.Caller(1)
	SystemLogger.Debug(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		field.WithContext(ctx))
}

func DebugWithAttr(ctx context.Context, msg string, options ...field.LogOptionFunc) {
	pc, filename, line, _ := runtime.Caller(1)
	optionsWithCtx := append([]field.LogOptionFunc{field.WithContext(ctx)}, options...)
	SystemLogger.Debug(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		optionsWithCtx...)
}

func Fatal(ctx context.Context, msg string) {
	pc, filename, line, _ := runtime.Caller(1)
	SystemLogger.Fatal(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		field.WithContext(ctx))
}

func FatalWithAttr(ctx context.Context, msg string, options ...field.LogOptionFunc) {
	pc, filename, line, _ := runtime.Caller(1)
	optionsWithCtx := append([]field.LogOptionFunc{field.WithContext(ctx)}, options...)
	SystemLogger.Fatal(fmt.Sprintf("%s:%d:%s: %v", filename, line, strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), "."), msg),
		optionsWithCtx...)
}

func InitLogExporter(serverInfo ServerInfo, setting LogSetting) {
	switch setting.LogExporter {
	case "console":
		consoleLogRunner(serverInfo, setting)
	case "http":
		httpLogRunner(serverInfo, setting)
	case "file":
		fileLogRunner(serverInfo, setting)
	}
}

// 初始化日志上报 AR 的日志器
func httpLogRunner(serverInfo ServerInfo, setting LogSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	// 1.初始化系统日志器，系统日志上报到 AnyRobot。
	systemLogClient := public.NewHTTPClient(
		public.WithAnyRobotURL(setting.HttpLogFeedIngesterUrl),
		public.WithCompression(1),
		public.WithTimeout(10*time.Second),
		public.WithRetry(true, 5*time.Second, 20*time.Second, 1*time.Minute))

	systemLogExporter := ar_log.NewExporter(systemLogClient)

	systemLogWriter := open_standard.OpenTelemetryWriter(
		encoder.NewJsonEncoderWithExporters(systemLogExporter),
		addServerResource4Log(resource.LogResource(), serverInfo))

	systemLogRunner := sdkRuntime.NewRuntime(systemLogWriter, field.NewSpanFromPool)

	systemLogRunner.SetUploadInternalAndMaxLog(
		time.Duration(setting.LogLoadInternal)*time.Second,
		setting.LogLoadMaxLog)

	// 运行SystemLogger日志器。
	go systemLogRunner.Run()

	SystemLogger.SetLevel(spanLog.AllLevel)
	SystemLogger.SetRuntime(systemLogRunner)
}

// 初始化日志输出到控制台的 exporter
func consoleLogRunner(serverInfo ServerInfo, setting LogSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	// 1.初始化系统日志器，系统日志在控制台输出。
	systemLogExporter := exporter.GetRealTimeExporter()

	systemLogWriter := open_standard.OpenTelemetryWriter(
		encoder.NewJsonEncoderWithExporters(systemLogExporter),
		addServerResource4Log(resource.LogResource(), serverInfo))

	systemLogRunner := sdkRuntime.NewRuntime(systemLogWriter, field.NewSpanFromPool)

	systemLogRunner.SetUploadInternalAndMaxLog(
		time.Duration(setting.LogLoadInternal)*time.Second,
		setting.LogLoadMaxLog)

	// 运行SystemLogger日志器。
	go systemLogRunner.Run()

	SystemLogger.SetLevel(spanLog.InfoLevel)
	SystemLogger.SetRuntime(systemLogRunner)
}

// 初始化日志输出到文件的 exporter
func fileLogRunner(serverInfo ServerInfo, setting LogSetting) {
	public.SetServiceInfo(serverInfo.ServerName, serverInfo.ServerVersion, POD_NAME)

	// 1.初始化系统日志器，系统日志写入文件。
	systemLogClient := public.NewFileClient("./o11y_log.json")

	systemLogExporter := ar_log.NewExporter(systemLogClient)

	systemLogWriter := open_standard.OpenTelemetryWriter(
		encoder.NewJsonEncoderWithExporters(systemLogExporter),
		addServerResource4Log(resource.LogResource(), serverInfo))

	systemLogRunner := sdkRuntime.NewRuntime(systemLogWriter, field.NewSpanFromPool)

	systemLogRunner.SetUploadInternalAndMaxLog(
		time.Duration(setting.LogLoadInternal)*time.Second,
		setting.LogLoadMaxLog)

	// 运行SystemLogger日志器。
	go systemLogRunner.Run()

	SystemLogger.SetLevel(spanLog.InfoLevel)
	SystemLogger.SetRuntime(systemLogRunner)
}
