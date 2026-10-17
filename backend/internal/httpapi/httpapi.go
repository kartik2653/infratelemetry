package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/domain"
	"chargingdatainfra/backend/internal/insights"
	"chargingdatainfra/backend/internal/store"
)

type API struct {
	log   *slog.Logger
	store *store.Store
	svc   *analytics.Service
}

func New(log *slog.Logger, st *store.Store, svc *analytics.Service) http.Handler {
	a := &API{log: log, store: st, svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", a.health)
	mux.HandleFunc("GET /api/v1/dashboard/summary", a.dashboardSummary)
	mux.HandleFunc("GET /api/v1/dashboard/trends", a.dashboardTrends)
	mux.HandleFunc("GET /api/v1/stations", a.stations)
	mux.HandleFunc("GET /api/v1/stations/{id}", a.station)
	mux.HandleFunc("GET /api/v1/stations/{id}/sessions", a.stationSessions)
	mux.HandleFunc("GET /api/v1/stations/{id}/faults", a.stationFaults)
	mux.HandleFunc("GET /api/v1/analytics/utilization", a.utilization)
	mux.HandleFunc("GET /api/v1/analytics/revenue", a.revenue)
	mux.HandleFunc("GET /api/v1/analytics/energy", a.energy)
	mux.HandleFunc("GET /api/v1/analytics/reliability", a.reliability)
	mux.HandleFunc("GET /api/v1/analytics/sessions", a.sessionAnalytics)
	mux.HandleFunc("GET /api/v1/sessions", a.sessions)
	mux.HandleFunc("GET /api/v1/alerts", a.listAlerts)
	mux.HandleFunc("PATCH /api/v1/alerts/{id}", a.patchAlert)
	mux.HandleFunc("GET /api/v1/insights", a.listInsights)
	return a.wrap(mux)
}

func (a *API) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		rw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		a.log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rw.status, "ms", time.Since(start).Milliseconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

type meta struct {
	From       *time.Time        `json:"from,omitempty"`
	To         *time.Time        `json:"to,omitempty"`
	AsOf       time.Time         `json:"as_of"`
	Timezone   string            `json:"timezone"`
	Seed       int64             `json:"seed"`
	Disclaimer string            `json:"disclaimer"`
	Page       int               `json:"page,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Total      int               `json:"total,omitempty"`
	TotalPages int               `json:"total_pages,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
}

type envelope struct {
	Demo bool `json:"demo"`
	Data any  `json:"data"`
	Meta meta `json:"meta"`
}

type errorEnvelope struct {
	Demo  bool `json:"demo"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *API) baseMeta() meta {
	ds := a.store.Dataset()
	return meta{AsOf: ds.AsOf, Timezone: ds.Timezone, Seed: ds.Seed, Disclaimer: ds.Disclaimer}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func writeErr(w http.ResponseWriter, status int, code, message string) {
	var body errorEnvelope
	body.Demo = true
	body.Error.Code = code
	body.Error.Message = message
	writeJSON(w, status, body)
}

func (a *API) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, analytics.ErrInvalidRange), errors.Is(err, analytics.ErrRangeTooLong):
		writeErr(w, http.StatusBadRequest, "INVALID_TIME_RANGE", err.Error())
	case errors.Is(err, analytics.ErrInvalidFilter), errors.Is(err, store.ErrInvalidStatus):
		writeErr(w, http.StatusBadRequest, "INVALID_QUERY", err.Error())
	case errors.Is(err, analytics.ErrUnknownStation), errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		a.log.Error("request failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "The request could not be completed")
	}
}

func (a *API) query(r *http.Request) (domain.Query, error) {
	ds := a.store.Dataset()
	from, to, err := parseRange(r, ds)
	if err != nil {
		return domain.Query{}, err
	}
	q := domain.Query{From: from, To: to, StationID: strings.TrimSpace(r.URL.Query().Get("station_id"))}
	ct := strings.TrimSpace(r.URL.Query().Get("charger_type"))
	if ct != "" && ct != "AC" && ct != "DC" {
		return domain.Query{}, analytics.ErrInvalidFilter
	}
	q.ChargerType = ct
	return q, nil
}

func parseRange(r *http.Request, ds domain.Dataset) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(ds.Timezone)
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	rawFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	rawTo := strings.TrimSpace(r.URL.Query().Get("to"))
	if rawFrom == "" && rawTo == "" {
		return ds.DataStart, ds.AsOf, nil
	}
	if rawFrom == "" || rawTo == "" {
		return time.Time{}, time.Time{}, analytics.ErrInvalidRange
	}
	from, err := parseInstant(rawFrom, loc, false)
	if err != nil {
		return time.Time{}, time.Time{}, analytics.ErrInvalidRange
	}
	to, err := parseInstant(rawTo, loc, true)
	if err != nil {
		return time.Time{}, time.Time{}, analytics.ErrInvalidRange
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, analytics.ErrInvalidRange
	}
	if to.Sub(from) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, analytics.ErrRangeTooLong
	}
	return from, to, nil
}

func parseInstant(v string, loc *time.Location, exclusiveEnd bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	d, err := time.ParseInLocation("2006-01-02", v, loc)
	if err != nil {
		return time.Time{}, err
	}
	if exclusiveEnd {
		return d.AddDate(0, 0, 1), nil
	}
	return d, nil
}

func pageParams(r *http.Request) (page, limit int, err error) {
	page, limit = 1, 25
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil || page < 1 {
			return 0, 0, analytics.ErrInvalidFilter
		}
	}
	if v := q.Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, analytics.ErrInvalidFilter
		}
	}
	return page, limit, nil
}

func paginate[T any](items []T, page, limit int) ([]T, int, int) {
	total := len(items)
	pages := 0
	if total > 0 {
		pages = (total + limit - 1) / limit
	}
	start := (page - 1) * limit
	if start >= total {
		return []T{}, total, pages
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total, pages
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	ds := a.store.Dataset()
	type idx struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Area string `json:"area"`
	}
	stations := make([]idx, 0, len(ds.Stations))
	for _, st := range ds.Stations {
		stations = append(stations, idx{ID: st.ID, Name: st.Name, Area: st.Area})
	}
	counts := map[string]int{}
	for _, c := range ds.Connectors {
		counts[string(c.Status)]++
	}
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: a.baseMeta(), Data: map[string]any{
		"status": "ok", "operator_name": ds.OperatorName, "network_name": ds.NetworkName,
		"as_of": ds.AsOf, "data_start": ds.DataStart, "timezone": ds.Timezone,
		"seed": ds.Seed, "days": ds.Days,
		"station_count": len(ds.Stations), "connector_count": len(ds.Connectors),
		"session_count": len(ds.Sessions), "fault_count": len(ds.Faults),
		"connectors_by_status": counts, "disclaimer": ds.Disclaimer, "stations": stations,
	}})
}

func (a *API) dashboardSummary(w http.ResponseWriter, r *http.Request) {
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	summary, err := a.svc.Summary(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	snap, err := a.svc.Compute(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	all := insights.FromSnapshot(snap)
	top := all
	if len(top) > 4 {
		top = top[:4]
	}
	var critical []domain.Alert
	for _, al := range a.store.Alerts() {
		if al.Status == string(domain.AlertResolved) {
			continue
		}
		if al.Severity == string(domain.SeverityCritical) {
			critical = append(critical, al)
		}
	}
	if len(critical) > 5 {
		critical = critical[:5]
	}
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: map[string]any{
		"summary": summary, "critical_alerts": critical, "insights": top,
	}})
}

func (a *API) dashboardTrends(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.Trends(ctx, q)
	})
}

func (a *API) utilization(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.Utilization(ctx, q)
	})
}

func (a *API) revenue(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.Revenue(ctx, q)
	})
}

func (a *API) energy(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.Energy(ctx, q)
	})
}

func (a *API) reliability(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.Reliability(ctx, q)
	})
}

func (a *API) sessionAnalytics(w http.ResponseWriter, r *http.Request) {
	a.withQuery(w, r, func(ctx context.Context, q domain.Query) (any, error) {
		return a.svc.SessionsAnalytics(ctx, q)
	})
}

func (a *API) listInsights(w http.ResponseWriter, r *http.Request) {
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	snap, err := a.svc.Compute(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: map[string]any{"insights": insights.FromSnapshot(snap)}})
}

func (a *API) withQuery(w http.ResponseWriter, r *http.Request, fn func(context.Context, domain.Query) (any, error)) {
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	data, err := fn(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	m.Filters = filtersOf(q)
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: data})
}

func filtersOf(q domain.Query) map[string]string {
	f := map[string]string{}
	if q.StationID != "" {
		f["station_id"] = q.StationID
	}
	if q.ChargerType != "" {
		f["charger_type"] = q.ChargerType
	}
	if len(f) == 0 {
		return nil
	}
	return f
}

func (a *API) stations(w http.ResponseWriter, r *http.Request) {
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	page, limit, err := pageParams(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	rows, err := a.svc.StationRows(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	sortField := r.URL.Query().Get("sort")
	switch sortField {
	case "", "name", "utilization", "uptime", "energy", "revenue", "margin", "sessions", "downtime":
	default:
		a.fail(w, analytics.ErrInvalidFilter)
		return
	}
	order := r.URL.Query().Get("order")
	switch order {
	case "", "asc", "desc":
	default:
		a.fail(w, analytics.ErrInvalidFilter)
		return
	}
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if search != "" {
		filtered := rows[:0:0]
		for _, row := range rows {
			blob := strings.ToLower(row.ID + " " + row.Name + " " + row.Area + " " + row.Address)
			if strings.Contains(blob, search) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	sortRows(rows, sortField, order)
	pageRows, total, pages := paginate(rows, page, limit)
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	m.Page, m.Limit, m.Total, m.TotalPages = page, limit, total, pages
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: map[string]any{"stations": pageRows}})
}

func sortRows(rows []analytics.StationRow, field, order string) {
	if field == "" {
		field = "name"
	}
	desc := strings.EqualFold(order, "desc")
	if order == "" && field != "name" {
		desc = true
	}
	less := func(i, j int) bool { return rows[i].ID < rows[j].ID }
	switch field {
	case "name":
		less = func(i, j int) bool { return rows[i].Name < rows[j].Name }
	case "utilization":
		less = func(i, j int) bool { return cmpPtr(rows[i].UtilizationPct, rows[j].UtilizationPct) < 0 }
	case "uptime":
		less = func(i, j int) bool { return cmpPtr(rows[i].UptimePct, rows[j].UptimePct) < 0 }
	case "energy":
		less = func(i, j int) bool { return rows[i].EnergyKWh < rows[j].EnergyKWh }
	case "revenue":
		less = func(i, j int) bool { return rows[i].RevenueINR < rows[j].RevenueINR }
	case "margin":
		less = func(i, j int) bool { return cmpPtr(rows[i].GrossMarginPct, rows[j].GrossMarginPct) < 0 }
	case "sessions":
		less = func(i, j int) bool { return rows[i].SessionCount < rows[j].SessionCount }
	case "downtime":
		less = func(i, j int) bool { return rows[i].DowntimeHours < rows[j].DowntimeHours }
	default:
		return
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if desc {
			return less(j, i)
		}
		return less(i, j)
	})
}

func cmpPtr(a, b *float64) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}

func (a *API) station(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := a.store.Station(id); !ok {
		a.fail(w, analytics.ErrUnknownStation)
		return
	}
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	q.StationID = id
	detail, err := a.svc.Station(r.Context(), id, q)
	if err != nil {
		a.fail(w, err)
		return
	}
	snap, err := a.svc.Compute(r.Context(), q)
	if err != nil {
		a.fail(w, err)
		return
	}
	mine := []insights.Insight{}
	for _, ins := range insights.FromSnapshot(snap) {
		for _, sid := range ins.StationIDs {
			if sid == id {
				mine = append(mine, ins)
				break
			}
		}
	}
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: struct {
		analytics.StationDetail
		Insights []insights.Insight `json:"insights"`
	}{StationDetail: detail, Insights: mine}})
}

func (a *API) stationSessions(w http.ResponseWriter, r *http.Request) {
	a.sessionList(w, r, r.PathValue("id"))
}

func (a *API) sessions(w http.ResponseWriter, r *http.Request) {
	a.sessionList(w, r, "")
}

func (a *API) sessionList(w http.ResponseWriter, r *http.Request, pathStation string) {
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	if pathStation != "" {
		if _, ok := a.store.Station(pathStation); !ok {
			a.fail(w, analytics.ErrUnknownStation)
			return
		}
		q.StationID = pathStation
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	switch status {
	case "", "completed", "failed", "interrupted", "in_progress":
	default:
		a.fail(w, analytics.ErrInvalidFilter)
		return
	}
	page, limit, err := pageParams(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	rows, err := a.svc.ListSessions(r.Context(), q, status)
	if err != nil {
		a.fail(w, err)
		return
	}
	pageRows, total, pages := paginate(rows, page, limit)
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	m.Page, m.Limit, m.Total, m.TotalPages = page, limit, total, pages
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: map[string]any{"sessions": pageRows}})
}

func (a *API) stationFaults(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := a.store.Station(id); !ok {
		a.fail(w, analytics.ErrUnknownStation)
		return
	}
	q, err := a.query(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	q.StationID = id
	page, limit, err := pageParams(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	detail, err := a.svc.Station(r.Context(), id, q)
	if err != nil {
		a.fail(w, err)
		return
	}
	pageRows, total, pages := paginate(detail.Faults, page, limit)
	m := a.baseMeta()
	m.From, m.To = &q.From, &q.To
	m.Page, m.Limit, m.Total, m.TotalPages = page, limit, total, pages
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: m, Data: map[string]any{"faults": pageRows}})
}

func (a *API) listAlerts(w http.ResponseWriter, r *http.Request) {
	severity := strings.TrimSpace(r.URL.Query().Get("severity"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	stationID := strings.TrimSpace(r.URL.Query().Get("station_id"))
	switch severity {
	case "", "critical", "warning", "info":
	default:
		a.fail(w, analytics.ErrInvalidFilter)
		return
	}
	switch status {
	case "", "open", "acknowledged", "resolved":
	default:
		a.fail(w, analytics.ErrInvalidFilter)
		return
	}
	if stationID != "" {
		if _, ok := a.store.Station(stationID); !ok {
			a.fail(w, analytics.ErrUnknownStation)
			return
		}
	}
	var out []domain.Alert
	for _, al := range a.store.Alerts() {
		if severity != "" && al.Severity != severity {
			continue
		}
		if status != "" && al.Status != status {
			continue
		}
		if category != "" && al.Category != category {
			continue
		}
		if stationID != "" && al.StationID != stationID {
			continue
		}
		out = append(out, al)
	}
	if out == nil {
		out = []domain.Alert{}
	}
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: a.baseMeta(), Data: map[string]any{"alerts": out}})
}

type patchBody struct {
	Status string `json:"status"`
}

func (a *API) patchAlert(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "Could not read the request body")
		return
	}
	var req patchBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "Body must be JSON with a status field")
		return
	}
	updated, err := a.store.UpdateAlert(r.PathValue("id"), strings.TrimSpace(req.Status))
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Demo: true, Meta: a.baseMeta(), Data: updated})
}
