-- Order Table
CREATE TABLE "order" (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_number   TEXT UNIQUE NOT NULL,
  customer_id    UUID REFERENCES customer(id) ON DELETE RESTRICT,
  status         TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled')),
  payment_method TEXT NOT NULL,
  total_cop      NUMERIC(14, 2) NOT NULL,
  notes          TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at   TIMESTAMPTZ
);
-- Note: 'order' is a reserved word, quoting locally to be safe, though many drivers handle it.
