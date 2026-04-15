package external

import (
	"testing"
)

func fptr(v float64) *float64 { return &v }

func TestLookupCKPrice(t *testing.T) {
	// Build a representative mini price map (mirrors the real CK key format)
	// Key format: "name|edition|variation|foil_suffix"
	ckMap := map[string]*float64{
		// Normal print
		"lightning bolt|magic 2010|non_foil":              fptr(0.50),
		"lightning bolt|magic 2010||non_foil":             fptr(0.50),
		// Showcase variant (same set, different variation)
		"lightning bolt|magic 2010|showcase|non_foil":     fptr(3.00),
		// Foil
		"lightning bolt|magic 2010||foil":                 fptr(2.50),
		// Different set (should NOT bleed into magic 2010 lookups)
		"lightning bolt|modern masters 2015||non_foil":    fptr(1.20),
		// Art card junk — must be skipped
		"lightning bolt|magic 2010|art card|non_foil":     fptr(99.99),
		// Token junk — must be skipped
		"lightning bolt|magic 2010|token|non_foil":        fptr(99.99),
		// Card from another TCG set
		"shock|ravnica||non_foil":                         fptr(0.10),
		"shock|ravnica remastered||non_foil":              fptr(0.15),
	}

	tests := []struct {
		name      string
		cardName  string
		ckEdition string
		variation string
		isFoil    bool
		wantPrice *float64
		wantNil   bool
	}{
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
			name:      "foil match",
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
			// art card entry in map should be ignored; falls back to base (score 2)
			wantPrice: fptr(0.50),
		},
		{
			name:      "no cross-set leakage: ravnica does not match ravnica remastered",
			cardName:  "Shock",
			ckEdition: "ravnica", // exact match should find only ravnica entry
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
			// any non-junk non-foil entry for lightning bolt qualifies
			wantNil: false, // should find something
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
			name:      "case insensitive name matching",
			cardName:  "LIGHTNING BOLT",
			ckEdition: "MAGIC 2010",
			variation: "",
			isFoil:    false,
			wantPrice: fptr(0.50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupCKPrice(tt.cardName, tt.ckEdition, tt.variation, tt.isFoil, ckMap)

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
