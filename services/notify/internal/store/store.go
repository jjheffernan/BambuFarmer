package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

// DefaultEventTypes for new channels.
var DefaultEventTypes = []string{"job.finished", "job.failed", "alert.raised"}

// Channel is an outbound webhook destination.
type Channel struct {
	ID           string    `json:"id"`
	SiteID       string    `json:"site_id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Enabled      bool      `json:"enabled"`
	EventTypes   []string  `json:"event_types"`
	HasAuth      bool      `json:"has_auth"`
	AuthHeader   string    `json:"-"` // never serialize on GET
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChannelCreate is the create payload.
type ChannelCreate struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	Enabled    *bool    `json:"enabled,omitempty"`
	EventTypes []string `json:"event_types,omitempty"`
	AuthHeader string   `json:"auth_header,omitempty"`
	SiteID     string   `json:"site_id,omitempty"`
}

// ChannelPatch is a partial update.
type ChannelPatch struct {
	Name       *string  `json:"name,omitempty"`
	URL        *string  `json:"url,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
	EventTypes []string `json:"event_types,omitempty"`
	AuthHeader *string  `json:"auth_header,omitempty"`
}

// Printer is a notify routing entry for an SN.
type Printer struct {
	ExternalSN string    `json:"external_sn"`
	SiteID     string    `json:"site_id"`
	Name       string    `json:"name"`
	ChannelIDs []string  `json:"channel_ids"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PrinterUpsert creates or replaces a printer route.
type PrinterUpsert struct {
	ExternalSN string   `json:"external_sn"`
	Name       string   `json:"name,omitempty"`
	ChannelIDs []string `json:"channel_ids,omitempty"`
	SiteID     string   `json:"site_id,omitempty"`
}

// PrinterPatch updates name and/or channel bindings.
type PrinterPatch struct {
	Name       *string  `json:"name,omitempty"`
	ChannelIDs []string `json:"channel_ids,omitempty"`
}

// Store backs notify channel + printer tables.
type Store struct {
	Pool *pgxpool.Pool
}

func normalizeTypes(in []string) []string {
	if len(in) == 0 {
		return append([]string{}, DefaultEventTypes...)
	}
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	if len(out) == 0 {
		return append([]string{}, DefaultEventTypes...)
	}
	return out
}

type scannable interface {
	Scan(dest ...any) error
}

func scanChannel(row scannable) (Channel, error) {
	var c Channel
	var typesRaw []byte
	var auth string
	err := row.Scan(&c.ID, &c.SiteID, &c.Name, &c.URL, &c.Enabled, &typesRaw, &auth, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	_ = json.Unmarshal(typesRaw, &c.EventTypes)
	if c.EventTypes == nil {
		c.EventTypes = []string{}
	}
	c.AuthHeader = auth
	c.HasAuth = strings.TrimSpace(auth) != ""
	return c, nil
}

// ListChannels returns channels for a site (empty site = all).
func (s *Store) ListChannels(ctx context.Context, siteID string) ([]Channel, error) {
	q := `SELECT id, site_id, name, url, enabled, event_types, auth_header, created_at, updated_at
		FROM notify_channels`
	args := []any{}
	if siteID != "" {
		q += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	q += ` ORDER BY name`
	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Channel
	for rows.Next() {
		c, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetChannel loads one channel by id.
func (s *Store) GetChannel(ctx context.Context, id string) (Channel, error) {
	c, err := scanChannel(s.Pool.QueryRow(ctx, `
		SELECT id, site_id, name, url, enabled, event_types, auth_header, created_at, updated_at
		FROM notify_channels WHERE id = $1`, id))
	if err == pgx.ErrNoRows {
		return c, fmt.Errorf("channel not found")
	}
	return c, err
}

// CreateChannel inserts a channel.
func (s *Store) CreateChannel(ctx context.Context, in ChannelCreate) (Channel, error) {
	name := strings.TrimSpace(in.Name)
	url := strings.TrimSpace(in.URL)
	if name == "" || url == "" {
		return Channel{}, fmt.Errorf("name and url required")
	}
	siteID := strings.TrimSpace(in.SiteID)
	if siteID == "" {
		siteID = "default"
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	types := normalizeTypes(in.EventTypes)
	typesJSON, _ := json.Marshal(types)
	id := "nch_" + db.NewID()
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO notify_channels (id, site_id, name, url, enabled, event_types, auth_header)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)`,
		id, siteID, name, url, enabled, string(typesJSON), strings.TrimSpace(in.AuthHeader))
	if err != nil {
		return Channel{}, err
	}
	return s.GetChannel(ctx, id)
}

// PatchChannel updates fields.
func (s *Store) PatchChannel(ctx context.Context, id string, in ChannelPatch) (Channel, error) {
	cur, err := s.GetChannel(ctx, id)
	if err != nil {
		return Channel{}, err
	}
	if in.Name != nil {
		cur.Name = strings.TrimSpace(*in.Name)
	}
	if in.URL != nil {
		cur.URL = strings.TrimSpace(*in.URL)
	}
	if in.Enabled != nil {
		cur.Enabled = *in.Enabled
	}
	if in.EventTypes != nil {
		cur.EventTypes = normalizeTypes(in.EventTypes)
	}
	if in.AuthHeader != nil {
		cur.AuthHeader = strings.TrimSpace(*in.AuthHeader)
	}
	if cur.Name == "" || cur.URL == "" {
		return Channel{}, fmt.Errorf("name and url required")
	}
	typesJSON, _ := json.Marshal(cur.EventTypes)
	_, err = s.Pool.Exec(ctx, `
		UPDATE notify_channels SET name=$2, url=$3, enabled=$4, event_types=$5::jsonb, auth_header=$6, updated_at=now()
		WHERE id=$1`, id, cur.Name, cur.URL, cur.Enabled, string(typesJSON), cur.AuthHeader)
	if err != nil {
		return Channel{}, err
	}
	return s.GetChannel(ctx, id)
}

// DeleteChannel removes a channel.
func (s *Store) DeleteChannel(ctx context.Context, id string) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM notify_channels WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// Counts returns channel and printer totals for a site.
func (s *Store) Counts(ctx context.Context, siteID string) (channels, printers int, err error) {
	qCh := `SELECT count(*) FROM notify_channels`
	qPr := `SELECT count(*) FROM notify_printers`
	args := []any{}
	if siteID != "" {
		qCh += ` WHERE site_id = $1`
		qPr += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	if err = s.Pool.QueryRow(ctx, qCh, args...).Scan(&channels); err != nil {
		return
	}
	err = s.Pool.QueryRow(ctx, qPr, args...).Scan(&printers)
	return
}

func scanPrinter(row scannable) (Printer, error) {
	var p Printer
	var idsRaw []byte
	err := row.Scan(&p.ExternalSN, &p.SiteID, &p.Name, &idsRaw, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	_ = json.Unmarshal(idsRaw, &p.ChannelIDs)
	if p.ChannelIDs == nil {
		p.ChannelIDs = []string{}
	}
	return p, nil
}

// ListPrinters returns printer routes.
func (s *Store) ListPrinters(ctx context.Context, siteID string) ([]Printer, error) {
	q := `SELECT external_sn, site_id, name, channel_ids, created_at, updated_at FROM notify_printers`
	args := []any{}
	if siteID != "" {
		q += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	q += ` ORDER BY external_sn`
	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Printer
	for rows.Next() {
		p, err := scanPrinter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPrinter loads one printer route.
func (s *Store) GetPrinter(ctx context.Context, sn string) (Printer, error) {
	p, err := scanPrinter(s.Pool.QueryRow(ctx, `
		SELECT external_sn, site_id, name, channel_ids, created_at, updated_at
		FROM notify_printers WHERE external_sn = $1`, sn))
	if err == pgx.ErrNoRows {
		return p, fmt.Errorf("printer not found")
	}
	return p, err
}

// UpsertPrinter creates or replaces a printer route.
func (s *Store) UpsertPrinter(ctx context.Context, in PrinterUpsert) (Printer, error) {
	sn := strings.TrimSpace(in.ExternalSN)
	if sn == "" {
		return Printer{}, fmt.Errorf("external_sn required")
	}
	siteID := strings.TrimSpace(in.SiteID)
	if siteID == "" {
		siteID = "default"
	}
	ids := in.ChannelIDs
	if ids == nil {
		ids = []string{}
	}
	idsJSON, _ := json.Marshal(ids)
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO notify_printers (external_sn, site_id, name, channel_ids)
		VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (external_sn) DO UPDATE SET
			site_id = EXCLUDED.site_id,
			name = EXCLUDED.name,
			channel_ids = EXCLUDED.channel_ids,
			updated_at = now()`,
		sn, siteID, strings.TrimSpace(in.Name), string(idsJSON))
	if err != nil {
		return Printer{}, err
	}
	return s.GetPrinter(ctx, sn)
}

// PatchPrinter updates a printer route.
func (s *Store) PatchPrinter(ctx context.Context, sn string, in PrinterPatch) (Printer, error) {
	cur, err := s.GetPrinter(ctx, sn)
	if err != nil {
		return Printer{}, err
	}
	if in.Name != nil {
		cur.Name = strings.TrimSpace(*in.Name)
	}
	if in.ChannelIDs != nil {
		cur.ChannelIDs = in.ChannelIDs
	}
	idsJSON, _ := json.Marshal(cur.ChannelIDs)
	_, err = s.Pool.Exec(ctx, `
		UPDATE notify_printers SET name=$2, channel_ids=$3::jsonb, updated_at=now()
		WHERE external_sn=$1`, sn, cur.Name, string(idsJSON))
	if err != nil {
		return Printer{}, err
	}
	return s.GetPrinter(ctx, sn)
}

// DeletePrinter removes a printer route.
func (s *Store) DeletePrinter(ctx context.Context, sn string) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM notify_printers WHERE external_sn = $1`, sn)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("printer not found")
	}
	return nil
}

// ChannelsForEvent returns enabled channels that should receive this event.
// Unregistered SN → all enabled matching event-type channels.
// Registered SN with empty channel_ids → all enabled matching channels.
// Registered SN with channel_ids → only those (still event-type filtered).
func (s *Store) ChannelsForEvent(ctx context.Context, siteID, eventType, externalSN string) ([]Channel, error) {
	all, err := s.ListChannels(ctx, siteID)
	if err != nil {
		return nil, err
	}
	var matched []Channel
	for _, c := range all {
		if !c.Enabled {
			continue
		}
		if !typeAllowed(c.EventTypes, eventType) {
			continue
		}
		matched = append(matched, c)
	}
	if externalSN == "" {
		return matched, nil
	}
	p, err := s.GetPrinter(ctx, externalSN)
	if err != nil {
		// unregistered → site-wide matched channels
		return matched, nil
	}
	if len(p.ChannelIDs) == 0 {
		return matched, nil
	}
	allow := map[string]bool{}
	for _, id := range p.ChannelIDs {
		allow[id] = true
	}
	var out []Channel
	for _, c := range matched {
		if allow[c.ID] {
			out = append(out, c)
		}
	}
	return out, nil
}

func typeAllowed(types []string, eventType string) bool {
	for _, t := range types {
		if t == eventType || t == "*" {
			return true
		}
	}
	return false
}
