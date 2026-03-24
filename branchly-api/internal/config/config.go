package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	MongoURI      string
	MongoDatabase string

	JWTSecret  []byte
	JWTTTLDays int

	EncryptionKey []byte

	RunnerURL string

	FrontendURL    string
	AllowedOrigins []string
	InternalSecret string

	CookieSecure        bool
	SessionCookieDomain string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}

	encHex := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))
	if encHex == "" {
		return nil, fmt.Errorf("config: ENCRYPTION_KEY is required")
	}
	encKey, err := parseEncryptionKey(encHex)
	if err != nil {
		return nil, fmt.Errorf("config: ENCRYPTION_KEY: %w", err)
	}

	ttlDays := 7
	if v := strings.TrimSpace(os.Getenv("JWT_TTL_DAYS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("config: JWT_TTL_DAYS must be a positive integer")
		}
		ttlDays = n
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	mongoURI := strings.TrimSpace(os.Getenv("MONGODB_URI"))
	if mongoURI == "" {
		return nil, fmt.Errorf("config: MONGODB_URI is required")
	}

	dbName := strings.TrimSpace(os.Getenv("MONGODB_DATABASE"))
	if dbName == "" {
		return nil, fmt.Errorf("config: MONGODB_DATABASE is required")
	}

	runnerURL := strings.TrimSpace(os.Getenv("RUNNER_URL"))
	if runnerURL == "" {
		return nil, fmt.Errorf("config: RUNNER_URL is required")
	}

	frontend := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if frontend == "" {
		return nil, fmt.Errorf("config: FRONTEND_URL is required")
	}

	originsStr := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if originsStr == "" {
		return nil, fmt.Errorf("config: ALLOWED_ORIGINS is required")
	}
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	internalSecret := strings.TrimSpace(os.Getenv("INTERNAL_API_SECRET"))
	if internalSecret == "" {
		return nil, fmt.Errorf("config: INTERNAL_API_SECRET is required")
	}

	cookieSecure := parseBoolEnv(os.Getenv("COOKIE_SECURE"))
	sessionCookieDomain := strings.TrimSpace(os.Getenv("SESSION_COOKIE_DOMAIN"))

	return &Config{
		Port:                port,
		MongoURI:            mongoURI,
		MongoDatabase:       dbName,
		JWTSecret:           []byte(jwtSecret),
		JWTTTLDays:          ttlDays,
		EncryptionKey:       encKey,
		RunnerURL:           strings.TrimSuffix(runnerURL, "/"),
		FrontendURL:         strings.TrimSuffix(frontend, "/"),
		AllowedOrigins:      origins,
		InternalSecret:      internalSecret,
		CookieSecure:        cookieSecure,
		SessionCookieDomain: sessionCookieDomain,
	}, nil
}

func parseBoolEnv(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseEncryptionKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if len(s) == 64 {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decode hex ENCRYPTION_KEY: %w", err)
		}
		if len(b) != 32 {
			return nil, fmt.Errorf("decoded ENCRYPTION_KEY must be 32 bytes")
		}
		return b, nil
	}
	if len(s) == 32 {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("ENCRYPTION_KEY must be 32 raw bytes or 64 hex characters")
}

func (c *Config) JWTTTL() time.Duration {
	return time.Duration(c.JWTTTLDays) * 24 * time.Hour
}
