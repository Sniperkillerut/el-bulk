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

	// ── Step 1: direct scryfall_id lookup ──────────────────────────────────────
	if scryfallID != "" {
		if cp, ok := ckPriceMap["scry:"+scryfallID+":"+foilSuffix]; ok && cp != nil {
			return cp
		}
	}

	// ── Step 2: scored name + edition + variation fallback ─────────────────────
	name = strings.ToLower(strings.TrimSpace(name))
	ckEdition = strings.ToLower(strings.TrimSpace(ckEdition))
	variation = strings.ToLower(strings.TrimSpace(variation))

	nameKeyPrefix := name + "|"
	foilKeySuffix := "|" + foilSuffix

	bestScore := -1
	var bestPrice *float64

	for k, cp := range ckPriceMap {
		if cp == nil {
			continue
		}
		// Skip non-name keys (scry:, ckid: prefixes)
		if strings.HasPrefix(k, "scry:") || strings.HasPrefix(k, "ckid:") {
			continue
		}
		if !strings.HasPrefix(k, nameKeyPrefix) || !strings.HasSuffix(k, foilKeySuffix) {
			continue
		}

		parts := strings.Split(k, "|")
		if len(parts) < 4 {
			continue
		}
		ckEditionEntry := parts[1]
		ckVariation := parts[2]

		// Skip junk entries that are often mispriced relative to the actual card
		if strings.Contains(ckVariation, "art card") ||
			strings.Contains(ckVariation, "token") ||
			strings.Contains(ckVariation, "gold-bordered") ||
			strings.Contains(ckVariation, "placeholder") {
			continue
		}

		// Exact edition match — NO fuzzy/substring matching to prevent cross-set leakage
		// (e.g. "ravnica" must not match "ravnica remastered").
		editionMatches := ckEdition != "" && ckEditionEntry == ckEdition

		var score int
		if editionMatches {
			switch {
			case ckVariation == variation:
				score = 3 // exact edition + exact variation
			case ckVariation == "" || ckVariation == "standard":
				score = 2 // exact edition + base/normal print
			default:
				score = 1 // exact edition, wrong variation
			}
		} else {
			score = 0 // cross-set fallback (only fires when ckEdition is empty/unknown)
		}

		// Update best: prefer higher score; break ties by lowest price.
		if score > bestScore || (score == bestScore && *cp < *bestPrice) {
			bestScore = score
			bestPrice = cp
		}
	}

	return bestPrice
}
