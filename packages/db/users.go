package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// Role constants (MVP subset of DESIGN_SPEC §10.2).
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// CreateUser inserts a local account. password is plaintext (hashed here).
func CreateUser(ctx context.Context, pool *pgxpool.Pool, username, password, displayName string, roles []string) (api.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return api.User{}, Validation("username required")
	}
	if len(roles) == 0 {
		roles = []string{RoleViewer}
	}
	for _, r := range roles {
		switch r {
		case RoleAdmin, RoleOperator, RoleViewer:
		default:
			return api.User{}, Validation("invalid role: " + r)
		}
	}
	hash, err := HashPassword(password)
	if err != nil {
		return api.User{}, err
	}
	id := newUserID()
	dn := displayName
	if dn == "" {
		dn = username
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, display_name, roles, active)
		VALUES ($1,$2,$3,$4,$5,TRUE)`,
		id, username, hash, dn, roles,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return api.User{}, api.Conflict("username already exists")
		}
		return api.User{}, err
	}
	return GetUserByID(ctx, pool, id)
}

// CountUsers returns how many users exist.
func CountUsers(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var n int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// GetUserByID loads a user (no password hash in returned struct).
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id string) (api.User, error) {
	u, _, err := getUser(ctx, pool, `id=$1`, id)
	return u, err
}

// GetUserByUsername loads by exact username (case-sensitive store; lookup lower).
func GetUserByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (api.User, error) {
	u, _, err := getUser(ctx, pool, `lower(username)=lower($1)`, username)
	return u, err
}

// AuthenticateUser verifies username/password and returns the user if active.
func AuthenticateUser(ctx context.Context, pool *pgxpool.Pool, username, password string) (api.User, error) {
	u, hash, err := getUser(ctx, pool, `lower(username)=lower($1)`, username)
	if err != nil {
		return api.User{}, api.Unauthorized("invalid username or password")
	}
	if !u.Active {
		return api.User{}, api.Unauthorized("account disabled")
	}
	ok, err := VerifyPassword(hash, password)
	if err != nil || !ok {
		return api.User{}, api.Unauthorized("invalid username or password")
	}
	return u, nil
}

func getUser(ctx context.Context, pool *pgxpool.Pool, where string, arg string) (api.User, string, error) {
	var u api.User
	var hash string
	var roles []string
	err := pool.QueryRow(ctx, `
		SELECT id, username, password_hash, display_name, roles, active, created_at, updated_at
		FROM users WHERE `+where, arg).Scan(
		&u.ID, &u.Username, &hash, &u.DisplayName, &roles, &u.Active, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return u, "", MapQueryError(err)
	}
	u.Roles = roles
	return u, hash, nil
}

// ListUsers returns all users ordered by username.
func ListUsers(ctx context.Context, pool *pgxpool.Pool) ([]api.User, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, username, display_name, roles, active, created_at, updated_at
		FROM users ORDER BY lower(username)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.User
	for rows.Next() {
		var u api.User
		var roles []string
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &roles, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Roles = roles
		out = append(out, u)
	}
	return out, rows.Err()
}

// UpdateUser patches roles and/or active flag. Empty roles = leave unchanged.
func UpdateUser(ctx context.Context, pool *pgxpool.Pool, id string, roles []string, active *bool) (api.User, error) {
	u, err := GetUserByID(ctx, pool, id)
	if err != nil {
		return u, err
	}
	if roles != nil {
		if len(roles) == 0 {
			return u, Validation("roles cannot be empty")
		}
		for _, r := range roles {
			switch r {
			case RoleAdmin, RoleOperator, RoleViewer:
			default:
				return u, Validation("invalid role: " + r)
			}
		}
		u.Roles = roles
	}
	if active != nil {
		u.Active = *active
	}
	_, err = pool.Exec(ctx, `
		UPDATE users SET roles=$1, active=$2, updated_at=now() WHERE id=$3`,
		u.Roles, u.Active, id)
	if err != nil {
		return u, err
	}
	return GetUserByID(ctx, pool, id)
}

// SetPassword updates password hash for a user (admin reset / bootstrap).
func SetPassword(ctx context.Context, pool *pgxpool.Pool, id, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	tag, err := pool.Exec(ctx, `UPDATE users SET password_hash=$1, updated_at=now() WHERE id=$2`, hash, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return api.NotFound("user not found")
	}
	return nil
}

// ChangePassword verifies the current password then sets a new one.
func ChangePassword(ctx context.Context, pool *pgxpool.Pool, id, currentPassword, newPassword string) error {
	_, hash, err := getUser(ctx, pool, `id=$1`, id)
	if err != nil {
		return err
	}
	ok, err := VerifyPassword(hash, currentPassword)
	if err != nil {
		return err
	}
	if !ok {
		return api.Unauthorized("current password is incorrect")
	}
	if currentPassword == newPassword {
		return Validation("new password must differ from current password")
	}
	return SetPassword(ctx, pool, id, newPassword)
}

func newUserID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "usr_" + hex.EncodeToString(b) + hex.EncodeToString([]byte(time.Now().UTC().Format("150405")))
}
