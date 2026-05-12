-- fn_get_product_facets
-- Returns a JSONB object containing counts for Condition, Foil, Treatment, Rarity, Language, Color, Collection, and Set.
-- Filter logic: always AND across dimensions. OR/AND mode only affects within-dimension
-- behavior for multi-value fields (Color, Collection, Format).
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
    p_is_admin BOOLEAN DEFAULT false,
    p_is_legendary TEXT DEFAULT '',
    p_is_land TEXT DEFAULT '',
    p_is_historic TEXT DEFAULT '',
    p_format TEXT DEFAULT '',
    p_frame_effects TEXT DEFAULT '',
    p_card_types TEXT DEFAULT ''
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
    v_format_arr TEXT[];
    v_frame_effects_arr TEXT[];
    v_card_types_arr TEXT[];
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
    v_format_arr := CASE WHEN p_format = '' THEN NULL ELSE string_to_array(LOWER(p_format), ',') END;
    v_frame_effects_arr := CASE WHEN p_frame_effects = '' THEN NULL ELSE string_to_array(LOWER(p_frame_effects), ',') END;
    v_card_types_arr := CASE WHEN p_card_types = '' THEN NULL ELSE string_to_array(p_card_types, ',') END;

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
    active_filters AS (
        SELECT 
            v_foil_arr IS NOT NULL as has_foil,
            v_treatment_arr IS NOT NULL as has_treatment,
            v_condition_arr IS NOT NULL as has_condition,
            v_rarity_arr IS NOT NULL as has_rarity,
            v_language_arr IS NOT NULL as has_language,
            v_color_arr IS NOT NULL as has_color,
            v_collection_arr IS NOT NULL as has_collection,
            v_set_name_arr IS NOT NULL as has_set,
            p_is_legendary != '' as has_legendary,
            p_is_land != '' as has_land,
            p_is_historic != '' as has_historic,
            v_format_arr IS NOT NULL as has_format,
            v_frame_effects_arr IS NOT NULL as has_frame_effects,
            v_card_types_arr IS NOT NULL as has_card_types
    ),
    all_filtered AS (
        SELECT *,
               -- Single-value fields: always OR within category
               (v_foil_arr IS NULL OR LOWER(foil_treatment) = ANY(v_foil_arr)) as match_foil,
               (v_treatment_arr IS NULL OR LOWER(card_treatment) = ANY(v_treatment_arr) OR (full_art AND 'full_art' = ANY(v_treatment_arr)) OR (textless AND 'textless' = ANY(v_treatment_arr))) as match_treatment,
               (v_condition_arr IS NULL OR UPPER(condition) = ANY(v_condition_arr)) as match_condition,
               (v_rarity_arr IS NULL OR LOWER(rarity) = ANY(v_rarity_arr)) as match_rarity,
               (v_language_arr IS NULL OR LOWER(language) = ANY(v_language_arr)) as match_language,
               (v_set_name_arr IS NULL OR set_name = ANY(v_set_name_arr)) as match_set,
               (p_is_legendary = '' OR (p_is_legendary = 'true' AND is_legendary = true) OR (p_is_legendary = 'false' AND is_legendary = false)) as match_legendary,
               (p_is_land = '' OR (p_is_land = 'true' AND is_land = true) OR (p_is_land = 'false' AND is_land = false)) as match_land,
               (p_is_historic = '' OR (p_is_historic = 'true' AND is_historic = true) OR (p_is_historic = 'false' AND is_historic = false)) as match_historic,
               -- Multi-value fields: OR/AND within category based on p_filter_logic
                (v_color_arr IS NULL OR (
                    CASE WHEN p_filter_logic = 'and' 
                    THEN (SELECT bool_and(color_identity ILIKE '%' || c || '%') FROM unnest(v_color_arr) c)
                    ELSE (SELECT bool_or(color_identity ILIKE '%' || c || '%') FROM unnest(v_color_arr) c)
                    END
                )) as match_color,
               (v_collection_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (SELECT COUNT(DISTINCT cc.slug) FROM product_category pc JOIN custom_category cc ON pc.category_id = cc.id WHERE pc.product_id = base_products.id AND cc.slug = ANY(v_collection_arr)) = array_length(v_collection_arr, 1)
                   ELSE EXISTS (SELECT 1 FROM product_category pc JOIN custom_category cc ON pc.category_id = cc.id WHERE pc.product_id = base_products.id AND cc.slug = ANY(v_collection_arr))
                   END
               )) as match_collection,
               (v_format_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (SELECT bool_and(legalities->>f = 'legal') FROM unnest(v_format_arr) f)
                   ELSE EXISTS (SELECT 1 FROM unnest(v_format_arr) f WHERE legalities->>f = 'legal')
                   END
               )) as match_format,
               (v_frame_effects_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (frame_effects @> to_jsonb(v_frame_effects_arr)) -- Approximate for AND, usually frame_effects is small
                   ELSE (frame_effects ?| v_frame_effects_arr)
                   END
               )) as match_frame_effects,
               (v_card_types_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (card_types @> to_jsonb(v_card_types_arr))
                   ELSE (card_types ?| v_card_types_arr)
                   END
               )) as match_card_types
        FROM base_products
    ),
    -- Always AND across dimensions
    filter_matches AS (
        SELECT *,
               ((NOT (SELECT has_foil FROM active_filters) OR match_foil) AND
                (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND
                (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND
                (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND
                (NOT (SELECT has_language FROM active_filters) OR match_language) AND
                (NOT (SELECT has_color FROM active_filters) OR match_color) AND
                (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND
                (NOT (SELECT has_set FROM active_filters) OR match_set) AND
                (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND
                (NOT (SELECT has_land FROM active_filters) OR match_land) AND
                (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND
                (NOT (SELECT has_format FROM active_filters) OR match_format) AND
                (NOT (SELECT has_frame_effects FROM active_filters) OR match_frame_effects) AND
                (NOT (SELECT has_card_types FROM active_filters) OR match_card_types))
               as match_all_filters
        FROM all_filtered
    ),
    dimension_matches AS (
        SELECT *,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_foil,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_treatment,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_condition,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_rarity,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_language,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_color,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_collection,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_set,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_format,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_card_types FROM active_filters) OR match_card_types)
               END as others_frame_effects,
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format) AND (NOT (SELECT has_frame_effects FROM active_filters) OR match_frame_effects)
               END as others_card_types
        FROM filter_matches
    ),
    f_condition AS (
        SELECT COALESCE(condition, 'unknown') as val, COUNT(*) as c FROM dimension_matches WHERE others_condition GROUP BY val
    ),
    f_foil AS (
        SELECT COALESCE(LOWER(foil_treatment), 'non_foil') as val, COUNT(*) as c FROM dimension_matches WHERE others_foil GROUP BY val
    ),
    f_rarity AS (
        SELECT COALESCE(LOWER(rarity), 'unknown') as val, COUNT(*) as c FROM dimension_matches WHERE others_rarity GROUP BY val
    ),
    f_language AS (
        SELECT COALESCE(LOWER(language), 'en') as val, COUNT(*) as c FROM dimension_matches WHERE others_language GROUP BY val
    ),
    f_color AS (
        SELECT 
            COUNT(*) FILTER (WHERE color_identity ILIKE '%W%') as w,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%U%') as u,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%B%') as b,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%R%') as r,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%G%') as g,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%C%') as c
        FROM dimension_matches WHERE others_color
    ),
    f_treatment AS (
        SELECT val, SUM(c) as c FROM (
            SELECT COALESCE(LOWER(card_treatment), 'normal') as val, COUNT(*) as c FROM dimension_matches WHERE others_treatment GROUP BY val
            UNION ALL
            SELECT 'full_art' as val, COUNT(*) as c FROM dimension_matches WHERE full_art = true AND others_treatment
            UNION ALL
            SELECT 'textless' as val, COUNT(*) as c FROM dimension_matches WHERE textless = true AND others_treatment
        ) t GROUP BY val
    ),
    f_collection AS (
        SELECT COALESCE(cc.slug, 'unknown') as val, COUNT(DISTINCT dimension_matches.id) as c
        FROM dimension_matches JOIN product_category pc ON dimension_matches.id = pc.product_id JOIN custom_category cc ON pc.category_id = cc.id
        WHERE others_collection GROUP BY val
    ),
    f_set_name AS (
        SELECT COALESCE(p.set_name, 'Unknown') as val, COUNT(*) as c, MAX(s.released_at) as release_date
        FROM dimension_matches p LEFT JOIN tcg_set s ON (LOWER(p.set_name) = LOWER(s.name) AND p.tcg = s.tcg) OR (LOWER(p.set_code) = LOWER(s.code) AND p.tcg = s.tcg)
        WHERE others_set GROUP BY val HAVING COUNT(*) > 0 ORDER BY release_date DESC NULLS LAST, val ASC LIMIT 50
    ),
    f_legendary AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_legendary = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_land AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_land = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_historic AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_historic = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_full_art AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE full_art = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_textless AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE textless = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_basic_land AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_basic_land = true AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_non_basic_land AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_land = true AND is_basic_land = false AND others_foil AND others_treatment AND others_rarity AND others_language AND others_color AND others_color AND others_collection AND others_set AND others_condition AND others_format
    ),
    f_format AS (
        SELECT f as val, COUNT(*) as c FROM dimension_matches, unnest(ARRAY['commander', 'modern', 'standard', 'legacy', 'vintage', 'pauper', 'pioneer']) f
        WHERE legalities->>f = 'legal' AND others_format GROUP BY val
    ),
    f_card_types AS (
        SELECT val, SUM(c) as c FROM (
            SELECT jsonb_array_elements_text(card_types) as val, COUNT(*) as c 
            FROM dimension_matches 
            WHERE others_card_types 
            GROUP BY val
        ) t GROUP BY val
    ),
    f_frame_effects AS (
        SELECT fe as val, COUNT(*) as c FROM dimension_matches, jsonb_array_elements_text(frame_effects) fe WHERE others_frame_effects GROUP BY val
    )
    SELECT jsonb_build_object(
        'condition', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_condition),
        'foil', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_foil),
        'rarity', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_rarity),
        'language', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_language),
        'color', (SELECT jsonb_build_object('W', w, 'U', u, 'B', b, 'R', r, 'G', g, 'C', c) FROM f_color),
        'treatment', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_treatment),
        'frame_effects', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_frame_effects),
        'card_types', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_card_types),
        'collection', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_collection),
        'set_name', (SELECT COALESCE(jsonb_agg(jsonb_build_object('id', val, 'label', val, 'count', c)), '[]'::jsonb) FROM f_set_name),
        'is_legendary', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_legendary),
        'is_land', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_land),
        'is_historic', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_historic),
        'full_art', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_full_art),
        'textless', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_textless),
        'is_basic_land', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_basic_land),
        'is_non_basic_land', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_non_basic_land),
        'format', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_format)
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
