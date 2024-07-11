package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var logger *zap.Logger

// InitLogger 初始化自定义的 zap 日志记录器
func InitLogger(logLevel string) {
	// 日志分割
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fmt.Sprintf("./runtime/%s.log", time.Now().Format("2006-01-02")),
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
	}

	// creat zap core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(getEncodeConfig()),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(lumberjackLogger),
		),
		parseLogLevel(logLevel),
	)

	// make zap Log
	logger = zap.New(core, zap.AddCaller())

	zap.ReplaceGlobals(logger)
}

func getEncodeConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "t",
		LevelKey:       "l",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// parseLogLevel 将字符串日志级别解析为 zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
