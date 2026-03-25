import { Product, ProductListResponse } from './types';
import { remoteLogger } from './remoteLogger';

const isServer = typeof window === 'undefined';
const API_BASE = isServer 
  ? (process.env.INTERNAL_API_URL || 'http://backend:8080')
  : ''; // Use relative path in browser to trigger Next.js rewrites

async function logAndThrow(res: Response, defaultMsg: string): Promise<never> {
  let errorMessage = defaultMsg;
  try {
    const data = await res.clone().json();
    errorMessage = data.error || defaultMsg;
  } catch (e) {
    errorMessage = res.statusText || defaultMsg;
  }

  remoteLogger.error(`API Error [${res.status}]: ${errorMessage}`, {
    url: String(res.url),
    status: Number(res.status),
    statusText: String(res.statusText || res.status),
  });

  const error = new Error(errorMessage);
  (error as any)._remoteLogged = true;
  if (res.status === 401) throw error;
  throw error;
}

export interface ProductFilters {
  tcg?: string;
  category?: string;
  search?: string;
  foil?: string;
  treatment?: string;
  condition?: string;
  collection?: string;
  storage_id?: string;
  page?: number;
  page_size?: number;
  rarity?: string;
  language?: string;
  color?: string;
}

export async function fetchProducts(filters: ProductFilters = {}): Promise<ProductListResponse> {
  const params = new URLSearchParams();

  Object.entries(filters).forEach(([key, val]) => {
    if (val !== undefined && val !== '' && val !== null) {
      params.set(key, String(val));
    }
  });

  const res = await fetch(`${API_BASE}/api/products?${params.toString()}`, {
    cache: 'no-store',
  });

  if (!res.ok) await logAndThrow(res, 'Failed to fetch products');
  return res.json();
}

export async function fetchProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/products/${id}`, { cache: 'no-store' });
  if (!res.ok) await logAndThrow(res, 'Product not found');
  return res.json();
}

export async function fetchCategories(): Promise<import('./types').CustomCategory[]> {
  const res = await fetch(`${API_BASE}/api/categories`, { cache: 'no-store' });
  if (!res.ok) return [];
  return res.json();
}

export async function fetchTCGs(): Promise<string[]> {
  const res = await fetch(`${API_BASE}/api/tcgs`, { cache: 'no-store' });
  if (!res.ok) return [];
  const data = await res.json();
  return data.tcgs || [];
}

export async function fetchPublicSettings(): Promise<import('./types').Settings> {
  const res = await fetch(`${API_BASE}/api/settings`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch settings');
  return res.json();
}

// Admin API (requires token)
export async function adminLogin(username: string, password: string): Promise<string> {
  const res = await fetch(`${API_BASE}/api/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) await logAndThrow(res, 'Invalid credentials');
  const data = await res.json();
  return data.token;
}

export async function adminFetchProducts(token: string, filters: ProductFilters = {}): Promise<ProductListResponse> {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, val]) => {
    if (val !== undefined && val !== '') params.set(key, String(val));
  });
  const res = await fetch(`${API_BASE}/api/admin/products?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) await logAndThrow(res, 'Failed to fetch products');
  return res.json();
}

export async function adminCreateProduct(token: string, data: Partial<Product>): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/admin/products`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to create product');
  return res.json();
}

export async function adminUpdateProduct(token: string, id: string, data: Partial<Product>): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/admin/products/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to update product');
  return res.json();
}

export async function adminDeleteProduct(token: string, id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/products/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) await logAndThrow(res, 'Failed to delete product');
}

export async function adminBulkCreateProducts(token: string, products: import('./types').BulkProductInput[]): Promise<{ message: string; count: number }> {
  const res = await fetch(`${API_BASE}/api/admin/products/bulk`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(products),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to bulk create products');
  return res.json();
}

// ---------------------------------------------------------------------------
// External card lookup (Scryfall for MTG, Pokémon TCG API for Pokémon)
// ---------------------------------------------------------------------------

export interface CardLookupResult {
  name?: string;
  image_url: string;
  set_name: string;
  set_code: string;
  collector_number?: string;
  price_tcgplayer?: number;  // USD
  price_cardmarket?: number; // EUR

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
}

export async function lookupMTGCard(
  token: string,
  name: string,
  set?: string,
  cn?: string,
  foil?: string,
): Promise<CardLookupResult> {
  const params = new URLSearchParams({ name });
  if (set) params.set('set', set);
  if (cn) params.set('cn', cn);
  if (foil) params.set('foil', foil);
  const res = await fetch(`${API_BASE}/api/admin/lookup/mtg?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) await logAndThrow(res, 'MTG lookup failed');
  return res.json();
}

export async function adminBatchLookupMTG(
  token: string,
  identifiers: { name?: string; set?: string; cn?: string }[]
): Promise<CardLookupResult[]> {
  const res = await fetch(`${API_BASE}/api/admin/lookup/mtg/batch`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ identifiers }),
  });
  if (!res.ok) await logAndThrow(res, 'MTG batch lookup failed');
  return res.json();
}

export async function lookupPokemonCard(
  token: string,
  name: string,
  set?: string,
): Promise<CardLookupResult> {
  const params = new URLSearchParams({ name });
  if (set) params.set('set', set);
  const res = await fetch(`${API_BASE}/api/admin/lookup/pokemon?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) await logAndThrow(res, 'Pokémon lookup failed');
  return res.json();
}

// ---------------------------------------------------------------------------
// Admin: exchange rate settings
// ---------------------------------------------------------------------------

export async function getAdminSettings(token: string): Promise<import('./types').Settings> {
  const res = await fetch(`${API_BASE}/api/admin/settings`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to load settings');
  return res.json();
}

export async function updateAdminSettings(
  token: string,
  settings: Partial<import('./types').Settings>,
): Promise<import('./types').Settings> {
  const res = await fetch(`${API_BASE}/api/admin/settings`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(settings),
  });
  if (!res.ok) throw new Error('Failed to update settings');
  return res.json();
}

// ---------------------------------------------------------------------------
// Admin: price refresh
// ---------------------------------------------------------------------------

export async function triggerPriceRefresh(
  token: string,
): Promise<{ updated: number; errors: number }> {
  const res = await fetch(`${API_BASE}/api/admin/prices/refresh`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Price refresh failed');
  return res.json();
}

// ---------------------------------------------------------------------------
// Admin: Storage Locations
// ---------------------------------------------------------------------------

export async function adminFetchStorage(token: string): Promise<import('./types').StoredIn[]> {
  const res = await fetch(`${API_BASE}/api/admin/storage`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    throw new Error('Failed to fetch storage locations');
  }
  return res.json();
}

export async function adminCreateStorage(token: string, name: string): Promise<import('./types').StoredIn> {
  const res = await fetch(`${API_BASE}/api/admin/storage`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to create storage location');
  return res.json();
}

export async function adminUpdateStorage(token: string, id: string, name: string): Promise<import('./types').StoredIn> {
  const res = await fetch(`${API_BASE}/api/admin/storage/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to update storage location');
  return res.json();
}

export async function adminDeleteStorage(token: string, id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/storage/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) await logAndThrow(res, 'Failed to delete storage location');
}

// ---------------------------------------------------------------------------
// Admin: Custom Categories
// ---------------------------------------------------------------------------

export async function adminFetchCategories(token: string): Promise<import('./types').CustomCategory[]> {
  const res = await fetch(`${API_BASE}/api/admin/categories`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    throw new Error('Failed to fetch custom categories');
  }
  return res.json();
}

export async function adminCreateCategory(
  token: string, 
  name: string, 
  slug?: string, 
  is_active: boolean = true,
  show_badge: boolean = true,
  searchable: boolean = true
): Promise<import('./types').CustomCategory> {
  const res = await fetch(`${API_BASE}/api/admin/categories`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable }),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to create custom category');
  return res.json();
}

export async function adminUpdateCategory(
  token: string, 
  id: string, 
  name: string, 
  slug?: string, 
  is_active?: boolean,
  show_badge?: boolean,
  searchable?: boolean
): Promise<import('./types').CustomCategory> {
  const res = await fetch(`${API_BASE}/api/admin/categories/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable }),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to update custom category');
  return res.json();
}

export async function adminDeleteCategory(token: string, id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/categories/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) await logAndThrow(res, 'Failed to delete custom category');
}

// ---------------------------------------------------------------------------
// Admin: Product Storage Management
// ---------------------------------------------------------------------------

export async function adminUpdateProductStorage(token: string, productId: string, updates: import('./types').ProductStorageInput[]): Promise<import('./types').StorageLocation[]> {
  const res = await fetch(`${API_BASE}/api/admin/products/${productId}/storage`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(updates),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to update product storage');
  return res.json();
}

// ---------------------------------------------------------------------------
// Orders (Public)
// ---------------------------------------------------------------------------

export async function createOrder(data: import('./types').CreateOrderRequest): Promise<{ order_number: string; order_id: string; total_cop: number; status: string }> {
  const res = await fetch(`${API_BASE}/api/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to place order');
  return res.json();
}

// ---------------------------------------------------------------------------
// Admin: Orders
// ---------------------------------------------------------------------------

export async function adminFetchOrders(token: string, filters: { status?: string; search?: string; page?: number; page_size?: number } = {}): Promise<import('./types').OrderListResponse> {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, val]) => {
    if (val !== undefined && val !== '') params.set(key, String(val));
  });
  const res = await fetch(`${API_BASE}/api/admin/orders?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    throw new Error('Failed to fetch orders');
  }
  return res.json();
}

export async function adminFetchOrderDetail(token: string, id: string): Promise<import('./types').OrderDetail> {
  const res = await fetch(`${API_BASE}/api/admin/orders/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    throw new Error('Failed to fetch order detail');
  }
  return res.json();
}

export async function adminUpdateOrder(token: string, id: string, data: { status?: string; items?: { id: string; quantity: number }[] }): Promise<import('./types').OrderDetail> {
  const res = await fetch(`${API_BASE}/api/admin/orders/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to update order');
  return res.json();
}

export async function adminCompleteOrder(token: string, id: string, decrements: { product_id: string; stored_in_id: string; quantity: number }[]): Promise<import('./types').OrderDetail> {
  const res = await fetch(`${API_BASE}/api/admin/orders/${id}/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Rev bearer ${token}` },
    body: JSON.stringify({ decrements }),
  });
  if (!res.ok) await logAndThrow(res, 'Failed to complete order');
  return res.json();
}
