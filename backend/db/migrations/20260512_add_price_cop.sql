-- Phase 1: Add price_cop materialized column and automatic calculation logic

-- 1. Add column
ALTER TABLE product ADD COLUMN IF NOT EXISTS price_cop NUMERIC(12, 2) DEFAULT 0;

-- 2. Create calculation function
CREATE OR REPLACE FUNCTION fn_calculate_price_cop(
    p_price_reference NUMERIC,
    p_price_source TEXT,
    p_price_cop_override NUMERIC
) RETURNS NUMERIC AS $$
DECLARE
    v_usd_rate NUMERIC;
    v_eur_rate NUMERIC;
    v_ck_rate NUMERIC;
BEGIN
    -- Return override if present
    IF p_price_cop_override IS NOT NULL AND p_price_cop_override > 0 THEN
        RETURN p_price_cop_override;
    END IF;

    -- Return 0 if no reference
    IF p_price_reference IS NULL OR p_price_reference = 0 THEN
        RETURN 0;
    END IF;

    -- Fetch rates from setting table (using COALESCE to fallback to defaults if keys missing)
    v_usd_rate := COALESCE((SELECT value::NUMERIC FROM setting WHERE key = 'usd_to_cop_rate'), 4200);
    v_eur_rate := COALESCE((SELECT value::NUMERIC FROM setting WHERE key = 'eur_to_cop_rate'), 4600);
    v_ck_rate  := COALESCE((SELECT value::NUMERIC FROM setting WHERE key = 'ck_to_cop_rate'), 4000);

    -- Calculate based on source
    CASE LOWER(p_price_source)
        WHEN 'tcgplayer' THEN RETURN p_price_reference * v_usd_rate;
        WHEN 'cardkingdom' THEN RETURN p_price_reference * v_ck_rate;
        WHEN 'cardmarket' THEN RETURN p_price_reference * v_eur_rate;
        ELSE RETURN 0;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- 3. Create trigger function
CREATE OR REPLACE FUNCTION fn_tr_product_price_cop() RETURNS TRIGGER AS $$
BEGIN
    NEW.price_cop := fn_calculate_price_cop(NEW.price_reference, NEW.price_source, NEW.price_cop_override);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. Attach trigger
DROP TRIGGER IF EXISTS tr_product_price_cop ON product;
CREATE TRIGGER tr_product_price_cop
    BEFORE INSERT OR UPDATE OF price_reference, price_source, price_cop_override ON product
    FOR EACH ROW
    EXECUTE FUNCTION fn_tr_product_price_cop();

-- 5. Backfill existing data
UPDATE product 
SET price_cop = fn_calculate_price_cop(price_reference, price_source, price_cop_override);
