package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/branchly/branchly-runner/internal/infra"
	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	MongoURI          string
	MongoDatabase     string
	EncryptionKey     []byte
	RunnerSecret      string
	MaxConcurrentJobs int
	WorkDir           string
	ShutdownTimeout   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8081"
	}
	mongoURI := strings.TrimSpace(os.Getenv("MONGODB_URI"))
	if mongoURI == "" {
		return nil, fmt.Errorf("config: MONGODB_URI is required")
	}
	dbName := strings.TrimSpace(os.Getenv("MONGODB_DATABASE"))
	if dbName == "" {
		return nil, fmt.Errorf("config: MONGODB_DATABASE is required")
	}
	encHex := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))
	if encHex == "" {
		return nil, fmt.Errorf("config: ENCRYPTION_KEY is required")
	}
	encKey, err := infra.ParseEncryptionKey(encHex)
	if err != nil {
		return nil, fmt.Errorf("config: ENCRYPTION_KEY: %w", err)
	}
	runnerSecret := strings.TrimSpace(os.Getenv("RUNNER_SECRET"))
	if runnerSecret == "" {
		runnerSecret = strings.TrimSpace(os.Getenv("INTERNAL_API_SECRET"))
	}
	if runnerSecret == "" {
		return nil, fmt.Errorf("config: RUNNER_SECRET or INTERNAL_API_SECRET is required")
	}
	maxJobs := 3
	if v := strings.TrimSpace(os.Getenv("MAX_CONCURRENT_JOBS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("config: MAX_CONCURRENT_JOBS must be a positive integer")
		}
		maxJobs = n
	}
	workDir := strings.TrimSpace(os.Getenv("WORK_DIR"))
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "branchly-jobs")
	}
	shutdownTimeout := 8 * time.Minute
	if v := strings.TrimSpace(os.Getenv("SHUTDOWN_TIMEOUT")); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < time.Second {
			return nil, fmt.Errorf("config: SHUTDOWN_TIMEOUT must be a valid duration >= 1s (e.g. 8m, 480s)")
		}
		shutdownTimeout = d
	}
	return &Config{
		Port:              port,
		MongoURI:          mongoURI,
		MongoDatabase:     dbName,
		EncryptionKey:     encKey,
		RunnerSecret:      runnerSecret,
		MaxConcurrentJobs: maxJobs,
		WorkDir:           workDir,
		ShutdownTimeout:   shutdownTimeout,
	}, nil
}
