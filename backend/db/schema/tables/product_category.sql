-- Product Category Mapping
CREATE TABLE product_category (
  product_id  UUID REFERENCES product(id) ON DELETE CASCADE,
  category_id UUID REFERENCES custom_category(id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, category_id)
);

CREATE INDEX idx_pc_category_id ON product_category(category_id);
