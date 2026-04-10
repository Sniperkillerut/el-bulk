package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/external"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

// LookupHandler handles card image/metadata lookups from external APIs.
type LookupHandler struct{}

func NewLookupHandler() *LookupHandler {
	return &LookupHandler{}
}

// GET /api/admin/lookup/mtg?name=<name>&set=<setCode>&foil=<foilTreatment>
func (h *LookupHandler) MTG(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering LookupHandler.MTG | Query: %s", r.URL.RawQuery)
	name := r.URL.Query().Get("name")
	sid := r.URL.Query().Get("sid")
	setCode := r.URL.Query().Get("set")
	foil := r.URL.Query().Get("foil") // e.g. "non_foil", "foil", "etched_foil"
	cn := r.URL.Query().Get("cn")

	if name == "" && sid == "" {
		render.Error(w, "query param 'name' or 'sid' is required", http.StatusBadRequest)
		return
	}

	result, err := external.LookupMTGCard(sid, name, setCode, cn, foil)
	if err != nil {
		logger.Error("MTG lookup failed: %v", err)
		if err.Error() == "card not found" {
			render.Error(w, "card not found", http.StatusNotFound)
			return
		}
		render.Error(w, "scryfall lookup failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	render.Success(w, result)
}

// POST /api/admin/lookup/mtg/batch
func (h *LookupHandler) BatchMTG(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering LookupHandler.BatchMTG")
	var input struct {
		Identifiers []external.CardIdentifier `json:"identifiers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Failed to decode batch lookup input: %v", err)
		render.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	results, err := external.BatchLookupMTGCard(input.Identifiers)
	if err != nil {
		logger.Error("Batch MTG lookup failed: %v", err)
		render.Error(w, "scryfall batch lookup failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	render.Success(w, results)
}


// GET /api/admin/lookup/pokemon?name=<name>&set=<setID>
func (h *LookupHandler) Pokemon(w http.ResponseWriter, r *http.Request) {
	logger.Trace("Entering LookupHandler.Pokemon | Query: %s", r.URL.RawQuery)
	name := r.URL.Query().Get("name")
	setID := r.URL.Query().Get("set")

	if name == "" {
		render.Error(w, "query param 'name' is required", http.StatusBadRequest)
		return
	}

	result, err := external.LookupPokemonCard(name, setID)
	if err != nil {
		logger.Error("Pokemon lookup failed: %v", err)
		if err.Error() == "card not found" {
			render.Error(w, "card not found", http.StatusNotFound)
			return
		}
		render.Error(w, "pokémon TCG API lookup failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	render.Success(w, result)
}
