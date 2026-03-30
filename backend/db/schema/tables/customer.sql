-- Customer Table
CREATE TABLE IF NOT EXISTS customer (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  first_name      TEXT NOT NULL,
  last_name       TEXT NOT NULL,
  email           TEXT,
  phone           TEXT UNIQUE NOT NULL,
  id_number       TEXT,
  address         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
