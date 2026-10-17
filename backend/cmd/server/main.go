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

	"chargingdatainfra/backend/internal/alerts"
	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/config"
	"chargingdatainfra/backend/internal/domain"
	"chargingdatainfra/backend/internal/httpapi"
	"chargingdatainfra/backend/internal/seed"
	"chargingdatainfra/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}
	logger.Info("generating demo network", "seed", cfg.Seed, "days", cfg.Days, "as_of", cfg.AsOf.Format(time.RFC3339))
	ds := seed.Generate(seed.Options{Seed: cfg.Seed, Days: cfg.Days, AsOf: cfg.AsOf, Location: cfg.Location})
	st := store.New(ds)
	svc := analytics.New(st)
	ctx := context.Background()
	asOf := ds.AsOf
	last30, err := svc.Compute(ctx, domain.Query{From: asOf.AddDate(0, 0, -30), To: asOf})
	if err != nil {
		logger.Error("alert window", "err", err)
		os.Exit(1)
	}
	last14, err := svc.Compute(ctx, domain.Query{From: asOf.AddDate(0, 0, -14), To: asOf})
	if err != nil {
		logger.Error("alert window", "err", err)
		os.Exit(1)
	}
	last7, err := svc.Compute(ctx, domain.Query{From: asOf.AddDate(0, 0, -7), To: asOf})
	if err != nil {
		logger.Error("alert window", "err", err)
		os.Exit(1)
	}
	prior, err := svc.Compute(ctx, domain.Query{From: asOf.AddDate(0, 0, -30), To: asOf.AddDate(0, 0, -7)})
	if err != nil {
		logger.Error("alert window", "err", err)
		os.Exit(1)
	}
	generated := alerts.Generate(asOf, st.Faults(), last30, last14, last7, prior)
	st.SetAlerts(generated)
	logger.Info("demo network ready",
		"stations", len(ds.Stations), "connectors", len(ds.Connectors),
		"sessions", len(ds.Sessions), "faults", len(ds.Faults), "alerts", len(generated),
		"disclaimer", ds.Disclaimer)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(logger, st, svc),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		logger.Info("listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
