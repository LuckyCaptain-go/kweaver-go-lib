package logger

import (
	"go.uber.org/zap"
)

var (
	globalLog = InitLogger(LogSetting{
		LogLevel: "debug",
	})
)

func InitGlobalLogger(setting LogSetting) {
	globalLog = InitLogger(setting)
}

func GetLogger() *zap.SugaredLogger {
	return globalLog
}

func Debug(args ...interface{}) {
	globalLog.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	globalLog.Debugf(template, args...)
}

func Info(args ...interface{}) {
	globalLog.Info(args...)
}

func Infof(template string, args ...interface{}) {
	globalLog.Infof(template, args...)
}

func Warn(args ...interface{}) {
	globalLog.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	globalLog.Warnf(template, args...)
}

func Error(args ...interface{}) {
	globalLog.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	globalLog.Errorf(template, args...)
}

func Panic(args ...interface{}) {
	globalLog.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	globalLog.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	globalLog.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	globalLog.Fatalf(template, args...)
}
