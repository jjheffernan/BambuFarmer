package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Bus publishes and consumes Redis Streams envelopes.
type Bus struct {
	rdb *redis.Client
}

// Dial opens a Redis client from REDIS_URL (redis://host:6379/0).
func Dial(ctx context.Context, redisURL string) (*Bus, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is empty")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Bus{rdb: rdb}, nil
}

// Close closes the underlying client.
func (b *Bus) Close() error {
	if b == nil || b.rdb == nil {
		return nil
	}
	return b.rdb.Close()
}

// Ping checks connectivity.
func (b *Bus) Ping(ctx context.Context) error {
	return b.rdb.Ping(ctx).Err()
}

// Publish appends an envelope to a stream (maxlen ~10k approximate).
func (b *Bus) Publish(ctx context.Context, stream string, env Envelope) (string, error) {
	if env.TS.IsZero() {
		env.TS = time.Now().UTC()
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return "", err
	}
	id, err := b.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10_000,
		Approx: true,
		Values: map[string]any{
			"type":    env.Type,
			"payload": string(raw),
		},
	}).Result()
	if err != nil {
		return "", err
	}
	return id, nil
}

// Handler processes one envelope; return error to skip XACK (retry later).
type Handler func(ctx context.Context, id string, env Envelope) error

// ConsumeLoop reads a consumer group forever until ctx cancel.
// Creates the group with MkStream if missing (start from '$' for new groups — pass "0" via startID for history).
func (b *Bus) ConsumeLoop(ctx context.Context, stream, group, consumer, startID string, h Handler) error {
	if startID == "" {
		startID = "$"
	}
	err := b.rdb.XGroupCreateMkStream(ctx, stream, group, startID).Err()
	if err != nil && !isBusyGroup(err) {
		return fmt.Errorf("xgroup create %s/%s: %w", stream, group, err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		streams, err := b.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    32,
			Block:    2 * time.Second,
		}).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// brief backoff on transient errors
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}
		for _, s := range streams {
			for _, msg := range s.Messages {
				env, err := decodeMessage(msg.Values)
				if err != nil {
					_ = b.rdb.XAck(ctx, stream, group, msg.ID).Err()
					continue
				}
				if err := h(ctx, msg.ID, env); err != nil {
					continue // leave pending for retry
				}
				_ = b.rdb.XAck(ctx, stream, group, msg.ID).Err()
			}
		}
	}
}

func decodeMessage(values map[string]any) (Envelope, error) {
	var env Envelope
	raw, _ := values["payload"].(string)
	if raw == "" {
		return env, fmt.Errorf("missing payload")
	}
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return env, err
	}
	return env, nil
}

func isBusyGroup(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return len(s) >= 9 && s[:9] == "BUSYGROUP"
}

// NewEnvelope builds a typed envelope with UTC timestamp.
func NewEnvelope(typ, source, siteID string, payload map[string]any) Envelope {
	if payload == nil {
		payload = map[string]any{}
	}
	return Envelope{
		Type:    typ,
		Source:  source,
		SiteID:  siteID,
		TS:      time.Now().UTC(),
		Payload: payload,
	}
}

// Consumer groups (DESIGN_SPEC §15).
const (
	GroupCore         = "core"
	GroupAnalytics    = "analytics"
	GroupIntegrations = "integrations"
	GroupGatewayFanout = "gateway-fanout"
)
