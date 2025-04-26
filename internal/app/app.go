package app

import (
	"context"
	"fmt"
	"github.com/NordCoder/Story/config"
	storypb "github.com/NordCoder/Story/generated/api/story"
	"github.com/NordCoder/Story/internal/controller"
	"github.com/NordCoder/Story/internal/infrastructure/redis"
	logger2 "github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/internal/usecase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(httpCfg *config.HTTPConfig, logger *zap.Logger) error {
	ctx := context.Background()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger2.PromMiddleware)
	r.Use(middleware.RequestID)
	r.Use(logger2.LoggerMiddleware(logger))

	r.Use(middleware.Timeout(parseDurationOr(httpCfg.Timeouts.Read, 5*time.Second) + parseDurationOr(httpCfg.Timeouts.Write, 10*time.Second)))

	//if httpCfg.CORS.Enabled { todo understand this smart thing
	//	r.Use(cors.Handler(cors.Options{
	//		AllowedOrigins:   httpCfg.CORS.AllowedOrigins,
	//		AllowedMethods:   httpCfg.CORS.AllowedMethods,
	//		AllowedHeaders:   httpCfg.CORS.AllowedHeaders,
	//		AllowCredentials: httpCfg.CORS.AllowCredentials,
	//		MaxAge:           int((parseDurationOr(httpCfg.CORS.MaxAge, 24*time.Hour)).Seconds()),
	//	}))
	//}

	r.Get(httpCfg.Endpoints.Health, func(w http.ResponseWriter, _ *http.Request) { // todo make better healthcheck
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get(httpCfg.Endpoints.Metrics, promhttp.Handler().ServeHTTP)
	r.Mount(httpCfg.Endpoints.Pprof, middleware.Profiler())

	redisClient, err := redis.NewRedisClient()
	if err != nil {
		logger.Fatal("failed to start redis client", zap.Error(err))
	}

	factRepo := redis.NewFactRepository(redisClient, 5*time.Hour)

	ctrl := controller.New(usecase.NewFactUseCase(factRepo))

	grpcSrv := grpc.NewServer()
	storypb.RegisterStoryServer(grpcSrv, ctrl)
	lis, err := net.Listen("tcp", ":"+httpCfg.GrpcPort)
	if err != nil {
		return err
	}

	go func() {
		err := grpcSrv.Serve(lis)
		if err != nil {
			logger.Fatal("failed to start grpc server", zap.Error(err))
		}
	}()
	logger.Info("grpc server listening", zap.String("port", httpCfg.GrpcPort))

	gw := runtime.NewServeMux()
	if err := storypb.RegisterStoryHandlerFromEndpoint(ctx, gw, httpCfg.GrpcHost+":"+httpCfg.GrpcHost, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}); err != nil {
		return err
	}
	r.Mount("/", gw)

	addr := fmt.Sprintf("%s:%s", httpCfg.Host, httpCfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  parseDurationOr(httpCfg.Timeouts.Read, 5*time.Second),
		WriteTimeout: parseDurationOr(httpCfg.Timeouts.Write, 10*time.Second),
		IdleTimeout:  parseDurationOr(httpCfg.Timeouts.Idle, 120*time.Second),
	}
	logger.Info("HTTP server listening", zap.String("addr", addr))

	//go func() { // todo understand this smart thing
	//	if httpCfg.TLS.Enabled {
	//		logger.Info("Starting HTTPS", zap.String("addr", addr))
	//		_ = srv.ListenAndServeTLS(httpCfg.TLS.CertFile, httpCfg.TLS.KeyFile)
	//	} else {
	//		logger.Info("Starting HTTP", zap.String("addr", addr))
	//		_ = srv.ListenAndServe()
	//	}
	//}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received")

	ctxShut, cancel := context.WithTimeout(ctx, parseDurationOr(httpCfg.Timeouts.ShutdownGracePeriod, 15*time.Second))
	defer cancel()

	if err := srv.Shutdown(ctxShut); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
	grpcSrv.GracefulStop()

	return nil
}

func parseDurationOr(s string, d time.Duration) time.Duration {
	if parsed, err := time.ParseDuration(s); err == nil {
		return parsed
	}
	return d
}
