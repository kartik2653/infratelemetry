package analytics

import (
	"context"

	"chargingdatainfra/backend/internal/domain"
)

func (s *Service) Summary(ctx context.Context, q domain.Query) (Summary, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return Summary{}, err
	}
	return summaryFrom(snap), nil
}

func (s *Service) Trends(ctx context.Context, q domain.Query) (Trends, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return Trends{}, err
	}
	return trendsFrom(snap), nil
}

func (s *Service) StationRows(ctx context.Context, q domain.Query) ([]StationRow, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return nil, err
	}
	return snap.rows(), nil
}

func (s *Service) Station(ctx context.Context, id string, q domain.Query) (StationDetail, error) {
	q.StationID = id
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return StationDetail{}, err
	}
	if len(snap.Stations) == 0 {
		for _, st := range s.src.Stations() {
			if st.ID == id {
				return detailFrom(snap, StationMetric{
					Station: st, Totals: Totals{FaultsByCategory: map[string]int{}},
					StatusCounts: map[string]int{}, WorstStatus: string(domain.StatusAvailable),
				}), nil
			}
		}
		return StationDetail{}, ErrUnknownStation
	}
	return detailFrom(snap, snap.Stations[0]), nil
}

func (s *Service) Utilization(ctx context.Context, q domain.Query) (UtilizationReport, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return UtilizationReport{}, err
	}
	return utilizationFrom(snap), nil
}

func (s *Service) Revenue(ctx context.Context, q domain.Query) (RevenueReport, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return RevenueReport{}, err
	}
	return revenueFrom(snap), nil
}

func (s *Service) Energy(ctx context.Context, q domain.Query) (EnergyReport, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return EnergyReport{}, err
	}
	return energyFrom(snap), nil
}

func (s *Service) Reliability(ctx context.Context, q domain.Query) (ReliabilityReport, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return ReliabilityReport{}, err
	}
	return reliabilityFrom(snap), nil
}

func (s *Service) SessionsAnalytics(ctx context.Context, q domain.Query) (SessionReport, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return SessionReport{}, err
	}
	return sessionReportFrom(snap), nil
}

func (s *Service) ListSessions(ctx context.Context, q domain.Query, status string) ([]SessionDTO, error) {
	snap, err := s.Compute(ctx, q)
	if err != nil {
		return nil, err
	}
	return snap.sessionDTOs(status), nil
}
