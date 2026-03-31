-- fn_get_product_facets
-- Returns a JSONB object containing counts for Condition, Foil, Treatment, Rarity, Language, Color, and Collection.
-- Each count respects all active filters *except* its own dimension (standard faceted navigation behavior).
CREATE OR REPLACE FUNCTION fn_get_product_facets(
    p_tcg TEXT DEFAULT '',
    p_category TEXT DEFAULT '',
    p_search TEXT DEFAULT '',
    p_storage_id TEXT DEFAULT '',
    p_foil TEXT DEFAULT '',
    p_treatment TEXT DEFAULT '',
    p_condition TEXT DEFAULT '',
    p_rarity TEXT DEFAULT '',
    p_language TEXT DEFAULT '',
    p_color TEXT DEFAULT '',
    p_collection TEXT DEFAULT '',
    p_is_admin BOOLEAN DEFAULT false
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    WITH base_products AS (
        -- This logic must match backend/handlers/products.go:buildFilters
        SELECT 
            p.*
        FROM product p
        LEFT JOIN tcg t ON p.tcg = t.id
        LEFT JOIN product_storage ps ON p.id = ps.product_id
        WHERE 
            (p_tcg = '' OR p.tcg = p_tcg)
            AND (p_category = '' OR p.category = p_category)
            AND (p_search = '' OR (
                p.name % p_search OR 
                p.name ILIKE '%' || p_search || '%' OR 
                COALESCE(p.set_name, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.set_code, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.artist, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.collector_number, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.oracle_text, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.type_line, '') ILIKE '%' || p_search || '%' OR
                COALESCE(p.promo_type, '') ILIKE '%' || p_search || '%'
            ))
            AND (p_storage_id = '' OR (ps.storage_id::text = p_storage_id AND ps.quantity > 0))
            AND (p_is_admin OR (t.is_active IS NULL OR t.is_active = true))
    ),
    -- Individual facet calculations
    f_condition AS (
        SELECT COALESCE(condition, 'unknown') as val, COUNT(DISTINCT base_products.id) as c 
        FROM base_products
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
          AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
          AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
          AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
        GROUP BY val
    ),
    f_foil AS (
        SELECT COALESCE(foil_treatment, 'unknown') as val, COUNT(DISTINCT base_products.id) as c 
        FROM base_products
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
          AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
          AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
          AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
        GROUP BY val
    ),
    f_rarity AS (
        SELECT COALESCE(LOWER(rarity), 'unknown') as val, COUNT(DISTINCT base_products.id) as c 
        FROM base_products
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
          AND (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
          AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
          AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
        GROUP BY val
    ),
    f_language AS (
        SELECT COALESCE(LOWER(language), 'unknown') as val, COUNT(DISTINCT base_products.id) as c 
        FROM base_products
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
          AND (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
          AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
          AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
        GROUP BY val
    ),
    f_color AS (
        SELECT 
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%W%') as w,
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%U%') as u,
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%B%') as b,
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%R%') as r,
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%G%') as g,
            COUNT(DISTINCT base_products.id) FILTER (WHERE color_identity ILIKE '%C%') as c
        FROM base_products
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
          AND (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
          AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
          AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
    ),
    f_treatment AS (
        -- Treatment is tricky because it involves both card_treatment text and full_art/textless flags
        SELECT val, SUM(c) as c FROM (
            SELECT LOWER(card_treatment) as val, COUNT(DISTINCT base_products.id) as c 
            FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
              AND (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
              AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
              AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
              AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
              AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
            GROUP BY val
            UNION ALL
            SELECT 'full_art' as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE full_art = true AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ','))) -- Respect collection filter for extra flags too
            UNION ALL
            SELECT 'textless' as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE textless = true AND (p_collection = '' OR c_col.slug = ANY(string_to_array(p_collection, ',')))
        ) t GROUP BY val
    ),
    f_collection AS (
        SELECT c_col.slug as val, COUNT(DISTINCT base_products.id) as c
        FROM base_products
        JOIN product_category pc_col ON base_products.id = pc_col.product_id
        JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE (p_condition = '' OR condition = ANY(string_to_array(p_condition, ',')))
          AND (p_foil = '' OR foil_treatment = ANY(string_to_array(p_foil, ',')))
          AND (p_treatment = '' OR card_treatment = ANY(string_to_array(p_treatment, ',')))
          AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(p_rarity, ',')))
          AND (p_language = '' OR LOWER(language) = ANY(string_to_array(p_language, ',')))
          AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
        -- NOTE: Do NOT filter p_collection here, to show all categories in the filter list
        GROUP BY val
    )
    SELECT jsonb_build_object(
        'condition', (SELECT jsonb_object_agg(val, c) FROM f_condition),
        'foil', (SELECT jsonb_object_agg(val, c) FROM f_foil),
        'rarity', (SELECT jsonb_object_agg(val, c) FROM f_rarity),
        'language', (SELECT jsonb_object_agg(val, c) FROM f_language),
        'color', (SELECT jsonb_build_object('W', w, 'U', u, 'B', b, 'R', r, 'G', g, 'C', c) FROM f_color),
        'treatment', (SELECT jsonb_object_agg(val, c) FROM f_treatment),
        'collection', (SELECT jsonb_object_agg(val, c) FROM f_collection)
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
