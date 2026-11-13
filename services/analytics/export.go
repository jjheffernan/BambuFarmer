package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// exportBundle holds rollup series for JSON/CSV export.
type exportBundle struct {
	Utilization []map[string]any `json:"utilization"`
	Costs       []map[string]any `json:"costs"`
	Failures    []map[string]any `json:"failures"`
}

func exportFormat(r *http.Request) string {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if q == "csv" || q == "json" {
		return q
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	if strings.Contains(accept, "text/csv") {
		return "csv"
	}
	return "json"
}

func exportKind(r *http.Request) string {
	k := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	switch k {
	case "utilization", "costs", "failures", "all":
		return k
	default:
		return "all"
	}
}

func writeExport(w http.ResponseWriter, format, kind string, bundle exportBundle) error {
	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="analytics-`+kind+`.csv"`)
		return writeExportCSV(w, kind, bundle)
	default:
		w.Header().Set("Content-Type", "application/json")
		out := map[string]any{}
		switch kind {
		case "utilization":
			out["utilization"] = bundle.Utilization
		case "costs":
			out["costs"] = bundle.Costs
		case "failures":
			out["failures"] = bundle.Failures
		default:
			out["utilization"] = bundle.Utilization
			out["costs"] = bundle.Costs
			out["failures"] = bundle.Failures
		}
		return json.NewEncoder(w).Encode(out)
	}
}

func writeExportCSV(w io.Writer, kind string, bundle exportBundle) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	switch kind {
	case "costs":
		return writeMapsCSV(cw, []string{"day", "site_id", "material_cost", "machine_cost", "labor_cost", "jobs"}, bundle.Costs)
	case "failures":
		return writeMapsCSV(cw, []string{"day", "site_id", "code", "count"}, bundle.Failures)
	case "utilization":
		return writeMapsCSV(cw, []string{"day", "site_id", "external_sn", "idle_s", "run_s", "fail_s", "samples"}, bundle.Utilization)
	default:
		if err := cw.Write([]string{"section", "day", "site_id", "external_sn", "idle_s", "run_s", "fail_s", "samples", "material_cost", "machine_cost", "labor_cost", "jobs", "code", "count"}); err != nil {
			return err
		}
		for _, r := range bundle.Utilization {
			if err := cw.Write([]string{
				"utilization", str(r["day"]), str(r["site_id"]), str(r["external_sn"]),
				str(r["idle_s"]), str(r["run_s"]), str(r["fail_s"]), str(r["samples"]),
				"", "", "", "", "", "",
			}); err != nil {
				return err
			}
		}
		for _, r := range bundle.Costs {
			if err := cw.Write([]string{
				"costs", str(r["day"]), str(r["site_id"]), "",
				"", "", "", "",
				str(r["material_cost"]), str(r["machine_cost"]), str(r["labor_cost"]), str(r["jobs"]),
				"", "",
			}); err != nil {
				return err
			}
		}
		for _, r := range bundle.Failures {
			if err := cw.Write([]string{
				"failures", str(r["day"]), str(r["site_id"]), "",
				"", "", "", "",
				"", "", "", "",
				str(r["code"]), str(r["count"]),
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	}
}

func writeMapsCSV(cw *csv.Writer, cols []string, rows []map[string]any) error {
	if err := cw.Write(cols); err != nil {
		return err
	}
	for _, r := range rows {
		rec := make([]string, len(cols))
		for i, c := range cols {
			rec[i] = str(r[c])
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
