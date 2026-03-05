package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/role-hub/internal/config"
	"github.com/JsonLee12138/agent-team/role-hub/internal/ingest"
	"github.com/JsonLee12138/agent-team/role-hub/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "serve":
		if err := runServe(); err != nil {
			log.Fatalf("serve failed: %v", err)
		}
	case "migrate":
		if err := runMigrate(); err != nil {
			log.Fatalf("migrate failed: %v", err)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("role-hub commands:\n  serve   start HTTP API server\n  migrate apply database migrations")
}

func runServe() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := storage.Open(ctx, cfg.DBDialect, cfg.DBDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db, cfg.DBDialect); err != nil {
		return err
	}

	store, err := storage.NewStore(db, cfg.DBDialect)
	if err != nil {
		return err
	}

	limiter := ingest.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	handler := ingest.NewHandler(store, cfg.MaxBodyBytes, cfg.MaxResultsCount, limiter)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/ingest", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("role-hub listening on %s", cfg.HTTPAddr)
	return server.ListenAndServe()
}

func runMigrate() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := storage.Open(ctx, cfg.DBDialect, cfg.DBDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := storage.ApplyMigrations(ctx, db, cfg.DBDialect); err != nil {
		return err
	}
	log.Printf("migrations applied")
	return nil
}
