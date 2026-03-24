package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/handler"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/repository"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	mongoClient, err := infra.ConnectMongo(ctx, cfg.MongoURI)
	if err != nil {
		slog.Error("mongo connect failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	db := mongoClient.Database(cfg.MongoDatabase)
	if err := infra.EnsureIndexes(ctx, db); err != nil {
		slog.Error("mongo indexes failed", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	repoRepo := repository.NewConnectedRepositoryRepository(db)
	jobRepo := repository.NewJobRepository(db)
	jobLogRepo := repository.NewJobLogRepository(db)

	authSvc := service.NewAuthService(cfg, userRepo)
	repoSvc := service.NewRepositoryService(cfg, userRepo, repoRepo)
	runner := infra.NewRunnerClient(cfg.RunnerURL, cfg.RunnerSecret)
	jobSvc := service.NewJobService(cfg, jobRepo, jobLogRepo, repoRepo, userRepo, runner)

	repoH := handler.NewRepositoryHandler(repoSvc)
	jobH := handler.NewJobHandler(jobSvc)
	sseH := handler.NewSSEHandler(jobSvc)
	internalH := handler.NewInternalHandler(jobSvc)
	internalAuthH := handler.NewInternalAuthHandler(authSvc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLog())
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	r.GET("/health", func(c *gin.Context) {
		respond.JSONOK(c, gin.H{"status": "ok"})
	})

	protected := r.Group("")
	protected.Use(middleware.AuthJWT(authSvc))
	{
		protected.GET("/repositories", repoH.List)
		protected.POST("/repositories", repoH.Connect)
		protected.DELETE("/repositories/:id", repoH.Delete)
		protected.GET("/repositories/github", repoH.ListGitHub)

		protected.GET("/jobs", jobH.List)
		protected.POST("/jobs", jobH.Create)
		protected.GET("/jobs/:id", jobH.Get)
		protected.GET("/jobs/:id/logs", sseH.StreamJobLogs)
	}

	internal := r.Group("/internal")
	internal.Use(middleware.InternalAPI(cfg.InternalSecret))
	{
		internal.POST("/auth/upsert", internalAuthH.Upsert)
		internal.POST("/jobs/:id/status", internalH.UpdateStatus)
		internal.POST("/jobs/:id/logs", internalH.AppendLog)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("api listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
