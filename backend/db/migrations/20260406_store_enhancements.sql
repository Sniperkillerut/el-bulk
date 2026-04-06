-- Store Enhancements Migration
-- 1. Updates to the 'order' table
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS subtotal_cop NUMERIC(14, 2) NOT NULL DEFAULT 0;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS shipping_cop NUMERIC(14, 2) NOT NULL DEFAULT 0;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS tax_cop NUMERIC(14, 2) NOT NULL DEFAULT 0;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS tracking_number TEXT;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS tracking_url TEXT;
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS is_local_pickup BOOLEAN NOT NULL DEFAULT false;

-- Sync subtotal for existing orders (total = subtotal + shipping + tax)
UPDATE "order" SET subtotal_cop = total_cop WHERE subtotal_cop = 0;

-- Update status constraint for 'order' table
-- Note: 'shipped' and 'ready_for_pickup' are the new statuses
ALTER TABLE "order" DROP CONSTRAINT IF EXISTS order_status_check;
ALTER TABLE "order" ADD CONSTRAINT order_status_check 
  CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled', 'shipped', 'ready_for_pickup'));

-- 2. Updates to the 'product' table
ALTER TABLE product ADD COLUMN IF NOT EXISTS cost_basis_cop NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- 3. Updates to the 'setting' table
INSERT INTO setting (key, value) VALUES ('flat_shipping_fee_cop', '15000') ON CONFLICT (key) DO NOTHING;
