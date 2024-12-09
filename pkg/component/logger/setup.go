package logger

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
	"time"
)

type Logger struct {
	once        sync.Once
	initialized bool
}

func (l *Logger) Setup() error {
	var setupErr error
	l.once.Do(func() {
		if l.initialized {
			setupErr = fmt.Errorf("logger already initialized")
			return
		}
		// 日志分割
		lumberjackLogger := &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s.log", config.LogDir, time.Now().Format("2006-01-02")),
			MaxSize:    100,
			MaxBackups: 7,
			MaxAge:     30,
			Compress:   true,
		}

		// level debug 优先于日志等级设置，如果开启了debug，则将日志等级设置为debug
		level := config.LogLevel
		if config.Debug {
			level = "debug"
		}
		// 配置多写入器 (终端 + 文件)
		var ws []zapcore.WriteSyncer
		if config.Debug {
			ws = append(ws, zapcore.AddSync(os.Stdout))
		}
		ws = append(ws, zapcore.AddSync(lumberjackLogger))

		// creat zap core
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(getEncodeConfig()),
			//zapcore.NewConsoleEncoder(getEncodeConfig()),
			zapcore.NewMultiWriteSyncer(ws...),
			parseLogLevel(level),
		)

		// make zap Log
		logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))
		zap.ReplaceGlobals(logger)
		l.initialized = true
	})
	return setupErr
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
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
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
