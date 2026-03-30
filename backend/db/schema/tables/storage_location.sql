-- Storage Location Table
CREATE TABLE IF NOT EXISTS storage_location (
  id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL
);
