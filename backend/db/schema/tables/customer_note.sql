CREATE TABLE IF NOT EXISTS customer_note (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customer(id) ON DELETE CASCADE NOT NULL,
    order_id UUID REFERENCES "order"(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    admin_id UUID REFERENCES admin(id) ON DELETE SET NULL, -- Track who added the note
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_note_customer ON customer_note(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_note_order ON customer_note(order_id);
