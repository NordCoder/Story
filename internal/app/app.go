package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NordCoder/Story/config"
	storypb "github.com/NordCoder/Story/generated/api/story"
	"github.com/NordCoder/Story/internal/controller"
	"github.com/NordCoder/Story/internal/infrastructure/redis"
	"github.com/NordCoder/Story/internal/infrastructure/wikipedia"
	mylogger "github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/internal/usecase"
	"github.com/NordCoder/Story/services/authorization/repository"
	authusecase "github.com/NordCoder/Story/services/authorization/usecase"
	"github.com/NordCoder/Story/services/prefetch"
	"github.com/NordCoder/Story/services/prefetch/category"
	prefetcherconfig "github.com/NordCoder/Story/services/prefetch/config"
	recusecase "github.com/NordCoder/Story/services/recommendation/usecase"
	"github.com/go-chi/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// todo fix code: initialize auth / rec services and put it on work
// todo create separate functions to create services

func initMetrics() *mylogger.Metrics {
	metrics := mylogger.NewMetrics()
	metrics.Init()
	return metrics
}

func Run(httpCfg *config.HTTPConfig, logger *zap.Logger) error {
	ctx := context.Background()

	metrics := initMetrics()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(metrics.PromMiddleware)
	r.Use(middleware.RequestID)
	r.Use(mylogger.LoggerMiddleware(logger))

	r.Use(middleware.Timeout(parseDurationOr(httpCfg.Timeouts.Read, 5*time.Second) + parseDurationOr(httpCfg.Timeouts.Write, 10*time.Second)))

	if httpCfg.CORS.Enabled {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   httpCfg.CORS.AllowedOrigins,
			AllowedMethods:   httpCfg.CORS.AllowedMethods,
			AllowedHeaders:   httpCfg.CORS.AllowedHeaders,
			AllowCredentials: httpCfg.CORS.AllowCredentials,
			MaxAge:           int((parseDurationOr(httpCfg.CORS.MaxAge, 24*time.Hour)).Seconds()),
		}))
	}

	r.Handle(httpCfg.Endpoints.Metrics, metrics.Handler())
	r.Mount(httpCfg.Endpoints.Pprof, middleware.Profiler())

	wiki := wikipedia.NewWikiMock()

	redisClient, err := redis.NewRedisClient()
	if err != nil {
		logger.Fatal("failed to start redis client", zap.Error(err))
	}

	factRepo := redis.NewFactRepository(redisClient, 5*time.Hour)

	// Лайвнесс: просто проверка, жив ли процесс
	r.Get(httpCfg.Endpoints.Liveness, controller.LiveHandler)

	// Рединесс: проверка реальных зависимостей
	readinessHandler := controller.NewReadinessHandler()

	// Добавляем сюда все важные зависимости
	readinessHandler.AddDependency("redis", factRepo)
	readinessHandler.AddDependency("wikipedia", wiki)

	// Регистрируем обработчик /ready
	readinessHandler.RegisterRoutes(r, httpCfg.Endpoints.Readiness)

	// provider init
	wwiiProvider := category.NewWWIICategoryProvider()

	// prefetcher init
	prefetchConfig, err := prefetcherconfig.NewPrefetcherConfig()
	if err != nil {
		logger.Fatal("failed to start prefetcher", zap.Error(err))
	}
	prefetcher := prefetch.NewPrefetcher(prefetchConfig, wiki, factRepo, logger, wwiiProvider)
	go func() {
		err = prefetcher.Run(ctx)
		if err != nil {
			logger.Fatal("failed to start prefetcher", zap.Error(err))
		}
	}()

	// main controller init
	authRepo := repository.NewAuthRepository()
	authService := authusecase.NewAuthService(authRepo)

	recService := recusecase.NewRecService(authService, wwiiProvider)

	ctrl := controller.New(usecase.NewFactUseCase(factRepo, recService))

	grpcSrv := grpc.NewServer()
	storypb.RegisterStoryServer(grpcSrv, ctrl)

	// server start
	lis, err := net.Listen("tcp", ":"+httpCfg.GrpcPort)
	if err != nil {
		logger.Fatal("failed to start tcp listener", zap.Error(err))
	}

	go func() {
		err := grpcSrv.Serve(lis)
		if err != nil {
			logger.Fatal("failed to start grpc server", zap.Error(err))
		}
	}()
	logger.Info("grpc server listening", zap.String("port", httpCfg.GrpcPort))

	gw := runtime.NewServeMux()
	if err := storypb.RegisterStoryHandlerFromEndpoint(ctx, gw, httpCfg.GrpcHost+":"+httpCfg.GrpcPort, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}); err != nil {
		return err
	}
	r.Handle("/*", gw)

	addr := fmt.Sprintf("%s:%s", httpCfg.Host, httpCfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  parseDurationOr(httpCfg.Timeouts.Read, 5*time.Second),
		WriteTimeout: parseDurationOr(httpCfg.Timeouts.Write, 10*time.Second),
		IdleTimeout:  parseDurationOr(httpCfg.Timeouts.Idle, 120*time.Second),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()
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
