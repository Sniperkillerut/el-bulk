-- Product Storage (Inventory)
CREATE TABLE product_storage (
  product_id   UUID REFERENCES product(id) ON DELETE CASCADE,
  storage_id   UUID REFERENCES storage_location(id) ON DELETE CASCADE,
  quantity     INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  PRIMARY KEY (product_id, storage_id)
);

CREATE INDEX idx_ps_storage_id ON product_storage(storage_id) WHERE quantity > 0;
