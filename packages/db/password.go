package db

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP-ish defaults for interactive logins).
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// HashPassword returns an encoded Argon2id hash: argon2id$v=19$m=...,t=...,p=...$salt$hash
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", Validation("password required")
	}
	if len(password) < 8 {
		return "", Validation("password must be at least 8 characters")
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks password against an encoded hash from HashPassword.
func VerifyPassword(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" {
		return false, fmt.Errorf("unsupported password hash format")
	}
	var version int
	if _, err := fmt.Sscanf(parts[1], "v=%d", &version); err != nil {
		return false, err
	}
	var mem, timeCost, threads uint32
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &mem, &timeCost, &threads); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, timeCost, mem, uint8(threads), uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return false, nil
	}
	return true, nil
}
