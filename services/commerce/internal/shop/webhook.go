package shop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Order is the store webhook body. Fields we do not use are ignored.
type Order struct {
	Name       string `json:"name"`
	ExternalID string `json:"external_id"`
	ArtifactID string `json:"artifact_id,omitempty"`
	SpoolID    string `json:"spool_id,omitempty"`
}

// Result is what the webhook returns.
type Result struct {
	Queued bool   `json:"queued"`
	JobID  string `json:"job_id,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Register mounts POST /api/v1/commerce/webhook.
func Register(mux *http.ServeMux, coreURL string) {
	mux.HandleFunc("POST /api/v1/commerce/webhook", func(w http.ResponseWriter, r *http.Request) {
		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		res, status, err := QueueJob(r.Context(), coreURL, nil, order)
		if err != nil {
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, status, res)
	})

	mux.HandleFunc("POST /api/v1/commerce/orders/status", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID string `json:"order_id"`
			Status  string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		res, code, err := MarkStatus(body.OrderID, body.Status)
		if err != nil {
			writeJSON(w, code, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, code, res)
	})
}

// QueueJob POSTs /api/v1/jobs on core. An empty core URL accepts the webhook and does not queue.
func QueueJob(ctx context.Context, coreURL string, client *http.Client, order Order) (Result, int, error) {
	coreURL = strings.TrimRight(strings.TrimSpace(coreURL), "/")
	if coreURL == "" {
		return Result{Queued: false, Reason: "CORE_URL unset"}, http.StatusOK, nil
	}
	name := strings.TrimSpace(order.Name)
	if name == "" {
		name = strings.TrimSpace(order.ExternalID)
	}
	if name == "" {
		name = "shop-order"
	}
	body, err := json.Marshal(map[string]string{
		"name":        name,
		"artifact_id": strings.TrimSpace(order.ArtifactID),
		"spool_id":    strings.TrimSpace(order.SpoolID),
	})
	if err != nil {
		return Result{}, http.StatusInternalServerError, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, coreURL+"/api/v1/jobs", bytes.NewReader(body))
	if err != nil {
		return Result{}, http.StatusBadGateway, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return Result{}, http.StatusBadGateway, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return Result{}, http.StatusBadGateway, errStatus(res.Status)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(raw, &created)
	return Result{Queued: true, JobID: created.ID}, http.StatusOK, nil
}

// Shipment is an echoed pack or ship status. Nothing is stored.
type Shipment struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// MarkStatus accepts packed or shipped for an order id.
func MarkStatus(orderID, status string) (Shipment, int, error) {
	orderID = strings.TrimSpace(orderID)
	status = strings.ToLower(strings.TrimSpace(status))
	if orderID == "" {
		return Shipment{}, http.StatusBadRequest, fmt.Errorf("order_id required")
	}
	if status != "packed" && status != "shipped" {
		return Shipment{}, http.StatusBadRequest, fmt.Errorf("status must be packed or shipped")
	}
	return Shipment{OrderID: orderID, Status: status}, http.StatusOK, nil
}

type statusErr string

func (e statusErr) Error() string { return string(e) }

func errStatus(status string) error { return statusErr("core jobs: " + status) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
