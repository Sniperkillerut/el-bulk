package external

import (
	"testing"
)

func TestResolveMTGPrice(t *testing.T) {
	// Mock Scryfall Data
	scryMap := make(map[PriceKey]CardMetadata)
	idMap := make(map[string]CardMetadata)
	ckMap := make(map[string]*float64)

	// Standard Sol Ring (from the bulk data)
	solRingStd := CardMetadata{
		ScryfallID:   "std_id",
		TCGPlayerUSD: ptr(1.00),
	}
	scryMap[PriceKey{Name: "sol ring", SetCode: "lea", Collector: "269", Foil: "non_foil"}] = solRingStd
	scryMap[PriceKey{Name: "sol ring", SetCode: "lea", Collector: "", Foil: "non_foil"}] = solRingStd
	scryMap[PriceKey{Name: "sol ring", SetCode: "", Collector: "", Foil: "non_foil"}] = solRingStd // Global fallback 1
	idMap["std_id"] = solRingStd

	// Specialty Sol Ring (processed later in bulk?)
	solRingPremium := CardMetadata{
		ScryfallID:   "premium_id",
		TCGPlayerUSD: ptr(100.00),
	}
	scryMap[PriceKey{Name: "sol ring", SetCode: "pr", Collector: "1", Foil: "foil"}] = solRingPremium
	// Note: in OUR NEW logic, scryMap[PriceKey{Name: "sol ring", SetCode: "", Collector: "", Foil: "foil"}]
	// will be set for the FIRST foil version we see.

	// CardKingdom Mock
	ckStd := 0.90
	ckPremium := 99.00
	ckMap["scry:std_id:non_foil"] = &ckStd
	ckMap["scry:premium_id:foil"] = &ckPremium

	t.Run("ID Match Wins", func(t *testing.T) {
		res := ResolveMTGPrice("std_id", "Sol Ring", "", "", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" || *res.TCGPlayerUSD != 1.00 || *res.CardKingdomUSD != 0.90 {
			t.Errorf("Expected standard Sol Ring, got %+v", res)
		}
	})

	t.Run("Name+Set+Collector Match", func(t *testing.T) {
		res := ResolveMTGPrice("", "Sol Ring", "lea", "269", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" || *res.TCGPlayerUSD != 1.00 {
			t.Errorf("Expected standard Sol Ring by set match, got %+v", res)
		}
	})

	t.Run("Global Fallback Match", func(t *testing.T) {
		res := ResolveMTGPrice("", "Sol Ring", "", "", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" || *res.TCGPlayerUSD != 1.00 {
			t.Errorf("Expected standard Sol Ring by global fallback, got %+v", res)
		}
	})

	t.Run("CK Multiple Foil IDs Match", func(t *testing.T) {
		res := ResolveMTGPrice("premium_id", "Sol Ring", "", "", "foil", "", "", "", scryMap, idMap, ckMap)
		if *res.CardKingdomUSD != 99.00 {
			t.Errorf("Expected premium CK price by ID+Foil, got %v", res.CardKingdomUSD)
		}
	})

	t.Run("Ripple Foil Specialty Match", func(t *testing.T) {
		// Mock CK map with ripple foil variation
		ripplePrice := 5.99
		ckMap["Sol Ring|Modern Horizons 3 Commander|ripple foil|foil"] = &ripplePrice
		ckMap["scry:ripple_id:foil"] = &ripplePrice

		idMap["ripple_id"] = CardMetadata{
			ScryfallID: "ripple_id",
		}

		res := ResolveMTGPrice("ripple_id", "Sol Ring", "m3c", "305", "ripple_foil", "", "Modern Horizons 3 Commander", "ripple foil", scryMap, idMap, ckMap)
		if res.CardKingdomUSD == nil || *res.CardKingdomUSD != 5.99 {
			t.Errorf("Expected Ripple Foil price 5.99, got %v", res.CardKingdomUSD)
		}
	})
}

func ptr(f float64) *float64 { return &f }
