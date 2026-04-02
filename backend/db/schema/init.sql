-- El Bulk TCG Store - Master Schema Initialization
-- Order is critical for foreign key dependencies

-- 1. Extensions
\i extensions.sql

-- 1.5 Utility Functions
\i functions/fn_update_updated_at.sql

-- 2. Independent Tables
\i tables/migration.sql
\i tables/setting.sql
\i tables/tcg.sql
\i tables/storage_location.sql
\i tables/customer.sql
\i tables/customer_auth.sql
\i tables/admin.sql
\i tables/custom_category.sql
\i tables/bounty.sql
\i tables/client_request.sql
\i tables/notice.sql
\i tables/newsletter_subscriber.sql
\i tables/bounty_offer.sql
\i tables/theme.sql
\i tables/translation.sql

-- 3. Dependent Tables
\i tables/product.sql
\i tables/deck_card.sql
\i tables/product_category.sql
\i tables/product_storage.sql
\i tables/order.sql
\i tables/order_item.sql
\i tables/customer_note.sql

-- 4. Functions & Stored Procedures
\i functions/fn_update_product_stock.sql
\i functions/fn_get_product_detail.sql
\i functions/fn_bulk_upsert_product.sql
\i functions/fn_place_order.sql
\i functions/fn_complete_order.sql
\i functions/fn_fulfill_bounty_offer.sql
\i functions/fn_get_product_facets.sql
\i functions/fn_submit_client_request.sql
\i functions/fn_submit_bounty_offer.sql

-- 5. Views
\i views/view_product_enriched.sql
\i views/view_order_list.sql
\i views/view_order_item_enriched.sql

-- 5. Triggers
\i triggers/trg_product_updated_at.sql
\i triggers/trg_bounty_updated_at.sql
\i triggers/trg_setting_updated_at.sql
\i triggers/trg_sync_product_stock.sql
\i triggers/trg_bounty_fulfillment.sql
