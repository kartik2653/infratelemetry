// Package seed builds a deterministic synthetic charging network.
// The random source is seeded, and fault plans are explicit, so two
// processes with the same options produce the same sessions.
package seed

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"chargingdatainfra/backend/internal/domain"
)

type Options struct {
	Seed     int64
	Days     int
	AsOf     time.Time
	Location *time.Location
}

type generator struct {
	opt       Options
	rng       *rand.Rand
	dataStart time.Time
	sessions  []domain.Session
	faults    []domain.Fault
	maint     []domain.MaintenanceWindow
	telemetry []domain.TelemetrySample
	seqSes    int
	seqFault  int
	seqMaint  int
	seqTel    int
	blocked   map[string][]domainInterval
}

type domainInterval struct {
	start time.Time
	end   time.Time
}

func Generate(opt Options) domain.Dataset {
	if opt.Location == nil {
		loc, err := time.LoadLocation(domain.Timezone)
		if err != nil {
			loc = time.FixedZone("IST", 5*3600+30*60)
		}
		opt.Location = loc
	}
	if opt.Days <= 0 {
		opt.Days = 90
	}
	if opt.Seed == 0 {
		opt.Seed = 20261017
	}
	asOfDay := time.Date(opt.AsOf.Year(), opt.AsOf.Month(), opt.AsOf.Day(), 0, 0, 0, 0, opt.Location)
	g := &generator{
		opt:       opt,
		rng:       rand.New(rand.NewSource(opt.Seed)),
		dataStart: asOfDay.AddDate(0, 0, -opt.Days),
		blocked:   map[string][]domainInterval{},
	}
	specs := blueprint()
	stations := make([]domain.Station, 0, len(specs))
	connectors := make([]domain.Connector, 0, 60)
	byStation := map[string][]domain.Connector{}
	commissioned := time.Date(2024, 6, 1, 0, 0, 0, 0, opt.Location)

	for _, spec := range specs {
		st := domain.Station{
			ID: spec.ID, Name: spec.Name, Area: spec.Area, Address: spec.Address,
			City: "Bengaluru", Latitude: spec.Lat, Longitude: spec.Lng,
			CommissionedAt: commissioned,
			PeakTariffINR:  spec.PeakTariff, OffPeakTariffINR: spec.OffPeakTariff,
			PeakEnergyCostINR: spec.PeakCost, OffPeakEnergyCostINR: spec.OffPeakCost,
		}
		stations = append(stations, st)
		labels := []string{"A", "B", "C", "D", "E"}
		for i, cs := range spec.Conns {
			c := domain.Connector{
				ID:           fmt.Sprintf("%s-%s", spec.ID, labels[i]),
				StationID:    spec.ID,
				Label:        labels[i],
				ChargerType:  cs.Type,
				Standard:     cs.Standard,
				RatedPowerKW: cs.KW,
				Status:       domain.StatusAvailable,
			}
			connectors = append(connectors, c)
			byStation[spec.ID] = append(byStation[spec.ID], c)
			g.blocked[c.ID] = nil
		}
	}

	g.planMaintenance(stations)
	g.planFaults(byStation)
	for _, spec := range specs {
		st := stations[indexStation(stations, spec.ID)]
		g.sessionsForStation(spec, st, byStation[spec.ID])
	}
	g.occupy(byStation)
	g.deriveStatus(connectors)

	return domain.Dataset{
		OperatorName: "NammaCharge Grid",
		NetworkName:  "NammaCharge Bengaluru Network",
		Disclaimer:   domain.Disclaimer,
		Seed:         opt.Seed,
		Days:         opt.Days,
		Timezone:     domain.Timezone,
		AsOf:         opt.AsOf,
		DataStart:    g.dataStart,
		Stations:     stations,
		Connectors:   connectors,
		Sessions:     g.sessions,
		Faults:       g.faults,
		Maintenance:  g.maint,
		Telemetry:    g.telemetry,
	}
}

func indexStation(stations []domain.Station, id string) int {
	for i := range stations {
		if stations[i].ID == id {
			return i
		}
	}
	return 0
}

func (g *generator) planMaintenance(stations []domain.Station) {
	loc := g.opt.Location
	t := time.Date(g.dataStart.Year(), g.dataStart.Month(), 5, 2, 0, 0, 0, loc)
	if t.Before(g.dataStart) {
		t = t.AddDate(0, 1, 0)
	}
	for !t.After(g.opt.AsOf) {
		end := t.Add(3 * time.Hour)
		if end.After(g.dataStart) && t.Before(g.opt.AsOf) {
			for _, st := range stations {
				g.seqMaint++
				g.maint = append(g.maint, domain.MaintenanceWindow{
					ID:        fmt.Sprintf("MNT-%03d", g.seqMaint),
					StationID: st.ID,
					Start:     t,
					End:       end,
					Reason:    "Scheduled preventive maintenance",
				})
				g.blockStation(st.ID, t, end)
			}
		}
		t = time.Date(t.Year(), t.Month()+1, 5, 2, 0, 0, 0, loc)
	}
}

func (g *generator) planFaults(byStation map[string][]domain.Connector) {
	// Hebbal: alternating connector faults on A and C. B carries the long open fault.
	for day := 1; day <= 84; day += 3 {
		label := "A"
		if (day/3)%2 == 1 {
			label = "C"
		}
		durH := 8 + (day % 6)
		start := g.at(day, 11, 0)
		end := start.Add(time.Duration(durH) * time.Hour)
		if !end.Before(g.opt.AsOf) {
			continue
		}
		g.addFault(byStation["BLR-006"], "BLR-006-"+label, domain.FaultConnector, "warning", start, &end,
			"Connector failed to latch and was returned to service after a site visit")
	}
	openStart := g.opt.AsOf.Add(-120 * time.Hour)
	g.addFault(byStation["BLR-006"], "BLR-006-B", domain.FaultConnector, "critical", openStart, nil,
		"Connector B is locked out after repeated contactor faults and is still open")

	// Rajajinagar: one station-wide power-supply outage, resolved.
	ps := g.at(62, 6, 0)
	pe := ps.Add(72 * time.Hour)
	g.addFault(byStation["BLR-014"], "", domain.FaultPowerSupply, "critical", ps, &pe,
		"Upstream supply failed for the whole site. All connectors were offline until the feeder was restored")

	// JP Nagar: recurring communication dropouts plus one open incident.
	for day := 4; day <= 80; day += 7 {
		start := g.at(day, 16, 30)
		end := start.Add(6 * time.Hour)
		if !end.Before(g.opt.AsOf) {
			continue
		}
		g.addFault(byStation["BLR-015"], "BLR-015-A", domain.FaultCommunication, "warning", start, &end,
			"Charger stopped reporting to the network and rejected new sessions until the link returned")
	}
	commStart := g.opt.AsOf.Add(-4 * time.Hour)
	g.addFault(byStation["BLR-015"], "BLR-015-A", domain.FaultCommunication, "warning", commStart, nil,
		"Connector A is not reporting status and is unavailable for new sessions")

	// Marathahalli payment terminal, including an open incident.
	for _, day := range []int{18, 47} {
		start := g.at(day, 9, 0)
		end := start.Add(8 * time.Hour)
		g.addFault(byStation["BLR-010"], "BLR-010-B", domain.FaultPayment, "warning", start, &end,
			"Payment terminal on connector B rejected authorisations")
	}
	payStart := g.opt.AsOf.Add(-10 * time.Hour)
	g.addFault(byStation["BLR-010"], "BLR-010-B", domain.FaultPayment, "warning", payStart, nil,
		"Payment terminal on connector B is still declining authorisations")

	// KR Puram temperature, one resolved and one open.
	ts := g.at(40, 13, 0)
	te := ts.Add(5 * time.Hour)
	g.addFault(byStation["BLR-018"], "BLR-018-A", domain.FaultTemperature, "critical", ts, &te,
		"DC cabinet temperature alarm stopped charging until the cabinet cooled")
	tempOpen := g.opt.AsOf.Add(-3 * time.Hour)
	g.addFault(byStation["BLR-018"], "BLR-018-A", domain.FaultTemperature, "critical", tempOpen, nil,
		"Temperature alarm is active on the 150 kW cabinet")

	// A short emergency stop and a handful of isolated incidents so downtime
	// is concentrated but not exclusive to three sites.
	es := g.at(22, 19, 10)
	ee := es.Add(70 * time.Minute)
	g.addFault(byStation["BLR-013"], "BLR-013-A", domain.FaultEmergency, "critical", es, &ee,
		"Emergency stop was pressed and reset after inspection")

	g.isolated(byStation, "BLR-001", "BLR-001-B", domain.FaultCommunication, "warning", 28, 8, 2,
		"Brief backhaul dropout")
	g.isolated(byStation, "BLR-003", "BLR-003-A", domain.FaultPowerSupply, "critical", 44, 7, 3,
		"Site supply breaker tripped and was reset")
	g.isolated(byStation, "BLR-011", "BLR-011-A", domain.FaultConnector, "warning", 36, 15, 4,
		"Single connector fault, repaired the same day")
	g.isolated(byStation, "BLR-016", "BLR-016-A", domain.FaultConnector, "warning", 55, 12, 5,
		"Cable cooling fault cleared after a restart")
	g.isolated(byStation, "BLR-020", "BLR-020-C", domain.FaultCommunication, "warning", 70, 18, 2,
		"Network switch reboot interrupted reporting")
	g.isolated(byStation, "BLR-008", "BLR-008-A", domain.FaultConnector, "warning", 33, 10, 3,
		"CCS latch replaced")
}

func (g *generator) isolated(byStation map[string][]domain.Connector, station, conn string, cat domain.FaultCategory, sev string, day, hour, durH int, desc string) {
	start := g.at(day, hour, 0)
	end := start.Add(time.Duration(durH) * time.Hour)
	if !start.Before(g.opt.AsOf) || !end.After(g.dataStart) {
		return
	}
	if end.After(g.opt.AsOf) {
		end = g.opt.AsOf
	}
	g.addFault(byStation[station], conn, cat, sev, start, &end, desc)
}

func (g *generator) addFault(conns []domain.Connector, connectorID string, cat domain.FaultCategory, sev string, start time.Time, end *time.Time, desc string) {
	if len(conns) == 0 || !start.Before(g.opt.AsOf) {
		return
	}
	g.seqFault++
	stationID := conns[0].StationID
	g.faults = append(g.faults, domain.Fault{
		ID: fmt.Sprintf("FLT-%03d", g.seqFault), StationID: stationID, ConnectorID: connectorID,
		Category: cat, Severity: sev, Start: start, End: end, Description: desc,
	})
	stop := g.opt.AsOf
	if end != nil {
		stop = *end
	}
	if connectorID == "" {
		g.blockStation(stationID, start, stop)
		return
	}
	g.blocked[connectorID] = append(g.blocked[connectorID], domainInterval{start, stop})
}

func (g *generator) blockStation(stationID string, start, end time.Time) {
	// Station id is a prefix of connector ids: BLR-014-A.
	for id := range g.ensureKnown(stationID) {
		g.blocked[id] = append(g.blocked[id], domainInterval{start, end})
	}
}

// ensureKnown returns connector ids already present in the block map for a
// station. Connectors are registered when maintenance is planned, which runs
// first and blocks every connector. Faults therefore always find them.
func (g *generator) ensureKnown(stationID string) map[string]struct{} {
	out := map[string]struct{}{}
	prefix := stationID + "-"
	for id := range g.blocked {
		if len(id) > len(prefix) && id[:len(prefix)] == prefix {
			out[id] = struct{}{}
		}
	}
	return out
}

func (g *generator) at(day, hour, min int) time.Time {
	d := g.dataStart.AddDate(0, 0, day)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, min, 0, 0, g.opt.Location)
}

func (g *generator) sessionsForStation(spec stationSpec, st domain.Station, conns []domain.Connector) {
	for _, c := range conns {
		var cursor time.Time
		for day := g.dataStart; day.Before(g.opt.AsOf); day = day.AddDate(0, 0, 1) {
			mean := g.meanFor(spec, c, day)
			n := sampleCount(g.rng, mean)
			if n == 0 {
				continue
			}
			weights := g.weightsFor(spec, day)
			cands := make([]time.Time, 0, n)
			for i := 0; i < n; i++ {
				h := pickHour(g.rng, weights)
				m := g.rng.Intn(60)
				t := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, g.opt.Location)
				if t.Before(g.dataStart) || !t.Before(g.opt.AsOf) {
					continue
				}
				cands = append(cands, t)
			}
			sort.Slice(cands, func(i, j int) bool { return cands[i].Before(cands[j]) })
			for _, start := range cands {
				if start.Before(cursor) {
					continue
				}
				if g.rng.Float64() < spec.PayFail {
					reason := "payment_failure"
					if g.rng.Float64() < 0.15 {
						reason = "communication_failure"
					}
					end := start.Add(3 * time.Minute)
					if !end.After(start) || !start.Before(g.opt.AsOf) {
						continue
					}
					if end.After(g.opt.AsOf) {
						end = g.opt.AsOf
					}
					if g.overlapsBlock(c.ID, start, end) {
						continue
					}
					g.addSession(spec, st, c, start, end, 0, domain.SessionFailed, reason, 0)
					cursor = end.Add(8 * time.Minute)
					continue
				}
				chargeMin := durationMin(c, g.rng)
				plugMin := chargeMin
				if c.ChargerType == domain.ChargerAC && spec.HogAC > 0 && g.rng.Float64() < spec.HogAC {
					plugMin = chargeMin + 300 + g.rng.Intn(180)
				}
				if c.ChargerType == domain.ChargerDC && spec.HogDC > 0 && g.rng.Float64() < spec.HogDC {
					chargeMin = 100 + g.rng.Intn(40)
					plugMin = chargeMin
				}
				status := domain.SessionCompleted
				reason := ""
				if g.rng.Float64() < spec.Interrupt {
					cut := 0.35 + g.rng.Float64()*0.3
					chargeMin = int(math.Max(5, float64(chargeMin)*cut))
					plugMin = chargeMin
					status = domain.SessionInterrupted
					reason = "connector_failure"
					if g.rng.Float64() < 0.25 {
						reason = "power_supply_failure"
					}
				}
				end := start.Add(time.Duration(plugMin) * time.Minute)
				chargeEnd := start.Add(time.Duration(chargeMin) * time.Minute)
				if end.After(g.opt.AsOf) {
					if start.Add(8 * time.Minute).After(g.opt.AsOf) {
						continue
					}
					end = g.opt.AsOf
					if chargeEnd.After(end) {
						chargeEnd = end
					}
					status = domain.SessionInProgress
					reason = ""
				}
				if g.overlapsBlock(c.ID, start, end) {
					continue
				}
				g.addSession(spec, st, c, start, end, int64(chargeEnd.Sub(start).Seconds()), status, reason, loadFactor(c, g.rng))
				cursor = end.Add(8 * time.Minute)
			}
		}
	}
}

func (g *generator) occupy(byStation map[string][]domain.Connector) {
	wanted := []string{"BLR-001-A", "BLR-003-A", "BLR-007-A", "BLR-011-A", "BLR-013-A", "BLR-020-A", "BLR-002-C", "BLR-005-A", "BLR-008-B", "BLR-019-A"}
	lookup := map[string]domain.Connector{}
	var stations = map[string]stationSpec{}
	for _, spec := range blueprint() {
		stations[spec.ID] = spec
	}
	for _, list := range byStation {
		for _, c := range list {
			lookup[c.ID] = c
		}
	}
	for _, id := range wanted {
		c, ok := lookup[id]
		if !ok {
			continue
		}
		start := g.opt.AsOf.Add(-28 * time.Minute)
		if g.overlapsBlock(c.ID, start, g.opt.AsOf) || g.overlapsSession(c.ID, start, g.opt.AsOf) {
			continue
		}
		spec := stations[c.StationID]
		st := domain.Station{
			ID: spec.ID, PeakTariffINR: spec.PeakTariff, OffPeakTariffINR: spec.OffPeakTariff,
			PeakEnergyCostINR: spec.PeakCost, OffPeakEnergyCostINR: spec.OffPeakCost,
		}
		g.addSession(spec, st, c, start, g.opt.AsOf, int64(28*time.Minute.Seconds()), domain.SessionInProgress, "", loadFactor(c, g.rng))
	}
}

func (g *generator) overlapsSession(connectorID string, start, end time.Time) bool {
	for _, s := range g.sessions {
		if s.ConnectorID != connectorID {
			continue
		}
		if start.Before(s.End) && s.Start.Before(end) {
			return true
		}
	}
	return false
}

func (g *generator) overlapsBlock(connectorID string, start, end time.Time) bool {
	for _, iv := range g.blocked[connectorID] {
		if start.Before(iv.end) && iv.start.Before(end) {
			return true
		}
	}
	return false
}

func (g *generator) addSession(spec stationSpec, st domain.Station, c domain.Connector, start, plugEnd time.Time, chargeSec int64, status domain.SessionStatus, reason string, load float64) {
	if !plugEnd.After(start) {
		return
	}
	plugSec := int64(plugEnd.Sub(start).Seconds())
	if chargeSec > plugSec {
		chargeSec = plugSec
	}
	if status == domain.SessionFailed {
		chargeSec = 0
	}
	var energy, avg, peak, tariff, revenue, cost float64
	if chargeSec > 0 && status != domain.SessionFailed {
		chargeH := float64(chargeSec) / 3600
		avg = c.RatedPowerKW * load
		energy = avg * chargeH
		capKWh := 48.0
		if c.ChargerType == domain.ChargerAC {
			capKWh = 30
		}
		if energy > capKWh {
			energy = capKWh
		}
		if spec.DerateFromDay > 0 {
			day := int(start.Sub(g.dataStart).Hours() / 24)
			if day >= spec.DerateFromDay {
				energy *= 0.55
				avg *= 0.55
			}
		}
		if chargeH > 0 {
			avg = energy / chargeH
		}
		peak = math.Min(c.RatedPowerKW, avg*1.16)
		if domain.IsPeakHour(start, g.opt.Location) {
			tariff = st.PeakTariffINR
			cost = energy * st.PeakEnergyCostINR
		} else {
			tariff = st.OffPeakTariffINR
			cost = energy * st.OffPeakEnergyCostINR
		}
		revenue = energy * tariff
		energy = domain.Round2(energy)
		revenue = domain.Round2(revenue)
		cost = domain.Round2(cost)
		avg = domain.Round2(avg)
		peak = domain.Round2(peak)
		// Recompute from rounded energy so revenue matches the stored kWh.
		revenue = domain.Round2(energy * tariff)
		cost = domain.Round2(energy * costRate(st, start, g.opt.Location))
	}
	g.seqSes++
	ses := domain.Session{
		ID: fmt.Sprintf("SES-%06d", g.seqSes), StationID: c.StationID, ConnectorID: c.ID,
		Start: start, End: plugEnd, DurationSec: plugSec, ChargeDurationSec: chargeSec,
		EnergyKWh: energy, AvgPowerKW: avg, PeakPowerKW: peak,
		TariffINRPerKWh: tariff, RevenueINR: revenue, EnergyCostINR: cost,
		Status: status, TerminationReason: reason,
		ChargerType: c.ChargerType, RatedPowerKW: c.RatedPowerKW,
	}
	g.sessions = append(g.sessions, ses)
	g.appendTelemetry(ses)
}

func costRate(st domain.Station, start time.Time, loc *time.Location) float64 {
	if domain.IsPeakHour(start, loc) {
		return st.PeakEnergyCostINR
	}
	return st.OffPeakEnergyCostINR
}

func (g *generator) appendTelemetry(ses domain.Session) {
	if ses.ChargeDurationSec <= 0 || ses.EnergyKWh <= 0 {
		return
	}
	chargeEnd := ses.Start.Add(time.Duration(ses.ChargeDurationSec) * time.Second)
	if chargeEnd.After(ses.End) {
		chargeEnd = ses.End
	}
	total := chargeEnd.Sub(ses.Start).Seconds()
	if total <= 0 {
		return
	}
	var slices []domain.TelemetrySample
	var summed float64
	for cur := ses.Start; cur.Before(chargeEnd); {
		next := nextHour(cur, g.opt.Location)
		if next.After(chargeEnd) {
			next = chargeEnd
		}
		sec := next.Sub(cur).Seconds()
		energy := ses.EnergyKWh * sec / total
		power := 0.0
		if sec > 0 {
			power = energy / (sec / 3600)
		}
		g.seqTel++
		slices = append(slices, domain.TelemetrySample{
			ID: fmt.Sprintf("TEL-%07d", g.seqTel), StationID: ses.StationID, ConnectorID: ses.ConnectorID,
			SessionID: ses.ID, Timestamp: cur, PowerKW: domain.Round2(power), EnergyKWh: energy,
		})
		summed += energy
		cur = next
	}
	if len(slices) > 0 {
		slices[len(slices)-1].EnergyKWh += ses.EnergyKWh - summed
		slices[len(slices)-1].EnergyKWh = domain.Round2(slices[len(slices)-1].EnergyKWh)
		for i := range slices[:len(slices)-1] {
			slices[i].EnergyKWh = domain.Round2(slices[i].EnergyKWh)
		}
		// Rounding the earlier slices can leave a residual. Put it on the last sample.
		var got float64
		for i := range slices[:len(slices)-1] {
			got += slices[i].EnergyKWh
		}
		slices[len(slices)-1].EnergyKWh = domain.Round2(ses.EnergyKWh - got)
	}
	g.telemetry = append(g.telemetry, slices...)
}

func nextHour(t time.Time, loc *time.Location) time.Time {
	lt := t.In(loc)
	n := time.Date(lt.Year(), lt.Month(), lt.Day(), lt.Hour(), 0, 0, 0, loc).Add(time.Hour)
	if !n.After(t) {
		return n.Add(time.Hour)
	}
	return n
}

func (g *generator) deriveStatus(connectors []domain.Connector) {
	asOf := g.opt.AsOf
	open := map[string]domain.Fault{}
	var stationOpen []domain.Fault
	for _, f := range g.faults {
		if !f.Ongoing(asOf) || f.Start.After(asOf) {
			continue
		}
		if f.ConnectorID == "" {
			stationOpen = append(stationOpen, f)
			continue
		}
		open[f.ConnectorID] = f
	}
	active := map[string]bool{}
	for _, s := range g.sessions {
		if s.Status == domain.SessionInProgress && !s.Start.After(asOf) && s.End.Equal(asOf) {
			active[s.ConnectorID] = true
		}
	}
	maintNow := map[string]bool{}
	for _, m := range g.maint {
		if !m.Start.After(asOf) && m.End.After(asOf) {
			if m.ConnectorID == "" {
				maintNow[m.StationID] = true
			} else {
				maintNow[m.ConnectorID] = true
			}
		}
	}
	for i := range connectors {
		c := &connectors[i]
		if f, ok := open[c.ID]; ok {
			if domain.HardwareFault(f.Category) {
				c.Status = domain.StatusFaulted
			} else {
				c.Status = domain.StatusUnavailable
			}
			continue
		}
		stationFaulted := false
		stationUnavail := false
		for _, f := range stationOpen {
			if f.StationID != c.StationID {
				continue
			}
			if domain.HardwareFault(f.Category) {
				stationFaulted = true
			} else {
				stationUnavail = true
			}
		}
		if stationFaulted {
			c.Status = domain.StatusFaulted
			continue
		}
		if stationUnavail || maintNow[c.StationID] || maintNow[c.ID] {
			c.Status = domain.StatusUnavailable
			continue
		}
		if active[c.ID] {
			c.Status = domain.StatusOccupied
			continue
		}
		c.Status = domain.StatusAvailable
	}
}

func (g *generator) meanFor(spec stationSpec, c domain.Connector, day time.Time) float64 {
	base := spec.MeanAC
	if c.ChargerType == domain.ChargerDC {
		base = spec.MeanDC
	}
	if base <= 0 {
		return 0
	}
	wd := day.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return base * spec.WeekendMul
	}
	return base * spec.WeekdayMul
}

func (g *generator) weightsFor(spec stationSpec, day time.Time) [24]int {
	if spec.Weight == "weekend_bias" {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			return weight("midday")
		}
		return weight("flat")
	}
	return weight(spec.Weight)
}

func weight(name string) [24]int {
	var w [24]int
	for i := range w {
		w[i] = 1
	}
	set := func(hours []int, v int) {
		for _, h := range hours {
			w[h] = v
		}
	}
	switch name {
	case "commute":
		set([]int{0, 1, 2, 3, 4, 5}, 0)
		set([]int{6, 22, 23}, 1)
		set([]int{7, 11, 16, 21}, 3)
		set([]int{8, 9, 10, 17, 18, 19}, 14)
		set([]int{12, 13, 14, 15, 20}, 2)
	case "evening":
		set([]int{0, 1, 2, 3, 4, 5}, 0)
		set([]int{6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 1)
		set([]int{16, 22}, 4)
		set([]int{17, 21}, 8)
		set([]int{18, 19, 20}, 22)
	case "office":
		for i := range w {
			w[i] = 0
		}
		set([]int{8, 18}, 3)
		set([]int{9, 10, 11, 12, 13, 14, 15, 16, 17}, 12)
	case "busy":
		set([]int{0, 1, 2, 3, 4, 5}, 0)
		set([]int{6, 22, 23}, 2)
		set([]int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, 6)
	case "highway":
		set([]int{0, 1, 2, 3, 4}, 1)
		set([]int{5, 6, 21, 22, 23}, 2)
		set([]int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 5)
	case "midday":
		set([]int{0, 1, 2, 3, 4, 5, 6}, 0)
		set([]int{7, 8, 9, 10, 18, 19, 20, 21}, 2)
		set([]int{11, 12, 13, 14, 15, 16}, 12)
	case "flat":
		set([]int{0, 1, 2, 3, 4, 5}, 1)
		for h := 6; h <= 22; h++ {
			w[h] = 2
		}
	}
	return w
}

func pickHour(r *rand.Rand, w [24]int) int {
	total := 0
	for _, v := range w {
		total += v
	}
	if total <= 0 {
		return r.Intn(24)
	}
	n := r.Intn(total)
	for h, v := range w {
		n -= v
		if n < 0 {
			return h
		}
	}
	return 23
}

func sampleCount(r *rand.Rand, mean float64) int {
	if mean <= 0 {
		return 0
	}
	v := mean + r.NormFloat64()*math.Sqrt(mean)*0.45
	if v < 0.5 {
		return 0
	}
	n := int(math.Round(v))
	if n > 18 {
		n = 18
	}
	return n
}

func durationMin(c domain.Connector, r *rand.Rand) int {
	switch {
	case c.ChargerType == domain.ChargerDC && c.RatedPowerKW >= 100:
		return 16 + r.Intn(22)
	case c.ChargerType == domain.ChargerDC:
		return 22 + r.Intn(26)
	case c.RatedPowerKW >= 22:
		return 45 + r.Intn(60)
	default:
		return 55 + r.Intn(80)
	}
}

func loadFactor(c domain.Connector, r *rand.Rand) float64 {
	switch {
	case c.ChargerType == domain.ChargerDC && c.RatedPowerKW >= 100:
		return 0.42 + r.Float64()*0.22
	case c.ChargerType == domain.ChargerDC:
		return 0.58 + r.Float64()*0.25
	default:
		return 0.75 + r.Float64()*0.2
	}
}
