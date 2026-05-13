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
        p_id := NULLIF(item->>'id', '')::uuid;

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
            scryfall_id, legalities, cost_basis_cop, frame_effects, updated_at
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
            NULLIF(item->>'scryfall_id', '')::uuid,
            item->'legalities',
            COALESCE((COALESCE(item->>'cost_basis_cop', item->>'cost_basis'))::numeric, 0),
            item->'frame_effects',
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
            frame_effects = EXCLUDED.frame_effects,
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
            SELECT p_id, NULLIF(cat::text, '')::uuid
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
                SELECT p_id, NULLIF(si->>'stored_in_id', '')::uuid, (si->>'quantity')::int
                FROM jsonb_array_elements(item->'storage_items') AS si
                WHERE (si->>'quantity')::int > 0
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity;
            -- Priority 2: top-level stored_in_id + stock/quantity
            ELSIF item ? 'stored_in_id' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                VALUES (p_id, NULLIF(item->>'stored_in_id', '')::uuid, COALESCE((item->>'stock')::int, (item->>'quantity')::int, 0))
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = EXCLUDED.quantity;
            END IF;
        ELSE
            -- Priority 1: storage_items array
            IF item ? 'storage_items' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                SELECT p_id, NULLIF(si->>'stored_in_id', '')::uuid, (si->>'quantity')::int
                FROM jsonb_array_elements(item->'storage_items') AS si
                WHERE (si->>'quantity')::int > 0
                ON CONFLICT (product_id, storage_id) DO UPDATE SET quantity = product_storage.quantity + EXCLUDED.quantity;
            -- Priority 2: top-level stored_in_id + stock/quantity
            ELSIF item ? 'stored_in_id' THEN
                INSERT INTO product_storage (product_id, storage_id, quantity)
                VALUES (p_id, NULLIF(item->>'stored_in_id', '')::uuid, COALESCE((item->>'stock')::int, (item->>'quantity')::int, 0))
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
                image_url, legalities, foil_treatment, card_treatment, scryfall_id, frame_effects
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
                NULLIF(dc->>'scryfall_id', '')::uuid,
                dc->'frame_effects'
            FROM jsonb_array_elements(item->'deck_cards') AS dc;
        END IF;

        upserted_id := p_id;
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
