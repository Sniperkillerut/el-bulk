import { Product, ProductListResponse } from './types';

const isServer = typeof window === 'undefined';
const API_BASE = isServer 
  ? (process.env.INTERNAL_API_URL || 'http://backend:8080')
  : ''; // Use relative path in browser to trigger Next.js rewrites

export interface ProductFilters {
  tcg?: string;
  category?: string;
  search?: string;
  foil?: string;
  treatment?: string;
  condition?: string;
  featured?: boolean;
  storage_id?: string;
  page?: number;
  page_size?: number;
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

  if (!res.ok) throw new Error(`Failed to fetch products: ${res.statusText}`);
  return res.json();
}

export async function fetchProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/products/${id}`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Product not found');
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
  if (!res.ok) throw new Error('Invalid credentials');
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
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    throw new Error(`Failed to fetch products: ${res.statusText}`);
  }
  return res.json();
}

export async function adminCreateProduct(token: string, data: Partial<Product>): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/admin/products`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create product');
  }
  return res.json();
}

export async function adminUpdateProduct(token: string, id: string, data: Partial<Product>): Promise<Product> {
  const res = await fetch(`${API_BASE}/api/admin/products/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to update product');
  }
  return res.json();
}

export async function adminDeleteProduct(token: string, id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/products/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to delete product');
}

// ---------------------------------------------------------------------------
// External card lookup (Scryfall for MTG, Pokémon TCG API for Pokémon)
// ---------------------------------------------------------------------------

export interface CardLookupResult {
  image_url: string;
  set_name: string;
  set_code: string;
  collector_number?: string;
  price_tcgplayer?: number;  // USD
  price_cardmarket?: number; // EUR
}

export async function lookupMTGCard(
  token: string,
  name: string,
  set?: string,
  foil?: string,
): Promise<CardLookupResult> {
  const params = new URLSearchParams({ name });
  if (set) params.set('set', set);
  if (foil) params.set('foil', foil);
  const res = await fetch(`${API_BASE}/api/admin/lookup/mtg?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'MTG lookup failed');
  }
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
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'Pokémon lookup failed');
  }
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
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to create storage location');
  }
  return res.json();
}

export async function adminUpdateStorage(token: string, id: string, name: string): Promise<import('./types').StoredIn> {
  const res = await fetch(`${API_BASE}/api/admin/storage/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to update storage location');
  }
  return res.json();
}

export async function adminDeleteStorage(token: string, id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/storage/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to delete storage location');
  }
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
  if (!res.ok) {
    if (res.status === 401) throw new Error('401 Unauthorized');
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to update product storage');
  }
  return res.json();
}
