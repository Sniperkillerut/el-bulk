-- Trigger to update updated_at on bounty
CREATE TRIGGER trg_bounty_updated_at
BEFORE UPDATE ON bounty
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
