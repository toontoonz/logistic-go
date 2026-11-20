package demo

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/toontoonz/logistic-go/backend/internal/service"
)

const (
	tick      = 700 * time.Millisecond
	demoKmh   = 480.0
	maxMisses = 5
)

// Simulator posts GPS from a goroutine so the browser can leave the
// simulator page. A tab timer would stop as soon as that page unmounted.
type Simulator struct {
	locations *service.LocationService

	mu       sync.Mutex
	cancel   context.CancelFunc
	driverID string
}

// New builds a simulator. Nothing moves until Start.
func New(locations *service.LocationService) *Simulator {
	return &Simulator{locations: locations}
}

// Status reports whether a route is currently being published.
type Status struct {
	Running  bool   `json:"running"`
	DriverID string `json:"driverId,omitempty"`
}

// Status returns the current run, if any.
func (s *Simulator) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Status{Running: s.cancel != nil, DriverID: s.driverID}
}

// Start begins posting points for driverID. Any previous run is stopped.
// The first point is written before this returns, so a missing driver
// comes back as an error instead of a silent goroutine.
func (s *Simulator) Start(driverID string) error {
	first := bangkokRoute[0]
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := s.locations.Ingest(ctx, driverID, first.lat, first.lng)
	cancel()
	if err != nil {
		return err
	}

	s.Stop()

	runCtx, runCancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancel = runCancel
	s.driverID = driverID
	s.mu.Unlock()

	// This goroutine outlives the HTTP request. Stop or process
	// shutdown cancels runCtx, and the select below leaves the loop.
	go s.loop(runCtx, driverID)
	return nil
}

// Stop ends the current run. Safe to call when nothing is running.
func (s *Simulator) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.driverID = ""
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *Simulator) loop(ctx context.Context, driverID string) {
	total := pathLength(bangkokRoute)
	step := demoKmh * 1000 / 3600 * tick.Seconds()
	dist := step
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	misses := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if dist >= total {
				dist = 0
			}
			point := pointAlong(bangkokRoute, dist)
			dist += step
			reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := s.locations.Ingest(reqCtx, driverID, point.lat, point.lng)
			cancel()
			if err != nil {
				misses++
				slog.Error("demo ping", "err", err, "driver", driverID)
				if misses >= maxMisses {
					s.Stop()
					return
				}
				continue
			}
			misses = 0
		}
	}
}
