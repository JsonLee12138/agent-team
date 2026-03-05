package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config controls role-hub runtime behavior.
type Config struct {
	HTTPAddr        string
	DBDialect       string
	DBDSN           string
	RateLimitRPS    float64
	RateLimitBurst  int
	MaxBodyBytes    int64
	MaxResultsCount int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        getenvDefault("ROLE_HUB_HTTP_ADDR", ":8080"),
		DBDialect:       getenvDefault("ROLE_HUB_DB_DIALECT", "postgres"),
		DBDSN:           getenvDefault("ROLE_HUB_DB_DSN", ""),
		RateLimitRPS:    getenvDefaultFloat("ROLE_HUB_RATE_LIMIT_RPS", 5),
		RateLimitBurst:  getenvDefaultInt("ROLE_HUB_RATE_LIMIT_BURST", 10),
		MaxBodyBytes:    getenvDefaultInt64("ROLE_HUB_MAX_BODY_BYTES", 1<<20), // 1MB
		MaxResultsCount: getenvDefaultInt("ROLE_HUB_MAX_RESULTS", 500),
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
