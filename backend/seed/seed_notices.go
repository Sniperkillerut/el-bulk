package main

import (
	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

func seedNotices(db *sqlx.DB) {
	logger.Info("📰 Seeding notices (blog posts)...")

	type Notice struct {
		Title       string
		Slug        string
		HTML        string
		ImageURL    string
		IsPublished bool
		DaysAgo     int
	}

	notices := []Notice{
		{
			Title:       "🎉 Now in Stock: Modern Horizons 3 Collector Boosters!",
			Slug:        "mh3-collector-boosters-arrival",
			IsPublished: true, DaysAgo: 3,
			ImageURL: "https://cards.scryfall.io/art_crop/front/0/3/0337a311-2c41-4438-b70e-6b87ba04cc3c.jpg",
			HTML: `<h2>The Most Anticipated Set of 2024 Has Arrived</h2>
<p>We just received a fresh allocation of <strong>Modern Horizons 3 Collector Booster Boxes</strong>. These are in extremely limited supply — only 3 boxes in stock.</p>
<ul>
  <li>12 Collector Boosters per box</li>
  <li>Exclusive serialized card treatments</li>
  <li>New Eldrazi titans, broken artifacts, and new planeswalkers</li>
</ul>
<p>Price: <strong>$1.950.000 COP</strong> per box. Stop by or order online while they last.</p>
<div data-card-preview="Ulamog, the Defiler" data-set="MH3"></div>
<p>👉 <a href="/sealed/mtg">Browse all MTG sealed →</a></p>`,
		},
		{
			Title:       "Bloomburrow Play Booster Box — Now Available",
			Slug:        "bloomburrow-play-booster-arrival",
			IsPublished: true, DaysAgo: 7,
			ImageURL: "https://cards.scryfall.io/art_crop/front/3/f/3f5ac5af-10a4-4afe-a76c-5e9eb5fdc06b.jpg",
			HTML: `<h2>Critter Season Has Begun!</h2>
<p><strong>Bloomburrow</strong> has arrived! We have Play Booster Boxes in stock ready to be cracked open.</p>
<p>The set features adorable animal civilizations and powerful limited archetypes. It's been a sleeper hit with Commander players for its powerful spells and unique token producers.</p>
<ul>
  <li>36 Play Boosters per box</li>
  <li>Plenty of foil mythics and special treatments to open</li>
</ul>
<p>Price: <strong>$680.000 COP</strong> per box.</p>
<p>We're also opening a draft pod this Saturday at 2pm! Limited spots. Ask us on WhatsApp to reserve.</p>`,
		},
		{
			Title:       "Saturday Commander Night — Every Week from 4pm",
			Slug:        "commander-night-weekly",
			IsPublished: true, DaysAgo: 14,
			ImageURL: "https://images.unsplash.com/photo-1550745165-9bc0b252726f?q=80&w=800",
			HTML: `<h2>Weekly Commander Nights at El Bulk</h2>
<p>Starting this week, we're hosting <strong>Commander Night every Saturday from 4:00pm - 8:00pm</strong> at the store.</p>
<h3>Format</h3>
<ul>
  <li>Open pods — any power level welcome</li>
  <li>4-player tables preferred</li>
  <li>Prizes for fun/interesting moments, not just winning</li>
</ul>
<h3>Rules</h3>
<ul>
  <li>No $0-cost infinite combos in the first game with strangers</li>
  <li>Proxies allowed for cards over $50.000 COP</li>
  <li>Bring your own deck or borrow one of ours</li>
</ul>
<p>No entry fee. Just show up and play!</p>`,
		},
		{
			Title:       "Pokémon SV: Stellar Crown Singles Now Listed",
			Slug:        "stellar-crown-singles-listed",
			IsPublished: true, DaysAgo: 4,
			ImageURL: "https://images.pokemontcg.io/sv7/logo.png",
			HTML: `<h2>Stellar Crown Singles Have Been Sorted and Priced!</h2>
<p>After cracking several boxes of <strong>Stellar Crown</strong>, we've finished sorting and pricing the singles. All cards are now available on the store.</p>
<p>Highlights from the pull sessions:</p>
<ul>
  <li>Pikachu ex (Stellar) — Beautiful artwork, strong demand already</li>
  <li>Terapagos ex & Terapagos ex (Special Illustration Rare)</li>
  <li>Multiple support trainer SIRs</li>
</ul>
<p>👉 <a href="/singles/pokemon">Browse all Pokémon singles →</a></p>
<p>We're doing a <strong>buy 3 singles, get 10% off</strong> promotion through the end of the month!</p>`,
		},
		{
			Title:       "El Bulk Now Accepts Pokémon TCG Pocket Trades",
			Slug:        "tcg-pocket-trade-program",
			IsPublished: true, DaysAgo: 21,
			ImageURL: "https://images.unsplash.com/photo-1498887960847-2a5e46312788?q=80&w=800",
			HTML: `<h2>Trade Your Digital Cards for Physical Ones</h2>
<p>With Pokémon TCG Pocket becoming huge, we're now experimenting with a <strong>Pocket ↔ Physical Trade Program</strong>.</p>
<p>Bring in your spare Pocket bourglass points or digital card duplicates and we'll give you credit toward physical cards in our store.</p>
<h3>How It Works</h3>
<ol>
  <li>Show us your Pocket collection on your phone</li>
  <li>We evaluate the trade value</li>
  <li>Get store credit for physical singles or sealed</li>
</ol>
<p>This is still experimental — come in and let's figure it out together.</p>`,
		},
		{
			Title:       "New Arrivals: Dragon Shield and Ultimate Guard Sleeves",
			Slug:        "new-sleeves-accessories-arrival",
			IsPublished: true, DaysAgo: 8,
			ImageURL: "https://www.dragonshield.com/wp-content/uploads/2020/05/AT-10001-DragonShield-MatteJetBlack.jpg",
			HTML: `<h2>Fresh Accessories Stock Just Arrived!</h2>
<p>We got a fresh shipment of <strong>Dragon Shield</strong> and <strong>Ultimate Guard</strong> products this week!</p>
<h3>New in Stock:</h3>
<ul>
  <li>Dragon Shield Matte — Jet Black, Navy Blue, Red (100ct)</li>
  <li>Dragon Shield Art — Rayquaza Classic (100ct)</li>
  <li>Ultimate Guard Flip'n'Tray 100+ — Blue &amp; Black</li>
  <li>KMC Perfect Fit Sleeves — Clear (100ct)</li>
</ul>
<p>👉 <a href="/accessories">Browse all accessories →</a></p>
<p>Bundle deal: Buy 3 packs of sleeves and get a free deck box (Gamegenic Squire while supplies last).</p>`,
		},
		{
			Title:       "DRAFT: Yu-Gi-Oh! Master Duel Night — Coming Soon",
			Slug:        "yugioh-master-duel-night-draft",
			IsPublished: false, DaysAgo: 1, // DRAFT — for testing draft visibility
			HTML: `<h2>[DRAFT] This notice isn't published yet</h2>
<p>We're planning a Yu-Gi-Oh! Master Duel tournament night. Details TBD.</p>
<p>This is just a draft to test that unpublished notices don't appear on the storefront.</p>`,
		},
		{
			Title:       "We Buy Bulk! Our Current Prices",
			Slug:        "bulk-buying-prices-2025",
			IsPublished: true, DaysAgo: 30,
			ImageURL: "https://images.unsplash.com/photo-1585507252242-11fe632c26e8?q=80&w=800",
			HTML: `<h2>Got Cardboard? We Buy It All!</h2>
<p>Cleaning out your old boxes? We buy bulk cards at fair prices — no appointment needed. Just walk in!</p>
<table>
  <thead>
    <tr><th>Type</th><th>Price</th></tr>
  </thead>
  <tbody>
    <tr><td>Bulk Commons & Uncommons</td><td>$20.000 COP / 1,000</td></tr>
    <tr><td>Bulk Rares & Mythics</td><td>$1.000 COP / card</td></tr>
    <tr><td>Junk Rare Lots</td><td>$12.000 COP / 100</td></tr>
    <tr><td>Foil Commons & Uncommons</td><td>$40.000 COP / 500</td></tr>
    <tr><td>Basic Lands</td><td>$4.000 COP / 200</td></tr>
  </tbody>
</table>
<p>Large lots (1,000+ cards) may receive bonus offers. Store credit offers up to 25% more!</p>
<p>We accept MTG (all sets), Pokémon (English only), Lorcana, and One Piece.</p>`,
		},
	}

	for _, n := range notices {
		imageURL := nilIfEmpty(n.ImageURL)
		_, err := db.Exec(`
			INSERT INTO notice (title, slug, content_html, featured_image_url, is_published, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (slug) DO UPDATE SET
				title = EXCLUDED.title,
				content_html = EXCLUDED.content_html,
				featured_image_url = EXCLUDED.featured_image_url,
				is_published = EXCLUDED.is_published
		`, n.Title, n.Slug, n.HTML, imageURL, n.IsPublished, daysAgoFixed(n.DaysAgo))
		if err != nil {
			logger.Error("Failed to seed notice '%s': %v", n.Slug, err)
		}
	}
	logger.Info("✅ %d notices seeded (%d published, 1 draft)", len(notices), len(notices)-1)
}
