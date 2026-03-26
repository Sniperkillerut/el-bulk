-- Trigger to update updated_at on setting
CREATE TRIGGER trg_setting_updated_at
BEFORE UPDATE ON setting
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
