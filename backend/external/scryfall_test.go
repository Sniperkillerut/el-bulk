package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupMTGCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards/m11/1" {
			// Exact match
			card := scryfallCard{
				Name:            "Lightning Bolt",
				Set:             "m11",
				CollectorNumber: "1",
				Lang:            "en",
			}
			json.NewEncoder(w).Encode(card)
			return
		}
		if r.URL.Path == "/cards/584837f0-1709-467a-8df1-0b8dc08f9146" {
			// Direct ID match
			card := scryfallCard{
				ID:              "584837f0-1709-467a-8df1-0b8dc08f9146",
				Name:            "Lightning Bolt",
				Set:             "m11",
				CollectorNumber: "1",
				Lang:            "en",
			}
			json.NewEncoder(w).Encode(card)
			return
		}
		if r.URL.Path == "/cards/named" {
			// Named exact or fuzzy
			exact := r.URL.Query().Get("exact")
			if exact == "Lightning Bolt" {
				card := scryfallCard{Name: "Lightning Bolt", Set: "m11"}
				json.NewEncoder(w).Encode(card)
				return
			}
			fuzzy := r.URL.Query().Get("fuzzy")
			if fuzzy == "bolt" {
				card := scryfallCard{Name: "Lightning Bolt", Set: "m11"}
				json.NewEncoder(w).Encode(card)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"object":"error","code":"not_found","status":404,"details":"No card found"}`))
	}))
	defer server.Close()

	originalBase := ScryfallBase
	ScryfallBase = server.URL
	defer func() { ScryfallBase = originalBase }()

	t.Run("Direct ID Lookup", func(t *testing.T) {
		res, err := LookupMTGCard(context.Background(), "584837f0-1709-467a-8df1-0b8dc08f9146", "", "", "", "non_foil")
		assert.NoError(t, err)
		assert.Equal(t, "Lightning Bolt", res.Name)
		assert.Equal(t, "584837f0-1709-467a-8df1-0b8dc08f9146", res.ScryfallID)
	})

	t.Run("Exact Match", func(t *testing.T) {
		res, err := LookupMTGCard(context.Background(), "", "Lightning Bolt", "m11", "1", "non_foil")
		assert.NoError(t, err)
		assert.Equal(t, "Lightning Bolt", res.Name)
		assert.Equal(t, "m11", *res.SetCode)
	})

	t.Run("Fuzzy Match", func(t *testing.T) {
		res, err := LookupMTGCard(context.Background(), "", "bolt", "", "", "non_foil")
		assert.NoError(t, err)
		assert.Equal(t, "Lightning Bolt", res.Name)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := LookupMTGCard(context.Background(), "", "nonexistent", "xxx", "999", "non_foil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not found")
	})

	t.Run("Empty Input", func(t *testing.T) {
		_, err := LookupMTGCard(context.Background(), "", "", "", "", "non_foil")
		assert.Error(t, err)
	})

	t.Run("Multi-faced Card", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			card := scryfallCard{
				Name: "Delver of Secrets",
				CardFaces: []struct {
					Name       string `json:"name"`
					TypeLine   string `json:"type_line"`
					OracleText string `json:"oracle_text"`
					Artist     string `json:"artist"`
					ImageURIs  struct {
						Normal string `json:"normal"`
						Large  string `json:"large"`
						PNG    string `json:"png"`
					} `json:"image_uris"`
				}{
					{Name: "Delver", OracleText: "Look at top", ImageURIs: struct {
						Normal string `json:"normal"`
						Large  string `json:"large"`
						PNG    string `json:"png"`
					}{Normal: "face1.png"}},
					{Name: "Insectile Aberration", OracleText: "Fly"},
				},
			}
			json.NewEncoder(w).Encode(card)
		}))
		defer server.Close()

		originalBase := ScryfallBase
		ScryfallBase = server.URL
		defer func() { ScryfallBase = originalBase }()

		res, err := LookupMTGCard(context.Background(), "", "Delver of Secrets", "ISD", "51", "non_foil")
		assert.NoError(t, err)
		assert.Contains(t, *res.OracleText, "Look at top")
		assert.Equal(t, "face1.png", res.ImageURL)
	})

	t.Run("Foil Prices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usd := "1.50"
			usdFoil := "5.00"
			usdEtched := "10.00"
			card := scryfallCard{
				Name: "Foil Card",
				Prices: struct {
					USD       *string `json:"usd"`
					USDFoil   *string `json:"usd_foil"`
					USDEtched *string `json:"usd_etched"`
					EUR       *string `json:"eur"`
					EURFoil   *string `json:"eur_foil"`
				}{
					USD: &usd, USDFoil: &usdFoil, USDEtched: &usdEtched,
				},
			}
			json.NewEncoder(w).Encode(card)
		}))
		defer server.Close()

		originalBase := ScryfallBase
		ScryfallBase = server.URL
		defer func() { ScryfallBase = originalBase }()

		res, err := LookupMTGCard(context.Background(), "", "Foil Card", "", "", "foil")
		assert.NoError(t, err)
		assert.Equal(t, 5.0, *res.PriceTCGPlayer)

		resEtched, _ := LookupMTGCard(context.Background(), "", "Foil Card", "", "", "etched_foil")
		assert.Equal(t, 10.0, *resEtched.PriceTCGPlayer)
	})
}

func TestBatchLookupMTGCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards/collection" {
			resp := scryfallCollectionResponse{
				Data: []scryfallCard{
					{Name: "Card 1", Set: "m11"},
					{Name: "Card 2", Set: "m11"},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	originalBase := ScryfallBase
	ScryfallBase = server.URL
	defer func() { ScryfallBase = originalBase }()

	t.Run("Success", func(t *testing.T) {
		ids := []CardIdentifier{
			{Name: "Card 1", SetCode: "m11"},
			{Name: "Card 2", SetCode: "m11"},
		}
		res, err := BatchLookupMTGCard(context.Background(), ids)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(res))
		assert.Equal(t, "Card 1", res[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		res, err := BatchLookupMTGCard(context.Background(), nil)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestScryfall_MappingEdgeCases(t *testing.T) {
	t.Run("Mapping Types", func(t *testing.T) {
		card := &scryfallCard{
			Name:          "Karn",
			TypeLine:      "Legendary Artifact Creature — Golem",
			ColorIdentity: []string{"W", "U"},
		}
		res := mapScryfallToResult(card, "non_foil")
		assert.True(t, res.IsLegendary)
		assert.True(t, res.IsHistoric)
		assert.Equal(t, "W,U", *res.ColorIdentity)
	})

	t.Run("Mapping Land", func(t *testing.T) {
		card := &scryfallCard{
			TypeLine: "Basic Land — Island",
		}
		res := mapScryfallToResult(card, "non_foil")
		assert.True(t, res.IsLand)
		assert.True(t, res.IsBasicLand)
	})
}

func TestScryfall_ErrorPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.URL.Path == "/decode-fail" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldBase := ScryfallBase
	ScryfallBase = server.URL
	defer func() { ScryfallBase = oldBase }()

	t.Run("Internal Server Error", func(t *testing.T) {
		_, err := scryfallGet(context.Background(), ScryfallBase+"/error", "non_foil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("Decode Error", func(t *testing.T) {
		_, err := scryfallGet(context.Background(), ScryfallBase+"/decode-fail", "non_foil")
		assert.Error(t, err)
	})
}
