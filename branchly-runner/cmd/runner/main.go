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

	"github.com/branchly/branchly-runner/internal/agent/claudecode"
	"github.com/branchly/branchly-runner/internal/config"
	"github.com/branchly/branchly-runner/internal/executor"
	"github.com/branchly/branchly-runner/internal/handler"
	"github.com/branchly/branchly-runner/internal/infra"
	"github.com/branchly/branchly-runner/internal/pool"
	"github.com/branchly/branchly-runner/internal/repository"
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

	db := mongoClient.Database(cfg.MongoDatabase)
	jobRepo := repository.NewJobRepository(db)
	jobLogRepo := repository.NewJobLogRepository(db)
	agent := claudecode.New()
	ex := executor.NewExecutor(agent, jobRepo, jobLogRepo, cfg.EncryptionKey, cfg.WorkDir)
	p := pool.New(cfg.MaxConcurrentJobs)
	jobsH := handler.NewJobsHandler(cfg.RunnerSecret, p, ex)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/jobs", jobsH.PostJob)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("runner listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown failed", "error", err)
	}
	slog.Info("waiting for job pool")
	p.Wait()
	_ = mongoClient.Disconnect(context.Background())
	slog.Info("runner stopped")
}
