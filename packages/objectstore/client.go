package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client puts/gets artifact bytes in a bucket.
type Client struct {
	mc     *minio.Client
	bucket string
}

// Dial builds a client from Config. Returns nil, nil when disabled.
func Dial(cfg Config) (*Client, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("objectstore: access/secret key required")
	}
	host, secure, err := parseEndpoint(cfg.Endpoint, cfg.Secure)
	if err != nil {
		return nil, err
	}
	mc, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: dial: %w", err)
	}
	return &Client{mc: mc, bucket: cfg.Bucket}, nil
}

// EnsureBucket creates the bucket if missing (idempotent).
func (c *Client) EnsureBucket(ctx context.Context) error {
	if c == nil {
		return nil
	}
	ok, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
}

// Put uploads object_key bytes.
func (c *Client) Put(ctx context.Context, key string, body []byte) error {
	if c == nil {
		return fmt.Errorf("objectstore: nil client")
	}
	key, err := sanitizeKey(key)
	if err != nil {
		return err
	}
	_, err = c.mc.PutObject(ctx, c.bucket, key, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	return err
}

// Delete removes object_key from the bucket. Missing keys are ignored.
func (c *Client) Delete(ctx context.Context, key string) error {
	if c == nil {
		return fmt.Errorf("objectstore: nil client")
	}
	key, err := sanitizeKey(key)
	if err != nil {
		return err
	}
	return c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}

// Get opens object_key for reading.
func (c *Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if c == nil {
		return nil, fmt.Errorf("objectstore: nil client")
	}
	key, err := sanitizeKey(key)
	if err != nil {
		return nil, err
	}
	obj, err := c.mc.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// Stat to surface NoSuchKey early.
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, err
	}
	return obj, nil
}

func parseEndpoint(endpoint string, secureHint bool) (host string, secure bool, err error) {
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		return "", false, fmt.Errorf("objectstore: empty endpoint")
	}
	if !strings.Contains(ep, "://") {
		return strings.TrimSuffix(ep, "/"), secureHint, nil
	}
	u, err := url.Parse(ep)
	if err != nil {
		return "", false, err
	}
	if u.Host == "" {
		return "", false, fmt.Errorf("objectstore: bad endpoint")
	}
	secure = u.Scheme == "https"
	return u.Host, secure, nil
}

func sanitizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" || strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("objectstore: invalid object_key")
	}
	return key, nil
}
