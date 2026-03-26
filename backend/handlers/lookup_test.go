package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/el-bulk/backend/external"
	"github.com/stretchr/testify/assert"
)

func TestLookupHandler_MTG(t *testing.T) {
	// Mock Scryfall
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards/named" {
			name := r.URL.Query().Get("exact")
			if name == "" {
				name = r.URL.Query().Get("fuzzy")
			}

			if name == "Black Lotus" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"name": "Black Lotus", "set_name": "Limited Edition Alpha", "set": "lea", "collector_number": "232", "image_uris": {"png": " lotus.png"}}`)
				return
			}
			if name == "Not Found" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Override ScryfallBase
	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	h := NewLookupHandler()

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/lookup/mtg?name=Black+Lotus", nil)
		rr := httptest.NewRecorder()
		h.MTG(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res external.CardLookupResult
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, "Black Lotus", res.Name)
	})

	t.Run("Missing Name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/lookup/mtg", nil)
		rr := httptest.NewRecorder()
		h.MTG(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/lookup/mtg?name=Not+Found", nil)
		rr := httptest.NewRecorder()
		h.MTG(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestLookupHandler_BatchMTG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards/collection" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"data": [{"name": "Card 1"}, {"name": "Card 2"}]}`)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldBase := external.ScryfallBase
	external.ScryfallBase = server.URL
	defer func() { external.ScryfallBase = oldBase }()

	h := NewLookupHandler()

	t.Run("Success", func(t *testing.T) {
		input := struct {
			Identifiers []external.CardIdentifier `json:"identifiers"`
		}{
			Identifiers: []external.CardIdentifier{{Name: "Card 1"}},
		}
		body, _ := json.Marshal(input)
		req, _ := http.NewRequest("POST", "/api/admin/lookup/mtg/batch", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BatchMTG(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res []external.CardLookupResult
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Len(t, res, 2)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/admin/lookup/mtg/batch", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()
		h.BatchMTG(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestLookupHandler_Pokemon_Error(t *testing.T) {
	h := NewLookupHandler()
	t.Run("Missing Name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/lookup/pokemon", nil)
		rr := httptest.NewRecorder()
		h.Pokemon(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Pokemon Bad Gateway", func(t *testing.T) {
		oldPokemonBase := external.PokemonTCGBase
		external.PokemonTCGBase = "http://invalid-url"
		defer func() { external.PokemonTCGBase = oldPokemonBase }()

		req, _ := http.NewRequest("GET", "/api/admin/lookup/pokemon?name=Pikachu", nil)
		rr := httptest.NewRecorder()
		h.Pokemon(rr, req)
		assert.Equal(t, http.StatusBadGateway, rr.Code)
	})

	t.Run("Pokemon Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		oldPokemonBase := external.PokemonTCGBase
		external.PokemonTCGBase = server.URL
		defer func() { external.PokemonTCGBase = oldPokemonBase }()

		req, _ := http.NewRequest("GET", "/api/admin/lookup/pokemon?name=Missing", nil)
		rr := httptest.NewRecorder()
		h.Pokemon(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestLookupHandler_Pokemon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards" {
			q := r.URL.Query().Get("q")
			if q == `name:"Charizard"` {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"data": [{"id": "base1-4", "name": "Charizard", "number": "4", "set": {"name": "Base Set", "id": "base1"}, "images": {"large": "char.png"}}], "totalCount": 1}`)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldBase := external.PokemonTCGBase
	external.PokemonTCGBase = server.URL
	defer func() { external.PokemonTCGBase = oldBase }()

	h := NewLookupHandler()

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/lookup/pokemon?name=Charizard", nil)
		rr := httptest.NewRecorder()
		h.Pokemon(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res external.CardLookupResult
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Equal(t, "Charizard", res.Name)
	})
}
