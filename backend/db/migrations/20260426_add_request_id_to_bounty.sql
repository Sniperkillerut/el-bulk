ALTER TABLE bounty ADD COLUMN request_id UUID REFERENCES client_request(id) ON DELETE SET NULL;
CREATE INDEX idx_bounty_request_id ON bounty(request_id);
