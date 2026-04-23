-- fn_get_product_facets
-- Returns a JSONB object containing counts for Condition, Foil, Treatment, Rarity, Language, Color, Collection, and Set.
-- Optimized to use array operators (ANY) instead of unnest loops for performance.
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
    p_set_name TEXT DEFAULT '',
    p_in_stock BOOLEAN DEFAULT true,
    p_filter_logic TEXT DEFAULT 'or',
    p_is_admin BOOLEAN DEFAULT false
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
    v_foil_arr TEXT[];
    v_treatment_arr TEXT[];
    v_condition_arr TEXT[];
    v_rarity_arr TEXT[];
    v_language_arr TEXT[];
    v_color_arr TEXT[];
    v_collection_arr TEXT[];
    v_set_name_arr TEXT[];
BEGIN
    -- Pre-parse filters into arrays for faster matching
    v_foil_arr := CASE WHEN p_foil = '' THEN NULL ELSE string_to_array(LOWER(p_foil), ',') END;
    v_treatment_arr := CASE WHEN p_treatment = '' THEN NULL ELSE string_to_array(LOWER(p_treatment), ',') END;
    v_condition_arr := CASE WHEN p_condition = '' THEN NULL ELSE string_to_array(UPPER(p_condition), ',') END;
    v_rarity_arr := CASE WHEN p_rarity = '' THEN NULL ELSE string_to_array(LOWER(p_rarity), ',') END;
    v_language_arr := CASE WHEN p_language = '' THEN NULL ELSE string_to_array(LOWER(p_language), ',') END;
    v_color_arr := CASE WHEN p_color = '' THEN NULL ELSE string_to_array(UPPER(p_color), ',') END;
    v_collection_arr := CASE WHEN p_collection = '' THEN NULL ELSE string_to_array(LOWER(p_collection), ',') END;
    v_set_name_arr := CASE WHEN p_set_name = '' THEN NULL ELSE string_to_array(p_set_name, ',') END;

    WITH base_products AS MATERIALIZED (
        SELECT p.*
        FROM product p
        LEFT JOIN tcg t ON p.tcg = t.id
        WHERE 
            (p_tcg = '' OR LOWER(p.tcg) = LOWER(p_tcg))
            AND (p_category = '' OR p.category = p_category)
            AND (p_search = '' OR p.search_vector @@ websearch_to_tsquery('english', p_search))
            AND (p_storage_id = '' OR EXISTS (
                SELECT 1 FROM product_storage ps 
                WHERE ps.product_id = p.id AND ps.storage_id::text = p_storage_id AND ps.quantity > 0
            ))
            AND (NOT p_in_stock OR p.stock > 0)
            AND (p_is_admin OR (t.is_active IS NULL OR t.is_active = true))
    ),
    all_filtered AS MATERIALIZED (
        -- Standard filtering applies to all facets except their own dimension
        SELECT *,
               (v_foil_arr IS NULL OR LOWER(foil_treatment) = ANY(v_foil_arr)) as match_foil,
               (v_treatment_arr IS NULL OR LOWER(card_treatment) = ANY(v_treatment_arr) OR (full_art AND 'full_art' = ANY(v_treatment_arr)) OR (textless AND 'textless' = ANY(v_treatment_arr))) as match_treatment,
               (v_condition_arr IS NULL OR condition = ANY(v_condition_arr)) as match_condition,
               (v_rarity_arr IS NULL OR LOWER(rarity) = ANY(v_rarity_arr)) as match_rarity,
               (v_language_arr IS NULL OR LOWER(language) = ANY(v_language_arr)) as match_language,
               (v_color_arr IS NULL OR (
                   SELECT bool_or(color_identity ILIKE '%' || c || '%') FROM unnest(v_color_arr) c
               )) as match_color,
               (v_collection_arr IS NULL OR EXISTS (
                   SELECT 1 FROM product_category pc JOIN custom_category cc ON pc.category_id = cc.id 
                   WHERE pc.product_id = base_products.id AND cc.slug = ANY(v_collection_arr)
               )) as match_collection,
               (v_set_name_arr IS NULL OR set_name = ANY(v_set_name_arr)) as match_set
        FROM base_products
    ),
    f_condition AS (
        SELECT COALESCE(condition, 'unknown') as val, COUNT(*) as c FROM all_filtered
        WHERE match_foil AND match_treatment AND match_rarity AND match_language AND match_color AND match_collection AND match_set
        GROUP BY val
    ),
    f_foil AS (
        SELECT COALESCE(LOWER(foil_treatment), 'non_foil') as val, COUNT(*) as c FROM all_filtered
        WHERE match_condition AND match_treatment AND match_rarity AND match_language AND match_color AND match_collection AND match_set
        GROUP BY val
    ),
    f_rarity AS (
        SELECT COALESCE(LOWER(rarity), 'unknown') as val, COUNT(*) as c FROM all_filtered
        WHERE match_condition AND match_foil AND match_treatment AND match_language AND match_color AND match_collection AND match_set
        GROUP BY val
    ),
    f_language AS (
        SELECT COALESCE(LOWER(language), 'en') as val, COUNT(*) as c FROM all_filtered
        WHERE match_condition AND match_foil AND match_treatment AND match_rarity AND match_color AND match_collection AND match_set
        GROUP BY val
    ),
    f_color AS (
        SELECT 
            COUNT(*) FILTER (WHERE color_identity ILIKE '%W%') as w,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%U%') as u,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%B%') as b,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%R%') as r,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%G%') as g,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%C%') as c
        FROM all_filtered
        WHERE match_condition AND match_foil AND match_treatment AND match_rarity AND match_language AND match_collection AND match_set
    ),
    f_treatment AS (
        SELECT val, SUM(c) as c FROM (
            SELECT COALESCE(LOWER(card_treatment), 'normal') as val, COUNT(*) as c FROM all_filtered
            WHERE match_condition AND match_foil AND match_rarity AND match_language AND match_color AND match_collection AND match_set
            GROUP BY val
            UNION ALL
            SELECT 'full_art' as val, COUNT(*) as c FROM all_filtered
            WHERE full_art = true AND match_condition AND match_foil AND match_rarity AND match_language AND match_color AND match_collection AND match_set
            UNION ALL
            SELECT 'textless' as val, COUNT(*) as c FROM all_filtered
            WHERE textless = true AND match_condition AND match_foil AND match_rarity AND match_language AND match_color AND match_collection AND match_set
        ) t GROUP BY val
    ),
    f_collection AS (
        SELECT COALESCE(cc.slug, 'unknown') as val, COUNT(DISTINCT all_filtered.id) as c
        FROM all_filtered
        JOIN product_category pc ON all_filtered.id = pc.product_id
        JOIN custom_category cc ON pc.category_id = cc.id
        WHERE match_condition AND match_foil AND match_treatment AND match_rarity AND match_language AND match_color AND match_set
        GROUP BY val
    ),
    f_set_name AS (
        SELECT 
            COALESCE(p.set_name, 'Unknown') as val, 
            COUNT(*) as c,
            MAX(s.released_at) as release_date
        FROM all_filtered p
        LEFT JOIN tcg_set s ON (LOWER(p.set_name) = LOWER(s.name) AND p.tcg = s.tcg) OR (LOWER(p.set_code) = LOWER(s.code) AND p.tcg = s.tcg)
        WHERE match_condition AND match_foil AND match_treatment AND match_rarity AND match_language AND match_color AND match_collection
        GROUP BY val
        HAVING COUNT(*) > 0
        ORDER BY release_date DESC NULLS LAST, val ASC
        LIMIT 50
    )
    SELECT jsonb_build_object(
        'condition', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_condition),
        'foil', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_foil),
        'rarity', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_rarity),
        'language', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_language),
        'color', (SELECT jsonb_build_object('W', w, 'U', u, 'B', b, 'R', r, 'G', g, 'C', c) FROM f_color),
        'treatment', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_treatment),
        'collection', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_collection),
        'set_name', (SELECT COALESCE(jsonb_agg(jsonb_build_object('id', val, 'label', val, 'count', c)), '[]'::jsonb) FROM f_set_name)
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
