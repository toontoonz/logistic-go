package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// OrderRepo persists parcel orders.
type OrderRepo struct {
	pool *pgxpool.Pool
}

// NewOrderRepo builds an OrderRepo.
func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

const orderSelect = `
	SELECT o.id, o.customer_name, o.pickup_address, o.dropoff_address,
	       o.pickup_lat, o.pickup_lng, o.dropoff_lat, o.dropoff_lng,
	       o.status, o.driver_id, d.name, o.created_at, o.updated_at
	FROM orders o
	LEFT JOIN drivers d ON d.id = o.driver_id`

// Create inserts an order.
func (r *OrderRepo) Create(ctx context.Context, o model.Order) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO orders (
			id, customer_name, pickup_address, dropoff_address,
			pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
			status, driver_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		o.ID, o.CustomerName, o.PickupAddress, o.DropoffAddress,
		o.PickupLat, o.PickupLng, o.DropoffLat, o.DropoffLng,
		o.Status, o.DriverID, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

// List returns orders newest first.
func (r *OrderRepo) List(ctx context.Context) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, orderSelect+` ORDER BY o.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// Get loads one order.
func (r *OrderRepo) Get(ctx context.Context, id string) (model.Order, error) {
	rows, err := r.pool.Query(ctx, orderSelect+` WHERE o.id = $1`, id)
	if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return model.Order{}, err
		}
		return model.Order{}, model.ErrNotFound
	}
	return scanOrder(rows)
}

// Assign attaches a driver. Pending orders become "assigned".
// Already-assigned orders can change driver until pickup.
func (r *OrderRepo) Assign(ctx context.Context, orderID, driverID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders
		SET driver_id = $2,
		    status = CASE WHEN status = 'pending' THEN 'assigned' ELSE status END,
		    updated_at = now()
		WHERE id = $1
		  AND status IN ('pending', 'assigned')`, orderID, driverID)
	if err != nil {
		return fmt.Errorf("assign order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// UpdateStatus writes a new status. The service has already checked
// that the transition is legal.
func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID, status string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders SET status = $2, updated_at = now() WHERE id = $1`, orderID, status)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (model.Order, error) {
	var o model.Order
	err := row.Scan(
		&o.ID, &o.CustomerName, &o.PickupAddress, &o.DropoffAddress,
		&o.PickupLat, &o.PickupLng, &o.DropoffLat, &o.DropoffLng,
		&o.Status, &o.DriverID, &o.DriverName, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Order{}, model.ErrNotFound
	}
	if err != nil {
		return model.Order{}, fmt.Errorf("scan order: %w", err)
	}
	return o, nil
}
