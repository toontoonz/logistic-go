package service

import (
	"context"
	"testing"
	"time"

	"github.com/toontoonz/logistic-go/backend/internal/model"
)

type memOrders struct {
	byID map[string]model.Order
}

func (m *memOrders) Create(_ context.Context, o model.Order) error {
	if m.byID == nil {
		m.byID = map[string]model.Order{}
	}
	m.byID[o.ID] = o
	return nil
}

func (m *memOrders) List(context.Context) ([]model.Order, error) {
	out := make([]model.Order, 0, len(m.byID))
	for _, o := range m.byID {
		out = append(out, o)
	}
	return out, nil
}

func (m *memOrders) Get(_ context.Context, id string) (model.Order, error) {
	o, ok := m.byID[id]
	if !ok {
		return model.Order{}, model.ErrNotFound
	}
	return o, nil
}

func (m *memOrders) Assign(_ context.Context, orderID, driverID string) error {
	o, ok := m.byID[orderID]
	if !ok || (o.Status != StatusPending && o.Status != StatusAssigned) {
		return model.ErrNotFound
	}
	o.DriverID = &driverID
	name := "Mali"
	o.DriverName = &name
	if o.Status == StatusPending {
		o.Status = StatusAssigned
	}
	o.UpdatedAt = time.Now().UTC()
	m.byID[orderID] = o
	return nil
}

func (m *memOrders) UpdateStatus(_ context.Context, orderID, status string) error {
	o, ok := m.byID[orderID]
	if !ok {
		return model.ErrNotFound
	}
	o.Status = status
	m.byID[orderID] = o
	return nil
}

type memDrivers struct {
	drivers map[string]model.Driver
}

func (m memDrivers) Get(_ context.Context, id string) (model.Driver, error) {
	d, ok := m.drivers[id]
	if !ok {
		return model.Driver{}, model.ErrNotFound
	}
	return d, nil
}

type memLocate struct {
	pos model.Position
	err error
}

func (m memLocate) Latest(context.Context, string) (model.Position, error) {
	if m.err != nil {
		return model.Position{}, m.err
	}
	return m.pos, nil
}

func TestCreateOrderPending(t *testing.T) {
	store := &memOrders{}
	svc := NewOrderService(store, memDrivers{}, memLocate{err: model.ErrNotFound})
	order, err := svc.Create(context.Background(), CreateOrderInput{
		CustomerName:   "Suda",
		PickupAddress:  "Lumphini Park",
		DropoffAddress: "Thong Lo",
		PickupLat:      13.73,
		PickupLng:      100.54,
		DropoffLat:     13.72,
		DropoffLng:     100.58,
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != StatusPending || order.ID == "" {
		t.Fatalf("unexpected order: %+v", order)
	}
	if _, err := svc.Create(context.Background(), CreateOrderInput{CustomerName: "Suda"}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAssignAndAdvance(t *testing.T) {
	store := &memOrders{}
	drivers := memDrivers{drivers: map[string]model.Driver{
		"drv-1": {ID: "drv-1", Name: "Mali"},
	}}
	svc := NewOrderService(store, drivers, memLocate{err: model.ErrNotFound})
	order, err := svc.Create(context.Background(), CreateOrderInput{
		CustomerName: "Suda", PickupAddress: "A", DropoffAddress: "B",
		PickupLat: 1, PickupLng: 2, DropoffLat: 3, DropoffLng: 4,
	})
	if err != nil {
		t.Fatal(err)
	}

	assigned, err := svc.Assign(context.Background(), order.ID, "drv-1")
	if err != nil {
		t.Fatal(err)
	}
	if assigned.Status != StatusAssigned || assigned.DriverID == nil || *assigned.DriverID != "drv-1" {
		t.Fatalf("assign result: %+v", assigned)
	}

	if _, err := svc.Assign(context.Background(), order.ID, "missing"); err == nil {
		t.Fatal("expected missing driver")
	}

	picked, err := svc.UpdateStatus(context.Background(), order.ID, StatusPickedUp)
	if err != nil {
		t.Fatal(err)
	}
	if picked.Status != StatusPickedUp {
		t.Fatalf("status = %s", picked.Status)
	}
	if _, err := svc.UpdateStatus(context.Background(), order.ID, StatusDelivered); err == nil {
		t.Fatal("expected skip to be rejected")
	}
	if _, err := svc.Assign(context.Background(), order.ID, "drv-1"); err == nil {
		t.Fatal("expected assign after pickup to fail")
	}
}

func TestGetIncludesDriverLocation(t *testing.T) {
	id := "ord-1"
	driverID := "drv-1"
	store := &memOrders{byID: map[string]model.Order{
		id: {ID: id, Status: StatusInTransit, DriverID: &driverID, CustomerName: "Suda"},
	}}
	pos := model.Position{DriverID: driverID, Lat: 13.7, Lng: 100.5}
	svc := NewOrderService(store, memDrivers{}, memLocate{pos: pos})
	detail, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if detail.DriverLocation == nil || detail.DriverLocation.Lat != 13.7 {
		t.Fatalf("location = %+v", detail.DriverLocation)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("expected not found")
	}
}
