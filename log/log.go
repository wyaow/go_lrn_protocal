package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger = newLogger()
var atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel) // 自动日志级别

// 初始化日志, 控制台输出
func newLogger() *zap.SugaredLogger {
	encodeCfg := zap.NewDevelopmentEncoderConfig()
	encodeCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encodeCfg)
	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), atomicLevel)

	return zap.New(core).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= zapcore.FatalLevel
		}))).Sugar()
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}

func SetLogLevel(level zapcore.Level) {
	atomicLevel.SetLevel(level)
}
