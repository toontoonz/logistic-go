// Package model holds the data shapes shared by HTTP, SQL, and Redis.
// Identifiers are plain strings (UUID text). That avoids a Postgres UUID
// codec while the project is still small.
package model

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a row or cache entry does not exist.
var ErrNotFound = errors.New("not found")

// Driver is a person who can be assigned orders and publish GPS points.
type Driver struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Vehicle   string    `json:"vehicle"`
	CreatedAt time.Time `json:"createdAt"`
	Location  *Position `json:"location,omitempty"`
}

// Position is one GPS sample. DriverName is denormalized so a map
// marker can render without a second lookup.
type Position struct {
	DriverID   string    `json:"driverId"`
	DriverName string    `json:"driverName,omitempty"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	RecordedAt time.Time `json:"recordedAt"`
}

// Order is a parcel moving from a pickup point to a dropoff point.
type Order struct {
	ID             string    `json:"id"`
	CustomerName   string    `json:"customerName"`
	PickupAddress  string    `json:"pickupAddress"`
	DropoffAddress string    `json:"dropoffAddress"`
	PickupLat      float64   `json:"pickupLat"`
	PickupLng      float64   `json:"pickupLng"`
	DropoffLat     float64   `json:"dropoffLat"`
	DropoffLng     float64   `json:"dropoffLng"`
	Status         string    `json:"status"`
	DriverID       *string   `json:"driverId,omitempty"`
	DriverName     *string   `json:"driverName,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// OrderDetail is an order plus the assigned driver's latest fix, if any.
type OrderDetail struct {
	Order
	DriverLocation *Position `json:"driverLocation,omitempty"`
}
