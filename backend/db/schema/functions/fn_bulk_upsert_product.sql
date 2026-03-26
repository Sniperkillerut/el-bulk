-- Bulk Upsert Product
-- Handles product creation/update, category mapping, and storage assignment in one call.
CREATE OR REPLACE FUNCTION fn_bulk_upsert_product(data jsonb)
RETURNS TABLE(product_id UUID) AS $$
DECLARE
    item jsonb;
    p_id UUID;
BEGIN
    FOR item IN SELECT * FROM jsonb_array_elements(data)
    LOOP
        -- Upsert product
        INSERT INTO product (
            id, name, tcg, category, set_name, set_code, collector_number, condition,
            foil_treatment, card_treatment, promo_type,
            price_reference, price_source, price_cop_override,
            image_url, description, language, color_identity, rarity, cmc,
            is_legendary, is_historic, is_land, is_basic_land, art_variation,
            oracle_text, artist, type_line, border_color, frame, full_art, textless,
            updated_at
        )
        VALUES (
            COALESCE((item->>'id')::uuid, gen_random_uuid()),
            item->>'name',
            item->>'tcg',
            item->>'category',
            item->>'set_name',
            item->>'set_code',
            item->>'collector_number',
            item->>'condition',
            COALESCE(item->>'foil_treatment', 'non_foil'),
            COALESCE(item->>'card_treatment', 'normal'),
            item->>'promo_type',
            (item->>'price_reference')::numeric,
            COALESCE(item->>'price_source', 'manual'),
            (item->>'price_cop_override')::numeric,
            item->>'image_url',
            item->>'description',
            COALESCE(item->>'language', 'en'),
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
            updated_at = now()
        RETURNING id INTO p_id;

        -- Categories
        DELETE FROM product_category WHERE product_id = p_id;
        IF item ? 'category_ids' THEN
            INSERT INTO product_category (product_id, category_id)
            SELECT p_id, (cat::text)::uuid
            FROM jsonb_array_elements_text(item->'category_ids') AS cat
            ON CONFLICT DO NOTHING;
        END IF;

        -- Storage
        DELETE FROM product_storage WHERE product_id = p_id;
        IF item ? 'storage_items' THEN
            INSERT INTO product_storage (product_id, storage_id, quantity)
            SELECT p_id, (si->>'stored_in_id')::uuid, (si->>'quantity')::int
            FROM jsonb_array_elements(item->'storage_items') AS si
            WHERE (si->>'quantity')::int > 0
            ON CONFLICT DO NOTHING;
        END IF;

        product_id := p_id;
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
