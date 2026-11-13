package jobs

import (
	"encoding/json"
	"net/http"

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
			items, err := db.ListJobs(r.Context(), pool, siteID, 50, db.JobListFilter{
				Query: r.URL.Query().Get("q"),
				State: r.URL.Query().Get("state"),
			})
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.Job{}
			}
			httpx.WriteJSON(w, items)
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
