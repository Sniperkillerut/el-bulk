-- Trigger to update updated_at on bounty
DROP TRIGGER IF EXISTS trg_bounty_updated_at ON bounty;
CREATE TRIGGER trg_bounty_updated_at
BEFORE UPDATE ON bounty
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
