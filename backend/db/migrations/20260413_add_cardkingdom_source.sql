-- Migration to add 'cardkingdom' as a valid price source
-- Drop the existing check constraint (Postgres usually names it table_column_check if anonymous)
-- We'll try to drop it by name, but to be safe we'll use a DO block to find it.

DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'product'::regclass 
      AND contype = 'c' 
      AND pg_get_constraintdef(oid) LIKE '%price_source%';

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE product DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Re-add the constraint with 'cardkingdom'
ALTER TABLE product ADD CONSTRAINT product_price_source_check 
    CHECK (price_source IN ('tcgplayer', 'cardmarket', 'cardkingdom', 'manual'));
