package handler_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/toontoonz/logistic-go/backend/internal/config"
	"github.com/toontoonz/logistic-go/backend/internal/handler"
	"github.com/toontoonz/logistic-go/backend/internal/model"
	"github.com/toontoonz/logistic-go/backend/internal/server"
	"github.com/toontoonz/logistic-go/backend/internal/service"
)

type trail struct{}

func (trail) Insert(context.Context, model.Position) error { return nil }
func (trail) Latest(context.Context, string) (model.Position, error) {
	return model.Position{}, model.ErrNotFound
}
func (trail) LatestAll(context.Context) ([]model.Position, error) { return nil, nil }
func (trail) History(context.Context, string, int) ([]model.Position, error) {
	return nil, nil
}

type noopCache struct{}

func (noopCache) Set(context.Context, model.Position) error { return nil }
func (noopCache) Get(context.Context, string) (model.Position, error) {
	return model.Position{}, model.ErrNotFound
}
func (noopCache) GetMany(context.Context, []string) (map[string]model.Position, error) {
	return map[string]model.Position{}, nil
}
func (noopCache) Remember(context.Context, string) error      { return nil }
func (noopCache) ActiveIDs(context.Context) ([]string, error) { return nil, nil }

type noDrivers struct{}

func (noDrivers) Get(context.Context, string) (model.Driver, error) {
	return model.Driver{}, model.ErrNotFound
}

type noBus struct{}

func (noBus) Broadcast([]byte) {}

func TestPostLocationRejectsImpossibleLatitude(t *testing.T) {
	locations := service.NewLocationService(trail{}, noopCache{}, noDrivers{}, noBus{})
	app := server.New(config.Config{CORSOrigins: []string{"http://localhost:3000"}}, handler.API{
		Locations: locations,
	})

	req, err := http.NewRequest(http.MethodPost, "/api/drivers/drv-1/location", strings.NewReader(`{"lat":120,"lng":100}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 2000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "latitude") {
		t.Fatalf("body = %s", body)
	}
}
