CREATE TABLE IF NOT EXISTS newsletter_subscriber (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    customer_id UUID REFERENCES customer(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_newsletter_email ON newsletter_subscriber(email);
CREATE INDEX IF NOT EXISTS idx_newsletter_customer ON newsletter_subscriber(customer_id);
