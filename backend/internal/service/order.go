package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// OrderStore is the persistence the order service needs.
//
// This is an interface, not a concrete Postgres type. Any value with
// these methods satisfies it — Go has no "implements" keyword. Tests
// pass an in-memory fake; main passes *repository.OrderRepo.
type OrderStore interface {
	Create(ctx context.Context, o model.Order) error
	List(ctx context.Context) ([]model.Order, error)
	Get(ctx context.Context, id string) (model.Order, error)
	Assign(ctx context.Context, orderID, driverID string) error
	UpdateStatus(ctx context.Context, orderID, status string) error
}

// DriverReader loads a driver. Location ingest and assignment both need it.
type DriverReader interface {
	Get(ctx context.Context, id string) (model.Driver, error)
}

// Locator reads the latest fix for a driver (Redis, then Postgres).
type Locator interface {
	Latest(ctx context.Context, driverID string) (model.Position, error)
}

// OrderService is the use-case layer for parcels.
type OrderService struct {
	orders  OrderStore
	drivers DriverReader
	locate  Locator
}

// NewOrderService builds an OrderService.
func NewOrderService(orders OrderStore, drivers DriverReader, locate Locator) *OrderService {
	return &OrderService{orders: orders, drivers: drivers, locate: locate}
}

// CreateOrderInput is the validated shape of POST /api/orders.
type CreateOrderInput struct {
	CustomerName   string
	PickupAddress  string
	DropoffAddress string
	PickupLat      float64
	PickupLng      float64
	DropoffLat     float64
	DropoffLng     float64
}

// Create stores a pending order.
func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (model.Order, error) {
	var err error
	if in.CustomerName, err = requireText("customerName", in.CustomerName, 80); err != nil {
		return model.Order{}, err
	}
	if in.PickupAddress, err = requireText("pickupAddress", in.PickupAddress, 200); err != nil {
		return model.Order{}, err
	}
	if in.DropoffAddress, err = requireText("dropoffAddress", in.DropoffAddress, 200); err != nil {
		return model.Order{}, err
	}
	if err = ValidateCoord(in.PickupLat, in.PickupLng); err != nil {
		return model.Order{}, err
	}
	if err = ValidateCoord(in.DropoffLat, in.DropoffLng); err != nil {
		return model.Order{}, err
	}

	now := time.Now().UTC()
	order := model.Order{
		ID:             uuid.NewString(),
		CustomerName:   in.CustomerName,
		PickupAddress:  in.PickupAddress,
		DropoffAddress: in.DropoffAddress,
		PickupLat:      in.PickupLat,
		PickupLng:      in.PickupLng,
		DropoffLat:     in.DropoffLat,
		DropoffLng:     in.DropoffLng,
		Status:         StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return model.Order{}, err
	}
	return order, nil
}

// List returns every order, newest first (the store chooses the order).
func (s *OrderService) List(ctx context.Context) ([]model.Order, error) {
	orders, err := s.orders.List(ctx)
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

// Get returns one order and, when a driver is assigned, their latest fix.
func (s *OrderService) Get(ctx context.Context, id string) (model.OrderDetail, error) {
	order, err := s.orders.Get(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.OrderDetail{}, notFound("order")
		}
		return model.OrderDetail{}, err
	}
	detail := model.OrderDetail{Order: order}
	if order.DriverID == nil || s.locate == nil {
		return detail, nil
	}
	pos, err := s.locate.Latest(ctx, *order.DriverID)
	if err == nil {
		detail.DriverLocation = &pos
		return detail, nil
	}
	if errors.Is(err, model.ErrNotFound) {
		return detail, nil
	}
	return model.OrderDetail{}, fmt.Errorf("driver location: %w", err)
}

// Assign attaches a driver while the parcel is still at the depot.
func (s *OrderService) Assign(ctx context.Context, orderID, driverID string) (model.Order, error) {
	driverID = trim(driverID)
	if driverID == "" {
		return model.Order{}, &InputError{Message: "driverId is required"}
	}
	order, err := s.orders.Get(ctx, orderID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Order{}, notFound("order")
		}
		return model.Order{}, err
	}
	if order.Status != StatusPending && order.Status != StatusAssigned {
		return model.Order{}, &ConflictError{Message: "driver can only be assigned before pickup"}
	}
	if _, err := s.drivers.Get(ctx, driverID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Order{}, notFound("driver")
		}
		return model.Order{}, err
	}
	if err := s.orders.Assign(ctx, orderID, driverID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Order{}, &ConflictError{Message: "order can no longer be assigned"}
		}
		return model.Order{}, err
	}
	updated, err := s.orders.Get(ctx, orderID)
	if err != nil {
		return model.Order{}, err
	}
	return updated, nil
}

// UpdateStatus moves an order one legal step forward.
func (s *OrderService) UpdateStatus(ctx context.Context, orderID, status string) (model.Order, error) {
	order, err := s.orders.Get(ctx, orderID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Order{}, notFound("order")
		}
		return model.Order{}, err
	}
	if !CanTransition(order.Status, status) {
		return model.Order{}, &ConflictError{Message: fmt.Sprintf("cannot change status from %s to %s", order.Status, status)}
	}
	if err := s.orders.UpdateStatus(ctx, orderID, status); err != nil {
		return model.Order{}, err
	}
	order.Status = status
	order.UpdatedAt = time.Now().UTC()
	return order, nil
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
