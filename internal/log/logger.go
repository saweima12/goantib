package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type TinyLogger struct {
	logFile *lumberjack.Logger
	syncer  zapcore.WriteSyncer
	logger  *zap.Logger
}

func (logger *TinyLogger) Sync() {
	logger.syncer.Sync()
}

var logInstance *TinyLogger

func InitLogger() error {
	logFile := lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	// fileSyncer
	writeSyncer := zapcore.AddSync(&logFile)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO, WARN, ERROR
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // 2024-12-12T10:00:00Z
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zap.InfoLevel,
	)

	// console syncer
	consoleSyncer := zapcore.AddSync(zapcore.Lock(os.Stdout))
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		consoleSyncer,
		zap.InfoLevel,
	)

	core := zapcore.NewTee(consoleCore, fileCore)
	zapLogger := zap.New(core)

	newLogger := &TinyLogger{
		logFile: &logFile,
		syncer:  writeSyncer,
		logger:  zapLogger,
	}

	logInstance = newLogger
	return nil
}

func DisposeLogger() {
	logInstance.logger.Sync()
}

func Logger() *zap.SugaredLogger {
	return logInstance.logger.Sugar()
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	switch level {
	case zapcore.DebugLevel:
		color = "\033[37m"
	case zapcore.InfoLevel:
		color = "\033[32m"
	case zapcore.WarnLevel:
		color = "\033[33m"
	case zapcore.ErrorLevel:
		color = "\033[31m"
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		color = "\033[35m"
	default:
		color = "\033[0m"
	}
	enc.AppendString(color + level.String() + "\033[0m") // 添加顏色並重置
}
