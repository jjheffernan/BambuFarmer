package auth

import "github.com/jjheffernan/BambuFarmer/packages/db"

// HashPassword hashes a plaintext password (Argon2id).
func HashPassword(password string) (string, error) {
	return db.HashPassword(password)
}

// VerifyPassword checks a plaintext password against an encoded hash.
func VerifyPassword(encoded, password string) (bool, error) {
	return db.VerifyPassword(encoded, password)
}
