package handlers

import (
	"net/http"

	"github.com/el-bulk/backend/external"
)

// LookupHandler handles card image/metadata lookups from external APIs.
type LookupHandler struct{}

func NewLookupHandler() *LookupHandler {
	return &LookupHandler{}
}

// GET /api/admin/lookup/mtg?name=<name>&set=<setCode>&foil=<foilTreatment>
func (h *LookupHandler) MTG(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	setCode := r.URL.Query().Get("set")
	foil := r.URL.Query().Get("foil") // e.g. "non_foil", "foil", "etched_foil"
	cn := r.URL.Query().Get("cn")

	if name == "" {
		jsonError(w, "query param 'name' is required", http.StatusBadRequest)
		return
	}

	result, err := external.LookupMTGCard(name, setCode, cn, foil)
	if err != nil {
		if err.Error() == "card not found" {
			jsonError(w, "card not found", http.StatusNotFound)
			return
		}
		jsonError(w, "scryfall lookup failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	jsonOK(w, result)
}


// GET /api/admin/lookup/pokemon?name=<name>&set=<setID>
func (h *LookupHandler) Pokemon(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	setID := r.URL.Query().Get("set")

	if name == "" {
		jsonError(w, "query param 'name' is required", http.StatusBadRequest)
		return
	}

	result, err := external.LookupPokemonCard(name, setID)
	if err != nil {
		if err.Error() == "card not found" {
			jsonError(w, "card not found", http.StatusNotFound)
			return
		}
		jsonError(w, "pokémon TCG API lookup failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	jsonOK(w, result)
}
