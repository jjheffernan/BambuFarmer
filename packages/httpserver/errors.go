package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// WriteError writes a JSON ErrorBody. Status comes from *api.Error when present;
// otherwise defaults to 500 with code internal.
func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, body := MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// MapError converts err into HTTP status + API error body.
func MapError(err error) (int, api.ErrorBody) {
	if err == nil {
		return http.StatusInternalServerError, api.ErrorBody{Error: "unknown error", Code: api.CodeInternal}
	}
	if ae, ok := api.AsError(err); ok {
		return ae.HTTPStatus(), ae.Body()
	}
	return http.StatusInternalServerError, api.ErrorBody{
		Error: err.Error(),
		Code:  api.CodeInternal,
	}
}
