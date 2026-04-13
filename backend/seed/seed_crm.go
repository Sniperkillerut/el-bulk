package main

import (
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedCRM(db *sqlx.DB, adminID string, customers []CustomerSeeded, bountyIDs []string) error {
	logger.Info("📋 Seeding CRM data (subscribers, notes, client requests, bounty offers)...")

	if len(customers) == 0 {
		logger.Warn("No customers available — skipping CRM seed")
		return nil
	}

	// ── Newsletter Subscribers ──────────────────────────────────────────────
	logger.Info("  📨 Seeding newsletter subscribers...")
	// First 18 customers are subscribers
	for i, c := range customers {
		if i >= 18 {
			break
		}
		var custID *string
		custID = &c.ID
		if _, err := db.Exec(`
			INSERT INTO newsletter_subscriber (email, customer_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO NOTHING
		`, c.Email, custID, daysAgo(randInt(5, 90))); err != nil {
			return fmt.Errorf("failed to seed subscriber '%s': %w", c.Email, err)
		}
	}
	// A few anonymous subscribers that never placed orders
	anonEmails := []string{
		"anoncard@proton.me",
		"magic_fan_bogota@gmail.com",
		"tcg_collector_co@outlook.com",
		"pokemon_master_cali@gmail.com",
		"cartas_cali@yahoo.com",
	}
	for _, email := range anonEmails {
		db.Exec(`
			INSERT INTO newsletter_subscriber (email, created_at)
			VALUES ($1, $2)
			ON CONFLICT (email) DO NOTHING
		`, email, daysAgo(randInt(10, 60)))
	}

	// ── Customer Notes ──────────────────────────────────────────────────────
	logger.Info("  📝 Seeding customer notes (journal of interactions)...")

	type Note struct {
		Idx     int
		Content string
	}
	notes := []Note{
		{0, "Cliente VIP — siempre paga puntualmente. Tiene preferencia por cartas en japonés."},
		{0, "Compró el Black Lotus proxy para display personal. No es para torneo."},
		{1, "Interesada en Commander. Preguntó por el deck de Elfos del Bulk."},
		{1, "Quiere sellado de Bloomburrow si llega más stock."},
		{2, "Jugador competitivo de Modern. Busca Ragavan NM en inglés."},
		{2, "Confirmó que recibió el order EB-000003 en buen estado."},
		{3, "Primera compra. Llegó por recomendación de Santiago (cliente #14)."},
		{4, "Comprador frecuente de Pokémon. Preguntó por torneos de Saturday Draft."},
		{4, "Trajo 5000 cartas de bulk — pagamos $100.000 COP. Cliente feliz."},
		{5, "Nuevo cliente — vio el store en Instagram."},
		{6, "Devolvió un Thoughtseize que estaba MP pero facturado como LP. Reembolso emitido."},
		{7, "Quiere que lo notifiquemos cuando llegue Force of Will en japonés."},
		{8, "Pidió que le separemos 2x Goblin Warchief para la próxima semana."},
		{9, "Pagó por transferencia pero el número de referencia no llegó. Pendiente verificar."},
		{10, "Viajó desde Medellín para esta compra. Muy agradecido con el servicio."},
	}

	for _, n := range notes {
		if n.Idx >= len(customers) {
			continue
		}
		c := customers[n.Idx]
		content := n.Content
		if _, err := db.Exec(`
			INSERT INTO customer_note (customer_id, content, admin_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, c.ID, content, adminID, daysAgo(randInt(1, 45))); err != nil {
			return fmt.Errorf("failed to seed customer note for %s: %w", c.ID, err)
		}
	}

	// ── Client Requests (from storefront form) ──────────────────────────────
	logger.Info("  🔍 Seeding client requests...")

	type ClientReq struct {
		CustomerIdx int // -1 = guest
		Name        string
		Contact     string
		CardName    string
		SetName     string
		Details     string
		Status      string
	}

	requests := []ClientReq{
		{0, "Sebastián Rodríguez", "3101234001", "Ragavan, Nimble Pilferer", "Modern Horizons 2",
			"NM en inglés, sin foil. Necesito 4.", "pending"},
		{1, "Valentina Gómez", "3201234002", "Umbra Mystic", "Bloomburrow",
			"Cualquier condición.", "solved"},
		{2, "Andrés Martínez", "@andres.mtcg", "Force of Will", "Alliances",
			"Versión japonesa preferiblemente.", "accepted"},
		{3, "Camila Torres", "3001234004", "Sheoldred, the Apocalypse", "Dominaria United",
			"Showcase o borderless, NM.", "pending"},
		{4, "Felipe Herrera", "fherrera.tcg@gmail.com", "The One Ring", "LTR",
			"Borderless. Presupuesto máx $450.000.", "rejected"},
		// Guest requests (no customer_id)
		{-1, "Alejandro Rincón", "+57 312 999 8877", "Pikachu VMAX", "Celebrations",
			"Para colección personal.", "pending"},
		{-1, "Isabella V.", "@isa_cartas", "Blue-Eyes White Dragon", "LOB 1st Ed",
			"En buen estado. Pago en efectivo.", "pending"},
		{-1, "Julia M.", "juliamtg@gmail.com", "Cyclonic Rift", "Commander 2014",
			"NM, no foil.", "solved"},
		{6, "Juan Carlos Moreno", "3171234007", "Liliana of the Veil", "Innistrad",
			"LP o NM. Para mazo moderno.", "accepted"},
		{5, "Laura Castillo", "3051234006", "Dragon Shield Matte Red",
			"", "¿Están en stock? Cuántas paquetes tienen?", "solved"},
		{-1, "Pedro Ortega", "3309998877", "Tarmogoyf", "Future Sight",
			"Condición LP, inglés.", "cancelled"},
		{7, "Diana Peña", "3061234008", "Moxen completos (5 moxen)",
			"Alpha o Beta", "Para colección enmarcada.", "pending"},
	}

	for _, req := range requests {
		var custID *string
		if req.CustomerIdx >= 0 && req.CustomerIdx < len(customers) {
			custID = &customers[req.CustomerIdx].ID
		}
		db.Exec(`
			INSERT INTO client_request (customer_id, customer_name, customer_contact, card_name, set_name, details, status, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, custID, req.Name, req.Contact, req.CardName, nilIfEmpty(req.SetName), nilIfEmpty(req.Details), req.Status, daysAgo(randInt(1, 45)))
	}

	// ── Bounty Offers ────────────────────────────────────────────────────────
	logger.Info("  💰 Seeding bounty offers...")

	if len(bountyIDs) == 0 {
		logger.Warn("  No bounties to make offers on — skipping bounty offers")
		return nil
	}

	type BountyOffer struct {
		BountyIdx   int
		CustomerIdx int
		Quantity    int
		Condition   string
		Status      string
		Notes       string
		AdminNotes  string
	}

	offers := []BountyOffer{
		{0, 5, 2, "NM", "pending", "Las tengo en mi colección personal, precio negociable.", ""},
		{0, 6, 1, "LP", "accepted", "Una sola copia. Muy buen estado visual.", "Precio acordado: $240k. Cliente trae el lunes."},
		{1, 7, 4, "NM", "pending", "Las compré sealed y no las necesito.", ""},
		{2, 8, 1, "NM", "fulfilled", "The One Ring NM sin foil.", "Comprada y agregada al inventario."},
		{3, 9, 1, "LP", "rejected", "Versión DMU, no 2XM.", "No cumple el criterio de set solicitado."},
		{4, 2, 2, "NM", "pending", "Foil showcase versión.", ""},
		{5, 10, 1, "MP", "pending", "Moderately played pero funcional.", ""},
		{6, 3, 3, "NM", "accepted", "Lightning Bolt promo de prerelease 30th.", "Aceptadas. Precio: $48k c/u."},
		{7, 11, 1, "NM", "cancelled", "No pude venir a la tienda.", ""},
		{8, 12, 1, "NM", "pending", "Charizard SAR con top loader.", ""},
		{9, 4, 1, "LP", "fulfilled", "Rayquaza VMax Alt Art. Ligero desgaste en las esquinas.", "Aceptada. Pagamos $700k COP."},
	}

	for _, o := range offers {
		bIdx := o.BountyIdx % len(bountyIDs)
		if o.CustomerIdx >= len(customers) {
			continue
		}
		cond := o.Condition
		notes := o.Notes
		adminNotes := o.AdminNotes

		db.Exec(`
			INSERT INTO bounty_offer (bounty_id, customer_id, quantity, condition, status, notes, admin_notes, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`,
			bountyIDs[bIdx], customers[o.CustomerIdx].ID, o.Quantity,
			nilIfEmpty(cond), o.Status,
			nilIfEmpty(notes), nilIfEmpty(adminNotes),
			daysAgo(randInt(1, 30)),
		)
	}

	logger.Info("✅ CRM data seeded")
	return nil
}
