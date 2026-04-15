package external

import "strings"

// LookupCKPrice finds the best CardKingdom price for a card using a scored matching
// strategy. It is the single source of truth for CK price resolution across the codebase.
//
// Parameters:
//   - name:       card name (case-insensitive; will be lowercased internally)
//   - ckEdition:  the CK set/edition name, already resolved — prefer the value from the
//                 ck_name DB column; fall back to NormalizeCKEdition(scryfall_set_name)
//   - variation:  CK variation string from MapFoilTreatmentToCKVariation (may be "")
//   - isFoil:     true if the card is any foil treatment
//   - ckPriceMap: the full CK price map from BuildCardKingdomPriceMap
//
// Scoring (higher = better match, first highest score wins):
//
//	3 — exact edition + exact variation
//	2 — exact edition + empty/"standard" variation  (the base/normal print)
//	1 — exact edition, wrong variation (e.g. showcase vs normal; last resort within set)
//	0 — wrong/unknown edition (cross-set fallback, only when edition is empty or unresolved)
//
// Junk entries (art cards, tokens, gold-bordered, placeholders) are always skipped.
// Returns nil if no match is found at any score level.
func LookupCKPrice(name, ckEdition, variation string, isFoil bool, ckPriceMap map[string]*float64) *float64 {
	name = strings.ToLower(strings.TrimSpace(name))
	ckEdition = strings.ToLower(strings.TrimSpace(ckEdition))
	variation = strings.ToLower(strings.TrimSpace(variation))

	nameKeyPrefix := name + "|"
	foilSuffix := "|non_foil"
	if isFoil {
		foilSuffix = "|foil"
	}

	bestScore := -1
	var bestPrice *float64

	for k, cp := range ckPriceMap {
		if cp == nil {
			continue
		}
		if !strings.HasPrefix(k, nameKeyPrefix) || !strings.HasSuffix(k, foilSuffix) {
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

		if score > bestScore {
			bestScore = score
			bestPrice = cp
		}
	}

	return bestPrice
}
