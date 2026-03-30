-- Custom Category (Collections)
CREATE TABLE IF NOT EXISTS custom_category (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT UNIQUE NOT NULL,
  slug       TEXT UNIQUE NOT NULL,
  is_active   BOOLEAN NOT NULL DEFAULT true,
  show_badge  BOOLEAN NOT NULL DEFAULT true,
  searchable  BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default categories
INSERT INTO custom_category (name, slug) VALUES 
  ('Featured', 'featured'),
  ('Sale', 'sale'),
  ('New Arrivals', 'new-arrivals'),
  ('Hot Items', 'hot-items')
ON CONFLICT (slug) DO NOTHING;
