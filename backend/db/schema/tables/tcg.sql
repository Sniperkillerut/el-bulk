-- TCG (Trading Card Game) Table
CREATE TABLE IF NOT EXISTS tcg (
  id         TEXT PRIMARY KEY, -- slug: mtg, pokemon, etc.
  name       TEXT NOT NULL,
  image_url  TEXT,
  is_active  BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default TCGs with premium banners
INSERT INTO tcg (id, name, image_url) VALUES
  ('mtg',      'Magic: The Gathering', '/tcgs/mtg_banner.png'),
  ('pokemon',  'Pokémon',             '/tcgs/pokemon_banner.png'),
  ('lorcana',  'Disney Lorcana',      '/tcgs/lorcana_banner.png'),
  ('onepiece', 'One Piece',           '/tcgs/one_piece_banner.png'),
  ('yugioh',   'Yu-Gi-Oh!',           '/tcgs/yugioh_banner.png'),
  ('starwars', 'Star Wars Unlimited',  '/tcgs/starwars_banner.png'),
  ('weiss',    'Weiss Schwarz',       '/tcgs/weiss_banner.png')
ON CONFLICT (id) DO UPDATE SET image_url = EXCLUDED.image_url;
