-- Admin Audit Log Table
CREATE TABLE IF NOT EXISTS admin_audit_log (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  admin_id       UUID REFERENCES admin(id) ON DELETE SET NULL,
  admin_username TEXT NOT NULL, -- Cached for historical record in case admin is deleted
  action         TEXT NOT NULL, -- e.g. CREATE_PRODUCT, UPDATE_SETTINGS, DELETE_CATEGORY
  resource_type  TEXT NOT NULL, -- e.g. product, setting, category, storage
  resource_id    TEXT,          -- ID of the affected resource
  details        JSONB,         -- Full before/after detail
  ip_address     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for performance on the dashboard
CREATE INDEX IF NOT EXISTS idx_audit_log_admin ON admin_audit_log(admin_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON admin_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_resource ON admin_audit_log(resource_type, resource_id);
