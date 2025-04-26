package logger

import (
	"context"
	"github.com/NordCoder/Story/config"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
)

type ctxLoggerKey struct{}

var Key = ctxLoggerKey{}

func Init(config *config.LoggerConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	var encoder zapcore.Encoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	var cores []zapcore.Core

	for _, path := range config.OutputPaths {
		cores = append(cores, newCore(config, path, encoder, func(l zapcore.Level) bool {
			return l < zapcore.ErrorLevel
		}))
	}

	for _, path := range config.ErrorOutputPaths {
		cores = append(cores, newCore(config, path, encoder, func(l zapcore.Level) bool {
			return l >= zapcore.ErrorLevel
		}))
	}

	combined := zapcore.NewTee(cores...)
	logger := zap.New(combined, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(logger)
	return logger, nil
}

func newCore(config *config.LoggerConfig, path string, encoder zapcore.Encoder, levelEnabler zap.LevelEnablerFunc) zapcore.Core {
	var ws zapcore.WriteSyncer
	if path == "stdout" {
		ws = zapcore.AddSync(os.Stdout)
	} else if path == "stderr" {
		ws = zapcore.AddSync(os.Stderr)
	} else {
		rotator := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    config.RotatorConfig.MaxFileSize,
			MaxBackups: config.RotatorConfig.MaxBackups,
			MaxAge:     config.RotatorConfig.MaxAge,
			Compress:   config.RotatorConfig.Compress,
		}
		ws = zapcore.AddSync(rotator)
	}
	return zapcore.NewCore(encoder, ws, levelEnabler)
}

func LoggerMiddleware(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())
			reqLogger := base.With(zap.String("request_id", reqID))
			ctx := context.WithValue(r.Context(), Key, reqLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	if lg, ok := ctx.Value(Key).(*zap.Logger); ok {
		return lg
	}
	return zap.L()
}
