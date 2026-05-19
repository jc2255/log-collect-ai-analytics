package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.SugaredLogger

// LogConfig 日志配置
type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// Init 初始化日志
func Init(level string, filename string) error {
	return InitWithConfig(LogConfig{
		Level:      level,
		Filename:   filename,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	})
}

// InitWithConfig 使用完整配置初始化日志
func InitWithConfig(cfg LogConfig) error {
	zapLevel := parseLevel(cfg.Level)

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

	// 文件输出（使用lumberjack支持日志轮转）
	if cfg.Filename != "" {
		// 自动创建日志目录
		if dir := filepath.Dir(cfg.Filename); dir != "" && dir != "." {
			os.MkdirAll(dir, 0755)
		}

		maxSize := cfg.MaxSize
		if maxSize <= 0 {
			maxSize = 100
		}
		maxBackups := cfg.MaxBackups
		if maxBackups <= 0 {
			maxBackups = 5
		}
		maxAge := cfg.MaxAge
		if maxAge <= 0 {
			maxAge = 30
		}

		writer := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   cfg.Compress,
		}
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.AddSync(writer), zapLevel))
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
