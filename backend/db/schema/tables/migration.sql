-- Migration Tracking Table
CREATE TABLE IF NOT EXISTS migration (
    name       TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
