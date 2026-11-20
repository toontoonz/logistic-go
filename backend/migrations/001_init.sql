-- Initial schema for orders, drivers, and GPS history.
-- Status values must stay in sync with internal/service/status.go.

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS drivers (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    phone      TEXT NOT NULL,
    vehicle    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS orders (
    id              TEXT PRIMARY KEY,
    customer_name   TEXT NOT NULL,
    pickup_address  TEXT NOT NULL,
    dropoff_address TEXT NOT NULL,
    pickup_lat      DOUBLE PRECISION NOT NULL,
    pickup_lng      DOUBLE PRECISION NOT NULL,
    dropoff_lat     DOUBLE PRECISION NOT NULL,
    dropoff_lng     DOUBLE PRECISION NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'assigned', 'picked_up', 'in_transit', 'delivered')),
    driver_id       TEXT REFERENCES drivers (id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);
CREATE INDEX IF NOT EXISTS orders_driver_idx ON orders (driver_id);

-- History is the durable log. lat/lng are stored twice on purpose:
-- the plain columns are what the API returns, and geom lets PostGIS
-- answer "who was near this point?" later without parsing JSON.
CREATE TABLE IF NOT EXISTS location_history (
    id          BIGSERIAL PRIMARY KEY,
    driver_id   TEXT NOT NULL REFERENCES drivers (id),
    geom        GEOGRAPHY(POINT, 4326) NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS location_history_driver_time_idx
    ON location_history (driver_id, recorded_at DESC);

CREATE INDEX IF NOT EXISTS location_history_geom_idx
    ON location_history USING GIST (geom);
