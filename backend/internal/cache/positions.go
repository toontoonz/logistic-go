// Package cache keeps each driver's latest GPS fix in Redis.
// Postgres keeps the full trail; Redis answers "where are they right now?"
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	json "github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/toontoonz/logistic-go/backend/internal/boot"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

const activeSet = "drivers:live"

// Positions is the Redis-backed latest-position cache.
type Positions struct {
	client *redis.Client
}

// New dials Redis, retrying while the container starts.
func New(ctx context.Context, redisURL string) (*Positions, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)

	err = boot.UntilReady(ctx, "redis", 40, time.Second, func(pingCtx context.Context) error {
		return client.Ping(pingCtx).Err()
	})
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Positions{client: client}, nil
}

// Close releases the connection pool.
func (p *Positions) Close() error {
	return p.client.Close()
}

// Ping checks that Redis is reachable.
func (p *Positions) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func posKey(driverID string) string {
	return "driver:pos:" + driverID
}

// Set stores the latest fix. No TTL: the value is overwritten by the
// next ping and should survive until then.
func (p *Positions) Set(ctx context.Context, pos model.Position) error {
	body, err := json.Marshal(pos)
	if err != nil {
		return fmt.Errorf("encode position: %w", err)
	}
	if err := p.client.Set(ctx, posKey(pos.DriverID), body, 0).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// Remember adds the driver to the set of drivers we have a fix for.
func (p *Positions) Remember(ctx context.Context, driverID string) error {
	if err := p.client.SAdd(ctx, activeSet, driverID).Err(); err != nil {
		return fmt.Errorf("redis sadd: %w", err)
	}
	return nil
}

// Get reads one latest fix.
func (p *Positions) Get(ctx context.Context, driverID string) (model.Position, error) {
	body, err := p.client.Get(ctx, posKey(driverID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return model.Position{}, model.ErrNotFound
	}
	if err != nil {
		return model.Position{}, fmt.Errorf("redis get: %w", err)
	}
	var pos model.Position
	if err := json.Unmarshal(body, &pos); err != nil {
		return model.Position{}, fmt.Errorf("decode position: %w", err)
	}
	return pos, nil
}

// GetMany reads many fixes in one round trip.
func (p *Positions) GetMany(ctx context.Context, ids []string) (map[string]model.Position, error) {
	out := make(map[string]model.Position, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = posKey(id)
	}
	values, err := p.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget: %w", err)
	}
	for _, value := range values {
		if value == nil {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		var pos model.Position
		if err := json.Unmarshal([]byte(text), &pos); err != nil {
			continue
		}
		out[pos.DriverID] = pos
	}
	return out, nil
}

// ActiveIDs lists drivers that have published at least one fix.
func (p *Positions) ActiveIDs(ctx context.Context) ([]string, error) {
	ids, err := p.client.SMembers(ctx, activeSet).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers: %w", err)
	}
	return ids, nil
}
