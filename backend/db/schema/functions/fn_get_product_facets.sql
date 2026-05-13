-- fn_get_product_facets
-- Returns a JSONB object containing counts for Condition, Foil, Treatment, Rarity, Language, Color, Collection, and Set.
-- Filter logic: always AND across dimensions in NARROW mode. OR across dimensions in BROAD mode.
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
    p_is_prepared TEXT DEFAULT '',
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
    -- Pre-parse filters into arrays for faster matching (handling potential spaces after commas)
    v_foil_arr := CASE WHEN p_foil = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_foil), '\s*,\s*') END;
    v_treatment_arr := CASE WHEN p_treatment = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_treatment), '\s*,\s*') END;
    v_condition_arr := CASE WHEN p_condition = '' THEN NULL ELSE regexp_split_to_array(UPPER(p_condition), '\s*,\s*') END;
    v_rarity_arr := CASE WHEN p_rarity = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_rarity), '\s*,\s*') END;
    v_language_arr := CASE WHEN p_language = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_language), '\s*,\s*') END;
    v_color_arr := CASE WHEN p_color = '' THEN NULL ELSE regexp_split_to_array(UPPER(p_color), '\s*,\s*') END;
    v_collection_arr := CASE WHEN p_collection = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_collection), '\s*,\s*') END;
    v_set_name_arr := CASE WHEN p_set_name = '' THEN NULL ELSE regexp_split_to_array(p_set_name, '\s*,\s*') END;
    v_format_arr := CASE WHEN p_format = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_format), '\s*,\s*') END;
    v_frame_effects_arr := CASE WHEN p_frame_effects = '' THEN NULL ELSE regexp_split_to_array(LOWER(p_frame_effects), '\s*,\s*') END;
    v_card_types_arr := CASE WHEN p_card_types = '' THEN NULL ELSE regexp_split_to_array(p_card_types, '\s*,\s*') END;

    -- Hoist active filter checks into variables for performance
    DECLARE
        has_foil BOOLEAN := v_foil_arr IS NOT NULL;
        has_treatment BOOLEAN := v_treatment_arr IS NOT NULL;
        has_condition BOOLEAN := v_condition_arr IS NOT NULL;
        has_rarity BOOLEAN := v_rarity_arr IS NOT NULL;
        has_language BOOLEAN := v_language_arr IS NOT NULL;
        has_color BOOLEAN := v_color_arr IS NOT NULL;
        has_collection BOOLEAN := v_collection_arr IS NOT NULL;
        has_set BOOLEAN := v_set_name_arr IS NOT NULL;
        has_properties BOOLEAN := (p_is_legendary != '' OR p_is_land != '' OR p_is_historic != '' OR p_is_prepared != '');
        has_format BOOLEAN := v_format_arr IS NOT NULL;
        has_frame_effects BOOLEAN := v_frame_effects_arr IS NOT NULL;
        has_card_types BOOLEAN := v_card_types_arr IS NOT NULL;
    BEGIN

    WITH base_products AS MATERIALIZED (
        SELECT p.*
        FROM product p
        LEFT JOIN tcg t ON p.tcg = t.id
        WHERE 
            (p_tcg = '' OR LOWER(p.tcg) = LOWER(p_tcg))
            AND (p_category = '' OR p.category = p_category)
            AND (p_search = '' OR (p.search_vector @@ websearch_to_tsquery('english', p_search) OR p.name ILIKE '%' || p_search || '%'))
            AND (p_storage_id = '' OR EXISTS (
                SELECT 1 FROM product_storage ps 
                WHERE ps.product_id = p.id AND ps.storage_id::text = p_storage_id AND ps.quantity > 0
            ))
            AND (NOT p_in_stock OR p.stock > 0)
            AND (p_is_admin OR (t.is_active IS NULL OR t.is_active = true))
    ),
    -- Optimization: Pre-calculate product-to-category mappings for faster collection matching
    categories_map AS (
        SELECT pc.product_id, array_agg(cc.slug) as slugs
        FROM product_category pc
        JOIN custom_category cc ON pc.category_id = cc.id
        WHERE EXISTS (SELECT 1 FROM base_products bp WHERE bp.id = pc.product_id)
        GROUP BY pc.product_id
    ),
    all_filtered AS (
        SELECT p.*,
               -- Pre-calculated arrays for native operator matching
               string_to_array(p.color_identity, ',') as color_identity_arr,
               COALESCE(cm.slugs, ARRAY[]::TEXT[]) as collection_slugs,
               -- Single-value fields: always OR within category
               (v_foil_arr IS NULL OR LOWER(p.foil_treatment) = ANY(v_foil_arr)) as match_foil,
               (v_treatment_arr IS NULL OR LOWER(p.card_treatment) = ANY(v_treatment_arr) OR (p.full_art AND 'full_art' = ANY(v_treatment_arr)) OR (p.textless AND 'textless' = ANY(v_treatment_arr))) as match_treatment,
               (v_condition_arr IS NULL OR UPPER(p.condition) = ANY(v_condition_arr)) as match_condition,
               (v_rarity_arr IS NULL OR LOWER(p.rarity) = ANY(v_rarity_arr)) as match_rarity,
               (v_language_arr IS NULL OR LOWER(p.language) = ANY(v_language_arr)) as match_language,
               (v_set_name_arr IS NULL OR p.set_name = ANY(v_set_name_arr)) as match_set,
               -- Grouped properties match logic (respects OR/AND mode)
               (
                 CASE WHEN p_filter_logic = 'and'
                 THEN (p_is_legendary = '' OR (p_is_legendary = 'true' AND p.is_legendary = true) OR (p_is_legendary = 'false' AND p.is_legendary = false)) AND
                      (p_is_land = '' OR (p_is_land = 'true' AND p.is_land = true) OR (p_is_land = 'false' AND p.is_land = false)) AND
                      (p_is_historic = '' OR (p_is_historic = 'true' AND p.is_historic = true) OR (p_is_historic = 'false' AND p.is_historic = false)) AND
                      (p_is_prepared = '' OR (p_is_prepared = 'true' AND p.is_prepared = true))
                 ELSE (p_is_legendary = 'true' AND p.is_legendary = true) OR
                      (p_is_land = 'true' AND p.is_land = true) OR
                      (p_is_historic = 'true' AND p.is_historic = true) OR
                      (p_is_prepared = 'true' AND p.is_prepared = true) OR
                      (p_is_legendary = '' AND p_is_land = '' AND p_is_historic = '' AND p_is_prepared = '')
                 END
               ) as match_properties,
               -- Multi-value fields: Optimized with array/JSONB operators
               (v_color_arr IS NULL OR (
                    CASE WHEN p_filter_logic = 'and' 
                    THEN (string_to_array(p.color_identity, ',') @> v_color_arr)
                    ELSE (string_to_array(p.color_identity, ',') && v_color_arr)
                    END
                )) as match_color,
               (v_collection_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (COALESCE(cm.slugs, ARRAY[]::TEXT[]) @> v_collection_arr)
                   ELSE (COALESCE(cm.slugs, ARRAY[]::TEXT[]) && v_collection_arr)
                   END
               )) as match_collection,
               (v_format_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (SELECT bool_and(p.legalities->>f = 'legal') FROM unnest(v_format_arr) f) -- Still needed for logic but limited rows
                   ELSE (p.legalities ?| v_format_arr)
                   END
               )) as match_format,
               (v_frame_effects_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (p.frame_effects @> to_jsonb(v_frame_effects_arr)) 
                   ELSE (p.frame_effects ?| v_frame_effects_arr)
                   END
               )) as match_frame_effects,
               (v_card_types_arr IS NULL OR (
                   CASE WHEN p_filter_logic = 'and'
                   THEN (p.card_types @> to_jsonb(v_card_types_arr))
                   ELSE (p.card_types ?| v_card_types_arr)
                   END
               )) as match_card_types
        FROM base_products p
        LEFT JOIN categories_map cm ON p.id = cm.product_id
    ),
    filter_matches AS (
        SELECT *,
               CASE WHEN p_filter_logic = 'and' THEN
                   ((NOT has_foil OR match_foil) AND
                    (NOT has_treatment OR match_treatment) AND
                    (NOT has_condition OR match_condition) AND
                    (NOT has_rarity OR match_rarity) AND
                    (NOT has_language OR match_language) AND
                    (NOT has_color OR match_color) AND
                    (NOT has_collection OR match_collection) AND
                    (NOT has_set OR match_set) AND
                    (NOT has_properties OR match_properties) AND
                    (NOT has_format OR match_format) AND
                    (NOT has_frame_effects OR match_frame_effects) AND
                    (NOT has_card_types OR match_card_types))
               ELSE
                   ((match_foil AND has_foil) OR
                    (match_treatment AND has_treatment) OR
                    (match_condition AND has_condition) OR
                    (match_rarity AND has_rarity) OR
                    (match_language AND has_language) OR
                    (match_color AND has_color) OR
                    (match_collection AND has_collection) OR
                    (match_set AND has_set) OR
                    (match_properties AND has_properties) OR
                    (match_format AND has_format) OR
                    (match_frame_effects AND has_frame_effects) OR
                    (match_card_types AND has_card_types) OR
                    (NOT has_foil AND NOT has_treatment AND NOT has_condition AND NOT has_rarity AND NOT has_language AND NOT has_color AND NOT has_collection AND NOT has_set AND NOT has_properties AND NOT has_format AND NOT has_frame_effects AND NOT has_card_types))
               END as match_all_filters
        FROM all_filtered
    ),
    dimension_matches AS (
        SELECT *,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_condition OR match_condition) AND (NOT has_treatment OR match_treatment) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_foil,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_treatment,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_condition,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_rarity,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_language,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_color,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_collection,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_set,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_properties,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_frame_effects OR match_frame_effects) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_format,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_card_types OR match_card_types)
               ELSE TRUE
               END as others_frame_effects,
               CASE WHEN p_filter_logic = 'and' THEN (NOT has_foil OR match_foil) AND (NOT has_treatment OR match_treatment) AND (NOT has_condition OR match_condition) AND (NOT has_rarity OR match_rarity) AND (NOT has_language OR match_language) AND (NOT has_color OR match_color) AND (NOT has_collection OR match_collection) AND (NOT has_set OR match_set) AND (NOT has_properties OR match_properties) AND (NOT has_format OR match_format) AND (NOT has_frame_effects OR match_frame_effects)
               ELSE TRUE
               END as others_card_types
        FROM filter_matches
    ),
    f_condition AS (
        SELECT condition as val, COUNT(*) as c FROM dimension_matches WHERE others_condition AND condition IS NOT NULL GROUP BY val
    ),
    f_foil AS (
        SELECT foil_treatment as val, COUNT(*) as c FROM dimension_matches WHERE others_foil AND foil_treatment IS NOT NULL GROUP BY val
    ),
    f_treatment AS (
        SELECT card_treatment as val, COUNT(*) as c FROM dimension_matches WHERE others_treatment AND card_treatment IS NOT NULL GROUP BY val
    ),
    f_rarity AS (
        SELECT rarity as val, COUNT(*) as c FROM dimension_matches WHERE others_rarity AND rarity IS NOT NULL GROUP BY val
    ),
    f_language AS (
        SELECT language as val, COUNT(*) as c FROM dimension_matches WHERE others_language AND language IS NOT NULL GROUP BY val
    ),
    f_color AS (
        SELECT 
            COUNT(*) FILTER (WHERE color_identity ILIKE '%W%') as w,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%U%') as u,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%B%') as b,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%R%') as r,
            COUNT(*) FILTER (WHERE color_identity ILIKE '%G%') as g,
            COUNT(*) FILTER (WHERE color_identity IS NULL OR color_identity = '') as c
        FROM dimension_matches 
        WHERE others_color
    ),
    f_collection AS (
        SELECT cc.slug as val, COUNT(DISTINCT p.id) as c 
        FROM dimension_matches p
        JOIN product_category pc ON p.id = pc.product_id
        JOIN custom_category cc ON pc.category_id = cc.id
        WHERE others_collection AND cc.slug IS NOT NULL GROUP BY val
    ),
    f_set_counts AS (
        SELECT COALESCE(p.set_name, 'Unknown') as val, COUNT(*) as c, MAX(p.set_code) as set_code, MAX(p.tcg) as tcg
        FROM dimension_matches p
        WHERE others_set GROUP BY val HAVING COUNT(*) > 0
    ),
    f_set_name AS (
        SELECT f.val, f.c, MAX(s.released_at) as release_date
        FROM f_set_counts f
        LEFT JOIN tcg_set s ON (LOWER(f.val) = LOWER(s.name) AND f.tcg = s.tcg) OR (LOWER(f.set_code) = LOWER(s.code) AND f.tcg = s.tcg)
        GROUP BY f.val, f.c
        ORDER BY release_date DESC NULLS LAST, f.val ASC
        LIMIT 50
    ),
    f_legendary AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_legendary = true AND others_properties
    ),
    f_land AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_land = true AND others_properties
    ),
    f_historic AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_historic = true AND others_properties
    ),
    f_prepared AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE is_prepared = true AND others_properties
    ),
    f_full_art AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE full_art = true AND others_properties
    ),
    f_textless AS (
        SELECT 'true' as val, COUNT(*) as c FROM dimension_matches WHERE textless = true AND others_properties
    ),
    f_land_type AS (
        SELECT 'basic' as val, COUNT(*) as c FROM dimension_matches WHERE is_basic_land = true AND others_properties
        UNION ALL
        SELECT 'non-basic' as val, COUNT(*) as c FROM dimension_matches WHERE is_land = true AND is_basic_land = false AND others_properties
    ),
    f_format AS (
        SELECT f as val, COUNT(*) as c FROM dimension_matches, unnest(ARRAY['commander', 'modern', 'standard', 'legacy', 'vintage', 'pauper', 'pioneer']) f
        WHERE legalities->>f = 'legal' AND others_format GROUP BY val
    ),
    f_card_types AS (
        SELECT val, SUM(c) as c FROM (
            SELECT jsonb_array_elements_text(card_types) as val, COUNT(*) as c 
            FROM dimension_matches 
            WHERE others_card_types AND card_types IS NOT NULL
            GROUP BY val
        ) t WHERE val IS NOT NULL GROUP BY val
    ),
    f_frame_effects AS (
        SELECT fe as val, COUNT(*) as c FROM dimension_matches, jsonb_array_elements_text(frame_effects) fe WHERE others_frame_effects AND fe IS NOT NULL GROUP BY val
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
        'is_prepared', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_prepared),
        'is_land', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_land),
        'is_historic', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_historic),
        'full_art', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_full_art),
        'textless', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_textless),
        'land_type', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_land_type),
        'format', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_format)
    ) INTO result;

    RETURN result;
END; -- End of performance-optimization block
END; -- End of function
$$ LANGUAGE plpgsql STABLE;
