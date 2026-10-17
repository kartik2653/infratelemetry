package insights

import (
	"time"

	"chargingdatainfra/backend/internal/analytics"
)

// FromSnapshot converts a calculated snapshot into ranked recommendations.
func FromSnapshot(s *analytics.Snapshot) []Insight {
	if s == nil {
		return nil
	}
	days := s.Query.To.Sub(s.Query.From).Hours() / 24
	if days <= 0 {
		days = 1
	}
	util, utilOK := s.Network.Utilization()
	_, marginPct, marginOK := s.Network.Margin()
	revKWh, revOK := s.Network.RevenuePerKWh()
	costKWh, costOK := s.Network.CostPerKWh()
	net := Network{DowntimeHours: s.Network.Downtime.Hours()}
	if utilOK {
		net.Utilization = &util
	}
	if marginOK {
		net.MarginPct = &marginPct
	}
	if revOK && costOK {
		net.MarginPerKWh = revKWh - costKWh
	}
	stations := make([]Station, 0, len(s.Stations))
	for _, sm := range s.Stations {
		stations = append(stations, stationFact(sm, s.WeekdayDates, s.WeekendDates, days))
	}
	return Generate(net, stations, s.Query.From, s.Query.To)
}

func stationFact(sm analytics.StationMetric, weekdayDates, weekendDates int, days float64) Station {
	util, utilOK := sm.Utilization()
	up, upOK := sm.Uptime()
	_, marginPct, marginOK := sm.Margin()
	revKWh, revOK := sm.RevenuePerKWh()
	costKWh, costOK := sm.CostPerKWh()
	st := Station{
		ID: sm.Station.ID, Name: sm.Station.Name, Area: sm.Station.Area,
		HasDC: sm.HasDC, MaxRatedKW: sm.MaxRatedKW, ConnectorCount: len(sm.Connectors),
		Revenue: sm.Revenue, Energy: sm.Energy, Sessions: sm.Sessions,
		DowntimeHours: sm.Downtime.Hours(), ConnectorFaults: sm.ConnectorFaults, Faults: sm.Totals.Faults,
		PaymentFails: sm.PaymentFails, SessionsByHour: sm.HourSessionStarts,
		RevenueAtRisk: sm.RevenueAtRisk, AvailableHours: sm.Available.Hours(),
	}
	if utilOK {
		v := util
		st.Utilization = &v
	}
	if upOK {
		v := up
		st.Uptime = &v
	}
	if marginOK {
		v := marginPct
		st.MarginPct = &v
	}
	if revOK && costOK {
		st.MarginPerKWh = revKWh - costKWh
	}
	if sm.Sessions > 0 {
		st.AvgPlugMin = float64(sm.DurationSec) / float64(sm.Sessions) / 60
		st.AvgChargeMin = float64(sm.ChargeSec) / float64(sm.Sessions) / 60
	}
	if len(sm.Connectors) > 0 && days > 0 {
		st.EnergyPerConnectorPerDay = sm.Energy / float64(len(sm.Connectors)) / days
	}
	for wd := time.Sunday; wd <= time.Saturday; wd++ {
		if wd == time.Saturday || wd == time.Sunday {
			st.WeekendSessions += sm.DOWSessions[wd]
		} else {
			st.WeekdaySessions += sm.DOWSessions[wd]
		}
	}
	if weekdayDates > 0 {
		st.WeekdayPerDay = float64(st.WeekdaySessions) / float64(weekdayDates)
	}
	if weekendDates > 0 {
		st.WeekendPerDay = float64(st.WeekendSessions) / float64(weekendDates)
	}
	hour, peak, ok := busiest(sm.Totals)
	st.PeakHour = hour
	if ok {
		v := peak
		st.PeakHourUtil = &v
	}
	return st
}

func busiest(t analytics.Totals) (hour int, util float64, ok bool) {
	best := -1.0
	for h := 0; h < 24; h++ {
		if t.HourAvailable[h] < 10*time.Hour {
			continue
		}
		u, good := analytics.UtilizationPct(t.HourCharging[h], t.HourAvailable[h])
		if good && u > best {
			best = u
			hour = h
			util = u
			ok = true
		}
	}
	return hour, util, ok
}
