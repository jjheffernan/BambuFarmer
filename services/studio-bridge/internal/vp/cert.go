package vp

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadCertPEM returns the PEM certificate from CertDir (empty if missing).
func ReadCertPEM(dir string) ([]byte, error) {
	path := filepath.Join(dir, "cert.pem")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// CertFingerprintSHA256 returns colon-ish hex SHA-256 of the leaf cert.
func CertFingerprintSHA256(pemBytes []byte) (string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", fmt.Errorf("no pem block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cert.Raw)
	hexed := strings.ToUpper(hex.EncodeToString(sum[:]))
	var parts []string
	for i := 0; i+1 < len(hexed); i += 2 {
		parts = append(parts, hexed[i:i+2])
	}
	return strings.Join(parts, ":"), nil
}
