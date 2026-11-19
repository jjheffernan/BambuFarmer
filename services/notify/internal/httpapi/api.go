package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/export"
	"github.com/jjheffernan/BambuFarmer/services/notify/internal/store"
)

// IngestAllowlist is the set of types accepted by POST /events.
var IngestAllowlist = map[string]bool{
	events.TypeJobFinished: true,
	events.TypeJobFailed:   true,
	events.TypeAlertRaised: true,
}

// Handler mounts notify HTTP APIs.
type Handler struct {
	Store  *store.Store
	Bus    *events.Bus
	Export *export.Client
	SiteID string
}

// Register mounts routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/notify", h.status)
	mux.HandleFunc("POST /api/v1/notify/test", h.test)
	mux.HandleFunc("GET /api/v1/notify/channels", h.listChannels)
	mux.HandleFunc("POST /api/v1/notify/channels", h.createChannel)
	mux.HandleFunc("GET /api/v1/notify/channels/{id}", h.getChannel)
	mux.HandleFunc("PATCH /api/v1/notify/channels/{id}", h.patchChannel)
	mux.HandleFunc("DELETE /api/v1/notify/channels/{id}", h.deleteChannel)
	mux.HandleFunc("GET /api/v1/notify/printers", h.listPrinters)
	mux.HandleFunc("POST /api/v1/notify/printers", h.upsertPrinter)
	mux.HandleFunc("GET /api/v1/notify/printers/{sn}", h.getPrinter)
	mux.HandleFunc("PATCH /api/v1/notify/printers/{sn}", h.patchPrinter)
	mux.HandleFunc("DELETE /api/v1/notify/printers/{sn}", h.deletePrinter)
	mux.HandleFunc("POST /api/v1/notify/events", h.ingest)
}

func (h *Handler) site(r *http.Request) string {
	if s := strings.TrimSpace(r.URL.Query().Get("site_id")); s != "" {
		return s
	}
	if h.SiteID != "" {
		return h.SiteID
	}
	return "default"
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	site := h.site(r)
	ch, pr, err := h.Store.Counts(r.Context(), site)
	if err != nil {
		httpserver.WriteError(w, api.Internal(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"configured": ch > 0,
		"service":    "notify",
		"mode":       "channels",
		"site_id":    site,
		"channels":   ch,
		"printers":   pr,
		"events":     store.DefaultEventTypes,
		"note":       "CRUD /api/v1/notify/channels + /printers; POST /events ingest; stream fan-out for job.finished|job.failed|alert.raised. ntfy/Web Push later.",
	})
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID string `json:"channel_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	site := h.site(r)
	body := export.TestBody(site)
	var deliveries []export.Delivery
	if req.ChannelID != "" {
		ch, err := h.Store.GetChannel(r.Context(), req.ChannelID)
		if err != nil {
			httpserver.WriteError(w, api.NotFound(err.Error()))
			return
		}
		d := export.Delivery{ChannelID: ch.ID, Name: ch.Name}
		if err := h.Export.PostChannel(r.Context(), ch, body); err != nil {
			d.Error = err.Error()
		} else {
			d.OK = true
		}
		deliveries = []export.Delivery{d}
	} else {
		channels, err := h.Store.ListChannels(r.Context(), site)
		if err != nil {
			httpserver.WriteError(w, api.Internal(err.Error()))
			return
		}
		for _, ch := range channels {
			if !ch.Enabled {
				continue
			}
			d := export.Delivery{ChannelID: ch.ID, Name: ch.Name}
			if err := h.Export.PostChannel(r.Context(), ch, body); err != nil {
				d.Error = err.Error()
			} else {
				d.OK = true
			}
			deliveries = append(deliveries, d)
		}
	}
	sent := 0
	for _, d := range deliveries {
		if d.OK {
			sent++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sent":       sent > 0,
		"count":      sent,
		"type":       body.Type,
		"deliveries": deliveries,
	})
}

func (h *Handler) listChannels(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.ListChannels(r.Context(), h.site(r))
	if err != nil {
		httpserver.WriteError(w, api.Internal(err.Error()))
		return
	}
	if list == nil {
		list = []store.Channel{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"channels": list})
}

func (h *Handler) createChannel(w http.ResponseWriter, r *http.Request) {
	var in store.ChannelCreate
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpserver.WriteError(w, api.BadRequest("invalid JSON"))
		return
	}
	if strings.TrimSpace(in.SiteID) == "" {
		in.SiteID = h.site(r)
	}
	ch, err := h.Store.CreateChannel(r.Context(), in)
	if err != nil {
		httpserver.WriteError(w, api.BadRequest(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, ch)
}

func (h *Handler) getChannel(w http.ResponseWriter, r *http.Request) {
	ch, err := h.Store.GetChannel(r.Context(), r.PathValue("id"))
	if err != nil {
		httpserver.WriteError(w, api.NotFound(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (h *Handler) patchChannel(w http.ResponseWriter, r *http.Request) {
	var in store.ChannelPatch
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpserver.WriteError(w, api.BadRequest("invalid JSON"))
		return
	}
	ch, err := h.Store.PatchChannel(r.Context(), r.PathValue("id"), in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httpserver.WriteError(w, api.NotFound(err.Error()))
			return
		}
		httpserver.WriteError(w, api.BadRequest(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

func (h *Handler) deleteChannel(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DeleteChannel(r.Context(), r.PathValue("id")); err != nil {
		httpserver.WriteError(w, api.NotFound(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listPrinters(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.ListPrinters(r.Context(), h.site(r))
	if err != nil {
		httpserver.WriteError(w, api.Internal(err.Error()))
		return
	}
	if list == nil {
		list = []store.Printer{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"printers": list})
}

func (h *Handler) upsertPrinter(w http.ResponseWriter, r *http.Request) {
	var in store.PrinterUpsert
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpserver.WriteError(w, api.BadRequest("invalid JSON"))
		return
	}
	if strings.TrimSpace(in.SiteID) == "" {
		in.SiteID = h.site(r)
	}
	p, err := h.Store.UpsertPrinter(r.Context(), in)
	if err != nil {
		httpserver.WriteError(w, api.BadRequest(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) getPrinter(w http.ResponseWriter, r *http.Request) {
	p, err := h.Store.GetPrinter(r.Context(), r.PathValue("sn"))
	if err != nil {
		httpserver.WriteError(w, api.NotFound(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) patchPrinter(w http.ResponseWriter, r *http.Request) {
	var in store.PrinterPatch
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpserver.WriteError(w, api.BadRequest("invalid JSON"))
		return
	}
	p, err := h.Store.PatchPrinter(r.Context(), r.PathValue("sn"), in)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httpserver.WriteError(w, api.NotFound(err.Error()))
			return
		}
		httpserver.WriteError(w, api.BadRequest(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deletePrinter(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DeletePrinter(r.Context(), r.PathValue("sn")); err != nil {
		httpserver.WriteError(w, api.NotFound(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string         `json:"type"`
		SiteID  string         `json:"site_id"`
		Source  string         `json:"source"`
		Payload map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteError(w, api.BadRequest("invalid JSON"))
		return
	}
	typ := strings.TrimSpace(req.Type)
	if !IngestAllowlist[typ] {
		e := api.BadRequest("type not allowed")
		e.Details = map[string]any{
			"allowed": []string{events.TypeJobFinished, events.TypeJobFailed, events.TypeAlertRaised},
		}
		httpserver.WriteError(w, e)
		return
	}
	stream := events.StreamForType(typ)
	if stream == "" {
		httpserver.WriteError(w, api.BadRequest("unknown stream for type"))
		return
	}
	if h.Bus == nil {
		httpserver.WriteError(w, api.NotConfigured("REDIS_URL unset"))
		return
	}
	site := strings.TrimSpace(req.SiteID)
	if site == "" {
		site = h.site(r)
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "notify-ingest"
	}
	env := events.NewEnvelope(typ, source, site, req.Payload)
	id, err := h.Bus.Publish(r.Context(), stream, env)
	if err != nil {
		httpserver.WriteError(w, api.BadGateway(err.Error()))
		return
	}
	// Fan-out happens via the stream consumer (GroupNotify) — avoid double POST.
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":     id,
		"stream": stream,
		"type":   typ,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
