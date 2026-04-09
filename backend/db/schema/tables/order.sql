-- Order Table
CREATE TABLE IF NOT EXISTS "order" (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_number   TEXT UNIQUE NOT NULL,
  customer_id    UUID REFERENCES customer(id) ON DELETE RESTRICT,
  status         TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled', 'shipped', 'ready_for_pickup')),
  payment_method TEXT NOT NULL,
  subtotal_cop   NUMERIC(14, 2) NOT NULL DEFAULT 0,
  shipping_cop   NUMERIC(14, 2) NOT NULL DEFAULT 0,
  tax_cop        NUMERIC(14, 2) NOT NULL DEFAULT 0,
  total_cop      NUMERIC(14, 2) NOT NULL,
  tracking_number TEXT,
  tracking_url   TEXT,
  is_local_pickup BOOLEAN NOT NULL DEFAULT false,
  inventory_restored BOOLEAN NOT NULL DEFAULT false,
  notes          TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  confirmed_at   TIMESTAMPTZ,
  completed_at   TIMESTAMPTZ
);
-- Note: 'order' is a reserved word, quoting locally to be safe, though many drivers handle it.
