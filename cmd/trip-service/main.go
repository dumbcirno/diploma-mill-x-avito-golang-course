package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dumbcirno/diploma-mill-x-avito-golang-course/internal/api"
	"github.com/dumbcirno/diploma-mill-x-avito-golang-course/internal/config"
	spec "github.com/dumbcirno/diploma-mill-x-avito-golang-course/internal/generated"
	"github.com/dumbcirno/diploma-mill-x-avito-golang-course/internal/trip"
	"github.com/dumbcirno/diploma-mill-x-avito-golang-course/internal/tx"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := newPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, cfg.DatabaseConnectTimeout)
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	txm := tx.NewManager(pool)
	repo := trip.NewRepository(pool, cfg.DatabaseQueryTimeout)
	svc := trip.NewService(txm, repo, cfg.IdempotencyTTL)
	handler := api.NewHandler(svc, pool, cfg.DatabaseQueryTimeout)

	router := chi.NewRouter()
	spec.HandlerWithOptions(handler, spec.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			api.InvalidParam(w, r, err)
		},
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("trip-service listening on %s", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Printf("shutdown signal received, waiting up to %s", cfg.ShutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("forced shutdown after %s: %v", cfg.ShutdownTimeout, err)
			_ = server.Close()
		}
		return nil
	}
}

func newPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	poolCfg.MaxConns = cfg.DatabaseMaxConns
	poolCfg.MinConns = cfg.DatabaseMinConns
	poolCfg.MaxConnLifetime = cfg.DatabaseMaxConnLifetime
	poolCfg.ConnConfig.ConnectTimeout = cfg.DatabaseConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	return pool, nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}
