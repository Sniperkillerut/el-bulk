package external

import (
	"strings"
)

// ResolvedPrices holds the results from all pricing sources for a single card.
type ResolvedPrices struct {
	ScryfallID     string
	TCGPlayerUSD   *float64
	CardmarketEUR  *float64
	CardKingdomUSD *float64
	Metadata       *CardMetadata // Full Scryfall metadata if found
}

// ResolveMTGPrice is the single source of truth for MTG pricing resolution.
// It implements the mandatory hierarchy:
// 1. Scryfall ID (exact match)
// 2. Name | Set | Collector (exact printing match)
// 3. Name | Set | Foil | Variant (hierarchical fallback)
func ResolveMTGPrice(
	sid, name, setCode, cn, foil, cardTreatment, ckEdition, ckVariation string,
	scryMap map[PriceKey]CardMetadata,
	idMap map[string]CardMetadata,
	ckMap map[string]*float64,
) ResolvedPrices {
	result := ResolvedPrices{}

	// Normalization
	sid = strings.TrimSpace(sid)
	nameAttr := strings.ToLower(strings.TrimSpace(name))
	setAttr := strings.ToLower(strings.TrimSpace(setCode))
	cnAttr := strings.TrimSpace(cn)
	foilAttr := strings.ToLower(strings.TrimSpace(foil))
	isFoil := foilAttr != "" && foilAttr != "non_foil"

	// ── Step 1: Scryfall ID Match (Highest Priority) ──────────────────────────
	var meta CardMetadata
	found := false
	if sid != "" {
		meta, found = idMap[sid]
	}

	// ── Step 2: Name | Set | Collector (Exact Printing Match) ──────────────────
	if !found && nameAttr != "" && setAttr != "" && cnAttr != "" {
		key := PriceKey{Name: nameAttr, SetCode: setAttr, Collector: cnAttr, Foil: foilAttr}
		meta, found = scryMap[key]
	}

	// ── Step 3: Name | Set | Foil (Edition Fallback) ──────────────────────────
	if !found && nameAttr != "" && setAttr != "" {
		key := PriceKey{Name: nameAttr, SetCode: setAttr, Collector: "", Foil: foilAttr}
		meta, found = scryMap[key]
	}

	// ── Step 4: Name | Foil (Global Fallback) ──────────────────────────────────
	if !found && nameAttr != "" {
		key := PriceKey{Name: nameAttr, SetCode: "", Collector: "", Foil: foilAttr}
		meta, found = scryMap[key]
	}

	if found {
		result.ScryfallID = meta.ScryfallID
		result.TCGPlayerUSD = meta.TCGPlayerUSD
		result.CardmarketEUR = meta.CardmarketEUR
		// If the specific foil price exists in metadata (from BatchLookup), use it.
		// Note: BuildPriceMap already handles some of this in the TCGPlayerUSD field.
		result.Metadata = &meta
	}

	// ── CardKingdom Resolution ───────────────────────────────────────────────
	if name != "" {
		// Delegate ALL CK resolution to LookupCKPrice to ensure consistent hierarchy
		// (Perfect matches > ID matches > Heuristic matches).
		result.CardKingdomUSD = LookupCKPrice(sid, name, ckEdition, ckVariation, isFoil, ckMap)
	}

	return result
}
