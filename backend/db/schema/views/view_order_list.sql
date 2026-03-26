-- View: view_order_list
-- Consolidates order data with customer name and total item count
CREATE OR REPLACE VIEW view_order_list AS
SELECT o.*, 
       c.first_name || ' ' || c.last_name as customer_name,
       COALESCE((SELECT COUNT(*) FROM order_item WHERE order_id = o.id), 0) as item_count
FROM "order" o
JOIN customer c ON o.customer_id = c.id;
