-- Admin-configurable global settings (key/value)
CREATE TABLE setting (
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
  ('contact_hours', 'Mon - Sat: 11:00 AM - 7:00 PM')
ON CONFLICT (key) DO NOTHING;
