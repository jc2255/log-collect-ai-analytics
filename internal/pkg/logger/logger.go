package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.SugaredLogger

// Init 初始化日志
func Init(level string, filename string) error {
	zapLevel := parseLevel(level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel))

	// 文件输出
	if filename != "" {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), zapLevel))
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	globalLogger = logger.Sugar()
	return nil
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// L 获取全局logger
func L() *zap.SugaredLogger {
	if globalLogger == nil {
		// 默认logger
		l, _ := zap.NewDevelopment()
		globalLogger = l.Sugar()
	}
	return globalLogger
}

func Info(args ...interface{})                    { L().Info(args...) }
func Infof(template string, args ...interface{})  { L().Infof(template, args...) }
func Error(args ...interface{})                   { L().Error(args...) }
func Errorf(template string, args ...interface{}) { L().Errorf(template, args...) }
func Warn(args ...interface{})                    { L().Warn(args...) }
func Warnf(template string, args ...interface{})  { L().Warnf(template, args...) }
func Debug(args ...interface{})                   { L().Debug(args...) }
func Debugf(template string, args ...interface{}) { L().Debugf(template, args...) }
func Fatal(args ...interface{})                   { L().Fatal(args...) }
func Fatalf(template string, args ...interface{}) { L().Fatalf(template, args...) }
