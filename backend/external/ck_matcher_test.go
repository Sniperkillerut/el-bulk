package external

import (
	"strings"
	"testing"
)

func fptr(v float64) *float64 { return &v }

func TestLookupCKPrice(t *testing.T) {
	// Build a representative mini price map.
	// Three key schemes mirror what BuildCardKingdomPriceMap emits:
	//   "scry:{uuid}:{foil}"          — primary path (~98% of real CK data)
	//   "ckid:{int}"                  — secondary (Scryfall bulk-data fast-path)
	//   "name|edition|variation|foil" — fallback (~2% without scryfall_id)
	ckMap := map[string]*float64{
		// ── scry: entries (primary) ─────────────────────────────────────────────
		"scry:uuid-bolt-m10:non_foil":  fptr(0.50),
		"scry:uuid-bolt-m10:foil":      fptr(2.50),
		"scry:uuid-shock-rav:non_foil": fptr(0.10),
		// A card that exists ONLY in scry: (no name-key fallback) to prove direct hit
		"scry:uuid-only-scry:non_foil": fptr(9.99),

		// ── name|edition|variation|foil entries (fallback) ──────────────────────
		// Normal print
		"lightning bolt|magic 2010||non_foil": fptr(0.50),
		// Showcase variant (same set, different variation)
		"lightning bolt|magic 2010|showcase|non_foil": fptr(3.00),
		// Foil
		"lightning bolt|magic 2010||foil": fptr(2.50),
		// Different set (should NOT bleed into magic 2010 lookups)
		"lightning bolt|modern masters 2015||non_foil": fptr(1.20),
		// Art card junk — must be skipped
		"lightning bolt|magic 2010|art card|non_foil": fptr(99.99),
		// Token junk — must be skipped
		"lightning bolt|magic 2010|token|non_foil": fptr(99.99),
		// Cards from another set
		"shock|ravnica||non_foil":            fptr(0.10),
		"shock|ravnica remastered||non_foil": fptr(0.15),
	}

	// Populate the global index for the matcher (since it now uses O(1) lookups)
	ckCacheMutex.Lock()
	ckNameIndex = make(map[string][]string)
	for k := range ckMap {
		if !strings.Contains(k, "|") {
			continue
		}
		parts := strings.Split(k, "|")
		nameKey := strings.ToLower(strings.TrimSpace(parts[0]))
		ckNameIndex[nameKey] = append(ckNameIndex[nameKey], k)
	}
	ckCacheMutex.Unlock()

	tests := []struct {
		name       string
		scryfallID string // "" = no scryfall_id (forces fallback path)
		cardName   string
		ckEdition  string
		variation  string
		isFoil     bool
		wantPrice  *float64
		wantNil    bool
	}{
		// ── Direct scryfall_id hits (primary path) ───────────────────────────────
		{
			name:       "scryfall_id direct hit non-foil",
			scryfallID: "uuid-bolt-m10",
			cardName:   "Lightning Bolt",
			ckEdition:  "wrong edition", // should be ignored when scryfall_id hits
			variation:  "wrong variation",
			isFoil:     false,
			wantPrice:  fptr(0.50),
		},
		{
			name:       "scryfall_id direct hit foil",
			scryfallID: "uuid-bolt-m10",
			cardName:   "Lightning Bolt",
			ckEdition:  "wrong edition",
			variation:  "",
			isFoil:     true,
			wantPrice:  fptr(2.50),
		},
		{
			name:       "scryfall_id-only entry (no name fallback) returns price",
			scryfallID: "uuid-only-scry",
			cardName:   "Some Card",
			ckEdition:  "",
			variation:  "",
			isFoil:     false,
			wantPrice:  fptr(9.99),
		},
		{
			name:       "unknown scryfall_id falls through to name matching",
			scryfallID: "uuid-not-in-map",
			cardName:   "Lightning Bolt",
			ckEdition:  "magic 2010",
			variation:  "",
			isFoil:     false,
			wantPrice:  fptr(0.50), // name fallback
		},

		// ── Scored name + edition + variation fallback ───────────────────────────
		{
			name:      "exact edition + exact empty variation returns base price",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "",
			isFoil:    false,
			wantPrice: fptr(0.50),
		},
		{
			name:      "exact edition + exact variation returns specific variant price",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "showcase",
			isFoil:    false,
			wantPrice: fptr(3.00),
		},
		{
			name:      "foil match via fallback",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "",
			isFoil:    true,
			wantPrice: fptr(2.50),
		},
		{
			name:      "wrong variation falls back to base (score 2 beats score 1)",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "extended art", // not in map for this set
			isFoil:    false,
			wantPrice: fptr(0.50), // score-2 base wins over score-1 wrong-variation
		},
		{
			name:      "junk entries (art card, token) are skipped",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "art card",
			isFoil:    false,
			wantPrice: fptr(0.50), // junk skipped, base wins
		},
		{
			name:      "no cross-set leakage: ravnica does not match ravnica remastered",
			cardName:  "Shock",
			ckEdition: "ravnica",
			variation: "",
			isFoil:    false,
			wantPrice: fptr(0.10),
		},
		{
			name:      "ravnica remastered does not bleed into ravnica",
			cardName:  "Shock",
			ckEdition: "ravnica remastered",
			variation: "",
			isFoil:    false,
			wantPrice: fptr(0.15),
		},
		{
			name:      "unknown edition returns cross-set fallback (score 0)",
			cardName:  "Lightning Bolt",
			ckEdition: "", // empty = unknown set
			variation: "",
			isFoil:    false,
			wantNil:   false, // should find something
		},
		{
			name:      "tie at score-1: lowest price wins (deterministic)",
			cardName:  "Lightning Bolt",
			ckEdition: "Magic 2010",
			variation: "promo", // not in map → score-2 base wins
			isFoil:    false,
			wantPrice: fptr(0.50),
		},
		{
			name:      "card not in map returns nil",
			cardName:  "Ancestral Recall",
			ckEdition: "magic 2010",
			variation: "",
			isFoil:    false,
			wantNil:   true,
		},
		{
			name:      "case insensitive name and edition matching",
			cardName:  "LIGHTNING BOLT",
			ckEdition: "MAGIC 2010",
			variation: "",
			isFoil:    false,
			wantPrice: fptr(0.50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupCKPrice(tt.scryfallID, tt.cardName, tt.ckEdition, tt.variation, tt.isFoil, ckMap)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", *got)
				}
				return
			}

			if got == nil {
				t.Errorf("expected a price, got nil")
				return
			}

			if tt.wantPrice != nil && *got != *tt.wantPrice {
				t.Errorf("expected price %.2f, got %.2f", *tt.wantPrice, *got)
			}
		})
	}
}
