// Package config reads process settings. The demo clock defaults to a fixed
// instant so restarts reproduce the same network. Override DATA_AS_OF when
// you want a different window.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

type Config struct {
	Addr     string
	Seed     int64
	Days     int
	AsOf     time.Time
	Location *time.Location
}

func Load() (Config, error) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return Config{}, fmt.Errorf("timezone: %w", err)
	}
	cfg := Config{
		Addr:     ":8080",
		Seed:     20261017,
		Days:     90,
		AsOf:     time.Date(2026, 10, 17, 18, 30, 0, 0, loc),
		Location: loc,
	}
	if v := strings.TrimSpace(os.Getenv("ADDR")); v != "" {
		cfg.Addr = v
	} else if v := strings.TrimSpace(os.Getenv("PORT")); v != "" {
		if strings.HasPrefix(v, ":") {
			cfg.Addr = v
		} else {
			cfg.Addr = ":" + v
		}
	}
	if v := strings.TrimSpace(os.Getenv("DATA_SEED")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("DATA_SEED: %w", err)
		}
		cfg.Seed = n
	}
	if v := strings.TrimSpace(os.Getenv("DATA_DAYS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("DATA_DAYS: %w", err)
		}
		cfg.Days = n
	}
	if cfg.Days < 7 || cfg.Days > 180 {
		return Config{}, fmt.Errorf("DATA_DAYS must be between 7 and 180")
	}
	if v := strings.TrimSpace(os.Getenv("DATA_AS_OF")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return Config{}, fmt.Errorf("DATA_AS_OF must be RFC3339: %w", err)
		}
		cfg.AsOf = t.In(loc)
	}
	return cfg, nil
}
