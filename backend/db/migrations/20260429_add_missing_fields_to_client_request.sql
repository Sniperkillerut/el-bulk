-- Migration: Add missing metadata fields to client_request
-- These fields are used for better card identification and UX in the profile and admin views.

ALTER TABLE client_request 
  ADD COLUMN IF NOT EXISTS image_url TEXT,
  ADD COLUMN IF NOT EXISTS foil_treatment TEXT,
  ADD COLUMN IF NOT EXISTS card_treatment TEXT,
  ADD COLUMN IF NOT EXISTS set_code TEXT,
  ADD COLUMN IF NOT EXISTS collector_number TEXT,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Trigger to update updated_at
DROP TRIGGER IF EXISTS tr_client_request_updated_at ON client_request;
CREATE TRIGGER tr_client_request_updated_at
    BEFORE UPDATE ON client_request
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_updated_at();
