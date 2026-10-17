package analytics

import (
	"context"
	"errors"
	"time"

	"chargingdatainfra/backend/internal/domain"
)

var (
	ErrInvalidRange   = errors.New("to must be after from")
	ErrRangeTooLong   = errors.New("time range cannot exceed 366 days")
	ErrUnknownStation = errors.New("station not found")
	ErrInvalidFilter  = errors.New("charger_type must be AC or DC")
)

// Source is the read seam. The in-memory store implements it today; a
// database repository can replace it without changing calculations.
type Source interface {
	Dataset() domain.Dataset
	Stations() []domain.Station
	Connectors() []domain.Connector
	Sessions() []domain.Session
	Faults() []domain.Fault
	Maintenance() []domain.MaintenanceWindow
}

type Service struct {
	src Source
}

func New(src Source) *Service { return &Service{src: src} }

// Totals accumulates one scope (network, station, connector, or day).
// Durations are sums across connectors, so downtime is connector-hours.
type Totals struct {
	Charging    time.Duration
	Available   time.Duration
	Maintenance time.Duration
	Downtime    time.Duration
	Scheduled   time.Duration
	Up          time.Duration

	PeakCharging  time.Duration
	PeakAvailable time.Duration
	OffCharging   time.Duration
	OffAvailable  time.Duration

	HourCharging  [24]time.Duration
	HourAvailable [24]time.Duration

	Sessions     int
	Completed    int
	Failed       int
	Interrupted  int
	InProgress   int
	Energy       float64
	Revenue      float64
	Cost         float64
	DurationSec  int64
	ChargeSec    int64
	PaymentFails int
	LongSessions int

	HourEnergy        [24]float64
	HourRevenue       [24]float64
	HourSessionStarts [24]int
	DOWSessions       [7]int
	DOWEnergy         [7]float64

	Faults           int
	OpenFaults       int
	ConnectorFaults  int
	RepairTime       time.Duration
	RepairCount      int
	Failures         int
	FaultsByCategory map[string]int
	PeakObservedKW   float64
	RevenueAtRisk    float64
	UsedNetworkRate  bool
}

func (t Totals) Utilization() (float64, bool) {
	return UtilizationPct(t.Charging, t.Available)
}

func (t Totals) Uptime() (float64, bool) {
	return UptimePct(t.Up, t.Scheduled)
}

func (t Totals) Margin() (float64, float64, bool) {
	return GrossMargin(t.Revenue, t.Cost)
}

func (t Totals) AvgDurationMin() (float64, bool) {
	if t.Sessions <= 0 {
		return 0, false
	}
	return float64(t.DurationSec) / float64(t.Sessions) / 60, true
}

func (t Totals) AvgEnergy() (float64, bool) {
	if t.Sessions <= 0 {
		return 0, false
	}
	return t.Energy / float64(t.Sessions), true
}

func (t Totals) AvgPower() (float64, bool) {
	if t.ChargeSec <= 0 || t.Energy <= 0 {
		return 0, false
	}
	return t.Energy / (float64(t.ChargeSec) / 3600), true
}

func (t Totals) RevenuePerKWh() (float64, bool) { return Ratio(t.Revenue, t.Energy) }
func (t Totals) CostPerKWh() (float64, bool)    { return Ratio(t.Cost, t.Energy) }

type DayStat struct {
	Date        string
	Charging    time.Duration
	Available   time.Duration
	Maintenance time.Duration
	Downtime    time.Duration
	Scheduled   time.Duration
	Up          time.Duration
	Sessions    int
	Energy      float64
	Revenue     float64
	Cost        float64
}

func (d DayStat) Utilization() (float64, bool) {
	return UtilizationPct(d.Charging, d.Available)
}

func (d DayStat) Uptime() (float64, bool) {
	return UptimePct(d.Up, d.Scheduled)
}

type ConnectorMetric struct {
	Connector   domain.Connector
	StationName string
	Totals
}

type StationMetric struct {
	Station domain.Station
	Totals
	Connectors   []ConnectorMetric
	Daily        []DayStat
	Faults       []domain.Fault
	ChargerTypes []string
	HasDC        bool
	MaxRatedKW   float64
	SumRatedKW   float64
	WorstStatus  string
	StatusCounts map[string]int
}

type Snapshot struct {
	Query          domain.Query
	AsOf           time.Time
	Location       *time.Location
	Disclaimer     string
	OperatorName   string
	NetworkName    string
	Timezone       string
	Stations       []StationMetric
	Connectors     []ConnectorMetric
	Daily          []DayStat
	Sessions       []domain.Session
	Faults         []domain.Fault
	Network        Totals
	StatusCounts   map[string]int
	WeekdayDates   int
	WeekendDates   int
	LongestOpen    *domain.Fault
	StationCount   int
	ConnectorCount int
}

type connSeries struct {
	c         domain.Connector
	charging  []Interval
	maint     []Interval
	unplanned []Interval
	blocked   []Interval
}

func (s *Service) Compute(ctx context.Context, q domain.Query) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !q.To.After(q.From) {
		return nil, ErrInvalidRange
	}
	if q.To.Sub(q.From) > 366*24*time.Hour {
		return nil, ErrRangeTooLong
	}
	if q.ChargerType != "" && q.ChargerType != string(domain.ChargerAC) && q.ChargerType != string(domain.ChargerDC) {
		return nil, ErrInvalidFilter
	}
	ds := s.src.Dataset()
	loc, err := time.LoadLocation(ds.Timezone)
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	stationByID := map[string]domain.Station{}
	for _, st := range s.src.Stations() {
		stationByID[st.ID] = st
	}
	if q.StationID != "" {
		if _, ok := stationByID[q.StationID]; !ok {
			return nil, ErrUnknownStation
		}
	}

	var conns []domain.Connector
	for _, c := range s.src.Connectors() {
		if q.StationID != "" && c.StationID != q.StationID {
			continue
		}
		if q.ChargerType != "" && string(c.ChargerType) != q.ChargerType {
			continue
		}
		conns = append(conns, c)
	}
	connSet := map[string]domain.Connector{}
	for _, c := range conns {
		connSet[c.ID] = c
	}

	rawFault := map[string][]Interval{}
	rawMaint := map[string][]Interval{}
	var overlapping []domain.Fault
	seenFault := map[string]bool{}

	applyConn := func(stationID, connectorID string, iv Interval, dst map[string][]Interval) {
		if connectorID == "" {
			for _, c := range conns {
				if c.StationID == stationID {
					dst[c.ID] = append(dst[c.ID], iv)
				}
			}
			return
		}
		if _, ok := connSet[connectorID]; ok {
			dst[connectorID] = append(dst[connectorID], iv)
		}
	}

	for _, f := range s.src.Faults() {
		if q.StationID != "" && f.StationID != q.StationID {
			continue
		}
		end := f.FaultEnd(ds.AsOf)
		iv := Interval{Start: f.Start, End: end}
		if !iv.Valid() {
			continue
		}
		if f.ConnectorID != "" {
			if _, ok := connSet[f.ConnectorID]; !ok {
				continue
			}
		} else if !stationHasConnector(conns, f.StationID) {
			continue
		}
		applyConn(f.StationID, f.ConnectorID, iv, rawFault)
		if overlaps(f.Start, end, q.From, q.To) && !seenFault[f.ID] {
			seenFault[f.ID] = true
			overlapping = append(overlapping, f)
		}
	}
	for _, m := range s.src.Maintenance() {
		if q.StationID != "" && m.StationID != q.StationID {
			continue
		}
		iv := Interval{Start: m.Start, End: m.End}
		if !iv.Valid() {
			continue
		}
		if m.ConnectorID != "" {
			if _, ok := connSet[m.ConnectorID]; !ok {
				continue
			}
		} else if !stationHasConnector(conns, m.StationID) {
			continue
		}
		applyConn(m.StationID, m.ConnectorID, iv, rawMaint)
	}

	rawCharge := map[string][]Interval{}
	for _, ses := range s.src.Sessions() {
		if _, ok := connSet[ses.ConnectorID]; !ok {
			continue
		}
		iv, ok := chargeInterval(ses)
		if !ok {
			continue
		}
		rawCharge[ses.ConnectorID] = append(rawCharge[ses.ConnectorID], iv)
	}

	series := make([]connSeries, 0, len(conns))
	for _, c := range conns {
		faults := Merge(rawFault[c.ID])
		maint := Merge(rawMaint[c.ID])
		blocked := Merge(append(append([]Interval{}, faults...), maint...))
		series = append(series, connSeries{
			c:         c,
			charging:  Subtract(Merge(rawCharge[c.ID]), blocked),
			maint:     maint,
			unplanned: Subtract(faults, maint),
			blocked:   blocked,
		})
	}

	// Preserve station order from the dataset.
	var stationOrder []string
	seenStation := map[string]bool{}
	for _, c := range conns {
		if !seenStation[c.StationID] {
			seenStation[c.StationID] = true
			stationOrder = append(stationOrder, c.StationID)
		}
	}

	snap := &Snapshot{
		Query: q, AsOf: ds.AsOf, Location: loc,
		Disclaimer: ds.Disclaimer, OperatorName: ds.OperatorName,
		NetworkName: ds.NetworkName, Timezone: ds.Timezone,
		StatusCounts: map[string]int{
			string(domain.StatusAvailable): 0, string(domain.StatusOccupied): 0,
			string(domain.StatusUnavailable): 0, string(domain.StatusFaulted): 0,
		},
		Network:        Totals{FaultsByCategory: map[string]int{}},
		ConnectorCount: len(conns),
		StationCount:   len(stationOrder),
	}
	snap.WeekdayDates, snap.WeekendDates = countDayTypes(q.From, q.To, loc)

	stIndex := map[string]int{}
	for _, id := range stationOrder {
		st := stationByID[id]
		sm := StationMetric{
			Station: st,
			Totals:  Totals{FaultsByCategory: map[string]int{}},
			StatusCounts: map[string]int{
				string(domain.StatusAvailable): 0, string(domain.StatusOccupied): 0,
				string(domain.StatusUnavailable): 0, string(domain.StatusFaulted): 0,
			},
			WorstStatus: string(domain.StatusAvailable),
		}
		snap.Stations = append(snap.Stations, sm)
		stIndex[id] = len(snap.Stations) - 1
	}

	dayIndex := map[string]int{}
	for _, key := range eachDay(q.From, q.To, loc) {
		dayIndex[key] = len(snap.Daily)
		snap.Daily = append(snap.Daily, DayStat{Date: key})
	}
	stDay := map[string]map[string]*DayStat{}
	for _, id := range stationOrder {
		stDay[id] = map[string]*DayStat{}
		for _, key := range eachDay(q.From, q.To, loc) {
			stDay[id][key] = &DayStat{Date: key}
		}
	}

	connIndex := map[string]int{}
	for _, ser := range series {
		name := ""
		if st, ok := stationByID[ser.c.StationID]; ok {
			name = st.Name
		}
		cm := ConnectorMetric{
			Connector: ser.c, StationName: name,
			Totals: Totals{FaultsByCategory: map[string]int{}},
		}
		snap.Connectors = append(snap.Connectors, cm)
		connIndex[ser.c.ID] = len(snap.Connectors) - 1
		sm := &snap.Stations[stIndex[ser.c.StationID]]
		sm.Connectors = append(sm.Connectors, cm)
		sm.SumRatedKW += ser.c.RatedPowerKW
		if ser.c.RatedPowerKW > sm.MaxRatedKW {
			sm.MaxRatedKW = ser.c.RatedPowerKW
		}
		if ser.c.ChargerType == domain.ChargerDC {
			sm.HasDC = true
		}
		sm.ChargerTypes = appendType(sm.ChargerTypes, string(ser.c.ChargerType))
		status := string(ser.c.Status)
		sm.StatusCounts[status]++
		snap.StatusCounts[status]++
		sm.WorstStatus = worseStatus(sm.WorstStatus, status)

		walkSlots(q.From, q.To, loc, func(slot Interval, day string, hour int) {
			maint := OverlapSum(slot, ser.maint)
			unplanned := OverlapSum(slot, ser.unplanned)
			blocked := OverlapSum(slot, ser.blocked)
			charge := OverlapSum(slot, ser.charging)
			slotDur := slot.Duration()
			avail := slotDur - blocked
			if avail < 0 {
				avail = 0
			}
			if charge > avail {
				charge = avail
			}
			scheduled := slotDur - maint
			if scheduled < 0 {
				scheduled = 0
			}
			up := scheduled - unplanned
			if up < 0 {
				up = 0
			}
			peak := isPeakHourIndex(hour)
			addTime(&snap.Network, charge, avail, maint, unplanned, scheduled, up, hour, peak)
			addTime(&sm.Totals, charge, avail, maint, unplanned, scheduled, up, hour, peak)
			addTime(&snap.Connectors[connIndex[ser.c.ID]].Totals, charge, avail, maint, unplanned, scheduled, up, hour, peak)
			if di, ok := dayIndex[day]; ok {
				addDay(&snap.Daily[di], charge, avail, maint, unplanned, scheduled, up)
			}
			if bucket := stDay[ser.c.StationID][day]; bucket != nil {
				addDay(bucket, charge, avail, maint, unplanned, scheduled, up)
			}
		})
	}

	for _, ses := range s.src.Sessions() {
		if _, ok := connSet[ses.ConnectorID]; !ok {
			continue
		}
		if ses.Start.Before(q.From) || !ses.Start.Before(q.To) {
			continue
		}
		snap.Sessions = append(snap.Sessions, ses)
		sm := &snap.Stations[stIndex[ses.StationID]]
		addSession(&snap.Network, ses, loc)
		addSession(&sm.Totals, ses, loc)
		if ci, ok := connIndex[ses.ConnectorID]; ok {
			addSession(&snap.Connectors[ci].Totals, ses, loc)
		}
		key := ses.Start.In(loc).Format("2006-01-02")
		if di, ok := dayIndex[key]; ok {
			addSessionDay(&snap.Daily[di], ses)
		}
		if bucket := stDay[ses.StationID][key]; bucket != nil {
			addSessionDay(bucket, ses)
		}
		prorateEnergy(ses, q.From, q.To, loc, &snap.Network, sm)
	}
	for i := range snap.Stations {
		snap.Stations[i].Connectors = nil
	}
	for _, cm := range snap.Connectors {
		sm := &snap.Stations[stIndex[cm.Connector.StationID]]
		sm.Connectors = append(sm.Connectors, cm)
	}

	for _, f := range overlapping {
		sm := &snap.Stations[stIndex[f.StationID]]
		addFault(&snap.Network, f, ds.AsOf, q)
		addFault(&sm.Totals, f, ds.AsOf, q)
		sm.Faults = append(sm.Faults, f)
		if snap.LongestOpen == nil && f.Ongoing(ds.AsOf) {
			cp := f
			snap.LongestOpen = &cp
		} else if f.Ongoing(ds.AsOf) && snap.LongestOpen != nil {
			age := ds.AsOf.Sub(f.Start)
			if age > ds.AsOf.Sub(snap.LongestOpen.Start) {
				cp := f
				snap.LongestOpen = &cp
			}
		}
	}
	snap.Faults = overlapping

	netRate := 0.0
	if snap.Network.Available > 0 {
		netRate = snap.Network.Revenue / snap.Network.Available.Hours()
	}
	for i := range snap.Stations {
		sm := &snap.Stations[i]
		rate := netRate
		if sm.Available > 0 {
			rate = sm.Revenue / sm.Available.Hours()
		} else if sm.Downtime > 0 {
			sm.UsedNetworkRate = true
		}
		sm.RevenueAtRisk = sm.Downtime.Hours() * rate
		snap.Network.RevenueAtRisk += sm.RevenueAtRisk
		days := orderedDays(stDay[sm.Station.ID], snap.Daily)
		sm.Daily = days
	}
	return snap, nil
}

func stationHasConnector(conns []domain.Connector, stationID string) bool {
	for _, c := range conns {
		if c.StationID == stationID {
			return true
		}
	}
	return false
}

func overlaps(start, end, from, to time.Time) bool {
	return start.Before(to) && from.Before(end)
}

func chargeInterval(s domain.Session) (Interval, bool) {
	if s.Status == domain.SessionFailed {
		return Interval{}, false
	}
	sec := s.ChargeDurationSec
	if sec <= 0 {
		sec = s.DurationSec
	}
	if sec <= 0 {
		return Interval{}, false
	}
	end := s.Start.Add(time.Duration(sec) * time.Second)
	if s.End.After(s.Start) && end.After(s.End) {
		end = s.End
	}
	iv := Interval{Start: s.Start, End: end}
	if !iv.Valid() {
		return Interval{}, false
	}
	return iv, true
}

func addTime(t *Totals, charge, avail, maint, down, scheduled, up time.Duration, hour int, peak bool) {
	t.Charging += charge
	t.Available += avail
	t.Maintenance += maint
	t.Downtime += down
	t.Scheduled += scheduled
	t.Up += up
	if hour >= 0 && hour < 24 {
		t.HourCharging[hour] += charge
		t.HourAvailable[hour] += avail
	}
	if peak {
		t.PeakCharging += charge
		t.PeakAvailable += avail
	} else {
		t.OffCharging += charge
		t.OffAvailable += avail
	}
}

func addDay(d *DayStat, charge, avail, maint, down, scheduled, up time.Duration) {
	d.Charging += charge
	d.Available += avail
	d.Maintenance += maint
	d.Downtime += down
	d.Scheduled += scheduled
	d.Up += up
}

func addSession(t *Totals, ses domain.Session, loc *time.Location) {
	t.Sessions++
	switch ses.Status {
	case domain.SessionCompleted:
		t.Completed++
	case domain.SessionFailed:
		t.Failed++
	case domain.SessionInterrupted:
		t.Interrupted++
	case domain.SessionInProgress:
		t.InProgress++
	}
	t.Energy += ses.EnergyKWh
	t.Revenue += ses.RevenueINR
	t.Cost += ses.EnergyCostINR
	t.DurationSec += ses.DurationSec
	charge := ses.ChargeDurationSec
	if charge <= 0 && ses.Status != domain.SessionFailed {
		charge = ses.DurationSec
	}
	t.ChargeSec += charge
	if ses.Status == domain.SessionFailed && ses.TerminationReason == "payment_failure" {
		t.PaymentFails++
	}
	if longSession(ses) {
		t.LongSessions++
	}
	h := ses.Start.In(loc).Hour()
	t.HourSessionStarts[h]++
	wd := int(ses.Start.In(loc).Weekday())
	t.DOWSessions[wd]++
	t.DOWEnergy[wd] += ses.EnergyKWh
	sec := ses.ChargeDurationSec
	if sec <= 0 {
		sec = ses.DurationSec
	}
	if _, peak, ok := ObservedPower(ses.EnergyKWh, sec, ses.PeakPowerKW); ok && peak > t.PeakObservedKW {
		t.PeakObservedKW = peak
	}
}

func addSessionDay(d *DayStat, ses domain.Session) {
	d.Sessions++
	d.Energy += ses.EnergyKWh
	d.Revenue += ses.RevenueINR
	d.Cost += ses.EnergyCostINR
}

func prorateEnergy(ses domain.Session, from, to time.Time, loc *time.Location, net *Totals, sm *StationMetric) {
	iv, ok := chargeInterval(ses)
	if !ok || ses.EnergyKWh <= 0 {
		return
	}
	total := iv.Duration().Seconds()
	if total <= 0 {
		return
	}
	for cur := iv.Start; cur.Before(iv.End); {
		next := nextHour(cur, loc)
		if !next.After(cur) {
			break
		}
		if next.After(iv.End) {
			next = iv.End
		}
		segStart, segEnd := cur, next
		if from.After(segStart) {
			segStart = from
		}
		if to.Before(segEnd) {
			segEnd = to
		}
		if segEnd.After(segStart) {
			frac := segEnd.Sub(segStart).Seconds() / total
			hour := segStart.In(loc).Hour()
			net.HourEnergy[hour] += ses.EnergyKWh * frac
			net.HourRevenue[hour] += ses.RevenueINR * frac
			sm.HourEnergy[hour] += ses.EnergyKWh * frac
			sm.HourRevenue[hour] += ses.RevenueINR * frac
		}
		cur = next
	}
}

func addFault(t *Totals, f domain.Fault, asOf time.Time, q domain.Query) {
	t.Faults++
	t.Failures++
	if t.FaultsByCategory == nil {
		t.FaultsByCategory = map[string]int{}
	}
	t.FaultsByCategory[string(f.Category)]++
	if f.Category == domain.FaultConnector {
		t.ConnectorFaults++
	}
	if f.Ongoing(asOf) {
		t.OpenFaults++
		return
	}
	if f.End == nil {
		return
	}
	d := f.End.Sub(f.Start)
	if d > 0 {
		t.RepairTime += d
		t.RepairCount++
	}
}

func longSession(ses domain.Session) bool {
	if ses.Status == domain.SessionFailed {
		return false
	}
	plugMin := float64(ses.DurationSec) / 60
	if ses.ChargerType == domain.ChargerDC {
		return plugMin > 90
	}
	return plugMin > 300
}

func appendType(types []string, t string) []string {
	for _, existing := range types {
		if existing == t {
			return types
		}
	}
	return append(types, t)
}

func worseStatus(current, next string) string {
	rank := map[string]int{
		string(domain.StatusAvailable): 0, string(domain.StatusOccupied): 1,
		string(domain.StatusUnavailable): 2, string(domain.StatusFaulted): 3,
	}
	if rank[next] > rank[current] {
		return next
	}
	return current
}

func isPeakHourIndex(h int) bool {
	return (h >= 8 && h < 11) || (h >= 18 && h < 21)
}

func walkSlots(from, to time.Time, loc *time.Location, fn func(slot Interval, day string, hour int)) {
	t := from
	for t.Before(to) {
		end := nextHour(t, loc)
		if !end.After(t) {
			break
		}
		if end.After(to) {
			end = to
		}
		lt := t.In(loc)
		fn(Interval{Start: t, End: end}, lt.Format("2006-01-02"), lt.Hour())
		t = end
	}
}

func nextHour(t time.Time, loc *time.Location) time.Time {
	lt := t.In(loc)
	n := time.Date(lt.Year(), lt.Month(), lt.Day(), lt.Hour(), 0, 0, 0, loc).Add(time.Hour)
	if !n.After(t) {
		return n.Add(time.Hour)
	}
	return n
}

func eachDay(from, to time.Time, loc *time.Location) []string {
	start := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 0, 0, 0, 0, loc)
	var keys []string
	last := to.Add(-time.Nanosecond)
	for !start.After(last) {
		keys = append(keys, start.Format("2006-01-02"))
		start = start.AddDate(0, 0, 1)
		if len(keys) > 400 {
			break
		}
	}
	return keys
}

func countDayTypes(from, to time.Time, loc *time.Location) (weekday, weekend int) {
	for _, key := range eachDay(from, to, loc) {
		d, err := time.ParseInLocation("2006-01-02", key, loc)
		if err != nil {
			continue
		}
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			weekend++
		} else {
			weekday++
		}
	}
	return weekday, weekend
}

func orderedDays(m map[string]*DayStat, order []DayStat) []DayStat {
	out := make([]DayStat, 0, len(order))
	for _, d := range order {
		if row := m[d.Date]; row != nil {
			out = append(out, *row)
		}
	}
	return out
}
