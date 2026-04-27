package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupPokemonCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == `name:"Charizard" set.id:base1` {
			resp := pokemonAPIResponse{
				Data: []pokemonAPICard{
					{
						ID:     "base1-4",
						Name:   "Charizard",
						Number: "4",
						Set: struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						}{ID: "base1", Name: "Base Set"},
						Images: struct {
							Small string `json:"small"`
							Large string `json:"large"`
						}{Small: "small.png", Large: "large.png"},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Query().Get("q") == `name:"No Image"` {
			resp := pokemonAPIResponse{
				Data: []pokemonAPICard{
					{Name: "No Image", ID: "ni-1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	originalBase := PokemonTCGBase
	PokemonTCGBase = server.URL
	defer func() { PokemonTCGBase = originalBase }()

	t.Run("Success", func(t *testing.T) {
		res, err := LookupPokemonCard(context.Background(), "Charizard", "base1")
		assert.NoError(t, err)
		assert.Equal(t, "Charizard", res.Name)
		assert.Equal(t, "large.png", res.ImageURL)
	})

	t.Run("No Image Error", func(t *testing.T) {
		_, err := LookupPokemonCard(context.Background(), "No Image", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no image")
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := LookupPokemonCard(context.Background(), "Nonexistent", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not found")
	})

	t.Run("Empty Name", func(t *testing.T) {
		_, err := LookupPokemonCard(context.Background(), "", "")
		assert.Error(t, err)
	})
}
