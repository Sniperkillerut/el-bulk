DROP VIEW IF EXISTS view_order_list CASCADE;

CREATE VIEW view_order_list AS
SELECT o.*, 
       c.first_name || ' ' || c.last_name as customer_name,
       c.phone as customer_phone,
       c.email as customer_email,
       COALESCE((SELECT COUNT(*) FROM order_item WHERE order_id = o.id), 0) as item_count
FROM "order" o
JOIN customer c ON o.customer_id = c.id;
