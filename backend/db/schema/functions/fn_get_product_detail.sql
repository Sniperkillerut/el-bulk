-- Get Product Detail
-- Returns a consolidated JSONB object that maps directly to models.Product
CREATE OR REPLACE FUNCTION fn_get_product_detail(p_id UUID)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT to_jsonb(p) || jsonb_build_object(
        'stored_in', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                'stored_in_id', ps.storage_id,
                'name', sl.name,
                'quantity', ps.quantity
            ))
            FROM product_storage ps
            JOIN storage_location sl ON ps.storage_id = sl.id
            WHERE ps.product_id = p_id AND ps.quantity > 0
        ), '[]'::jsonb),
        'categories', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                'id', cc.id,
                'name', cc.name,
                'slug', cc.slug,
                'show_badge', cc.show_badge,
                'is_active', cc.is_active,
                'searchable', cc.searchable
            ))
            FROM product_category pc
            JOIN custom_category cc ON pc.category_id = cc.id
            WHERE pc.product_id = p_id
        ), '[]'::jsonb)
    ) INTO result
    FROM product p
    WHERE p.id = p_id;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
