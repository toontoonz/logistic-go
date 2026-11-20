package service

import (
	"context"
	"errors"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/toontoonz/logistic-go/backend/internal/model"
	"github.com/toontoonz/logistic-go/backend/internal/ws"
)

type memTrail struct {
	points []model.Position
}

func (m *memTrail) Insert(_ context.Context, p model.Position) error {
	m.points = append(m.points, p)
	return nil
}
func (m *memTrail) Latest(context.Context, string) (model.Position, error) {
	if len(m.points) == 0 {
		return model.Position{}, model.ErrNotFound
	}
	return m.points[len(m.points)-1], nil
}
func (m *memTrail) LatestAll(context.Context) ([]model.Position, error) {
	return m.points, nil
}
func (m *memTrail) History(context.Context, string, int) ([]model.Position, error) {
	return m.points, nil
}

type memCache struct {
	pos    map[string]model.Position
	setErr error
	active []string
}

func (m *memCache) Set(_ context.Context, p model.Position) error {
	if m.setErr != nil {
		return m.setErr
	}
	if m.pos == nil {
		m.pos = map[string]model.Position{}
	}
	m.pos[p.DriverID] = p
	return nil
}
func (m *memCache) Get(_ context.Context, id string) (model.Position, error) {
	p, ok := m.pos[id]
	if !ok {
		return model.Position{}, model.ErrNotFound
	}
	return p, nil
}
func (m *memCache) GetMany(_ context.Context, ids []string) (map[string]model.Position, error) {
	out := map[string]model.Position{}
	for _, id := range ids {
		if p, ok := m.pos[id]; ok {
			out[id] = p
		}
	}
	return out, nil
}
func (m *memCache) Remember(_ context.Context, id string) error {
	m.active = append(m.active, id)
	return nil
}
func (m *memCache) ActiveIDs(context.Context) ([]string, error) { return m.active, nil }

type memBus struct {
	msgs [][]byte
}

func (m *memBus) Broadcast(msg []byte) {
	m.msgs = append(m.msgs, append([]byte(nil), msg...))
}

func TestIngestRejectsBadCoordinate(t *testing.T) {
	trail := &memTrail{}
	svc := NewLocationService(trail, &memCache{}, memDrivers{drivers: map[string]model.Driver{
		"drv-1": {ID: "drv-1", Name: "Mali"},
	}}, &memBus{})
	if _, err := svc.Ingest(context.Background(), "drv-1", 120, 10); err == nil {
		t.Fatal("expected validation error")
	}
	if len(trail.points) != 0 {
		t.Fatal("trail was written for an invalid coordinate")
	}
}

func TestIngestStoresAndBroadcastsWhenCacheFails(t *testing.T) {
	trail := &memTrail{}
	bus := &memBus{}
	cache := &memCache{setErr: errors.New("redis down")}
	svc := NewLocationService(trail, cache, memDrivers{drivers: map[string]model.Driver{
		"drv-1": {ID: "drv-1", Name: "Mali"},
	}}, bus)

	pos, err := svc.Ingest(context.Background(), "drv-1", 13.73, 100.54)
	if err != nil {
		t.Fatal(err)
	}
	if pos.DriverName != "Mali" || len(trail.points) != 1 {
		t.Fatalf("pos %+v trail %d", pos, len(trail.points))
	}
	if len(cache.active) != 0 {
		t.Fatal("remember should be skipped when set fails")
	}
	if len(bus.msgs) != 1 {
		t.Fatalf("broadcasts = %d", len(bus.msgs))
	}
	var msg ws.PositionMessage
	if err := json.Unmarshal(bus.msgs[0], &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "position" || msg.Lat != 13.73 || msg.DriverName != "Mali" {
		t.Fatalf("message = %+v", msg)
	}
}

func TestIngestUnknownDriver(t *testing.T) {
	svc := NewLocationService(&memTrail{}, &memCache{}, memDrivers{drivers: map[string]model.Driver{}}, &memBus{})
	_, err := svc.Ingest(context.Background(), "missing", 13, 100)
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("err = %v", err)
	}
}
