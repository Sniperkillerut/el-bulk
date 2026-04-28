package external

import (
	"testing"
)

func TestResolveMTGPrice(t *testing.T) {
	// Mock Scryfall Data
	scryMap := make(map[PriceKey]CardMetadata)
	idMap := make(map[string]CardMetadata)
	ckMap := make(map[string]*float64)

	p1 := 1.00
	p3 := 100.00

	// Standard Sol Ring
	solRingStd := CardMetadata{
		ScryfallID:   "std_id",
		TCGPlayerUSD: &p1,
	}
	scryMap[PriceKey{Name: "sol ring", SetCode: "lea", Collector: "269"}] = solRingStd
	idMap["std_id"] = solRingStd

	// Specialty Sol Ring
	solRingPremium := CardMetadata{
		ScryfallID:       "premium_id",
		TCGPlayerUSDFoil: &p3,
	}
	scryMap[PriceKey{Name: "sol ring", SetCode: "pr", Collector: "1"}] = solRingPremium
	idMap["premium_id"] = solRingPremium

	t.Run("ID Match Wins", func(t *testing.T) {
		res := ResolveMTGPrice("std_id", "Sol Ring", "", "", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" || res.TCGPlayerUSD == nil || *res.TCGPlayerUSD != 1.00 {
			t.Errorf("Expected standard Sol Ring, got %+v", res)
		}
	})

	t.Run("Name+Set+Collector Match", func(t *testing.T) {
		res := ResolveMTGPrice("", "Sol Ring", "lea", "269", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" || res.TCGPlayerUSD == nil || *res.TCGPlayerUSD != 1.00 {
			t.Errorf("Expected standard Sol Ring by set match, got %+v", res)
		}
	})

	t.Run("Foil Match", func(t *testing.T) {
		res := ResolveMTGPrice("premium_id", "Sol Ring", "", "", "foil", "", "", "", scryMap, idMap, ckMap)
		if res.TCGPlayerUSD == nil || *res.TCGPlayerUSD != 100.00 {
			t.Errorf("Expected premium foil price, got %v", res.TCGPlayerUSD)
		}
	})

	t.Run("Hierarchical Fallback (Name+Set)", func(t *testing.T) {
		// Populate fallback in map manually for test
		scryMap[PriceKey{Name: "sol ring", SetCode: "lea"}] = solRingStd
		res := ResolveMTGPrice("", "Sol Ring", "lea", "", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" {
			t.Errorf("Expected fallback match by name+set, got %v", res.ScryfallID)
		}
	})

	t.Run("Global Fallback (Name)", func(t *testing.T) {
		scryMap[PriceKey{Name: "sol ring"}] = solRingStd
		res := ResolveMTGPrice("", "Sol Ring", "", "", "non_foil", "", "", "", scryMap, idMap, ckMap)
		if res.ScryfallID != "std_id" {
			t.Errorf("Expected global fallback match by name, got %v", res.ScryfallID)
		}
	})

	t.Run("Card Kingdom Hierarchical Resolution", func(t *testing.T) {
		ckMap := make(map[string]*float64)
		pCK := 159.99
		// Simulate Tempest Ancient Tomb in CK DB (Empty variation)
		ckMap["ancient tomb|tempest||non_foil"] = &pCK

		// Mock the NameIndex which LookupCKPrice needs
		ckCacheMutex.Lock()
		ckNameIndex = map[string][]string{
			"ancient tomb": {"ancient tomb|tempest||non_foil"},
		}
		ckCacheMutex.Unlock()

		// Case 1: Variation mismatch but edition match (Score 2 should win)
		// Requesting "retro" variation but CK only has ""
		res := ResolveMTGPrice("some_id", "Ancient Tomb", "tmp", "315", "non_foil", "legacy_border", "Tempest", "retro", scryMap, idMap, ckMap)
		if res.CardKingdomUSD == nil || *res.CardKingdomUSD != 159.99 {
			t.Errorf("Expected CK price 159.99 via hierarchical fallback, got %v", res.CardKingdomUSD)
		}

		// Case 2: Scryfall ID Match (Score 3 or direct match)
		pCKEos := 499.99
		ckMap["scry:eos_id:non_foil"] = &pCKEos
		res2 := ResolveMTGPrice("eos_id", "Ancient Tomb", "eos", "91", "non_foil", "borderless", "Edge of Eternities", "borderless galaxy foil", scryMap, idMap, ckMap)
		if res2.CardKingdomUSD == nil || *res2.CardKingdomUSD != 499.99 {
			t.Errorf("Expected CK price 499.99 via Scryfall ID match, got %v", res2.CardKingdomUSD)
		}
	})
}


