package artifacts

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/httpx"
)

// Register mounts artifact upload/list/download routes.
func Register(mux *http.ServeMux, pool *pgxpool.Pool, siteID string, store Store) {
	mux.HandleFunc("GET /api/v1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			items, err := db.ListArtifacts(r.Context(), pool, siteID, 50)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if items == nil {
				items = []api.Artifact{}
			}
			httpx.WriteJSON(w, items)
		})
	})
	mux.HandleFunc("POST /api/v1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			if err := r.ParseMultipartForm(64 << 20); err != nil {
				httpx.WriteError(w, api.BadRequest("multipart form required"))
				return
			}
			file, hdr, err := r.FormFile("file")
			if err != nil {
				httpx.WriteError(w, api.BadRequest("file field required"))
				return
			}
			defer file.Close()
			name := filepath.Base(hdr.Filename)
			if name == "" || name == "." {
				name = "plate.gcode.3mf"
			}
			body, err := io.ReadAll(io.LimitReader(file, 64<<20))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			estSec, _ := strconv.Atoi(r.FormValue("est_seconds"))
			estGrams, _ := strconv.ParseFloat(r.FormValue("est_grams"), 64)
			id := db.NewID()
			key := id + extFor(name)
			if err := store.Write(key, body); err != nil {
				httpx.WriteError(w, api.Wrap(http.StatusInternalServerError, api.CodeInternal, "store artifact", err))
				return
			}
			art, err := db.InsertArtifact(r.Context(), pool, siteID, api.Artifact{
				ID:         id,
				Filename:   name,
				ObjectKey:  key,
				Bytes:      int64(len(body)),
				EstSeconds: estSec,
				EstGrams:   estGrams,
			})
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, art)
		})
	})
	mux.HandleFunc("GET /api/v1/artifacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			art, err := db.GetArtifact(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			httpx.WriteJSON(w, art)
		})
	})
	mux.HandleFunc("GET /api/v1/artifacts/{id}/content", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			art, err := db.GetArtifact(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			f, err := store.Open(art.ObjectKey)
			if err != nil {
				httpx.WriteError(w, api.NotFound("artifact bytes missing"))
				return
			}
			defer f.Close()
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="`+art.Filename+`"`)
			_, _ = io.Copy(w, f)
		})
	})
	mux.HandleFunc("DELETE /api/v1/artifacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		httpx.RequireDB(w, pool, func() {
			art, err := db.GetArtifact(r.Context(), pool, siteID, r.PathValue("id"))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}
			if err := store.Remove(art.ObjectKey); err != nil {
				httpx.WriteError(w, api.Wrap(http.StatusInternalServerError, api.CodeInternal, "remove artifact bytes", err))
				return
			}
			if err := db.DeleteArtifact(r.Context(), pool, siteID, art.ID); err != nil {
				httpx.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func extFor(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".gcode.3mf") {
		return ".gcode.3mf"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return ".gcode.3mf"
	}
	return ext
}
