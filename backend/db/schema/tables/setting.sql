-- Admin-configurable global settings (key/value)
CREATE TABLE IF NOT EXISTS setting (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Base Data
INSERT INTO setting (key, value) VALUES
  ('usd_to_cop_rate', '4200'),
  ('eur_to_cop_rate', '4600'),
  ('contact_address', 'Cra. 15 # 76-54, Local 201, Centro Comercial Unilago, Bogotá'),
  ('contact_phone', '+57 300 000 0000'),
  ('contact_email', 'contact@el-bulk.co'),
  ('contact_instagram', 'el-bulk'),
  ('contact_hours', 'Mon - Sat: 11:00 AM - 7:00 PM'),
  ('default_theme_id', '00000000-0000-0000-0000-000000000001'),
  ('flat_shipping_fee_cop', '15000')
ON CONFLICT (key) DO NOTHING;
