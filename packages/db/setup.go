package db

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

//go:embed example_presets.json
var examplePresetsJSON []byte

// GetInstallState returns the singleton install/wizard row.
func GetInstallState(ctx context.Context, pool *pgxpool.Pool) (api.InstallState, error) {
	var s api.InstallState
	var completedAt *time.Time
	err := pool.QueryRow(ctx, `
		SELECT completed, completed_at, version, step, site_name, timezone, auth_mode,
		       farm_manager_url, farm_manager_transport, spoolman_url, manyfold_url,
		       admin_username, notes, updated_at
		FROM install_state WHERE id=1`).Scan(
		&s.Completed, &completedAt, &s.Version, &s.Step, &s.SiteName, &s.Timezone, &s.AuthMode,
		&s.FarmManagerURL, &s.FarmManagerTransport, &s.SpoolmanURL, &s.ManyfoldURL,
		&s.AdminUsername, &s.Notes, &s.UpdatedAt,
	)
	if err != nil {
		return s, err
	}
	s.CompletedAt = completedAt
	return s, nil
}

// SaveInstallDraft updates wizard fields without marking complete.
func SaveInstallDraft(ctx context.Context, pool *pgxpool.Pool, siteID string, req api.SetupRequest) (api.InstallState, error) {
	if req.Step == "" {
		req.Step = "draft"
	}
	auth := req.AuthMode
	if auth == "" {
		auth = "lan"
	}
	if auth == "local_insecure" {
		auth = "lan"
	}
	transport := req.FarmManagerTransport
	if transport == "" {
		transport = "fm_native"
	}
	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}
	name := req.SiteName
	if name == "" {
		name = "Default Site"
	}
	admin := req.AdminUsername
	if admin == "" {
		admin = "admin"
	}

	_, err := pool.Exec(ctx, `
		UPDATE install_state SET
			step=$1, site_name=$2, timezone=$3, auth_mode=$4,
			farm_manager_url=$5, farm_manager_transport=$6,
			spoolman_url=$7, manyfold_url=$8, admin_username=$9, notes=$10,
			updated_at=now()
		WHERE id=1`,
		req.Step, name, tz, auth,
		req.FarmManagerURL, transport,
		req.SpoolmanURL, req.ManyfoldURL, admin, req.Notes,
	)
	if err != nil {
		return api.InstallState{}, err
	}

	// Keep sites row in sync for display name / timezone.
	_, _ = pool.Exec(ctx, `
		UPDATE sites SET name=$1, timezone=$2 WHERE id=$3`, name, tz, siteID)

	if req.RateCard != nil {
		if _, err := UpsertRateCard(ctx, pool, siteID, *req.RateCard); err != nil {
			return api.InstallState{}, err
		}
	}
	return GetInstallState(ctx, pool)
}

// CompleteInstall marks the wizard done and applies final config.
// When no users exist yet, AdminPassword is required and creates the first admin.
func CompleteInstall(ctx context.Context, pool *pgxpool.Pool, siteID, appVersion string, req api.SetupRequest) (api.InstallState, error) {
	n, err := CountUsers(ctx, pool)
	if err != nil {
		return api.InstallState{}, err
	}
	if n == 0 {
		if req.AdminPassword == "" {
			return api.InstallState{}, Validation("admin_password required to finish setup")
		}
	}

	req.Step = "done"
	if _, err := SaveInstallDraft(ctx, pool, siteID, req); err != nil {
		return api.InstallState{}, err
	}
	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		UPDATE install_state SET completed=TRUE, completed_at=$1, version=$2, step='done', updated_at=$1
		WHERE id=1`, now, appVersion)
	if err != nil {
		return api.InstallState{}, err
	}

	if n == 0 {
		admin := req.AdminUsername
		if admin == "" {
			admin = "admin"
		}
		if _, err := CreateUser(ctx, pool, admin, req.AdminPassword, admin, []string{RoleAdmin}); err != nil {
			return api.InstallState{}, err
		}
	}

	if req.ImportExamplePresets {
		var bundle api.PresetBundle
		if err := json.Unmarshal(examplePresetsJSON, &bundle); err == nil {
			_ = ImportPresetBundle(ctx, pool, siteID, bundle)
		}
	}
	return GetInstallState(ctx, pool)
}

// ResetInstall clears completion (dev / re-run wizard). Does not wipe presets/jobs.
func ResetInstall(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		UPDATE install_state SET completed=FALSE, completed_at=NULL, step='welcome', updated_at=now()
		WHERE id=1`)
	return err
}
