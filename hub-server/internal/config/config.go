package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config controls role-hub runtime behavior.
type Config struct {
	HTTPAddr        string
	DBDialect       string
	DBDSN           string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBConnMaxLife   time.Duration
	DBConnMaxIdle   time.Duration
	DBTimeout       time.Duration
	RateLimitRPS    float64
	RateLimitBurst  int
	MaxBodyBytes    int64
	MaxResultsCount int
	MaxInflight     int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        getenvDefault("ROLE_HUB_HTTP_ADDR", ":8080"),
		DBDialect:       getenvDefault("ROLE_HUB_DB_DIALECT", "postgres"),
		DBDSN:           getenvDefault("ROLE_HUB_DB_DSN", ""),
		DBMaxOpenConns:  getenvDefaultInt("ROLE_HUB_DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:  getenvDefaultInt("ROLE_HUB_DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLife:   getenvDefaultDuration("ROLE_HUB_DB_CONN_MAX_LIFETIME", 30*time.Minute),
		DBConnMaxIdle:   getenvDefaultDuration("ROLE_HUB_DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		DBTimeout:       getenvDefaultDuration("ROLE_HUB_DB_TIMEOUT", 3*time.Second),
		RateLimitRPS:    getenvDefaultFloat("ROLE_HUB_RATE_LIMIT_RPS", 5),
		RateLimitBurst:  getenvDefaultInt("ROLE_HUB_RATE_LIMIT_BURST", 10),
		MaxBodyBytes:    getenvDefaultInt64("ROLE_HUB_MAX_BODY_BYTES", 1<<20), // 1MB
		MaxResultsCount: getenvDefaultInt("ROLE_HUB_MAX_RESULTS", 500),
		MaxInflight:     getenvDefaultInt("ROLE_HUB_MAX_INFLIGHT", 100),
	}

	if cfg.DBDialect != "postgres" && cfg.DBDialect != "sqlite" {
		return Config{}, fmt.Errorf("unsupported ROLE_HUB_DB_DIALECT: %s", cfg.DBDialect)
	}
	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_DSN is required")
	}
	if cfg.RateLimitRPS <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_RATE_LIMIT_RPS must be > 0")
	}
	if cfg.RateLimitBurst <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_RATE_LIMIT_BURST must be > 0")
	}
	if cfg.MaxBodyBytes <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_MAX_BODY_BYTES must be > 0")
	}
	if cfg.MaxResultsCount <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_MAX_RESULTS must be > 0")
	}
	if cfg.DBMaxOpenConns <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_MAX_OPEN_CONNS must be > 0")
	}
	if cfg.DBMaxIdleConns < 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_MAX_IDLE_CONNS must be >= 0")
	}
	if cfg.DBConnMaxLife < 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_CONN_MAX_LIFETIME must be >= 0")
	}
	if cfg.DBConnMaxIdle < 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_CONN_MAX_IDLE_TIME must be >= 0")
	}
	if cfg.DBTimeout <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_DB_TIMEOUT must be > 0")
	}
	if cfg.MaxInflight <= 0 {
		return Config{}, fmt.Errorf("ROLE_HUB_MAX_INFLIGHT must be > 0")
	}

	return cfg, nil
}

func getenvDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func getenvDefaultInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}

func getenvDefaultInt64(key string, def int64) int64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parsed
		}
	}
	return def
}

func getenvDefaultFloat(key string, def float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return def
}

func getenvDefaultDuration(key string, def time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return def
}
