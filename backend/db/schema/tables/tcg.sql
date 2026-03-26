-- TCG (Trading Card Game) Table
CREATE TABLE tcg (
  id         TEXT PRIMARY KEY, -- slug: mtg, pokemon, etc.
  name       TEXT NOT NULL,
  is_active  BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Default TCGs
INSERT INTO tcg (id, name) VALUES
  ('mtg', 'Magic: The Gathering'),
  ('pokemon', 'Pokémon'),
  ('lorcana', 'Disney Lorcana'),
  ('onepiece', 'One Piece'),
  ('yugioh', 'Yu-Gi-Oh!'),
  ('starwars', 'Star Wars Unlimited'),
  ('weiss', 'Weiss Schwarz')
ON CONFLICT (id) DO NOTHING;
