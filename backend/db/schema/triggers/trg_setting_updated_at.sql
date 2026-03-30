-- Trigger to update updated_at on setting
DROP TRIGGER IF EXISTS trg_setting_updated_at ON setting;
CREATE TRIGGER trg_setting_updated_at
BEFORE UPDATE ON setting
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
