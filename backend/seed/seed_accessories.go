package main

import (
	"fmt"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedAccessories(db *sqlx.DB, cats CategoryMap, storage StorageMap) ([]string, error) {
	logger.Info("🛡️  Seeding Accessories...")

	type Acc struct {
		Name       string
		TCG        string
		Price      float64
		Stock      int
		ImageURL   string
		StorageLoc string
		CatSlug    string
		Desc       string
	}

	items := []Acc{
		// ── Sleeves ─────────────────────────────────────────────────────────
		{
			"Dragon Shield Matte Sleeves - Jet Black (100ct)", "accessories",
			62000, 25,
			"https://www.dragonshield.com/wp-content/uploads/2020/05/AT-10001-DragonShield-MatteJetBlack.jpg",
			"Counter Display", "sale",
			"Premium Japanese-size inner sleeves with matte finish. Excellent for competitive play.",
		},
		{
			"Dragon Shield Matte Sleeves - Navy Blue (100ct)", "accessories",
			62000, 20,
			"https://www.dragonshield.com/wp-content/uploads/2020/05/AT-10017-DragonShield-MatteNavy.jpg",
			"Counter Display", "sale",
			"The iconic navy blue Dragon Shield sleeve. Durable and smooth for shuffling.",
		},
		{
			"Dragon Shield Matte Sleeves - Red (100ct)", "accessories",
			62000, 18,
			"https://www.dragonshield.com/wp-content/uploads/2020/05/AT-10002-DragonShield-MatteRed.jpg",
			"Counter Display", "new-arrivals",
			"Bold red matte finish. Perfect for aggressive Commander builds.",
		},
		{
			"Dragon Shield Art Sleeves - Rayquaza Classic (100ct)", "pokemon",
			85000, 12,
			"https://www.dragonshield.com/wp-content/uploads/2022/10/AT-17506-DragonShieldArtRayquazaClassic.jpg",
			"Counter Display", "hot-items",
			"Classic Rayquaza art in collaboration with Pokémon. Matte quality finish.",
		},
		{
			"KMC Perfect Fit Sleeves - Clear (100ct)", "accessories",
			28000, 50,
			"https://images.cardboardcrack.com/product/sleeves/kmc-perfect-fit.jpg",
			"Counter Display", "sale",
			"Japanese-size inner sleeves for double-sleeving. Crystal clear PVC.",
		},
		{
			"Ultimate Guard Epic Art Sleeves - MTG Plains (100ct)", "mtg",
			75000, 15,
			"https://ultimateguard.com/media/catalog/product/u/g/ug-matte-art-standard.jpg",
			"Counter Display", "new-arrivals",
			"Limited edition art sleeves featuring MTG basic land art. Standard size.",
		},

		// ── Deck Boxes ──────────────────────────────────────────────────────
		{
			"Ultimate Guard Flip'n'Tray 100+ - XenoSkin Blue", "accessories",
			125000, 8,
			"https://ultimateguard.com/media/catalog/product/f/l/flip-n-tray-blue.jpg",
			"Showcase B", "featured",
			"The tournament-standard deck box. Holds 100+ double-sleeved cards with magnetic closure.",
		},
		{
			"Ultimate Guard Flip'n'Tray 100+ - XenoSkin Black", "accessories",
			125000, 10,
			"https://ultimateguard.com/media/catalog/product/f/l/flip-n-tray-black.jpg",
			"Showcase B", "featured",
			"Sleek black edition of the premium Flip'n'Tray deck box.",
		},
		{
			"Gamegenic Squire 100+ Convertible - Red", "accessories",
			92000, 15,
			"https://cdn.gamegenic.com/wp-content/uploads/2021/07/Squire_100_Convertible_Red.png",
			"Showcase B", "sale",
			"Convertible deck box — remove the top to use as a tray. Holds 100+ double-sleeved cards.",
		},
		{
			"Legion Dragon Hide Binder (9-Pocket, 18 Pages)", "accessories",
			75000, 12,
			"https://images.cardboardcrack.com/product/binders/legion-dragon-hide-binder-black.jpg",
			"Showcase A", "featured",
			"Premium faux leather binder for your collection. 162 card capacity side-loading pockets.",
		},

		// ── Playmats ────────────────────────────────────────────────────────
		{
			"Ultra PRO Playmat - Jace, the Mind Sculptor (MTG)", "mtg",
			95000, 6,
			"https://cdn.shopify.com/s/files/1/0267/5084/7992/products/ultrapro-jace-playmat.jpg",
			"Storage Box 3", "staff-picks",
			"Official licensed playmat featuring Jace, the Mind Sculptor art. 24\" x 13.5\" rubber base.",
		},
		{
			"Ultra PRO Playmat - Pikachu (Pokémon)", "pokemon",
			88000, 8,
			"https://cdn.shopify.com/s/files/1/0267/5084/7992/products/ultrapro-pikachu-playmat.jpg",
			"Storage Box 3", "hot-items",
			"Pokémon-themed playmat featuring classic Pikachu art. Non-slip rubber base.",
		},
		{
			"Inked Gaming Custom Playmat - El Bulk Store Exclusive", "accessories",
			145000, 5,
			"https://images.unsplash.com/photo-1598214886806-c87b84b7078b?q=80&w=800&auto=format&fit=crop",
			"Showcase A", "featured",
			"El Bulk store-branded playmat. Exclusive design featuring the store logo and card art. Limited run.",
		},

		// ── Storage & Organization ───────────────────────────────────────────
		{
			"BCW Short Box - Fits 200 Standard Cards", "accessories",
			18000, 30,
			"https://images.cardboardcrack.com/product/storage/bcw-short-box.jpg",
			"Bulk Bin", "budget-builds",
			"Cardboard storage box that holds 200 standard-sized sleeved cards. Great for bulk storage.",
		},
		{
			"Ultimate Guard Boulder Deck Case 100+ - Solid White", "accessories",
			68000, 20,
			"https://ultimateguard.com/media/catalog/product/b/o/boulder-100-solid-white.jpg",
			"Counter Display", "new-arrivals",
			"Rock-solid polypropylene deck case. Holds 100+ double-sleeved cards.",
		},

		// ── Tokens & Accessories ─────────────────────────────────────────────
		{
			"Chessex Dice Set (7-Die Polyhedral) - Blue/Gold", "accessories",
			35000, 20,
			"https://images.unsplash.com/photo-1585507252242-11fe632c26e8?q=80&w=800&auto=format&fit=crop",
			"Counter Display", "sale",
			"7-die set in translucent blue with gold numbering. Essential for any RPG or complex card game.",
		},
		{
			"Magic Life Counter - Automatic Electric Counter", "mtg",
			42000, 15,
			"https://images.unsplash.com/photo-1591488320449-011701bb6704?q=80&w=800&auto=format&fit=crop",
			"Counter Display", "new-arrivals",
			"Automatic life counter for Magic: The Gathering. Counts from 40 down to 0.",
		},
	}

	var ids []string
	for i, item := range items {
		createdAt := daysAgo(randInt(3, 45))
		costBasis := item.Price * 0.55

		var pID string
		err := db.QueryRow(`
			INSERT INTO product (
				name, tcg, category, set_name, set_code,
				price_source, price_cop_override, stock,
				image_url, description, cost_basis_cop, created_at
			) VALUES ($1, $2, 'accessories', 'N/A', 'N/A', 'manual', $3, $4, $5, $6, $7, $8)
			RETURNING id
		`,
			item.Name, item.TCG, item.Price, item.Stock,
			item.ImageURL, item.Desc, costBasis, createdAt,
		).Scan(&pID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert accessory '%s': %w", item.Name, err)
		}
		ids = append(ids, pID)

		if sid, ok := storage[item.StorageLoc]; ok {
			db.Exec(`
				INSERT INTO product_storage (product_id, storage_id, quantity)
				VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
			`, pID, sid, item.Stock)
		}
		if catID, ok := cats[item.CatSlug]; ok {
			db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
		}
		// Every 3rd accessory is also "new-arrivals"
		if i%3 == 0 {
			if catID, ok := cats["new-arrivals"]; ok {
				db.Exec(`INSERT INTO product_category (product_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, pID, catID)
			}
		}
	}

	logger.Info("✅ %d accessories seeded", len(ids))
	return ids, nil
}
