package handler

import (
	"context"

	"github.com/toontoonz/logistic-go/backend/internal/demo"
	"github.com/toontoonz/logistic-go/backend/internal/service"
	"github.com/toontoonz/logistic-go/backend/internal/ws"
)

// Pinger is implemented by both the Postgres pool and the Redis cache.
type Pinger interface {
	Ping(ctx context.Context) error
}

// API groups the handlers. Services are interfaces' concrete values;
// the handler does not know whether Postgres or a fake is behind them.
type API struct {
	Orders    *service.OrderService
	Drivers   *service.DriverService
	Locations *service.LocationService
	Hub       *ws.Hub
	Demo      *demo.Simulator
	DB        Pinger
	Cache     Pinger
}
