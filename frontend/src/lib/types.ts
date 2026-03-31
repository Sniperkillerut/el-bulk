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
  collector_number?: string;
  promo_type?: string;

  // MTG Metadata
  language: string;
  color_identity?: string;
  rarity?: string;
  cmc?: number;
  is_legendary: boolean;
  is_historic: boolean;
  is_land: boolean;
  is_basic_land: boolean;
  art_variation?: string;
  oracle_text?: string;
  artist?: string;
  type_line?: string;
  border_color?: string;
  frame?: string;
  full_art: boolean;
  textless: boolean;

  created_at: string;
  updated_at: string;
  cart_count?: number;
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

export interface Facets {
  condition: Record<string, number>;
  foil: Record<string, number>;
  treatment: Record<string, number>;
  rarity: Record<string, number>;
  language: Record<string, number>;
  color: Record<string, number>;
  collection: Record<string, number>;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
  facets: Facets;
  query_time_ms: number;
}

export interface CartItem {
  product: Product;
  quantity: number;
}

export const FOIL_LABELS: Record<string, string> = {
  non_foil: 'Non-Foil',
  foil: 'Foil',
  etched_foil: 'Etched Foil',
  glossy: 'Glossy',
  surge_foil: 'Surge Foil',
  textured_foil: 'Textured Foil',
  galaxy_foil: 'Galaxy Foil',
  step_and_compleat: 'Step-and-Compleat Foil',
  oil_slick: 'Oil Slick Raised Foil',
  neon_ink: 'Neon Ink Foil',
  confetti_foil: 'Confetti Foil',
  double_rainbow: 'Double Rainbow Foil',
};

export interface UserProfile {
  id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  id_number?: string;
  address?: string;
  auth_provider?: string;
  auth_provider_id?: string;
  avatar_url?: string;
}

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

export const resolveLabel = (key: string, map: Record<string, string>) => {
  if (!key || key === 'none') return '';
  if (map[key]) return map[key];
  // Dynamic formatting for future-proofing
  let label = key.replace(/([a-z])([A-Z])/g, '$1 $2'); // camelCase
  label = label.replace(/oilslick/gi, 'Oil Slick');
  label = label.replace(/stepandcompleat/gi, 'Step-and-Compleat');
  label = label.replace(/silverfoil/gi, 'Silver Foil');
  // Avoid double spaces when _foil or camelCase already added a space/separator
  label = label.replace(/([a-z])foil$/i, '$1 Foil');
  label = label.replace(/_/g, ' '); 
  // Consolidate multiple spaces
  label = label.replace(/\s+/g, ' ');

  // Title Case, but preserve the hyphen in Step-and-Compleat
  return label.replace(/\b\w/g, l => l.toUpperCase())
    .replace(/Step-And-Compleat/g, 'Step-and-Compleat')
    .trim();
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

export interface TCG {
  id: string;   // slug: mtg, pokemon, etc.
  name: string;
  is_active: boolean;
  item_count?: number;
  created_at?: string;
}

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
  show_badge: boolean;
  searchable: boolean;
  item_count?: number;
  created_at?: string;
}

export interface CustomCategoryInput {
  name: string;
  slug?: string;
  is_active?: boolean;
  show_badge?: boolean;
  searchable?: boolean;
}

// ── Orders ──────────────────────────────────────────────

export interface Customer {
  id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone: string;
  id_number?: string;
  address?: string;
  created_at: string;
}

export interface Order {
  id: string;
  order_number: string;
  customer_id: string;
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled';
  payment_method: string;
  total_cop: number;
  notes?: string;
  created_at: string;
  completed_at?: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id?: string;
  product_name: string;
  product_set?: string;
  foil_treatment?: string;
  card_treatment?: string;
  condition?: string;
  unit_price_cop: number;
  quantity: number;
  stored_in_snapshot?: string;
}

export interface OrderItemDetail extends OrderItem {
  image_url?: string;
  stock: number;
  stored_in: StorageLocation[];
}

export interface OrderDetail {
  order: Order;
  customer: Customer;
  items: OrderItemDetail[];
}

export interface OrderWithCustomer extends Order {
  customer_name: string;
  item_count: number;
}

export interface OrderListResponse {
  orders: OrderWithCustomer[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateOrderRequest {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  id_number: string;
  address: string;
  payment_method: string;
  notes: string;
  items: { product_id: string; quantity: number }[];
}

export const PAYMENT_METHODS: Record<string, string> = {
  cash: 'Efectivo',
  transfer: 'Transferencia',
  nequi: 'Nequi',
  daviplata: 'Daviplata',
};

export const ORDER_STATUS_LABELS: Record<string, string> = {
  pending: 'Pendiente',
  confirmed: 'Confirmado',
  completed: 'Completado',
  cancelled: 'Cancelado',
};

export interface BulkProductInput extends Partial<Omit<Product, 'categories' | 'stored_in'>> {
  category_ids?: string[];
  storage_items?: StorageLocation[];
}

export interface ScryfallCard {
  id: string;
  name: string;
  set: string;
  set_name?: string;
  collector_number: string;
  image_uris?: {
    normal?: string;
    small?: string;
    large?: string;
    border_crop?: string;
  };
  card_faces?: Array<{
    image_uris?: {
      normal?: string;
    };
  }>;
  prices?: {
    usd?: string | null;
    usd_foil?: string | null;
    usd_etched?: string | null;
    eur?: string | null;
  };
  finishes?: string[];
  frame_effects?: string[];
  promo_types?: string[];
  type_line?: string;
  border_color?: string;
  full_art?: boolean;
  textless?: boolean;
  security_stamp?: string;
  artist?: string;
  frame?: string;
  promo?: boolean;
  digital?: boolean;
}

export interface Bounty {
  id: string;
  name: string;
  tcg: string;
  set_name?: string;
  condition?: Condition;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  collector_number?: string;
  promo_type?: string;
  language: string;
  target_price?: number;
  hide_price: boolean;
  quantity_needed: number;
  image_url?: string;
  price_source: PriceSource;
  price_reference?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface BountyInput {
  name: string;
  tcg: string;
  set_name?: string;
  condition?: Condition;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  collector_number?: string;
  promo_type?: string;
  language: string;
  target_price?: number;
  hide_price: boolean;
  quantity_needed: number;
  image_url?: string;
  price_source: PriceSource;
  price_reference?: number;
  is_active?: boolean;
}

export interface ClientRequest {
  id: string;
  customer_id?: string;
  customer_name: string;
  customer_contact: string;
  card_name: string;
  set_name?: string;
  details?: string;
  status: 'pending' | 'accepted' | 'rejected' | 'solved';
  created_at: string;
}

export interface ClientRequestInput {
  customer_id?: string;
  customer_name: string;
  customer_contact: string;
  card_name: string;
  set_name?: string;
  details?: string;
}

export interface BountyOffer {
  id: string;
  bounty_id: string;
  customer_id: string;
  customer_name: string;
  customer_contact: string;
  condition?: string;
  quantity: number;
  status: 'pending' | 'accepted' | 'rejected' | 'fulfilled';
  notes?: string;
  admin_notes?: string;
  created_at: string;
  updated_at: string;
  bounty_name?: string;
}

export interface BountyOfferInput {
  bounty_id: string;
  customer_name: string;
  customer_contact: string;
  condition?: Condition;
  quantity: number;
  notes?: string;
}

export interface BountyWithOffers extends Bounty {
  offers: BountyOffer[];
}

export interface Notice {
  id: string;
  title: string;
  slug: string;
  content_html: string;
  featured_image_url?: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface NoticeInput {
  title: string;
  slug: string;
  content_html: string;
  featured_image_url?: string;
  is_published: boolean;
}

export interface NewsletterSubscriber {
  id: string;
  email: string;
  customer_id?: string;
  first_name?: string;
  last_name?: string;
  created_at: string;
}

export interface CustomerNote {
  id: string;
  customer_id: string;
  order_id?: string;
  content: string;
  admin_id?: string;
  admin_name?: string;
  created_at: string;
}

export interface CustomerStats extends UserProfile {
  order_count: number;
  total_spend: number;
  is_subscriber: boolean;
  latest_note?: string | null;
  request_count: number;
  active_request_count: number;
  offer_count: number;
  active_offer_count: number;
  created_at: string;
}

export interface CustomerDetail extends UserProfile {
  orders: Order[];
  notes: CustomerNote[];
  requests: ClientRequest[];
  offers: BountyOffer[];
  is_subscriber: boolean;
}

export interface AdminStats {
  total_sku_records: number;
  query_speed_ms: number;
  database_size: string;
  cache_hit_ratio: number;
  active_connections: number;
  max_connections: number;
}
