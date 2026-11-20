package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// DriverStore persists drivers.
type DriverStore interface {
	Create(ctx context.Context, d model.Driver) error
	List(ctx context.Context) ([]model.Driver, error)
	Get(ctx context.Context, id string) (model.Driver, error)
}

// PositionLookup attaches latest fixes to driver records.
type PositionLookup interface {
	ForDrivers(ctx context.Context, ids []string) (map[string]model.Position, error)
	Latest(ctx context.Context, driverID string) (model.Position, error)
}

// DriverService is the use-case layer for drivers.
type DriverService struct {
	store  DriverStore
	locate PositionLookup
}

// NewDriverService builds a DriverService. locate may be nil in tests
// that never read a position.
func NewDriverService(store DriverStore, locate PositionLookup) *DriverService {
	return &DriverService{store: store, locate: locate}
}

// CreateDriverInput is the body of POST /api/drivers.
type CreateDriverInput struct {
	Name    string
	Phone   string
	Vehicle string
}

// Create registers a driver.
func (s *DriverService) Create(ctx context.Context, in CreateDriverInput) (model.Driver, error) {
	var err error
	if in.Name, err = requireText("name", in.Name, 80); err != nil {
		return model.Driver{}, err
	}
	if in.Phone, err = validatePhone(in.Phone); err != nil {
		return model.Driver{}, err
	}
	in.Vehicle, err = trimOptional(in.Vehicle, 80, "vehicle")
	if err != nil {
		return model.Driver{}, err
	}
	driver := model.Driver{
		ID:        uuid.NewString(),
		Name:      in.Name,
		Phone:     in.Phone,
		Vehicle:   in.Vehicle,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.Create(ctx, driver); err != nil {
		return model.Driver{}, err
	}
	return driver, nil
}

// List returns drivers with their latest position filled in when we have one.
func (s *DriverService) List(ctx context.Context) ([]model.Driver, error) {
	drivers, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	if drivers == nil {
		drivers = []model.Driver{}
	}
	if s.locate == nil || len(drivers) == 0 {
		return drivers, nil
	}
	ids := make([]string, len(drivers))
	for i, d := range drivers {
		ids[i] = d.ID
	}
	positions, err := s.locate.ForDrivers(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range drivers {
		if p, ok := positions[drivers[i].ID]; ok {
			drivers[i].Location = &p
		}
	}
	return drivers, nil
}

// Get returns one driver, plus a latest fix when one exists.
func (s *DriverService) Get(ctx context.Context, id string) (model.Driver, error) {
	driver, err := s.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Driver{}, notFound("driver")
		}
		return model.Driver{}, err
	}
	if s.locate == nil {
		return driver, nil
	}
	pos, err := s.locate.Latest(ctx, id)
	if err == nil {
		driver.Location = &pos
		return driver, nil
	}
	if errors.Is(err, model.ErrNotFound) {
		return driver, nil
	}
	return model.Driver{}, err
}

func trimOptional(value string, max int, field string) (string, error) {
	value = trim(value)
	if value == "" {
		return "", nil
	}
	return requireText(field, value, max)
}
