import { FoilTreatment, CardTreatment, ScryfallCard, PriceSource, DeckCard } from './types';

/**
 * Extracts the best image URL from a Scryfall card object,
 * handling double-faced cards as well.
 */
export function getScryfallImage(card: ScryfallCard | undefined): string {
  if (!card) return '';
  if (card.image_uris?.normal) return card.image_uris.normal;
  if (card.card_faces?.[0]?.image_uris?.normal) return card.card_faces[0].image_uris.normal;
  return '';
}

/**
 * Identifies the card treatment (Normal, Borderless, etc.) based on Scryfall metadata.
 */
export function resolveCardTreatment(card: ScryfallCard): CardTreatment {
  const fe = card.frame_effects || [];
  const pt = card.promo_types || [];

  if (fe.includes('serialized')) return 'serialized';
  if (pt.includes('stepandcompleat')) return 'step_and_compleat';
  if (pt.includes('judgegift')) return 'judge_promo';
  if (fe.includes('showcase')) return 'showcase';
  if (card.border_color === 'borderless') return 'borderless';
  if (fe.includes('extendedart')) return 'extended_art';
  if (fe.includes('retro')) return 'retro_frame';
  if (card.frame === '1997') return 'legacy_border';
  if (card.full_art) return 'full_art';
  if (card.textless) return 'textless';
  if (pt.includes('promo') || card.promo) return 'promo';

  return 'normal';
}

/**
 * Determines a friendly name for an art variation (e.g., "Alt Art", "Japanese").
 */
export function resolveArtVariation(card: ScryfallCard): string {
  const promo = card.promo_types || [];
  if (promo.includes('jpwalker')) return 'japanese';
  if (promo.includes('concept')) return 'concept_art';
  return '';
}

export function resolveFoilTreatment(card: ScryfallCard | undefined): FoilTreatment {
  if (!card) return 'non_foil';
  const pt = card.promo_types || [];

  if (pt.includes('oilslick')) return 'oil_slick';
  if (pt.includes('stepandcompleat')) return 'step_and_compleat';
  if (pt.includes('galaxyfoil')) return 'galaxy_foil';
  if (pt.includes('surgefoil')) return 'surge_foil';
  if (pt.includes('textured')) return 'textured_foil';
  if (pt.includes('confettifoil')) return 'confetti_foil';
  if (pt.includes('neonink')) return 'neon_ink';
  if (pt.includes('doublerainbow')) return 'double_rainbow';
  if (pt.includes('glossy')) return 'glossy';

  const finishes = card.finishes || [];
  if (finishes.includes('etched')) return 'etched_foil';
  if (finishes.includes('foil')) return 'foil';
  return 'non_foil';
}

/**
 * Extracts pricing data from Scryfall.
 */
export function extractPrices(card: ScryfallCard) {
  const p = card.prices || {};
  const usd = p.usd || p.usd_foil || p.usd_etched || '0';
  const eur = p.eur || '0';

  return {
    usd: parseFloat(usd as string),
    eur: parseFloat(eur as string)
  };
}

/**
 * Filter waterfall: 1. Get all unique treatments across all prints
 */
export function getTreatmentOptions(prints: ScryfallCard[], set: string): CardTreatment[] {
  const treatments = new Set<CardTreatment>();
  prints
    .filter(p => !set || p.set?.toLowerCase() === set.toLowerCase())
    .forEach(p => treatments.add(resolveCardTreatment(p)));
  return Array.from(treatments);
}

export interface ArtOption {
  cn: string;
  artist: string;
}

/**
 * Filter waterfall: 2. Get art variations for a chosen treatment (Returns ArtOption with cn and artist)
 */
export function getArtOptions(prints: ScryfallCard[], set: string, treatment: CardTreatment): ArtOption[] {
  const options = new Map<string, string>(); // cn -> artist
  prints
    .filter(p => p.set?.toLowerCase() === set.toLowerCase() && resolveCardTreatment(p) === treatment)
    .forEach(p => {
      if (p.collector_number) {
        options.set(p.collector_number, p.artist || 'Unknown');
      }
    });

  return Array.from(options.entries())
    .map(([cn, artist]) => ({ cn, artist }))
    .sort((a, b) => a.cn.localeCompare(b.cn, undefined, { numeric: true }));
}

/**
 * Filter waterfall: 3. Get promo types for a chosen treatment + art (cn)
 */
export function getPromoOptions(prints: ScryfallCard[], set: string, treatment: CardTreatment, cn: string): string[] {
  const options = new Set<string>();
  prints
    .filter(p => !set || p.set?.toLowerCase() === set.toLowerCase() && resolveCardTreatment(p) === treatment && (p.collector_number === cn || !cn))
    .forEach(p => {
      (p.promo_types || [])
        .forEach(pt => options.add(pt));
    });
  return Array.from(options);
}

/**
 * Filter waterfall: 4. Get foil options for a chosen combination
 */
export function getFoilOptions(prints: ScryfallCard[], set: string, treatment: CardTreatment, cn: string, promo: string): FoilTreatment[] {
  const options = new Set<FoilTreatment>();

  const matches = prints.filter(p => {
    const isSetMatch = !set || p.set?.toLowerCase() === set.toLowerCase();
    const isTreatmentMatch = resolveCardTreatment(p) === treatment;
    const isCnMatch = p.collector_number === cn || !cn;

    // Improved promo matching: check if the requested promo (single tag) is in the print's tags
    const isPromoMatch = !promo || promo === 'none' || (p.promo_types || []).includes(promo);

    return isSetMatch && isTreatmentMatch && isCnMatch && isPromoMatch;
  });

  // If no exact matches (e.g. promo not available for this cn), try relaxing the promo check
  const finalMatches = matches.length > 0 ? matches : prints.filter(p => {
    return (!set || p.set?.toLowerCase() === set.toLowerCase()) &&
      resolveCardTreatment(p) === treatment &&
      (p.collector_number === cn || !cn);
  });

  finalMatches.forEach(p => {
    if (p.finishes?.includes('nonfoil')) options.add('non_foil');
    const r = resolveFoilTreatment(p);
    if (r !== 'non_foil') options.add(r as FoilTreatment);
  });

  return Array.from(options);
}

/**
 * Legacy compatibility functions for ProductEditModal
 */
export function getTreatmentType(card: ScryfallCard): CardTreatment {
  return resolveCardTreatment(card);
}

export function applyPrintPrices(card: ScryfallCard | undefined, foil: FoilTreatment, source: PriceSource): number {
  if (!card) return 0;
  const p = card.prices || {};
  if (source === 'tcgplayer') {
    if (foil === 'foil') return parseFloat(p.usd_foil || '0');
    if (foil === 'etched_foil') return parseFloat(p.usd_etched || '0');
    return parseFloat(p.usd || '0');
  } else if (source === 'cardmarket') {
    return parseFloat(p.eur || '0');
  }
  return 0;
}

export function extractMTGMetadata(card: ScryfallCard | undefined) {
  if (!card) return {};
  return {
    language: 'en',
    color_identity: card.color_identity?.join('') || '',
    rarity: card.rarity || '',
    cmc: card.cmc || 0,
    collector_number: card.collector_number || '',
    promo_type: (card.promo_types || []).join(',') || 'none',
    is_legendary: card.type_line?.includes('Legendary') || false,
    is_land: card.type_line?.includes('Land') || false,
    is_basic_land: card.type_line?.includes('Basic Land') || false,
    oracle_text: card.oracle_text || '',
    artist: card.artist || '',
    type_line: card.type_line || '',
    border_color: card.border_color || '',
    frame: card.frame || '',
    full_art: card.full_art || false,
    textless: card.textless || false,
  };
}

export function findMatchingPrint(prints: ScryfallCard[], set: string, treatment: CardTreatment, cn: string, promo: string, foil: FoilTreatment): ScryfallCard | undefined {
  // Normalize foil for finish check
  const requestedFinish = foil === 'non_foil' ? 'nonfoil' : (foil === 'foil' ? 'foil' : (foil === 'etched_foil' ? 'etched' : foil));

  // 1. Try exact match (set + treatment + cn + promo + foil)
  let match = prints.find(p => {
    if (p.set?.toLowerCase() !== set.toLowerCase()) return false;
    if (resolveCardTreatment(p) !== treatment) return false;
    if (cn && p.collector_number !== cn) return false;
    if (promo && promo !== 'none' && !(p.promo_types || []).includes(promo)) return false;
    if (foil && !(p.finishes || []).includes(requestedFinish)) return false;
    return true;
  });

  // 2. Try match without foil (foil often isn't a separate record in Scryfall paper search)
  if (!match) {
    match = prints.find(p => {
      if (p.set?.toLowerCase() !== set.toLowerCase()) return false;
      if (resolveCardTreatment(p) !== treatment) return false;
      if (cn && p.collector_number !== cn) return false;
      if (promo && promo !== 'none' && !(p.promo_types || []).includes(promo)) return false;
      return true;
    });
  }

  // 3. Try match without promo (if promo was specified but not found)
  if (!match && promo && promo !== 'none') {
    match = prints.find(p => {
      if (p.set?.toLowerCase() !== set.toLowerCase()) return false;
      if (resolveCardTreatment(p) !== treatment) return false;
      if (cn && p.collector_number !== cn) return false;
      return true;
    });
  }

  // 4. Try match without cn (if treatment changed, cn might not match)
  if (!match) {
    match = prints.find(p => {
      if (p.set?.toLowerCase() !== set.toLowerCase()) return false;
      if (resolveCardTreatment(p) !== treatment) return false;
      return true;
    });
  }

  // 5. Final fallback to set match or first print
  return match || prints.find(p => p.set?.toLowerCase() === set.toLowerCase()) || prints[0];
}

export function getSuggestedPrice(card: ScryfallCard | undefined, foil: FoilTreatment, source: PriceSource, settings?: { usd_to_cop_rate: number, eur_to_cop_rate: number }): number | undefined {
  if (!card || !settings) return undefined;
  const ref = applyPrintPrices(card, foil, source);
  if (!ref) return 0;
  
  const rate = source === 'tcgplayer' ? settings.usd_to_cop_rate : settings.eur_to_cop_rate;
  if (!rate) return 0;

  // Round to nearest 100 COP as a standard
  return Math.round((ref * rate) / 100) * 100;
}

/**
 * Categorizes deck cards by their MTG type (ignoring subtypes)
 * and returns summary statistics.
 */
export function getDeckAnalytics(cards: DeckCard[]) {
  const counts: Record<string, number> = {
    Lands: 0,
    Creatures: 0,
    Instants: 0,
    Sorceries: 0,
    Artifacts: 0,
    Enchantments: 0,
    Planeswalkers: 0,
    Battles: 0,
    Other: 0
  };

  let total = 0;
  cards.forEach(card => {
    total += card.quantity;
    if (!card.type_line) {
      counts.Other += card.quantity;
      return;
    }
    
    // Split by Em-dash or hyphen and take the first part (types, not subtypes)
    const typeLine = card.type_line.split(/[—\-]/)[0].toLowerCase();
    
    if (typeLine.includes('land')) counts.Lands += card.quantity;
    else if (typeLine.includes('creature')) counts.Creatures += card.quantity;
    else if (typeLine.includes('instant')) counts.Instants += card.quantity;
    else if (typeLine.includes('sorcery')) counts.Sorceries += card.quantity;
    else if (typeLine.includes('artifact')) counts.Artifacts += card.quantity;
    else if (typeLine.includes('enchantment')) counts.Enchantments += card.quantity;
    else if (typeLine.includes('planeswalker')) counts.Planeswalkers += card.quantity;
    else if (typeLine.includes('battle')) counts.Battles += card.quantity;
    else counts.Other += card.quantity;
  });

  const summary = Object.entries(counts)
    .filter(([, count]) => count > 0)
    .map(([type, count]) => `${count} ${type.toLowerCase()}`)
    .join(' - ');

  return { total, counts, summary };
}
