package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr                string
	LogLevel                string
	ShutdownTimeout         time.Duration
	HTTPReadTimeout         time.Duration
	HTTPReadHeaderTimeout   time.Duration
	HTTPWriteTimeout        time.Duration
	HTTPIdleTimeout         time.Duration
	DatabaseURL             string
	DatabaseMaxConns        int32
	DatabaseMinConns        int32
	DatabaseMaxConnLifetime time.Duration
	DatabaseConnectTimeout  time.Duration
	DatabaseQueryTimeout    time.Duration
	IdempotencyTTL          time.Duration
}

func Load() (Config, error) {
	var missing []string
	require := func(name string) string {
		v := os.Getenv(name)
		if v == "" {
			missing = append(missing, name)
		}
		return v
	}

	httpAddr := require("HTTP_ADDR")
	logLevel := require("LOG_LEVEL")
	shutdownRaw := require("SHUTDOWN_TIMEOUT")
	readTimeoutRaw := require("HTTP_READ_TIMEOUT")
	readHeaderTimeoutRaw := require("HTTP_READ_HEADER_TIMEOUT")
	writeTimeoutRaw := require("HTTP_WRITE_TIMEOUT")
	idleTimeoutRaw := require("HTTP_IDLE_TIMEOUT")
	databaseURL := require("DATABASE_URL")
	maxConnsRaw := require("DATABASE_MAX_CONNS")
	minConnsRaw := require("DATABASE_MIN_CONNS")
	lifetimeRaw := require("DATABASE_MAX_CONN_LIFETIME")
	connectRaw := require("DATABASE_CONNECT_TIMEOUT")
	queryRaw := require("DATABASE_QUERY_TIMEOUT")

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("required environment variables are not set: %v", missing)
	}

	shutdownTimeout, err := time.ParseDuration(shutdownRaw)
	if err != nil {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
	}
	readTimeout, err := time.ParseDuration(readTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_READ_TIMEOUT: %w", err)
	}
	readHeaderTimeout, err := time.ParseDuration(readHeaderTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_READ_HEADER_TIMEOUT: %w", err)
	}
	writeTimeout, err := time.ParseDuration(writeTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_WRITE_TIMEOUT: %w", err)
	}
	idleTimeout, err := time.ParseDuration(idleTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_IDLE_TIMEOUT: %w", err)
	}

	maxConns, err := strconv.Atoi(maxConnsRaw)
	if err != nil {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONNS: %w", err)
	}
	minConns, err := strconv.Atoi(minConnsRaw)
	if err != nil {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNS: %w", err)
	}

	lifetime, err := time.ParseDuration(lifetimeRaw)
	if err != nil {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONN_LIFETIME: %w", err)
	}
	connectTimeout, err := time.ParseDuration(connectRaw)
	if err != nil {
		return Config{}, fmt.Errorf("DATABASE_CONNECT_TIMEOUT: %w", err)
	}
	queryTimeout, err := time.ParseDuration(queryRaw)
	if err != nil {
		return Config{}, fmt.Errorf("DATABASE_QUERY_TIMEOUT: %w", err)
	}

	ttl := 24 * time.Hour
	if raw := os.Getenv("IDEMPOTENCY_TTL"); raw != "" {
		ttl, err = time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("IDEMPOTENCY_TTL: %w", err)
		}
	}

	return Config{
		HTTPAddr:                httpAddr,
		LogLevel:                logLevel,
		ShutdownTimeout:         shutdownTimeout,
		HTTPReadTimeout:         readTimeout,
		HTTPReadHeaderTimeout:   readHeaderTimeout,
		HTTPWriteTimeout:        writeTimeout,
		HTTPIdleTimeout:         idleTimeout,
		DatabaseURL:             databaseURL,
		DatabaseMaxConns:        int32(maxConns),
		DatabaseMinConns:        int32(minConns),
		DatabaseMaxConnLifetime: lifetime,
		DatabaseConnectTimeout:  connectTimeout,
		DatabaseQueryTimeout:    queryTimeout,
		IdempotencyTTL:          ttl,
	}, nil
}
