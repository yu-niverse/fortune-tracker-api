package logger

import (
	"Fortune_Tracker_API/config"
	"os"
	"path"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	// Configuration for the logger encoder (how to format the logs)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder 	// change time format to ISO8601
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder	// change level format to CAPITAL
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Create logger file
	path := config.Viper.GetString("LOG_PATH")
	file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	fileWriteSyncer := zapcore.AddSync(file)

	// Create logger
	// (low) Debug -> Info -> Warn -> Error -> Panic -> Fatal (high)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.ErrorLevel), // you can change the output level here
		zapcore.NewCore(encoder, fileWriteSyncer, zapcore.DebugLevel),
	)
	Log = zap.New(core)
}

func Info(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Error(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	if Log == nil {
		return
	}
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	Log.Warn(message, fields...)
}

// Get the caller info for the log
func getCallerInfoForLog() (callerFields []zap.Field) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName)

	callerFields = append(callerFields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))
	return
}