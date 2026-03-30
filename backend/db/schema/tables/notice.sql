-- Notices Table (Blog/News)
CREATE TABLE IF NOT EXISTS notice (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    content_html TEXT NOT NULL,
    featured_image_url TEXT,
    is_published BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for slug lookups
CREATE INDEX IF NOT EXISTS idx_notice_slug ON notice(slug);

-- Trigger to update updated_at
CREATE TRIGGER tr_notice_updated_at
    BEFORE UPDATE ON notice
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_updated_at();
