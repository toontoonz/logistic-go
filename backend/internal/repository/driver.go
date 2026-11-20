package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// DriverRepo persists drivers.
type DriverRepo struct {
	pool *pgxpool.Pool
}

// NewDriverRepo builds a DriverRepo.
func NewDriverRepo(pool *pgxpool.Pool) *DriverRepo {
	return &DriverRepo{pool: pool}
}

// Create inserts a driver. The service chooses the id and timestamp.
func (r *DriverRepo) Create(ctx context.Context, d model.Driver) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO drivers (id, name, phone, vehicle, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		d.ID, d.Name, d.Phone, d.Vehicle, d.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert driver: %w", err)
	}
	return nil
}

// List returns every driver in alphabetical order.
func (r *DriverRepo) List(ctx context.Context) ([]model.Driver, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, phone, vehicle, created_at
		FROM drivers
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list drivers: %w", err)
	}
	defer rows.Close()

	drivers := make([]model.Driver, 0)
	for rows.Next() {
		var d model.Driver
		if err := rows.Scan(&d.ID, &d.Name, &d.Phone, &d.Vehicle, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan driver: %w", err)
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}

// Get loads one driver by id.
func (r *DriverRepo) Get(ctx context.Context, id string) (model.Driver, error) {
	var d model.Driver
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, phone, vehicle, created_at
		FROM drivers
		WHERE id = $1`, id).Scan(&d.ID, &d.Name, &d.Phone, &d.Vehicle, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Driver{}, model.ErrNotFound
	}
	if err != nil {
		return model.Driver{}, fmt.Errorf("get driver: %w", err)
	}
	return d, nil
}
