package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toontoonz/logistic-go/backend/internal/model"
)

// LocationRepo stores the GPS trail in PostGIS.
type LocationRepo struct {
	pool *pgxpool.Pool
}

// NewLocationRepo builds a LocationRepo.
func NewLocationRepo(pool *pgxpool.Pool) *LocationRepo {
	return &LocationRepo{pool: pool}
}

// Insert appends one sample. ST_MakePoint takes (longitude, latitude),
// the opposite of the way people usually say "lat, lng".
func (r *LocationRepo) Insert(ctx context.Context, p model.Position) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO location_history (driver_id, geom, lat, lng, recorded_at)
		VALUES (
			$1,
			ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
			$3, $2, $4
		)`, p.DriverID, p.Lng, p.Lat, p.RecordedAt)
	if err != nil {
		return fmt.Errorf("insert location: %w", err)
	}
	return nil
}

// Latest is the newest sample for one driver.
func (r *LocationRepo) Latest(ctx context.Context, driverID string) (model.Position, error) {
	var p model.Position
	err := r.pool.QueryRow(ctx, `
		SELECT l.driver_id, d.name, l.lat, l.lng, l.recorded_at
		FROM location_history l
		JOIN drivers d ON d.id = l.driver_id
		WHERE l.driver_id = $1
		ORDER BY l.recorded_at DESC
		LIMIT 1`, driverID).Scan(&p.DriverID, &p.DriverName, &p.Lat, &p.Lng, &p.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Position{}, model.ErrNotFound
	}
	if err != nil {
		return model.Position{}, fmt.Errorf("latest location: %w", err)
	}
	return p, nil
}

// LatestAll returns each driver's newest sample. DISTINCT ON keeps the
// first row per driver; the ORDER BY puts that first row latest-first.
func (r *LocationRepo) LatestAll(ctx context.Context) ([]model.Position, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (l.driver_id)
		       l.driver_id, d.name, l.lat, l.lng, l.recorded_at
		FROM location_history l
		JOIN drivers d ON d.id = l.driver_id
		ORDER BY l.driver_id, l.recorded_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("latest locations: %w", err)
	}
	defer rows.Close()

	out := make([]model.Position, 0)
	for rows.Next() {
		var p model.Position
		if err := rows.Scan(&p.DriverID, &p.DriverName, &p.Lat, &p.Lng, &p.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// History returns the newest samples for one driver, oldest last so a
// map can draw the trail in travel order. limit is applied in SQL.
func (r *LocationRepo) History(ctx context.Context, driverID string, limit int) ([]model.Position, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT driver_id, lat, lng, recorded_at
		FROM location_history
		WHERE driver_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2`, driverID, limit)
	if err != nil {
		return nil, fmt.Errorf("location history: %w", err)
	}
	defer rows.Close()

	// Newest first from SQL; reverse so callers walk the path in time.
	desc := make([]model.Position, 0)
	for rows.Next() {
		var p model.Position
		if err := rows.Scan(&p.DriverID, &p.Lat, &p.Lng, &p.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		desc = append(desc, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(desc)-1; i < j; i, j = i+1, j-1 {
		desc[i], desc[j] = desc[j], desc[i]
	}
	return desc, nil
}
