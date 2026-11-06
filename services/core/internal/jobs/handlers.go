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

// Register mounts job list/create/get/transition routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, bus *events.Bus, siteID string) {
	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListJobs(r.Context(), pool, siteID, 50)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			job, err := db.CreateJob(r.Context(), pool, siteID, req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if rc, err := db.GetRateCard(r.Context(), pool, siteID); err == nil {
				est := db.EstimateJobCost(rc, job.EstSeconds, job.EstGrams, job.Copies)
				_ = db.ApplyJobCost(r.Context(), pool, siteID, job.ID, est, 0, rc.Currency)
				job.CostEstimate = est
				job.Currency = rc.Currency
			}
			if bus != nil {
				env := events.NewEnvelope(events.TypeJobQueued, "core", siteID, map[string]any{
					"job_id": job.ID, "name": job.Name, "external_sn": job.ExternalSN,
				})
				_, _ = bus.Publish(r.Context(), events.StreamJob, env)
				if job.ExternalSN != "" {
					cmd := events.NewEnvelope("command.start_print", "core", siteID, map[string]any{
						"job_id": job.ID, "external_sn": job.ExternalSN,
					})
					_, _ = bus.Publish(r.Context(), events.CommandStream(siteID), cmd)
				}
			}
			httpx.WriteJSON(w, job)
		})
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			job, err := db.GetJob(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
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
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			job, err := db.TransitionJob(r.Context(), pool, siteID, r.PathValue("id"), body.State, "api", body.Reason)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			httpx.WriteJSON(w, job)
		})
	})
}
