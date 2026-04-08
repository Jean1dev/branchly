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

	agentpkg "github.com/branchly/branchly-runner/internal/agent"
	"github.com/branchly/branchly-runner/internal/agent/claudecode"
	"github.com/branchly/branchly-runner/internal/agent/gemini"
	"github.com/branchly/branchly-runner/internal/agent/mock"
	"github.com/branchly/branchly-runner/internal/config"
	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/executor"
	"github.com/branchly/branchly-runner/internal/gitprovider"
	"github.com/branchly/branchly-runner/internal/handler"
	"github.com/branchly/branchly-runner/internal/infra"
	"github.com/branchly/branchly-runner/internal/pool"
	"github.com/branchly/branchly-runner/internal/repository"
	"github.com/branchly/branchly-runner/internal/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	if cfg.Environment == "dev" {
		slog.Warn("running in dev mode — all agents are mocked; no real tokens will be used")
	} else {
		if cfg.AnthropicAPIKey == "" {
			slog.Info("ANTHROPIC_API_KEY not set — Claude Code jobs will use each user's own API key (BYOK)")
		}
		if cfg.GeminiAPIKey == "" {
			slog.Info("GEMINI_API_KEY not set — Gemini jobs will use each user's own API key (BYOK)")
		}
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
	repoRepo := repository.NewRepoRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	keyResolver := infra.NewKeyResolver(apiKeyRepo, cfg.EncryptionKey, map[domain.APIKeyProvider]string{
		domain.APIKeyProviderAnthropic: cfg.AnthropicAPIKey,
		domain.APIKeyProviderGoogle:    cfg.GeminiAPIKey,
	})

	var claudeAgent, geminiAgent domain.Agent
	if cfg.Environment == "dev" {
		claudeAgent = mock.New()
		geminiAgent = mock.New()
	} else {
		claudeAgent = claudecode.New()
		geminiAgent = gemini.New()
	}
	agentFactory := agentpkg.NewFactory(claudeAgent, geminiAgent)
	providerFactory := gitprovider.NewFactory()

	ex := executor.NewExecutor(
		agentFactory,
		providerFactory,
		jobRepo,
		jobLogRepo,
		repoRepo,
		integrationRepo,
		keyResolver,
		cfg.EncryptionKey,
		cfg.WorkDir,
	)
	p := pool.New(cfg.MaxConcurrentJobs)

	retryPoller := worker.NewRetryPoller(jobRepo, repoRepo, ex, p)
	go retryPoller.Start(ctx)

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
