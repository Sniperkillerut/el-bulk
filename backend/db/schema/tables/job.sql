CREATE TABLE IF NOT EXISTS job (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type         TEXT NOT NULL, -- e.g., 'price_refresh', 'scryfall_sync', 'csv_import'
  status       TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
  progress     INT NOT NULL DEFAULT 0, -- 0 to 100
  payload      JSONB, -- input parameters for the job
  result       JSONB, -- output data or summary
  error        TEXT, -- error message if failed
  admin_id     UUID REFERENCES admin(id), -- who started it
  started_at   TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_job_status ON job(status);
CREATE INDEX IF NOT EXISTS idx_job_created_at ON job(created_at DESC);
