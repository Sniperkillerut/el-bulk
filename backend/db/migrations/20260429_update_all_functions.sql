-- Updating all functions and triggers
-- fn_accept_client_request
-- Atomically accepts a client_request:
--   1. Find an existing active bounty matching card identity (via oracle_id or scryfall_id).
--   2. If none, create a new bounty.
--   3. Link the request to the bounty, increment bounty.quantity_needed, set request status = 'accepted'.
CREATE OR REPLACE FUNCTION fn_accept_client_request(
    p_request_id UUID
) RETURNS JSONB AS $$
DECLARE
    v_req         client_request%ROWTYPE;
    v_bounty_id   UUID;
    v_is_generic  BOOLEAN;
BEGIN
    -- Lock and fetch the request
    SELECT * INTO v_req FROM client_request WHERE id = p_request_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'request not found: %', p_request_id;
    END IF;

    v_is_generic := (v_req.match_type = 'any');

    IF v_is_generic THEN
        -- Find existing generic bounty (oracle_id match preferred, fallback to name)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE is_generic = true
          AND is_active = true
          AND (
            (oracle_id IS NOT NULL AND v_req.oracle_id IS NOT NULL AND oracle_id = v_req.oracle_id)
            OR 
            (lower(trim(name)) = lower(trim(v_req.card_name)) AND tcg = v_req.tcg)
          )
        ORDER BY (oracle_id IS NOT NULL AND v_req.oracle_id IS NOT NULL AND oracle_id = v_req.oracle_id) DESC, created_at DESC
        LIMIT 1;
    ELSE
        -- Find existing specific bounty (scryfall_id match preferred)
        SELECT id INTO v_bounty_id
        FROM bounty
        WHERE is_generic = false
          AND is_active = true
          AND (
            (scryfall_id IS NOT NULL AND v_req.scryfall_id IS NOT NULL AND v_req.scryfall_id != '' AND scryfall_id = v_req.scryfall_id::UUID)
            OR
            (lower(trim(name)) = lower(trim(v_req.card_name)) AND tcg = v_req.tcg AND (set_name IS NOT DISTINCT FROM v_req.set_name))
          )
        ORDER BY (scryfall_id IS NOT NULL AND v_req.scryfall_id IS NOT NULL AND v_req.scryfall_id != '' AND scryfall_id = v_req.scryfall_id::UUID) DESC, created_at DESC
        LIMIT 1;
    END IF;

    -- Create bounty if not found
    IF v_bounty_id IS NULL THEN
        INSERT INTO bounty (
            name, tcg, set_name, quantity_needed, is_active, is_generic,
            scryfall_id, oracle_id, image_url, set_code, collector_number,
            foil_treatment, card_treatment, language, hide_price, price_source
        ) VALUES (
            trim(v_req.card_name), v_req.tcg, v_req.set_name, v_req.quantity,
            true, v_is_generic,
            v_req.scryfall_id::UUID, v_req.oracle_id, v_req.image_url, v_req.set_code, v_req.collector_number,
            COALESCE(v_req.foil_treatment, 'non_foil'), COALESCE(v_req.card_treatment, 'normal'), 
            'en', false, 'tcgplayer'
        )
        RETURNING id INTO v_bounty_id;
    ELSE
        -- Increment quantity on existing bounty
        UPDATE bounty
        SET quantity_needed = quantity_needed + v_req.quantity,
            updated_at = now()
        WHERE id = v_bounty_id;
    END IF;

    -- Link request to bounty and mark as accepted
    UPDATE client_request
    SET bounty_id = v_bounty_id,
        status = 'accepted'
    WHERE id = p_request_id;

    RETURN jsonb_build_object(
        'bounty_id', v_bounty_id,
        'request_id', p_request_id,
        'is_generic', v_is_generic
    );
END;
$$ LANGUAGE plpgsql;
-- Bulk Upsert Product
-- Handles product creation/update, category mapping, and storage assignment in one call.
-- Updated to support "Attribute Matching" for CSV imports:
-- If no ID is provided, it attempts to find an existing variant to increment stock instead of creating duplicates.
CREATE OR REPLACE FUNCTION fn_bulk_upsert_product(data jsonb)
RETURNS TABLE(upserted_id UUID) AS $$
DECLARE
    item jsonb;
    p_id UUID;
    is_import BOOLEAN;
BEGIN
    FOR item IN SELECT * FROM jsonb_array_elements(data)
    LOOP
        is_import := (item->>'id' IS NULL);
        p_id := (item->>'id')::uuid;

        -- 1. Variant Matching for Imports
        -- If no ID is provided, try to find an exact variant match to avoid duplicates.
        -- Harden matching using NULLIF to handle empty strings as NULLs for optional attributes.
        IF is_import THEN
            SELECT id INTO p_id FROM product
            WHERE name = item->>'name'
              AND tcg = item->>'tcg'
              AND category = item->>'category'
              AND COALESCE(set_code, '') = COALESCE(NULLIF(item->>'set_code', ''), '')
              AND COALESCE(collector_number, '') = COALESCE(NULLIF(item->>'collector_number', ''), '')
              AND condition = item->>'condition'
              AND foil_treatment = COALESCE(NULLIF(item->>'foil_treatment', ''), 'non_foil')
              AND card_treatment = COALESCE(NULLIF(item->>'card_treatment', ''), 'normal')
              AND language = COALESCE(NULLIF(item->>'language', ''), 'en')
            LIMIT 1;
        END IF;

        -- 2. Upsert Product Metadata
        -- Use the found p_id or generate a new one if it's still NULL.
        INSERT INTO product (
            id, name, tcg, category, set_name, set_code, collector_number, condition,
            foil_treatment, card_treatment, promo_type,
            price_reference, price_source, price_cop_override,
            image_url, description, language, color_identity, rarity, cmc,
            is_legendary, is_historic, is_land, is_basic_land, art_variation,
            oracle_text, artist, type_line, border_color, frame, full_art, textless,
            scryfall_id, legalities, cost_basis_cop, updated_at
        )
        VALUES (
            COALESCE(p_id, gen_random_uuid()),
            item->>'name',
            item->>'tcg',
            item->>'category',
            item->>'set_name',
            item->>'set_code',
            item->>'collector_number',
            item->>'condition',
            COALESCE(NULLIF(item->>'foil_treatment', ''), 'non_foil'),
            COALESCE(NULLIF(item->>'card_treatment', ''), 'normal'),
            item->>'promo_type',
            (item->>'price_reference')::numeric,
            COALESCE(NULLIF(item->>'price_source', ''), 'manual'),
            (item->>'price_cop_override')::numeric,
            item->>'image_url',
            item->>'description',
            COALESCE(NULLIF(item->>'language', ''), 'en'),
            item->>'color_identity',
            item->>'rarity',
            (item->>'cmc')::numeric,
            COALESCE((item->>'is_legendary')::boolean, false),
            COALESCE((item->>'is_historic')::boolean, false),
            COALESCE((item->>'is_land')::boolean, false),
            COALESCE((item->>'is_basic_land')::boolean, false),
            item->>'art_variation',
            item->>'oracle_text',
            item->>'artist',
            item->>'type_line',
            item->>'border_color',
            item->>'frame',
            COALESCE((item->>'full_art')::boolean, false),
            COALESCE((item->>'textless')::boolean, false),
            (item->>'scryfall_id')::uuid,
            item->'legalities',
            COALESCE((COALESCE(item->>'cost_basis_cop', item->>'cost_basis'))::numeric, 0),
            now()
        )
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            tcg = EXCLUDED.tcg,
            category = EXCLUDED.category,
            set_name = EXCLUDED.set_name,
            set_code = EXCLUDED.set_code,
            collector_number = EXCLUDED.collector_number,
            condition = EXCLUDED.condition,
            foil_treatment = EXCLUDED.foil_treatment,
            card_treatment = EXCLUDED.card_treatment,
            promo_type = EXCLUDED.promo_type,
            price_reference = EXCLUDED.price_reference,
            price_source = EXCLUDED.price_source,
            price_cop_override = EXCLUDED.price_cop_override,
            image_url = EXCLUDED.image_url,
            description = EXCLUDED.description,
            language = EXCLUDED.language,
            color_identity = EXCLUDED.color_identity,
            rarity = EXCLUDED.rarity,
            cmc = EXCLUDED.cmc,
            is_legendary = EXCLUDED.is_legendary,
            is_historic = EXCLUDED.is_historic,
            is_land = EXCLUDED.is_land,
            is_basic_land = EXCLUDED.is_basic_land,
            art_variation = EXCLUDED.art_variation,
            oracle_text = EXCLUDED.oracle_text,
            artist = EXCLUDED.artist,
            type_line = EXCLUDED.type_line,
            border_color = EXCLUDED.border_color,
            frame = EXCLUDED.frame,
            full_art = EXCLUDED.full_art,
            textless = EXCLUDED.textless,
            scryfall_id = EXCLUDED.scryfall_id,
            legalities = EXCLUDED.legalities,
            cost_basis_cop = CASE WHEN is_import THEN product.cost_basis_cop + EXCLUDED.cost_basis_cop ELSE EXCLUDED.cost_basis_cop END,
            updated_at = now()
        RETURNING id INTO p_id;

        -- 3. Categories Tracking
        -- For manual edits (ID provided), we replace categories. 
        -- For imports (No ID provided), we just append new ones.
        IF NOT is_import THEN
            DELETE FROM product_category WHERE product_id = p_id;
        END IF;

        IF item ? 'category_ids' THEN
            INSERT INTO product_category (product_id, category_id)
            SELECT p_id, (cat::text)::uuid
            FROM jsonb_array_elements_text(item->'category_ids') AS cat
            ON CONFLICT DO NOTHING;
        END IF;

        -- 4. Storage & Stock Tracking
        -- Handles both 'storage_items' array and top-level 'stored_in_id' + 'stock' fallback.
        -- For manual edits (ID provided), we replace the full storage manifest.
        -- For imports (No ID provided), we ADD the new quantities to existing locations.
        IF NOT is_import THEN
            DELETE FROM product_storage WHERE product_id = p_id;
            
            -- Priority 1: storage_items array
            IF item ? 'storage_items' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                SELECT p_id, (si->>'stored_in_id')::uuid, (si->>'quantity')::int
                FROM jsonb_array_elements(item->'storage_items') AS si
                WHERE (si->>'quantity')::int > 0
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity;
            -- Priority 2: top-level stored_in_id + stock/quantity
            ELSIF item ? 'stored_in_id' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                VALUES (p_id, (item->>'stored_in_id')::uuid, COALESCE((item->>'stock')::int, (item->>'quantity')::int, 0))
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity;
            END IF;
        ELSE
            -- Priority 1: storage_items array
            IF item ? 'storage_items' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                SELECT p_id, (si->>'stored_in_id')::uuid, (si->>'quantity')::int
                FROM jsonb_array_elements(item->'storage_items') AS si
                WHERE (si->>'quantity')::int > 0
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;
            -- Priority 2: top-level stored_in_id + stock/quantity
            ELSIF item ? 'stored_in_id' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                VALUES (p_id, (item->>'stored_in_id')::uuid, COALESCE((item->>'stock')::int, (item->>'quantity')::int, 0))
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;
            END IF;
        END IF;

        -- 5. Deck Cards (Usually only for store_exclusives)
        -- Always replace manifest for deck cards as it's an architectural definition.
        DELETE FROM deck_card WHERE product_id = p_id;
        IF item ? 'deck_cards' THEN
            INSERT INTO deck_card (
                product_id, name, set_code, set_name, collector_number, rarity, quantity, 
                language, type_line, color_identity, cmc, is_legendary, is_historic, is_land, is_basic_land,
                art_variation, oracle_text, artist, border_color, frame, full_art, textless, promo_type,
                image_url, legalities, foil_treatment, card_treatment, scryfall_id
            )
            SELECT 
                p_id,
                dc->>'name',
                dc->>'set_code',
                dc->>'set_name',
                dc->>'collector_number',
                dc->>'rarity',
                COALESCE((dc->>'quantity')::int, 1),
                COALESCE(dc->>'language', 'en'),
                dc->>'type_line',
                dc->>'color_identity',
                (dc->>'cmc')::numeric,
                COALESCE((dc->>'is_legendary')::boolean, false),
                COALESCE((dc->>'is_historic')::boolean, false),
                COALESCE((dc->>'is_land')::boolean, false),
                COALESCE((dc->>'is_basic_land')::boolean, false),
                dc->>'art_variation',
                dc->>'oracle_text',
                dc->>'artist',
                dc->>'border_color',
                dc->>'frame',
                COALESCE((dc->>'full_art')::boolean, false),
                COALESCE((dc->>'textless')::boolean, false),
                dc->>'promo_type',
                dc->>'image_url',
                dc->'legalities',
                COALESCE(dc->>'foil_treatment', 'non_foil'),
                COALESCE(dc->>'card_treatment', 'normal'),
                (dc->>'scryfall_id')::uuid
            FROM jsonb_array_elements(item->'deck_cards') AS dc;
        END IF;

        upserted_id := p_id;
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- fn_cancel_client_request
-- Atomically cancels a client_request:
--   1. If status is 'accepted', decrement bounty.quantity_needed.
--   2. Set status = 'not_needed' and record reason.
CREATE OR REPLACE FUNCTION fn_cancel_client_request(
    p_request_id UUID,
    p_customer_id UUID,
    p_reason TEXT
) RETURNS VOID AS $$
DECLARE
    v_req client_request%ROWTYPE;
BEGIN
    -- Lock and fetch the request
    SELECT * INTO v_req FROM client_request WHERE id = p_request_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'request not found';
    END IF;

    -- Verify ownership
    IF v_req.customer_id IS NOT NULL AND v_req.customer_id != p_customer_id THEN
        RAISE EXCEPTION 'unauthorized: request belongs to another user';
    END IF;

    -- Only pending or accepted can be cancelled by user
    IF v_req.status != 'pending' AND v_req.status != 'accepted' THEN
        -- If already not_needed, do nothing
        IF v_req.status = 'not_needed' THEN
            RETURN;
        END IF;
        RAISE EXCEPTION 'request cannot be cancelled in current status: %', v_req.status;
    END IF;

    -- If accepted, subtract from associated bounty
    IF v_req.status = 'accepted' AND v_req.bounty_id IS NOT NULL THEN
        UPDATE bounty
        SET quantity_needed = GREATEST(0, quantity_needed - v_req.quantity),
            updated_at = now()
        WHERE id = v_req.bounty_id;
    END IF;

    -- Mark as not_needed
    UPDATE client_request
    SET status = 'not_needed',
        cancellation_reason = p_reason,
        updated_at = now()
    WHERE id = p_request_id;

END;
$$ LANGUAGE plpgsql;
-- Confirm Order
-- Atomically decrements stock from admin locations, removes from pending, and updates status.
CREATE OR REPLACE FUNCTION fn_confirm_order(
    p_order_id UUID,
    decrements jsonb
)
RETURNS VOID AS $$
DECLARE
    dec jsonb;
    v_status TEXT;
    v_item RECORD;
    v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
BEGIN
    SELECT status INTO v_status FROM "order" WHERE id = p_order_id;
    IF v_status = 'confirmed' OR v_status = 'completed' OR v_status = 'cancelled' THEN
        RAISE EXCEPTION 'Order is already processed (status: %)', v_status;
    END IF;

    -- 1. Decrement from 'pending' location (Increases product.stock)
    FOR v_item IN SELECT product_id, quantity FROM order_item WHERE order_id = p_order_id AND product_id IS NOT NULL AND quantity > 0
    LOOP
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - v_item.quantity)
        WHERE product_id = v_item.product_id AND storage_id = v_pending_id;
    END LOOP;

    -- 2. Decrement from specified admin locations (Decreases product.stock)
    FOR dec IN SELECT * FROM jsonb_array_elements(decrements)
    LOOP
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - (dec->>'quantity')::int)
        WHERE product_id = (dec->>'product_id')::uuid 
          AND storage_id = (dec->>'stored_in_id')::uuid;
    END LOOP;

    UPDATE "order" 
    SET status = 'confirmed',
        confirmed_at = now()
    WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;
-- fn_fulfill_bounty_offer
-- Atomically fulfills a bounty offer against selected client requests.
--   p_offer_id: the BountyOffer being accepted.
--   p_request_ids: UUID[] of ClientRequests to mark as solved.
CREATE OR REPLACE FUNCTION fn_fulfill_bounty_offer(
    p_offer_id    UUID,
    p_request_ids UUID[]
) RETURNS JSONB AS $$
DECLARE
    v_offer      bounty_offer%ROWTYPE;
    v_fulfilled  INT;
    v_new_qty    INT;
BEGIN
    SELECT * INTO v_offer FROM bounty_offer WHERE id = p_offer_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'offer not found: %', p_offer_id;
    END IF;

    -- Mark offer as accepted
    UPDATE bounty_offer
    SET status = 'accepted', updated_at = now()
    WHERE id = p_offer_id;

    -- Mark selected requests as solved (only those linked to this bounty)
    UPDATE client_request
    SET status = 'solved'
    WHERE id = ANY(p_request_ids)
      AND bounty_id = v_offer.bounty_id;

    GET DIAGNOSTICS v_fulfilled = ROW_COUNT;

    -- Decrement bounty quantity, deactivate if it hits 0
    UPDATE bounty
    SET quantity_needed = GREATEST(0, quantity_needed - v_fulfilled),
        is_active = (quantity_needed - v_fulfilled) > 0,
        updated_at = now()
    WHERE id = v_offer.bounty_id
    RETURNING quantity_needed INTO v_new_qty;

    RETURN jsonb_build_object(
        'offer_id',        p_offer_id,
        'bounty_id',       v_offer.bounty_id,
        'fulfilled',       v_fulfilled,
        'bounty_qty_left', v_new_qty
    );
END;
$$ LANGUAGE plpgsql;
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
                'searchable', cc.searchable,
                'bg_color', cc.bg_color,
                'text_color', cc.text_color,
                'icon', cc.icon
        ))
        FROM product_category pc
        JOIN custom_category cc ON pc.category_id = cc.id
        WHERE pc.product_id = p_id
    ), '[]'::jsonb),
    'deck_cards', COALESCE((
        SELECT jsonb_agg(jsonb_build_object(
            'id', dc.id,
            'name', dc.name,
            'set_code', dc.set_code,
            'collector_number', dc.collector_number,
            'quantity', dc.quantity,
            'type_line', dc.type_line,
            'image_url', dc.image_url,
            'foil_treatment', dc.foil_treatment,
            'card_treatment', dc.card_treatment,
            'rarity', dc.rarity,
            'art_variation', dc.art_variation,
            'scryfall_id', dc.scryfall_id
        ))
        FROM deck_card dc
        WHERE dc.product_id = p_id
    ), '[]'::jsonb)
) INTO result
FROM product p
WHERE p.id = p_id;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
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
    p_format TEXT DEFAULT ''
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
            v_format_arr IS NOT NULL as has_format
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
               )) as match_format
        FROM base_products
    ),
    -- Always AND across dimensions (both OR and AND mode)
    filter_matches AS (
        SELECT *,
               (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND
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
               (NOT (SELECT has_format FROM active_filters) OR match_format)
               as match_all_filters
        FROM all_filtered
    ),
    -- For each facet dimension: in AND mode, include ALL filters (so impossible options show count=0).
    -- In OR mode, exclude self dimension (standard faceted search: show alternatives).
    dimension_matches AS (
        SELECT *,
               -- Match others for Foil
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_foil,
               -- Match others for Treatment
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_treatment,
               -- Match others for Condition
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_condition,
               -- Match others for Rarity
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_rarity,
               -- Match others for Language
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_language,
               -- Match others for Color
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_color,
               -- Match others for Collection
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_collection,
               -- Match others for Set
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic) AND (NOT (SELECT has_format FROM active_filters) OR match_format)
               END as others_set,
               -- Match others for Format
               CASE WHEN p_filter_logic = 'and' THEN match_all_filters
               ELSE (NOT (SELECT has_foil FROM active_filters) OR match_foil) AND (NOT (SELECT has_treatment FROM active_filters) OR match_treatment) AND (NOT (SELECT has_condition FROM active_filters) OR match_condition) AND (NOT (SELECT has_rarity FROM active_filters) OR match_rarity) AND (NOT (SELECT has_language FROM active_filters) OR match_language) AND (NOT (SELECT has_color FROM active_filters) OR match_color) AND (NOT (SELECT has_collection FROM active_filters) OR match_collection) AND (NOT (SELECT has_set FROM active_filters) OR match_set) AND (NOT (SELECT has_legendary FROM active_filters) OR match_legendary) AND (NOT (SELECT has_land FROM active_filters) OR match_land) AND (NOT (SELECT has_historic FROM active_filters) OR match_historic)
               END as others_format
        FROM filter_matches
    ),
    f_condition AS (
        SELECT COALESCE(condition, 'unknown') as val, COUNT(*) as c FROM dimension_matches
        WHERE others_condition
        GROUP BY val
    ),
    f_foil AS (
        SELECT COALESCE(LOWER(foil_treatment), 'non_foil') as val, COUNT(*) as c FROM dimension_matches
        WHERE others_foil
        GROUP BY val
    ),
    f_rarity AS (
        SELECT COALESCE(LOWER(rarity), 'unknown') as val, COUNT(*) as c FROM dimension_matches
        WHERE others_rarity
        GROUP BY val
    ),
    f_language AS (
        SELECT COALESCE(LOWER(language), 'en') as val, COUNT(*) as c FROM dimension_matches
        WHERE others_language
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
        FROM dimension_matches
        WHERE others_color
    ),
    f_treatment AS (
        SELECT val, SUM(c) as c FROM (
            SELECT COALESCE(LOWER(card_treatment), 'normal') as val, COUNT(*) as c FROM dimension_matches
            WHERE others_treatment
            GROUP BY val
            UNION ALL
            SELECT 'full_art' as val, COUNT(*) as c FROM dimension_matches
            WHERE full_art = true AND others_treatment
            UNION ALL
            SELECT 'textless' as val, COUNT(*) as c FROM dimension_matches
            WHERE textless = true AND others_treatment
        ) t GROUP BY val
    ),
    f_collection AS (
        SELECT COALESCE(cc.slug, 'unknown') as val, COUNT(DISTINCT dimension_matches.id) as c
        FROM dimension_matches
        JOIN product_category pc ON dimension_matches.id = pc.product_id
        JOIN custom_category cc ON pc.category_id = cc.id
        WHERE others_collection
        GROUP BY val
    ),
    f_set_name AS (
        SELECT 
            COALESCE(p.set_name, 'Unknown') as val, 
            COUNT(*) as c,
            MAX(s.released_at) as release_date
        FROM dimension_matches p
        LEFT JOIN tcg_set s ON (LOWER(p.set_name) = LOWER(s.name) AND p.tcg = s.tcg) OR (LOWER(p.set_code) = LOWER(s.code) AND p.tcg = s.tcg)
        WHERE others_set
        GROUP BY val
        HAVING COUNT(*) > 0
        ORDER BY release_date DESC NULLS LAST, val ASC
        LIMIT 50
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
    f_format AS (
        SELECT f as val, COUNT(*) as c FROM dimension_matches, unnest(ARRAY['commander', 'modern', 'standard', 'legacy', 'vintage', 'pauper', 'pioneer']) f
        WHERE legalities->>f = 'legal' AND others_format
        GROUP BY val
    )
    SELECT jsonb_build_object(
        'condition', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_condition),
        'foil', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_foil),
        'rarity', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_rarity),
        'language', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_language),
        'color', (SELECT jsonb_build_object('W', w, 'U', u, 'B', b, 'R', r, 'G', g, 'C', c) FROM f_color),
        'treatment', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_treatment),
        'collection', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_collection),
        'set_name', (SELECT COALESCE(jsonb_agg(jsonb_build_object('id', val, 'label', val, 'count', c)), '[]'::jsonb) FROM f_set_name),
        'is_legendary', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_legendary),
        'is_land', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_land),
        'is_historic', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_historic),
        'format', (SELECT COALESCE(jsonb_object_agg(val, c), '{}'::jsonb) FROM f_format)
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql;
-- Place Order
-- Handles customer upsert, order creation, and item population in one transaction.
CREATE OR REPLACE FUNCTION fn_place_order(
    customer_data jsonb,
    order_items_data jsonb,
    order_meta jsonb
)
RETURNS TABLE(order_id UUID, order_number TEXT) AS $$
DECLARE
    v_customer_id UUID;
    v_order_id UUID;
    v_order_num TEXT;
BEGIN
    -- Guard: Ensure order_items_data is an array
    IF JSONB_TYPEOF(order_items_data) != 'array' THEN
        RAISE EXCEPTION 'order_items_data must be a JSONB array, got %', JSONB_TYPEOF(order_items_data);
    END IF;
    -- Upsert Customer logic based on whether ID is provided
    IF customer_data->>'id' IS NOT NULL AND customer_data->>'id' != '' THEN
        UPDATE customer SET
            first_name = customer_data->>'first_name',
            last_name = customer_data->>'last_name',
            email = customer_data->>'email',
            phone = customer_data->>'phone',
            id_number = customer_data->>'id_number',
            address = customer_data->>'address'
        WHERE id = (customer_data->>'id')::uuid
        RETURNING id INTO v_customer_id;
    END IF;

    IF v_customer_id IS NULL THEN
        INSERT INTO customer (first_name, last_name, email, phone, id_number, address)
        VALUES (
            customer_data->>'first_name',
            customer_data->>'last_name',
            customer_data->>'email',
            customer_data->>'phone',
            customer_data->>'id_number',
            customer_data->>'address'
        )
        ON CONFLICT (phone) DO UPDATE SET
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            email = EXCLUDED.email,
            id_number = EXCLUDED.id_number,
            address = EXCLUDED.address
        RETURNING id INTO v_customer_id;
    END IF;

    -- Create Order
    INSERT INTO "order" (
        order_number, customer_id, status, payment_method, 
        subtotal_cop, shipping_cop, tax_cop, total_cop, 
        is_local_pickup, is_priority, notes
    )
    VALUES (
        order_meta->>'order_number',
        v_customer_id,
        'pending',
        order_meta->>'payment_method',
        (order_meta->>'subtotal_cop')::numeric,
        (order_meta->>'shipping_cop')::numeric,
        (order_meta->>'tax_cop')::numeric,
        (order_meta->>'total_cop')::numeric,
        (order_meta->>'is_local_pickup')::boolean,
        (order_meta->>'is_priority')::boolean,
        order_meta->>'notes'
    )
    RETURNING id, "order".order_number INTO v_order_id, v_order_num;

    -- Insert Order Items
    INSERT INTO order_item (
        order_id, product_id, product_name, product_set, 
        foil_treatment, card_treatment, condition, unit_price_cop, quantity, stored_in_snapshot
    )
    SELECT 
        v_order_id,
        (oi->>'product_id')::uuid,
        oi->>'product_name',
        oi->>'product_set',
        oi->>'foil_treatment',
        oi->>'card_treatment',
        oi->>'condition',
        (oi->>'unit_price_cop')::numeric,
        (oi->>'quantity')::int,
        (oi->'stored_in_snapshot')
    FROM jsonb_array_elements(order_items_data) AS oi;

    -- Add to 'pending' storage location
    INSERT INTO product_storage (product_id, storage_id, quantity)
    SELECT 
        (oi->>'product_id')::uuid,
        (SELECT id FROM storage_location WHERE name = 'pending'),
        (oi->>'quantity')::int
    FROM jsonb_array_elements(order_items_data) AS oi
    WHERE oi->>'product_id' IS NOT NULL AND (oi->>'quantity')::int > 0
    ON CONFLICT (product_id, storage_id) 
    DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;

    order_id := v_order_id;
    order_number := v_order_num;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;
-- Restore Order Stock
-- Adds quantities back to specified physical storage locations for a cancelled order.
CREATE OR REPLACE FUNCTION fn_restore_order_stock(
    p_order_id UUID,
    increments jsonb
)
RETURNS VOID AS $$
DECLARE
    inc jsonb;
    v_status TEXT;
    v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
    v_check RECORD;
    v_restored BOOLEAN;
BEGIN
    SELECT status, inventory_restored INTO v_status, v_restored FROM "order" WHERE id = p_order_id;
    
    IF v_restored THEN
        RAISE EXCEPTION 'Inventory already restored for this order';
    END IF;

    IF v_status != 'cancelled' THEN
        RAISE EXCEPTION 'Only cancelled orders can be restored to stock (current status: %)', v_status;
    END IF;

    -- 1. Validate totals (every product in order must be fully restored)
    FOR v_check IN 
        SELECT 
            oi.product_id, 
            oi.product_name,
            oi.quantity as order_qty,
            COALESCE(inc.inc_qty, 0) as restored_qty
        FROM order_item oi
        LEFT JOIN (
            SELECT (elem->>'product_id')::uuid as pid, SUM((elem->>'quantity')::int) as inc_qty
            FROM jsonb_array_elements(increments) AS elem
            GROUP BY (elem->>'product_id')::uuid
        ) inc ON oi.product_id = inc.pid
        WHERE oi.order_id = p_order_id 
          AND oi.product_id IS NOT NULL 
          AND oi.quantity > 0
    LOOP
        IF v_check.restored_qty != v_check.order_qty THEN
            RAISE EXCEPTION 'Debes restaurar la cantidad total (%) para el producto % (asignado: %)', 
                v_check.order_qty, v_check.product_name, v_check.restored_qty;
        END IF;
    END LOOP;

    -- 2. Iterate through increments and add back to physical locations, and decrement from pending
    FOR inc IN SELECT * FROM jsonb_array_elements(increments)
    LOOP
        -- Physical storage update (increment)
        INSERT INTO product_storage (product_id, storage_id, quantity)
        VALUES (
            (inc->>'product_id')::uuid, 
            (inc->>'stored_in_id')::uuid, 
            (inc->>'quantity')::int
        )
        ON CONFLICT (product_id, storage_id) 
        DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;

        -- Pending storage update (decrement)
        UPDATE product_storage 
        SET quantity = GREATEST(0, quantity - (inc->>'quantity')::int)
        WHERE product_id = (inc->>'product_id')::uuid 
          AND storage_id = v_pending_id;
    END LOOP;

    -- 3. Mark the order as inventory restored
    UPDATE "order" SET inventory_restored = TRUE WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;
-- Submit Bounty Offer
-- Atomically handles customer lookup/linking and offer creation.
CREATE OR REPLACE FUNCTION fn_submit_bounty_offer(
    p_bounty_id UUID,
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_quantity INTEGER,
    p_condition TEXT DEFAULT NULL,
    p_notes TEXT DEFAULT NULL,
    p_status TEXT DEFAULT 'pending',
    p_customer_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_offer_id UUID;
    v_created_at TIMESTAMPTZ;
    v_bounty_name TEXT;
    v_first_name TEXT;
    v_last_name TEXT;
BEGIN
    -- Only lookup/create if no explicit ID provided
    IF v_customer_id IS NULL THEN
        -- Try to find an existing customer by email or phone
        SELECT id INTO v_customer_id 
        FROM customer 
        WHERE email = p_customer_contact OR phone = p_customer_contact 
        LIMIT 1;

        -- If no customer found, create one to ensure data integrity
        IF v_customer_id IS NULL THEN
        v_first_name := split_part(p_customer_name, ' ', 1);
        v_last_name := trim(substring(p_customer_name from char_length(v_first_name) + 1));
        
        IF v_last_name = '' THEN
            v_last_name := '-'; -- Last name is REQUIRED in our schema, so use a hyphen as placeholder if only one name given
        END IF;

        INSERT INTO customer (
            first_name, 
            last_name, 
            email, 
            phone
        ) VALUES (
            v_first_name, 
            v_last_name, 
            CASE WHEN p_customer_contact LIKE '%@%' THEN p_customer_contact ELSE NULL END,
            CASE WHEN p_customer_contact NOT LIKE '%@%' THEN p_customer_contact ELSE NULL END
        ) RETURNING id INTO v_customer_id;
        END IF;
    END IF;
    
    INSERT INTO bounty_offer (bounty_id, customer_id, quantity, condition, notes, status)
    VALUES (p_bounty_id, v_customer_id, p_quantity, p_condition, p_notes, COALESCE(p_status, 'pending'))
    RETURNING id, created_at INTO v_offer_id, v_created_at;

    -- Get bounty name for return object
    SELECT name INTO v_bounty_name FROM bounty WHERE id = p_bounty_id;
    
    RETURN jsonb_build_object(
        'id', v_offer_id,
        'bounty_id', p_bounty_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'bounty_name', v_bounty_name,
        'quantity', p_quantity,
        'condition', p_condition,
        'status', COALESCE(p_status, 'pending'),
        'notes', p_notes,
        'created_at', v_created_at,
        'updated_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
-- Submit Client Request
-- Atomically handles customer lookup/linking and request creation.
CREATE OR REPLACE FUNCTION fn_submit_client_request(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_card_name TEXT,
    p_set_name TEXT DEFAULT NULL,
    p_details TEXT DEFAULT NULL,
    p_quantity INTEGER DEFAULT 1,
    p_tcg TEXT DEFAULT 'mtg',
    p_customer_id UUID DEFAULT NULL,
    p_match_type TEXT DEFAULT 'exact',
    p_scryfall_id TEXT DEFAULT NULL,
    p_image_url TEXT DEFAULT NULL,
    p_foil_treatment TEXT DEFAULT NULL,
    p_card_treatment TEXT DEFAULT NULL,
    p_set_code TEXT DEFAULT NULL,
    p_collector_number TEXT DEFAULT NULL,
    p_oracle_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_request_id UUID;
    v_created_at TIMESTAMPTZ;
    v_first_name TEXT;
    v_last_name TEXT;
BEGIN
    -- Only lookup/create if no explicit ID provided
    IF v_customer_id IS NULL THEN
        -- Try to find an existing customer by email or phone
        SELECT id INTO v_customer_id 
        FROM customer 
        WHERE email = p_customer_contact OR phone = p_customer_contact 
        LIMIT 1;

        -- If no customer found, create one to ensure data integrity
        IF v_customer_id IS NULL THEN
            v_first_name := split_part(p_customer_name, ' ', 1);
            v_last_name := trim(substring(p_customer_name from char_length(v_first_name) + 1));
        
            IF v_last_name = '' THEN
                v_last_name := '-'; -- Last name is REQUIRED in our schema
            END IF;

            INSERT INTO customer (
                first_name, 
                last_name, 
                email, 
                phone
            ) VALUES (
                v_first_name, 
                v_last_name, 
                CASE WHEN p_customer_contact LIKE '%@%' THEN p_customer_contact ELSE NULL END,
                CASE WHEN p_customer_contact NOT LIKE '%@%' THEN p_customer_contact ELSE NULL END
            ) RETURNING id INTO v_customer_id;
        END IF;
    END IF;

    -- Insert the request
    INSERT INTO client_request (
        customer_id, customer_name, customer_contact, card_name, set_name, details, quantity, tcg, status,
        match_type, scryfall_id, image_url, foil_treatment, card_treatment, set_code, collector_number, oracle_id
    )
    VALUES (
        v_customer_id, p_customer_name, p_customer_contact, trim(p_card_name), p_set_name, p_details, p_quantity, lower(trim(p_tcg)), 'pending',
        p_match_type, p_scryfall_id, p_image_url, p_foil_treatment, p_card_treatment, p_set_code, p_collector_number, p_oracle_id
    )
    RETURNING id, created_at INTO v_request_id, v_created_at;
    
    RETURN jsonb_build_object(
        'id', v_request_id,
        'customer_id', v_customer_id,
        'customer_name', p_customer_name,
        'customer_contact', p_customer_contact,
        'card_name', trim(p_card_name),
        'set_name', p_set_name,
        'details', p_details,
        'quantity', p_quantity,
        'tcg', lower(trim(p_tcg)),
        'status', 'pending',
        'match_type', p_match_type,
        'scryfall_id', p_scryfall_id,
        'oracle_id', p_oracle_id,
        'image_url', p_image_url,
        'foil_treatment', p_foil_treatment,
        'card_treatment', p_card_treatment,
        'set_code', p_set_code,
        'collector_number', p_collector_number,
        'created_at', v_created_at
    );
END;
$$ LANGUAGE plpgsql;
-- Submit Client Requests Batch
-- Atomically handles customer lookup/linking and multiple request creations.
CREATE OR REPLACE FUNCTION fn_submit_client_requests_batch(
    p_customer_name TEXT,
    p_customer_contact TEXT,
    p_cards JSONB, -- Array of objects: {card_name, set_name, details, quantity, tcg, scryfall_id, oracle_id, ...}
    p_customer_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_customer_id UUID := p_customer_id;
    v_card JSONB;
    v_count INT := 0;
    v_first_name TEXT;
    v_last_name TEXT;
BEGIN
    -- Only lookup/create if no explicit ID provided
    IF v_customer_id IS NULL THEN
        -- Try to find an existing customer by email or phone
        SELECT id INTO v_customer_id 
        FROM customer 
        WHERE email = p_customer_contact OR phone = p_customer_contact 
        LIMIT 1;

        -- If no customer found, create one to ensure data integrity
        IF v_customer_id IS NULL THEN
            v_first_name := split_part(p_customer_name, ' ', 1);
            v_last_name := trim(substring(p_customer_name from char_length(v_first_name) + 1));
            
            IF v_last_name = '' THEN
                v_last_name := '-';
            END IF;

            INSERT INTO customer (
                first_name, 
                last_name, 
                email, 
                phone
            ) VALUES (
                v_first_name, 
                v_last_name, 
                CASE WHEN p_customer_contact LIKE '%@%' THEN p_customer_contact ELSE NULL END,
                CASE WHEN p_customer_contact NOT LIKE '%@%' THEN p_customer_contact ELSE NULL END
            ) RETURNING id INTO v_customer_id;
        END IF;
    END IF;

    -- Iterate over cards and insert
    FOR v_card IN SELECT * FROM jsonb_array_elements(p_cards) LOOP
        INSERT INTO client_request (
            customer_id, 
            customer_name, 
            customer_contact, 
            card_name, 
            set_name, 
            details, 
            quantity,
            tcg,
            status,
            match_type,
            scryfall_id,
            oracle_id,
            image_url,
            foil_treatment,
            card_treatment,
            set_code,
            collector_number
        ) VALUES (
            v_customer_id, 
            p_customer_name, 
            p_customer_contact, 
            trim(v_card->>'card_name'), 
            v_card->>'set_name', 
            v_card->>'details', 
            COALESCE((v_card->>'quantity')::INT, 1),
            COALESCE(lower(trim(v_card->>'tcg')), 'mtg'),
            'pending',
            COALESCE(v_card->>'match_type', 'any'),
            v_card->>'scryfall_id',
            (v_card->>'oracle_id')::UUID,
            v_card->>'image_url',
            v_card->>'foil_treatment',
            v_card->>'card_treatment',
            v_card->>'set_code',
            v_card->>'collector_number'
        );
        v_count := v_count + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'count', v_count,
        'customer_id', v_customer_id
    );
END;
$$ LANGUAGE plpgsql;
-- Sync product stock from product_storage sum (excluding 'pending', and subtracting 'pending')
CREATE OR REPLACE FUNCTION fn_update_product_stock()
RETURNS TRIGGER AS $$
DECLARE
  v_pid UUID;
  v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
BEGIN
  IF TG_OP = 'DELETE' THEN
    v_pid := OLD.product_id;
  ELSE
    v_pid := NEW.product_id;
  END IF;

  UPDATE product 
  SET stock = COALESCE((
    SELECT SUM(CASE WHEN storage_id = v_pending_id THEN -quantity ELSE quantity END)
    FROM product_storage 
    WHERE product_id = v_pid
  ), 0)
  WHERE id = v_pid;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- Generic update_updated_at function
CREATE OR REPLACE FUNCTION fn_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- Trigger for bounty fulfillment
DROP TRIGGER IF EXISTS trg_bounty_fulfillment ON bounty_offer;

CREATE TRIGGER trg_bounty_fulfillment
AFTER UPDATE ON bounty_offer
FOR EACH ROW
EXECUTE FUNCTION fn_fulfill_bounty_offer();
-- Trigger to update updated_at on bounty
DROP TRIGGER IF EXISTS trg_bounty_updated_at ON bounty;
CREATE TRIGGER trg_bounty_updated_at
BEFORE UPDATE ON bounty
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
-- Trigger to release pending stock when an order is cancelled
CREATE OR REPLACE FUNCTION fn_release_pending_stock()
RETURNS TRIGGER AS $$
DECLARE
   v_pending_id UUID := (SELECT id FROM storage_location WHERE name = 'pending');
   v_item RECORD;
BEGIN
   -- Only process if status changed TO cancelled, and it wasn't already completed/cancelled
   IF NEW.status = 'cancelled' AND OLD.status NOT IN ('cancelled', 'completed') THEN
      FOR v_item IN SELECT product_id, quantity FROM order_item WHERE order_id = NEW.id AND product_id IS NOT NULL AND quantity > 0 LOOP
         UPDATE product_storage 
         SET quantity = GREATEST(0, quantity - v_item.quantity)
         WHERE product_id = v_item.product_id AND storage_id = v_pending_id;
      END LOOP;
   END IF;
   
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_order_cancel ON "order";
CREATE TRIGGER trg_order_cancel
AFTER UPDATE OF status ON "order"
FOR EACH ROW
WHEN (NEW.status = 'cancelled' AND OLD.status != 'cancelled')
EXECUTE FUNCTION fn_release_pending_stock();
-- Trigger to update updated_at on product
DROP TRIGGER IF EXISTS trg_product_updated_at ON product;
CREATE TRIGGER trg_product_updated_at
BEFORE UPDATE ON product
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
-- Trigger to update updated_at on setting
DROP TRIGGER IF EXISTS trg_setting_updated_at ON setting;
CREATE TRIGGER trg_setting_updated_at
BEFORE UPDATE ON setting
FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();
-- Trigger to sync product stock from product_storage
DROP TRIGGER IF EXISTS trg_sync_product_stock ON product_storage;
CREATE TRIGGER trg_sync_product_stock
AFTER INSERT OR UPDATE OR DELETE ON product_storage
FOR EACH ROW EXECUTE FUNCTION fn_update_product_stock();
