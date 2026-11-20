package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	json "github.com/goccy/go-json"
	"github.com/toontoonz/logistic-go/backend/internal/model"
	"github.com/toontoonz/logistic-go/backend/internal/ws"
)

// LocationStore is the PostGIS trail.
type LocationStore interface {
	Insert(ctx context.Context, p model.Position) error
	Latest(ctx context.Context, driverID string) (model.Position, error)
	LatestAll(ctx context.Context) ([]model.Position, error)
	History(ctx context.Context, driverID string, limit int) ([]model.Position, error)
}

// PositionCache is the Redis "where are they right now" store.
type PositionCache interface {
	Set(ctx context.Context, p model.Position) error
	Get(ctx context.Context, driverID string) (model.Position, error)
	GetMany(ctx context.Context, ids []string) (map[string]model.Position, error)
	Remember(ctx context.Context, driverID string) error
	ActiveIDs(ctx context.Context) ([]string, error)
}

// Broadcaster pushes a JSON payload to every connected map.
type Broadcaster interface {
	Broadcast(msg []byte)
}

// LocationService accepts GPS pings and reads them back.
type LocationService struct {
	store   LocationStore
	cache   PositionCache
	drivers DriverReader
	bus     Broadcaster
}

// NewLocationService builds a LocationService.
func NewLocationService(store LocationStore, cache PositionCache, drivers DriverReader, bus Broadcaster) *LocationService {
	return &LocationService{store: store, cache: cache, drivers: drivers, bus: bus}
}

// Ingest records one GPS sample, caches it, and tells live maps.
func (s *LocationService) Ingest(ctx context.Context, driverID string, lat, lng float64) (model.Position, error) {
	if err := ValidateCoord(lat, lng); err != nil {
		return model.Position{}, err
	}
	driver, err := s.drivers.Get(ctx, driverID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Position{}, notFound("driver")
		}
		return model.Position{}, err
	}

	pos := model.Position{
		DriverID:   driver.ID,
		DriverName: driver.Name,
		Lat:        lat,
		Lng:        lng,
		RecordedAt: time.Now().UTC(),
	}
	// The trail in Postgres is the source of truth. Redis and the
	// websocket are best-effort so a cache blip does not lose history.
	if err := s.store.Insert(ctx, pos); err != nil {
		return model.Position{}, err
	}
	if err := s.cache.Set(ctx, pos); err != nil {
		slog.Error("cache position", "err", err, "driver", driver.ID)
	} else if err := s.cache.Remember(ctx, driver.ID); err != nil {
		slog.Error("remember driver", "err", err, "driver", driver.ID)
	}
	if s.bus != nil {
		body, err := json.Marshal(ws.PositionMessage{Type: "position", Position: pos})
		if err != nil {
			slog.Error("encode position", "err", err)
		} else {
			s.bus.Broadcast(body)
		}
	}
	return pos, nil
}

// Latest prefers Redis and falls back to the trail.
func (s *LocationService) Latest(ctx context.Context, driverID string) (model.Position, error) {
	pos, err := s.cache.Get(ctx, driverID)
	if err == nil {
		return pos, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		slog.Error("redis get", "err", err, "driver", driverID)
	}
	return s.store.Latest(ctx, driverID)
}

// ForDrivers loads latest fixes for a set of drivers, filling cache misses
// from Postgres.
func (s *LocationService) ForDrivers(ctx context.Context, ids []string) (map[string]model.Position, error) {
	many, err := s.cache.GetMany(ctx, ids)
	if err != nil || many == nil {
		if err != nil {
			slog.Error("redis mget", "err", err)
		}
		many = map[string]model.Position{}
	}
	if len(many) == len(ids) {
		return many, nil
	}
	all, dbErr := s.store.LatestAll(ctx)
	if dbErr != nil {
		if len(many) > 0 {
			return many, nil
		}
		return nil, dbErr
	}
	want := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	for _, p := range all {
		if _, ok := want[p.DriverID]; !ok {
			continue
		}
		if _, ok := many[p.DriverID]; !ok {
			many[p.DriverID] = p
		}
	}
	return many, nil
}

// Snapshot is what a map receives the moment it connects.
func (s *LocationService) Snapshot(ctx context.Context) ([]model.Position, error) {
	ids, err := s.cache.ActiveIDs(ctx)
	if err == nil && len(ids) > 0 {
		many, mErr := s.cache.GetMany(ctx, ids)
		if mErr == nil && len(many) > 0 {
			out := make([]model.Position, 0, len(many))
			for _, p := range many {
				out = append(out, p)
			}
			return out, nil
		}
	}
	all, err := s.store.LatestAll(ctx)
	if err != nil {
		return nil, err
	}
	if all == nil {
		all = []model.Position{}
	}
	return all, nil
}

// History returns the trail for one driver, oldest first.
func (s *LocationService) History(ctx context.Context, driverID string, limit int) ([]model.Position, error) {
	if _, err := s.drivers.Get(ctx, driverID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, notFound("driver")
		}
		return nil, err
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	points, err := s.store.History(ctx, driverID, limit)
	if err != nil {
		return nil, err
	}
	if points == nil {
		points = []model.Position{}
	}
	return points, nil
}

// Warm copies the newest database fix for each driver into Redis.
// Called once at startup so a Redis restart does not blank the map.
func (s *LocationService) Warm(ctx context.Context) error {
	all, err := s.store.LatestAll(ctx)
	if err != nil {
		return err
	}
	for _, p := range all {
		if err := s.cache.Set(ctx, p); err != nil {
			return err
		}
		if err := s.cache.Remember(ctx, p.DriverID); err != nil {
			return err
		}
	}
	return nil
}
