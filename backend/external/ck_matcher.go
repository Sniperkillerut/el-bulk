package external

import "strings"

// LookupCKPrice finds the best CardKingdom price for a card.
// It is the single source of truth for CK price resolution across the codebase.
//
// Lookup order (first hit wins):
//
//  1. Direct scryfall_id match  — "scry:{scryfallID}:{foil|non_foil}"
//     Covers ~98% of CK entries. Completely bypasses name/edition/variation logic.
//
//  2. Scored name + edition + variation fallback — for the ~2% of CK entries that
//     lack a scryfall_id, or when scryfallID is unknown to the caller.
//     Scoring:
//     3 — exact edition + exact variation
//     2 — exact edition + empty/"standard" variation (base/normal print)
//     1 — exact edition, wrong variation
//     0 — wrong/unknown edition (cross-set fallback, only when ckEdition is empty)
//
// Junk entries (art cards, tokens, gold-bordered, placeholders) are always skipped
// in the scored fallback.
// Returns nil if no match is found at any level.
func LookupCKPrice(scryfallID, name, ckEdition, variation string, isFoil bool, ckPriceMap map[string]*float64) *float64 {
	foilSuffix := "non_foil"
	if isFoil {
		foilSuffix = "foil"
	}

	// ── Step 1: Direct Scryfall ID Lookup ──────────────────────────────────────
	// This is the most precise link. If we have it, use it.
	if scryfallID != "" {
		if cp, ok := ckPriceMap["scry:"+scryfallID+":"+foilSuffix]; ok && cp != nil {
			return cp
		}
	}

	// ── Step 2: Scored Name + Edition + Variation fallback ─────────────────────
	// For cases without an ID or where the ID link is missing in CK data.
	cleanName := strings.ToLower(strings.TrimSpace(name))
	cleanEdition := strings.ToLower(strings.TrimSpace(ckEdition))
	cleanVariation := strings.ToLower(strings.TrimSpace(variation))

	foilKeySuffix := "|" + foilSuffix
	bestScore := -1
	var bestPrice *float64

	ckCacheMutex.RLock()
	potentialKeys := ckNameIndex[cleanName]
	ckCacheMutex.RUnlock()

	for _, k := range potentialKeys {
		cp := ckPriceMap[k]
		if cp == nil || !strings.HasSuffix(k, foilKeySuffix) {
			continue
		}

		parts := strings.Split(k, "|")
		if len(parts) < 4 {
			continue
		}
		ckEditionEntry := strings.ToLower(parts[1])
		ckVariationEntry := strings.ToLower(parts[2])

		// Skip junk entries
		if strings.Contains(ckVariationEntry, "art card") || strings.Contains(ckVariationEntry, "token") ||
			strings.Contains(ckVariationEntry, "gold-bordered") || strings.Contains(ckVariationEntry, "placeholder") {
			continue
		}

		editionMatches := cleanEdition != "" && ckEditionEntry == cleanEdition
		var score int
		if editionMatches {
			switch {
			case ckVariationEntry == cleanVariation:
				score = 3
			case ckVariationEntry == "" || ckVariationEntry == "standard":
				score = 2
			default:
				score = 1
			}
		} else {
			score = 0
		}

		if score > bestScore || (score == bestScore && *cp < *bestPrice) {
			bestScore = score
			bestPrice = cp
		}
	}

	return bestPrice
}
