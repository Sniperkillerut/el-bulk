import { FoilTreatment, CardTreatment, ScryfallCard, PriceSource, DeckCard, Condition } from './types';

/**
 * Mapping from Scryfall specialized finish tags to internal FoilTreatment values.
 */
export const SPECIALIZED_FOIL_MAP: Record<string, FoilTreatment> = {
  'ripplefoil': 'ripple_foil',
  'surgefoil': 'surge_foil',
  'galaxyfoil': 'galaxy_foil',
  'oilslick': 'oil_slick',
  'stepandcompleat': 'step_and_compleat',
  'textured': 'textured_foil',
  'confettifoil': 'confetti_foil',
  'doublerainbow': 'double_rainbow',
  'neonink': 'neon_ink',
  'platinumfoil': 'platinum_foil',
  'glossy': 'glossy',
  'holofoil': 'holo_foil',
  'etched': 'etched_foil',
  'foil': 'foil'
};

/**
 * Tags that mandate a specific foil finish and exclude non-foil.
 */
export const EXCLUSIVE_FOIL_TAGS = Object.keys(SPECIALIZED_FOIL_MAP);

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

  if (fe.includes('serialized')) return 'serialized';
  if (fe.includes('showcase')) return 'showcase';
  if (card.border_color === 'borderless') return 'borderless';
  if (fe.includes('extendedart')) return 'extended_art';
  if (fe.includes('retro')) return 'retro_frame';
  if (card.frame === '1997') return 'legacy_border';
  if (card.full_art) return 'full_art';
  if (card.textless) return 'textless';
  if ((card.promo_types || []).includes('promo') || card.promo) return 'promo';

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

export function resolveFoilTreatment(card: ScryfallCard | undefined, preferredFinish?: FoilTreatment): FoilTreatment {
  if (!card) return 'non_foil';
  const pt = card.promo_types || [];
  const finishes = card.finishes || [];

  // 0. If preferredFinish is valid for this card, keep it
  if (preferredFinish) {
    if (preferredFinish === 'non_foil' && finishes.includes('nonfoil')) return 'non_foil';
    if (preferredFinish === 'foil' && finishes.includes('foil')) return 'foil';
    if (preferredFinish === 'etched_foil' && finishes.includes('etched')) return 'etched_foil';

    // For specialized foils, check if the promo tag is present
    const tag = Object.keys(SPECIALIZED_FOIL_MAP).find(k => SPECIALIZED_FOIL_MAP[k] === preferredFinish);
    if (tag && pt.includes(tag)) return preferredFinish;
  }

  // 1. Check specialized foil tags using global map
  for (const [tag, treatment] of Object.entries(SPECIALIZED_FOIL_MAP)) {
    if (pt.includes(tag)) return treatment;
  }

  // 2. Check finishes
  if (finishes.includes('etched')) return 'etched_foil';
  if (finishes.includes('foil')) return 'foil';

  // 3. Fallback to non-foil (covers both explicit 'nonfoil' and cases where no finish is specified)
  return 'non_foil';
}

/**
 * For initial card selection/preview: determines the most "representative"
 * foil treatment if a specific one hasn't been chosen yet.
 */
export function getInitialFoilTreatment(card: ScryfallCard | undefined): FoilTreatment {
  if (!card) return 'non_foil';

  const finishes = card.finishes || [];
  if (finishes.includes('nonfoil')) return 'non_foil';

  const resolved = resolveFoilTreatment(card);
  if (resolved !== 'non_foil') return resolved;

  if (finishes.length === 1 && finishes.includes('foil')) return 'foil';
  if (finishes.length === 1 && finishes.includes('etched')) return 'etched_foil';

  return 'non_foil';
}

/**
 * Heuristically identifies a foil treatment from a raw string (e.g. from CSV).
 */
export function identifyFoilFromString(str: string | undefined): FoilTreatment {
  if (!str) return 'non_foil';
  const s = str.toLowerCase().trim();

  // Check specialized foils using global map
  for (const [tag, treatment] of Object.entries(SPECIALIZED_FOIL_MAP)) {
    // Some tags are shorter in common parlance (e.g. 'ripple' vs 'ripplefoil')
    // but the map has the full Scryfall tags. We'll check if the tag is in the string.
    if (s.includes(tag) || (tag.endsWith('foil') && s.includes(tag.replace('foil', '')))) {
      return treatment;
    }
  }

  return 'non_foil';
}

/**
 * Normalizes a condition string (e.g. from CSV) to the standard TCG codes.
 */
export function normalizeCondition(str: string | undefined): Condition {
  if (!str) return 'NM';
  const s = str.toLowerCase().trim();

  // NM: Near Mint, Mint, M, NM
  if (s === 'nm' || s === 'near mint' || s === 'mint' || s === 'm' || s === 'nearmint') return 'NM';

  // LP: Lightly Played, Excellent, EX, LP, PLD, Pld
  if (s === 'lp' || s === 'lightly played' || s === 'excellent' || s === 'ex' || s === 'pld' || s === 'lightlyplayed') return 'LP';

  // MP: Moderately Played, Good, GD, MP
  if (s === 'mp' || s === 'moderately played' || s === 'good' || s === 'gd' || s === 'moderatelyplayed') return 'MP';

  // HP: Heavily Played, Poor, PR, HP
  if (s === 'hp' || s === 'heavily played' || s === 'poor' || s === 'pr' || s === 'heavilyplayed') return 'HP';

  // DMG: Damaged, DMG
  if (s === 'dmg' || s === 'damaged') return 'DMG';

  // Catch-all fallbacks for common substrings if exact matches fail
  if (s.includes('damage')) return 'DMG';
  if (s.includes('heavily')) return 'HP';
  if (s.includes('moderate')) return 'MP';
  if (s.includes('lightly') || s.includes('ex')) return 'LP';
  if (s.includes('mint')) return 'NM';

  return 'NM'; // Default to NM
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
    .filter(p => {
      const isSetMatch = p.set?.toLowerCase() === set.toLowerCase();
      // For SLD (Secret Lair), we usually want to show all arts regardless of individual treatment matches,
      // as they are often grouped together in complex ways.
      if (set.toLowerCase() === 'sld') return isSetMatch;
      return isSetMatch && resolveCardTreatment(p) === treatment;
    })
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
export function getPromoOptions(prints: ScryfallCard[], set: string, treatment: CardTreatment, cn: string): { promos: string[], hasStandard: boolean } {
  const options = new Set<string>();
  let hasStandard = false;
  
  prints
    .filter(p => !set || p.set?.toLowerCase() === set.toLowerCase() && resolveCardTreatment(p) === treatment && (p.collector_number === cn || !cn))
    .forEach(p => {
      const pt = p.promo_types || [];
      // Filter out structural tags (already handled by treatment selector) 
      // AND foil tags (handled by foil selector, not promo selector)
      const redundantTags = ['showcase', 'borderless', 'normal', 'boosterfun', ...EXCLUSIVE_FOIL_TAGS];
      
      // A print supports 'Standard' if it has no promo tags, or only redundant ones
      const nonRedundant = pt.filter(t => !redundantTags.includes(t));
      if (nonRedundant.length === 0) {
        hasStandard = true;
      }

      nonRedundant.forEach(t => options.add(t));
    });

  return { 
    promos: Array.from(options),
    hasStandard 
  };
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
    const finishes = p.finishes || [];
    const pt = p.promo_types || [];

    if (finishes.includes('nonfoil')) {
      // Exclude non-foil if the card has an exclusive foil tag (surgefoil, galaxyfoil, etc.)
      const hasExclusiveTag = pt.some(t => EXCLUSIVE_FOIL_TAGS.includes(t));
      
      if (!hasExclusiveTag) {
        options.add('non_foil');
      }
    }
    
    if (finishes.includes('foil') || finishes.includes('etched')) {
      const resolved = resolveFoilTreatment(p);
      if (resolved === 'foil' || resolved === 'etched_foil') {
        options.add(resolved);
      } else {
        // It's a specialized foil (surge_foil, ripple_foil, etc.)
        // Show it if the card itself has the matching promo tag — these are card properties,
        // not user-selectable promos (e.g. surgefoil co-exists with universesbeyond on the same card)
        const matchingTag = Object.keys(SPECIALIZED_FOIL_MAP).find(tag => SPECIALIZED_FOIL_MAP[tag] === resolved);
        if (matchingTag && pt.includes(matchingTag)) {
          options.add(resolved);
        }
      }
    }
  });

  // If no specialized foils or standard foils matched the current promo selection,
  // we only add non_foil as a fallback if at least one print actually supports it.
  if (options.size === 0) {
    const supportsNonFoil = finalMatches.some(p => (p.finishes || []).includes('nonfoil'));
    if (supportsNonFoil) {
      options.add('non_foil');
    }
  }

  return Array.from(options);
}

/**
 * Legacy compatibility functions for ProductEditModal
 */
export function getTreatmentType(card: ScryfallCard): CardTreatment {
  return resolveCardTreatment(card);
}

export function applyPrintPrices(card: ScryfallCard | undefined, foil: FoilTreatment, source: PriceSource, currentPriceReference?: number | string): number {
  if (!card) return 0;
  const p = card.prices || {};

  if (source === 'tcgplayer') {
    if (foil === 'etched_foil') return parseFloat(p.usd_etched || '0');
    if (foil !== 'non_foil' && foil !== '') return parseFloat(p.usd_foil || '0');
    return parseFloat(p.usd || '0');
  } else if (source === 'cardmarket') {
    if (foil !== 'non_foil' && foil !== '') return parseFloat(p.eur_foil || p.eur || '0');
    return parseFloat(p.eur || '0');
  }

  return currentPriceReference !== undefined ? Number(currentPriceReference) : 0;
}

export interface MTGMetadata {
  language: string;
  color_identity: string;
  rarity: string;
  cmc: number | '';
  collector_number: string;
  set_code: string;
  promo_type: string;
  is_legendary: boolean;
  is_land: boolean;
  is_basic_land: boolean;
  oracle_text: string;
  artist: string;
  type_line: string;
  border_color: string;
  frame: string;
  full_art: boolean;
  textless: boolean;
  frame_effects: string[];
}

export function extractMTGMetadata(card: ScryfallCard | undefined): MTGMetadata {
  const defaults: MTGMetadata = {
    language: 'en',
    color_identity: '',
    rarity: '',
    cmc: 0,
    collector_number: '',
    set_code: '',
    promo_type: '',
    is_legendary: false,
    is_land: false,
    is_basic_land: false,
    oracle_text: '',
    artist: '',
    type_line: '',
    border_color: '',
    frame: '',
    full_art: false,
    textless: false,
    frame_effects: []
  };

  if (!card) return defaults;
  
  return {
    language: 'en',
    color_identity: card.color_identity?.join(',') || '',
    rarity: card.rarity || '',
    cmc: card.cmc ?? 0,
    collector_number: card.collector_number || '',
    set_code: card.set || '',
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
    frame_effects: card.frame_effects || [],
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

export function getSuggestedPrice(card: ScryfallCard | undefined, foil: FoilTreatment, source: PriceSource, settings?: { usd_to_cop_rate: number, eur_to_cop_rate: number, ck_to_cop_rate?: number }): number | undefined {
  if (!card || !settings) return undefined;
  const ref = applyPrintPrices(card, foil, source);
  if (!ref) return 0;

  let rate: number;
  switch (source) {
    case 'tcgplayer':
      rate = settings.usd_to_cop_rate;
      break;
    case 'cardkingdom':
      // CK has its own rate; fall back to usd_to_cop_rate if not set for backwards compat
      rate = settings.ck_to_cop_rate ?? settings.usd_to_cop_rate;
      break;
    default:
      rate = settings.eur_to_cop_rate;
  }
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

  const groups: Record<string, DeckCard[]> = {
    Creatures: [],
    Instants: [],
    Sorceries: [],
    Artifacts: [],
    Enchantments: [],
    Planeswalkers: [],
    Battles: [],
    Lands: [],
    Other: []
  };

  let total = 0;
  cards.forEach(card => {
    total += card.quantity;
    if (!card.type_line) {
      counts.Other += card.quantity;
      groups.Other.push(card);
      return;
    }

    const typeLine = card.type_line.split(/[—\-]/)[0].toLowerCase();

    if (typeLine.includes('land')) { counts.Lands += card.quantity; groups.Lands.push(card); }
    else if (typeLine.includes('creature')) { counts.Creatures += card.quantity; groups.Creatures.push(card); }
    else if (typeLine.includes('instant')) { counts.Instants += card.quantity; groups.Instants.push(card); }
    else if (typeLine.includes('sorcery')) { counts.Sorceries += card.quantity; groups.Sorceries.push(card); }
    else if (typeLine.includes('artifact')) { counts.Artifacts += card.quantity; groups.Artifacts.push(card); }
    else if (typeLine.includes('enchantment')) { counts.Enchantments += card.quantity; groups.Enchantments.push(card); }
    else if (typeLine.includes('planeswalker')) { counts.Planeswalkers += card.quantity; groups.Planeswalkers.push(card); }
    else if (typeLine.includes('battle')) { counts.Battles += card.quantity; groups.Battles.push(card); }
    else { counts.Other += card.quantity; groups.Other.push(card); }
  });

  const summary = Object.entries(counts)
    .filter(([, count]) => count > 0)
    .map(([type, count]) => `${count} ${type.toLowerCase()}`)
    .join(' - ');

  return { total, counts, summary, groups };
}

/**
 * Filters a comma-separated string of promo tags to remove redundant or misleading
 * information based on the current card's treatment and foil finish.
 */
export function filterPromoTags(promoType: string | undefined, foilTreatment: string, cardTreatment: string): string[] {
  if (!promoType || promoType === 'none') return [];
  const foilTags = EXCLUSIVE_FOIL_TAGS;

  return promoType.split(',').filter(t => {
    const s = t?.trim();
    if (!s) return false;
    const normalized = s.toLowerCase().replace(/[^a-z0-9]/g, '');

    // Filter out foil-related tags if the product is non-foil
    if (foilTreatment === 'non_foil' && foilTags.includes(normalized)) return false;

    // Filter out redundant treatment tags
    if (normalized === 'showcase' && cardTreatment === 'showcase') return false;
    if (normalized === 'extendedart' && cardTreatment === 'extended_art') return false;
    if (normalized === 'borderless' && cardTreatment === 'borderless') return false;
    if (normalized === 'boosterfun') return false;

    return true;
  });
}
