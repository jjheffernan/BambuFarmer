package jobs

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts job list/create/get/reprint/transition/dispatch routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, bus *events.Bus, siteID string) {
	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			f, limit := parseJobListFilter(r)
			items, err := db.ListJobs(r.Context(), pool, siteID, limit, f)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.Job{}
			}
			if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
				writeJobsCSV(w, items)
				return
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("GET /api/v1/jobs/history", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			f, limit := parseJobListFilter(r)
			if f.State == "" {
				// Print-history default: terminal + recent active states.
				// Empty still means all; clients pass state when narrowing.
			}
			if limit < 100 {
				limit = 200
			}
			items, err := db.ListJobs(r.Context(), pool, siteID, limit, f)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.Job{}
			}
			if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
				writeJobsCSV(w, items)
				return
			}
			httpx.WriteJSON(w, map[string]any{"jobs": items, "count": len(items)})
		})
	})
	mux.HandleFunc("POST /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var req api.CreateJobRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			respondCreated(w, r, pool, bus, siteID, req)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/queue", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				Direction string `json:"direction"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			job, err := db.ReorderQueuedJob(r.Context(), pool, siteID, r.PathValue("id"), body.Direction)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/reprint", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			src, err := db.GetJob(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			req, err := ReprintRequest(src)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			respondCreated(w, r, pool, bus, siteID, req)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				Reason string `json:"reason"`
			}
			if r.ContentLength != 0 {
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					httpx.WriteError(w, api.BadRequest("invalid JSON body"))
					return
				}
			}
			job, err := CancelOne(r.Context(), pool, bus, siteID, r.PathValue("id"), body.Reason)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			job, err := db.GetJob(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListJobEvents(r.Context(), pool, siteID, r.PathValue("id"), 100)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/dispatch", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				ExternalSN string `json:"external_sn"`
			}
			if r.ContentLength != 0 {
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					httpx.WriteError(w, api.BadRequest("invalid JSON body"))
					return
				}
			}
			job, err := DispatchOne(r.Context(), pool, bus, siteID, r.PathValue("id"), body.ExternalSN)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/reassign", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				ExternalSN string `json:"external_sn"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			job, err := ReassignOne(r.Context(), pool, siteID, r.PathValue("id"), body.ExternalSN, "api")
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/transition", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			var body struct {
				State  string `json:"state"`
				Reason string `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				httpx.WriteError(w, api.BadRequest("invalid JSON body"))
				return
			}
			if body.State == "" {
				httpx.WriteError(w, api.BadRequest("state required"))
				return
			}
			job, err := db.TransitionJob(r.Context(), pool, siteID, r.PathValue("id"), body.State, "api", body.Reason)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
}

func parseJobListFilter(r *http.Request) (db.JobListFilter, int) {
	q := r.URL.Query()
	f := db.JobListFilter{
		Query:      q.Get("q"),
		State:      q.Get("state"),
		ExternalSN: firstNonEmpty(q.Get("external_sn"), q.Get("sn")),
	}
	if from, err := parseDayOrTime(q.Get("from")); err == nil && from != nil {
		f.From = from
	}
	if to, err := parseDayOrTime(q.Get("to")); err == nil && to != nil {
		// Inclusive end day: bump date-only to next midnight.
		if len(q.Get("to")) == 10 {
			n := to.Add(24 * time.Hour)
			f.To = &n
		} else {
			f.To = to
		}
	}
	limit := 50
	if n, err := strconv.Atoi(q.Get("limit")); err == nil && n > 0 {
		limit = n
		if limit > 500 {
			limit = 500
		}
	}
	return f, limit
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseDayOrTime(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return &t, nil
	}
	return nil, api.BadRequest("invalid from/to time")
}

func writeJobsCSV(w http.ResponseWriter, items []api.Job) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="print-history.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "name", "state", "external_sn", "artifact_id", "est_seconds", "est_grams", "cost_estimate", "cost_actual", "currency", "created_at", "started_at", "finished_at", "error"})
	for _, j := range items {
		_ = cw.Write([]string{
			j.ID, j.Name, j.State, j.ExternalSN, j.ArtifactID,
			strconv.Itoa(j.EstSeconds), strconv.FormatFloat(j.EstGrams, 'f', -1, 64),
			strconv.FormatFloat(j.CostEstimate, 'f', -1, 64), strconv.FormatFloat(j.CostActual, 'f', -1, 64),
			j.Currency, j.CreatedAt.Format(time.RFC3339), fmtTime(j.StartedAt), fmtTime(j.FinishedAt), j.Error,
		})
	}
	cw.Flush()
}

func fmtTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// respondCreated inserts a queued job, re-estimates cost, and dispatches when the target SN is idle.
func respondCreated(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool, bus *events.Bus, siteID string, req api.CreateJobRequest) {
	job, err := db.CreateJob(r.Context(), pool, siteID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if rc, err := db.GetRateCard(r.Context(), pool, siteID); err == nil {
		est := db.EstimateJobCost(rc, job.EstSeconds, job.EstGrams, job.Copies)
		_ = db.ApplyJobCost(r.Context(), pool, siteID, job.ID, est, 0, rc.Currency)
		job.CostEstimate = est
		job.Currency = rc.Currency
	}
	PublishJobQueued(r.Context(), bus, siteID, job)
	if job.ExternalSN != "" {
		if idle, err := db.ListIdlePrinters(r.Context(), pool, siteID); err == nil {
			for _, p := range idle {
				if p.ExternalSN == job.ExternalSN {
					if dispatched, err := DispatchOne(r.Context(), pool, bus, siteID, job.ID, job.ExternalSN); err == nil {
						job = dispatched
					}
					break
				}
			}
		}
	}
	httpx.WriteJSON(w, job)
}
