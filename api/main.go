// SPDX-License-Identifier: Apache-2.0
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

	"github.com/kikudesuyo/error-agent/api/library"
	"github.com/kikudesuyo/error-agent/api/route"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	ingestToken := os.Getenv("INGEST_TOKEN")
	applicationEnvironment := os.Getenv("APP_ENV")
	if databaseURL == "" || ingestToken == "" {
		slog.Error("DATABASE_URL and INGEST_TOKEN must be configured")
		os.Exit(1)
	}
	if err := library.ValidateIngestToken(ingestToken); err != nil {
		slog.Error("ingest token configuration is invalid")
		os.Exit(1)
	}
	if err := library.ValidateDatabaseTransport(databaseURL, applicationEnvironment); err != nil {
		slog.Error("database transport configuration is invalid")
		os.Exit(1)
	}

	db, err := library.OpenDatabase(databaseURL)
	if err != nil {
		slog.Error("database initialization failed")
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("database pool initialization failed")
		os.Exit(1)
	}
	defer sqlDB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           route.NewMainRoute(db, ingestToken),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrorList := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "port", port)
		serverErrorList <- server.ListenAndServe()
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrorList:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server stopped unexpectedly")
			os.Exit(1)
		}
	case <-signalContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			slog.Error("HTTP server shutdown failed")
			os.Exit(1)
		}
	}
}
