// Package store keeps the synthetic network in memory.
// Sessions, faults, and maintenance are immutable after New.
// Alerts are the only mutable collection. A later database-backed repository
// can replace this type without changing the analytics calculations.
package store

import (
	"errors"
	"sync"
	"time"

	"chargingdatainfra/backend/internal/domain"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidStatus = errors.New("invalid status")
)

type Store struct {
	ds     domain.Dataset
	alerts []domain.Alert
	mu     sync.RWMutex
}

func New(ds domain.Dataset) *Store {
	return &Store{ds: ds}
}

func (s *Store) Dataset() domain.Dataset { return s.ds }

func (s *Store) Stations() []domain.Station { return s.ds.Stations }

func (s *Store) Connectors() []domain.Connector { return s.ds.Connectors }

func (s *Store) Sessions() []domain.Session { return s.ds.Sessions }

func (s *Store) Faults() []domain.Fault { return s.ds.Faults }

func (s *Store) Maintenance() []domain.MaintenanceWindow { return s.ds.Maintenance }

func (s *Store) Telemetry() []domain.TelemetrySample { return s.ds.Telemetry }

func (s *Store) Station(id string) (domain.Station, bool) {
	for _, st := range s.ds.Stations {
		if st.ID == id {
			return st, true
		}
	}
	return domain.Station{}, false
}

func (s *Store) SetAlerts(alerts []domain.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append([]domain.Alert(nil), alerts...)
}

func (s *Store) Alerts() []domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Alert, len(s.alerts))
	copy(out, s.alerts)
	return out
}

func (s *Store) UpdateAlert(id, status string) (domain.Alert, error) {
	switch domain.AlertStatus(status) {
	case domain.AlertOpen, domain.AlertAcknowledged, domain.AlertResolved:
	default:
		return domain.Alert{}, ErrInvalidStatus
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alerts {
		if s.alerts[i].ID == id {
			s.alerts[i].Status = status
			now := time.Now().UTC()
			s.alerts[i].UpdatedAt = &now
			return s.alerts[i], nil
		}
	}
	return domain.Alert{}, ErrNotFound
}
