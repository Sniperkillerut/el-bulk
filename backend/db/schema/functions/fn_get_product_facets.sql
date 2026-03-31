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
    p_filter_logic TEXT DEFAULT 'or',
    p_is_admin BOOLEAN DEFAULT false
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    WITH base_products AS (
        -- Mandatory context filters: only show cards that actually have stock (>0)
        SELECT 
            p.*
        FROM product p
        LEFT JOIN tcg t ON p.tcg = t.id
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
            AND (p_storage_id = '' OR EXISTS (
                SELECT 1 FROM product_storage ps 
                WHERE ps.product_id = p.id AND ps.storage_id::text = p_storage_id AND ps.quantity > 0
            ))
            AND EXISTS (SELECT 1 FROM product_storage ps2 WHERE ps2.product_id = p.id AND ps2.quantity > 0)
            AND (p_is_admin OR (t.is_active IS NULL OR t.is_active = true))
    ),
    -- Individual facet calculations
    f_condition AS (
        SELECT COALESCE(condition, 'unknown') as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
            AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
            AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
            AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
          ELSE
            true
          END
        GROUP BY val
    ),
    f_foil AS (
        SELECT COALESCE(LOWER(foil_treatment), 'non_foil') as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
            AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
            AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
            AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
          ELSE
            true
          END
        GROUP BY val
    ),
    f_rarity AS (
        SELECT COALESCE(LOWER(rarity), 'unknown') as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
            AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
            AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
            AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
          ELSE
            true
          END
        GROUP BY val
    ),
    f_language AS (
        SELECT COALESCE(LOWER(language), 'en') as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
        LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
        LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
            AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
            AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
            AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
          ELSE
            true
          END
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
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
            AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
            AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
            AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
          ELSE
            true
          END
    ),
    f_treatment AS (
        SELECT val, SUM(c) as c FROM (
            SELECT COALESCE(LOWER(card_treatment), 'normal') as val, COUNT(DISTINCT base_products.id) as c 
            FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE 
              CASE WHEN p_filter_logic = 'and' THEN
                (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
                AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
                AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
                AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
                AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
                AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
              ELSE
                true
              END
            GROUP BY val
            UNION ALL
            SELECT 'full_art' as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE full_art = true AND 
              CASE WHEN p_filter_logic = 'and' THEN
                (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
                AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
                AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
                AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
                AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
                AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
              ELSE true END
            UNION ALL
            SELECT 'textless' as val, COUNT(DISTINCT base_products.id) as c FROM base_products 
            LEFT JOIN product_category pc_col ON base_products.id = pc_col.product_id
            LEFT JOIN custom_category c_col ON pc_col.category_id = c_col.id
            WHERE textless = true AND 
              CASE WHEN p_filter_logic = 'and' THEN
                (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
                AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
                AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
                AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
                AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
                AND (p_collection = '' OR LOWER(c_col.slug) = ANY(string_to_array(LOWER(p_collection), ',')))
              ELSE true END
        ) t GROUP BY val
    ),
    f_collection AS (
        SELECT COALESCE(c_col.slug, 'unknown') as val, COUNT(DISTINCT base_products.id) as c
        FROM base_products
        JOIN product_category pc_col ON base_products.id = pc_col.product_id
        JOIN custom_category c_col ON pc_col.category_id = c_col.id
        WHERE 
          CASE WHEN p_filter_logic = 'and' THEN
            (p_condition = '' OR LOWER(condition) = ANY(string_to_array(LOWER(p_condition), ',')))
            AND (p_foil = '' OR LOWER(foil_treatment) = ANY(string_to_array(LOWER(p_foil), ',')))
            AND (p_treatment = '' OR (LOWER(card_treatment) = ANY(string_to_array(LOWER(p_treatment), ',')) OR (LOWER(p_treatment) LIKE '%full_art%' AND full_art = true) OR (LOWER(p_treatment) LIKE '%textless%' AND textless = true)))
            AND (p_rarity = '' OR LOWER(rarity) = ANY(string_to_array(LOWER(p_rarity), ',')))
            AND (p_language = '' OR LOWER(language) = ANY(string_to_array(LOWER(p_language), ',')))
            AND (p_color = '' OR color_identity ILIKE '%' || p_color || '%')
          ELSE
            true
          END
        GROUP BY val
    )
    SELECT jsonb_build_object(
        'condition', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_condition),
        'foil', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_foil),
        'rarity', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_rarity),
        'language', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_language),
        'color', (SELECT jsonb_build_object('W', w, 'U', u, 'B', b, 'R', r, 'G', g, 'C', c) FROM f_color),
        'treatment', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_treatment),
        'collection', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_collection)
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql;

