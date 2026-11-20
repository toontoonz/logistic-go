// Command server is the Lane API: orders, drivers, and live GPS.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toontoonz/logistic-go/backend/internal/cache"
	"github.com/toontoonz/logistic-go/backend/internal/config"
	"github.com/toontoonz/logistic-go/backend/internal/demo"
	"github.com/toontoonz/logistic-go/backend/internal/handler"
	"github.com/toontoonz/logistic-go/backend/internal/repository"
	"github.com/toontoonz/logistic-go/backend/internal/server"
	"github.com/toontoonz/logistic-go/backend/internal/service"
	"github.com/toontoonz/logistic-go/backend/internal/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// NotifyContext cancels ctx when the process receives Ctrl+C or
	// SIGTERM (what Docker sends on `compose stop`). Passing ctx into
	// database calls lets those calls abort instead of running on
	// after the process has been asked to leave.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := repository.Migrate(ctx, pool, cfg.MigrationsDir); err != nil {
		return err
	}

	positions, err := cache.New(ctx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer positions.Close()

	driverRepo := repository.NewDriverRepo(pool)
	orderRepo := repository.NewOrderRepo(pool)
	locationRepo := repository.NewLocationRepo(pool)

	hub := ws.NewHub()
	// The hub keeps running until HTTP shutdown has unregistered
	// sockets. Cancelling it with the signal context would deadlock:
	// handlers would block forever trying to send on a hub that has
	// already exited.
	hubCtx, hubCancel := context.WithCancel(context.Background())
	defer hubCancel()
	go hub.Run(hubCtx)

	locations := service.NewLocationService(locationRepo, positions, driverRepo, hub)
	if err := locations.Warm(ctx); err != nil {
		return fmt.Errorf("warm position cache: %w", err)
	}
	drivers := service.NewDriverService(driverRepo, locations)
	orders := service.NewOrderService(orderRepo, driverRepo, locations)
	simulator := demo.New(locations)

	app := server.New(cfg, handler.API{
		Orders:    orders,
		Drivers:   drivers,
		Locations: locations,
		Hub:       hub,
		Demo:      simulator,
		DB:        pool,
		Cache:     positions,
	})

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		errCh <- app.Listen(cfg.HTTPAddr)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen: %w", err)
		}
		return errors.New("server stopped")
	}

	simulator.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
