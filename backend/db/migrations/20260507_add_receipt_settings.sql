-- Migration: Add receipt and branding settings to setting table
INSERT INTO setting (key, value) VALUES 
('receipt_auto_email', 'true'),
('receipt_footer_text', 'Thank you for your purchase at El Bulk! If you have any questions, contact us at info@elbulk.com'),
('store_logo_url', '')
ON CONFLICT (key) DO NOTHING;
