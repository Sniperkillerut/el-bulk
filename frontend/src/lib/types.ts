export type FoilTreatment = string;
export type CardTreatment = string;

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
  stored_in?: StorageLocation[];
  categories?: CustomCategory[];
  image_url?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

/** Admin-configurable exchange rates. */
export interface Settings {
  usd_to_cop_rate: number; // TCGPlayer prices
  eur_to_cop_rate: number; // Cardmarket prices
  contact_address: string;
  contact_phone: string;
  contact_email: string;
  contact_instagram: string;
  contact_hours: string;
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

export const FOIL_LABELS: Record<string, string> = {
  non_foil: 'Non-Foil',
  foil: 'Foil',
  holo_foil: 'Holo Foil',
  platinum_foil: 'Platinum Foil',
  ripple_foil: 'Surge Foil',
  etched_foil: 'Etched Foil',
  galaxy_foil: 'Galaxy Foil',
  galaxyfoil: 'Galaxy Foil',
  textured_foil: 'Textured Foil',
  glossy: 'Glossy',
  confetti_foil: 'Confetti Foil',
  invisible_ink: 'Invisible Ink',
};

export const TREATMENT_LABELS: Record<string, string> = {
  normal: 'Regular',
  full_art: 'Full Art (Regular)',
  extended_art: 'Extended Art',
  borderless: 'Borderless',
  showcase: 'Showcase',
  legacy_border: 'Classic Border',
  retro_frame: 'Retro Frame',
  textless: 'Textless',
  judge_promo: 'Judge Promo',
  promo: 'Promo',
  alternate_art: 'Alternate Art',
  step_and_compleat: 'Step-and-Compleat',
  serialized: 'Serialized/Numbered',
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

export interface StoredIn {
  id: string;
  name: string;
  item_count?: number;
}

export interface StorageLocation {
  stored_in_id: string;
  name: string;
  quantity: number;
}

export interface ProductStorageInput {
  stored_in_id: string;
  quantity: number;
}

export interface CustomCategory {
  id: string;
  name: string;
  slug: string;
  is_active: boolean;
  item_count?: number;
  created_at?: string;
}

export interface CustomCategoryInput {
  name: string;
  slug?: string;
  is_active?: boolean;
}
