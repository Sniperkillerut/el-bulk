import { Product, StoredIn, CustomCategory, Condition, FoilTreatment, CardTreatment, PriceSource, DeckCard } from '@/lib/types';

export type TabId = 'variant' | 'pricing' | 'deck';

export interface FormState {
  id?: string;
  name: string;
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories' | 'store_exclusives';
  set_name: string;
  set_code: string;
  condition: Condition;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  price: number;
  price_reference: number | '';
  price_source: PriceSource;
  price_cop_override: number | '';
  stock: number;
  image_url: string;
  description: string;
  collector_number: string;
  promo_type: string;
  language: string;
  color_identity: string;
  rarity: string;
  cmc: number | '';
  is_legendary: boolean;
  is_historic: boolean;
  is_land: boolean;
  is_basic_land: boolean;
  art_variation: string;
  oracle_text: string;
  artist: string;
  type_line: string;
  border_color: string;
  frame: string;
  full_art: boolean;
  textless: boolean;
  category_ids: string[];
  storage_items: { stored_in_id: string; quantity: number }[];
  deck_cards: DeckCard[];
}

export interface ProductEditModalProps {
  editProduct: Product | null;
  token: string;
  storageLocations: StoredIn[];
  categories: CustomCategory[];
  tcgs: any[];
  settings?: any;
  storageFilter?: string;
  onClose: () => void;
  onSaved: () => void;
  onSaveAndNew?: (lastForm: { tcg: string; category: string; condition: string; storageIds: string[] }) => void;
}
