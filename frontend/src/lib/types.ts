export type FoilTreatment =
  | 'non_foil' | 'foil' | 'holo_foil' | 'platinum_foil'
  | 'ripple_foil' | 'etched_foil' | 'galaxy_foil';

export type CardTreatment =
  | 'normal' | 'full_art' | 'extended_art' | 'borderless' | 'showcase'
  | 'legacy_border' | 'textless' | 'judge_promo' | 'promo' | 'alternate_art';

export type Condition = 'NM' | 'LP' | 'MP' | 'HP' | 'DMG';

export type PriceSource = 'tcgplayer' | 'cardmarket' | 'manual';

export interface Product {
  id: string;
  name: string;
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories';
  set_name?: string;
  set_code?: string;
  condition?: Condition;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  // Pricing: price is the computed COP value returned by the backend
  price: number;
  price_reference?: number;    // raw USD or EUR from external source
  price_source: PriceSource;   // which source/currency price_reference is in
  price_cop_override?: number; // admin's explicit COP override
  stock: number;
  image_url?: string;
  description?: string;
  featured: boolean;
  created_at: string;
  updated_at: string;
}

/** Admin-configurable exchange rates. */
export interface Settings {
  usd_to_cop_rate: number; // TCGPlayer prices
  eur_to_cop_rate: number; // Cardmarket prices
}


export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
}

export interface CartItem {
  product: Product;
  quantity: number;
}

export const FOIL_LABELS: Record<FoilTreatment, string> = {
  non_foil: 'Non-Foil',
  foil: 'Foil',
  holo_foil: 'Holo Foil',
  platinum_foil: 'Platinum Foil',
  ripple_foil: 'Ripple Foil',
  etched_foil: 'Etched Foil',
  galaxy_foil: 'Galaxy Foil',
};

export const TREATMENT_LABELS: Record<CardTreatment, string> = {
  normal: 'Regular',
  full_art: 'Full Art',
  extended_art: 'Extended Art',
  borderless: 'Borderless',
  showcase: 'Showcase',
  legacy_border: 'Legacy Border',
  textless: 'Textless',
  judge_promo: 'Judge Promo',
  promo: 'Promo',
  alternate_art: 'Alternate Art',
};

export const TCG_LABELS: Record<string, string> = {
  mtg: 'Magic: The Gathering',
  pokemon: 'Pokémon',
  lorcana: 'Disney Lorcana',
  onepiece: 'One Piece',
  yugioh: 'Yu-Gi-Oh!',
};

export const TCG_SHORT: Record<string, string> = {
  mtg: 'MTG',
  pokemon: 'Pokémon',
  lorcana: 'Lorcana',
  onepiece: 'One Piece',
  yugioh: 'Yu-Gi-Oh!',
};

export const KNOWN_TCGS = ['mtg', 'pokemon', 'lorcana', 'onepiece', 'yugioh'];
