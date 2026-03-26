-- Order Item Table
CREATE TABLE order_item (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id       UUID REFERENCES "order"(id) ON DELETE CASCADE,
  product_id     UUID REFERENCES product(id) ON DELETE SET NULL,
  product_name   TEXT NOT NULL,
  product_set    TEXT,
  foil_treatment TEXT,
  card_treatment TEXT,
  condition      TEXT,
  unit_price_cop NUMERIC(14, 2) NOT NULL,
  quantity       INTEGER NOT NULL,
  stored_in_snapshot JSONB
);
