import { Product, ProductListResponse, NewsletterSubscriber, CustomerStats, CustomerDetail, AdminStats } from './types';
import { remoteLogger } from './remoteLogger';

const isServer = typeof window === 'undefined';
const API_BASE = isServer 
  ? (process.env.INTERNAL_API_URL || 'http://backend:8080')
  : (process.env.NEXT_PUBLIC_API_URL || ''); 

async function logAndThrow(res: Response, defaultMsg: string): Promise<never> {
  let errorMessage = defaultMsg;
  try {
    const data = await res.clone().json();
    errorMessage = data.error || defaultMsg;
  } catch {
    errorMessage = res.statusText || defaultMsg;
  }

  remoteLogger.error(`API Error [${res.status}]: ${errorMessage}`, {
    url: String(res.url),
    status: Number(res.status),
    statusText: String(res.statusText || res.status),
  });

  const error = new Error(errorMessage);
  (error as Error & { _remoteLogged?: boolean })._remoteLogged = true;
  if (res.status === 401) throw error;
  throw error;
}

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined | null>;
}

async function apiFetch<T>(endpoint: string, options: FetchOptions = {}, _token?: string): Promise<T> {
  const { params, headers: customHeaders, ...rest } = options;
  
  const url = new URL(`${API_BASE}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`);
  if (params) {
    Object.entries(params).forEach(([key, val]) => {
      if (val !== undefined && val !== '' && val !== null) {
        url.searchParams.set(key, String(val));
      }
    });
  }

  const headers = new Headers(customHeaders);
  // Manual token header removed - now using secure cookies automatically managed by the browser
  if (!headers.has('Content-Type') && (rest.method === 'POST' || rest.method === 'PUT' || rest.method === 'PATCH')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(url.toString(), {
    credentials: 'include',
    ...rest,
    headers,
  });

  if (!res.ok) {
    await logAndThrow(res, `API error on ${endpoint}`);
  }

  if (res.status === 204) return {} as T;
  return res.json();
}

export interface ProductFilters {
  [key: string]: string | number | boolean | undefined | null;
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
  sort_by?: string;
  sort_dir?: string;
}

export async function fetchProducts(filters: ProductFilters = {}): Promise<ProductListResponse> {
  return apiFetch<ProductListResponse>('/api/products', { params: filters, cache: 'no-store' });
}

export async function fetchProduct(id: string): Promise<Product> {
  return apiFetch<Product>(`/api/products/${id}`, { cache: 'default' });
}

// Metadata caching (in-memory session)
const metadataCache = new Map<string, { data: unknown; timestamp: number }>();
const CACHE_TTL = 300000; // 5 minutes

function getCached<T>(key: string): T | null {
  const entry = metadataCache.get(key);
  if (entry && Date.now() - entry.timestamp < CACHE_TTL) return entry.data as T;
  return null;
}

function setCached(key: string, data: unknown) {
  metadataCache.set(key, { data, timestamp: Date.now() });
}

export async function fetchCategories(): Promise<import('./types').CustomCategory[]> {
  const cached = getCached<import('./types').CustomCategory[]>('categories');
  if (cached) return cached;

  try {
    const data = await apiFetch<import('./types').CustomCategory[]>('/api/categories', { cache: 'default' });
    setCached('categories', data);
    return data;
  } catch {
    return [];
  }
}

export async function fetchTCGs(activeOnly: boolean = true): Promise<import('./types').TCG[]> {
  const key = `tcgs_${activeOnly}`;
  const cached = getCached<import('./types').TCG[]>(key);
  if (cached) return cached;

  try {
    const data = await apiFetch<import('./types').TCG[] | { tcgs: import('./types').TCG[] }>('/api/tcgs', { params: { active_only: activeOnly }, cache: 'default' });
    const tcgs = Array.isArray(data) ? data : (data.tcgs || []);
    setCached(key, tcgs);
    return tcgs;
  } catch {
    return [];
  }
}

export async function fetchPublicSettings(): Promise<import('./types').Settings> {
  const cached = getCached<import('./types').Settings>('settings');
  if (cached) return cached;

  const data = await apiFetch<import('./types').Settings>('/api/settings', { cache: 'default' });
  setCached('settings', data);
  return data;
}

// User Auth API (uses cookies)
export async function userFetchMe(): Promise<import('./types').UserProfile | null> {
  try {
    return await apiFetch<import('./types').UserProfile>('/api/auth/me');
  } catch {
    return null;
  }
}

export async function userLogout(): Promise<void> {
  await apiFetch('/api/auth/logout', { method: 'POST' });
}

// Admin API (uses cookies)
export async function adminLogin(username: string, password: string): Promise<string> {
  const data = await apiFetch<{ token: string }>('/api/admin/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
  return data.token;
}

export async function adminLogout(): Promise<void> {
  await apiFetch('/api/admin/logout', { method: 'POST' });
}

export async function adminFetchProducts(token: string, filters: ProductFilters = {}): Promise<ProductListResponse> {
  return apiFetch<ProductListResponse>('/api/admin/products', { params: filters, cache: 'no-store' }, token);
}

export async function adminCreateProduct(token: string, data: Partial<Product>): Promise<Product> {
  return apiFetch<Product>('/api/admin/products', {
    method: 'POST',
    body: JSON.stringify(data),
  }, token);
}

export async function adminUpdateProduct(token: string, id: string, data: Partial<Product>): Promise<Product> {
  return apiFetch<Product>(`/api/admin/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }, token);
}

export async function adminDeleteProduct(token: string, id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/products/${id}`, { method: 'DELETE' }, token);
}

export async function adminBulkCreateProducts(token: string, products: import('./types').BulkProductInput[]): Promise<{ message: string; count: number }> {
  return apiFetch<{ message: string; count: number }>('/api/admin/products/bulk', {
    method: 'POST',
    body: JSON.stringify(products),
  }, token);
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
  return apiFetch<CardLookupResult>('/api/admin/lookup/mtg', {
    params: { name, set, cn, foil },
  }, token);
}

export async function adminBatchLookupMTG(
  token: string,
  identifiers: { name?: string; set?: string; cn?: string }[]
): Promise<CardLookupResult[]> {
  return apiFetch<CardLookupResult[]>('/api/admin/lookup/mtg/batch', {
    method: 'POST',
    body: JSON.stringify({ identifiers }),
  }, token);
}

export async function lookupPokemonCard(
  token: string,
  name: string,
  set?: string,
): Promise<CardLookupResult> {
  return apiFetch<CardLookupResult>('/api/admin/lookup/pokemon', {
    params: { name, set },
  }, token);
}

// ---------------------------------------------------------------------------
// Admin: exchange rate settings
// ---------------------------------------------------------------------------

export async function getAdminSettings(token: string): Promise<import('./types').Settings> {
  const cached = getCached<import('./types').Settings>('admin_settings');
  if (cached) return cached;

  const data = await apiFetch<import('./types').Settings>('/api/admin/settings', { cache: 'default' }, token);
  setCached('admin_settings', data);
  return data;
}

export async function updateAdminSettings(
  token: string,
  settings: Partial<import('./types').Settings>,
): Promise<import('./types').Settings> {
  return apiFetch<import('./types').Settings>('/api/admin/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  }, token);
}

// ---------------------------------------------------------------------------
// Admin: price refresh
// ---------------------------------------------------------------------------

export async function triggerPriceRefresh(
  token: string,
): Promise<{ updated: number; errors: number }> {
  return apiFetch<{ updated: number; errors: number }>('/api/admin/prices/refresh', {
    method: 'POST',
  }, token);
}

// adminFetchStats removed from here to avoid duplication

// ---------------------------------------------------------------------------
// Admin: Storage Locations
// ---------------------------------------------------------------------------

export async function adminFetchStorage(token: string): Promise<import('./types').StoredIn[]> {
  const cached = getCached<import('./types').StoredIn[]>('admin_storage');
  if (cached) return cached;

  const data = await apiFetch<import('./types').StoredIn[]>('/api/admin/storage', { cache: 'default' }, token);
  setCached('admin_storage', data);
  return data;
}

export async function adminCreateStorage(token: string, name: string): Promise<import('./types').StoredIn> {
  const data = await apiFetch<import('./types').StoredIn>('/api/admin/storage', {
    method: 'POST',
    body: JSON.stringify({ name }),
  }, token);
  metadataCache.delete('admin_storage');
  return data;
}

export async function adminUpdateStorage(token: string, id: string, name: string): Promise<import('./types').StoredIn> {
  const data = await apiFetch<import('./types').StoredIn>(`/api/admin/storage/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  }, token);
  metadataCache.delete('admin_storage');
  return data;
}

export async function adminDeleteStorage(token: string, id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/storage/${id}`, { method: 'DELETE' }, token);
  metadataCache.delete('admin_storage');
}

// ---------------------------------------------------------------------------
// Admin: Custom Categories
// ---------------------------------------------------------------------------

export async function adminFetchCategories(token: string): Promise<import('./types').CustomCategory[]> {
  const cached = getCached<import('./types').CustomCategory[]>('admin_categories');
  if (cached) return cached;

  const data = await apiFetch<import('./types').CustomCategory[]>('/api/admin/categories', { cache: 'default' }, token);
  setCached('admin_categories', data);
  return data;
}

export async function adminCreateCategory(
  token: string, 
  name: string, 
  slug?: string, 
  is_active: boolean = true,
  show_badge: boolean = true,
  searchable: boolean = true
): Promise<import('./types').CustomCategory> {
  const data = await apiFetch<import('./types').CustomCategory>('/api/admin/categories', {
    method: 'POST',
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable }),
  }, token);
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
  return data;
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
  const data = await apiFetch<import('./types').CustomCategory>(`/api/admin/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable }),
  }, token);
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
  return data;
}

export async function adminDeleteCategory(token: string, id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/categories/${id}`, { method: 'DELETE' }, token);
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
}

// ---------------------------------------------------------------------------
// Admin: TCG Management
// ---------------------------------------------------------------------------

export async function adminFetchTCGs(token: string): Promise<import('./types').TCG[]> {
  const cached = getCached<import('./types').TCG[]>('admin_tcgs');
  if (cached) return cached;

  const data = await apiFetch<import('./types').TCG[]>('/api/admin/tcgs', { cache: 'default' }, token);
  setCached('admin_tcgs', data);
  return data;
}

export async function adminCreateTCG(token: string, id: string, name: string): Promise<import('./types').TCG> {
  const data = await apiFetch<import('./types').TCG>('/api/admin/tcgs', {
    method: 'POST',
    body: JSON.stringify({ id, name }),
  }, token);
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
  return data;
}

export async function adminUpdateTCG(token: string, id: string, name: string, is_active: boolean): Promise<import('./types').TCG> {
  const data = await apiFetch<import('./types').TCG>(`/api/admin/tcgs/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, is_active }),
  }, token);
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
  return data;
}

export async function adminDeleteTCG(token: string, id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/tcgs/${id}`, { method: 'DELETE' }, token);
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
}

export async function adminSyncSets(token: string): Promise<{ count: number; last_sync: string }> {
  const data = await apiFetch<{ count: number; last_sync: string }>('/api/admin/tcgs/sync-sets', {
    method: 'POST',
  }, token);
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('settings');
  metadataCache.delete('admin_settings');
  return data;
}

// ---------------------------------------------------------------------------
// Admin: Product Storage Management
// ---------------------------------------------------------------------------

export async function adminUpdateProductStorage(token: string, productId: string, updates: import('./types').ProductStorageInput[]): Promise<import('./types').StorageLocation[]> {
  return apiFetch<import('./types').StorageLocation[]>(`/api/admin/products/${productId}/storage`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  }, token);
}

// ---------------------------------------------------------------------------
// Orders (Public)
// ---------------------------------------------------------------------------

export async function createOrder(data: import('./types').CreateOrderRequest): Promise<{ order_number: string; order_id: string; total_cop: number; status: string }> {
  return apiFetch<{ order_number: string; order_id: string; total_cop: number; status: string }>('/api/orders', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// ---------------------------------------------------------------------------
// Admin: Orders
// ---------------------------------------------------------------------------

export async function adminFetchOrders(token: string, filters: { status?: string; search?: string; page?: number; page_size?: number } = {}): Promise<import('./types').OrderListResponse> {
  return apiFetch<import('./types').OrderListResponse>('/api/admin/orders', { params: filters, cache: 'no-store' }, token);
}

export async function adminFetchOrderDetail(token: string, id: string): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}`, { cache: 'no-store' }, token);
}

export async function adminUpdateOrder(token: string, id: string, data: { status?: string; items?: { id: string; quantity: number }[] }): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }, token);
}

export async function adminCompleteOrder(token: string, id: string, decrements: { product_id: string; stored_in_id: string; quantity: number }[]): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}/complete`, {
    method: 'POST',
    body: JSON.stringify({ decrements }),
  }, token);
}

export async function adminFetchStats(token: string): Promise<AdminStats> {
  return apiFetch<AdminStats>('/api/admin/stats', { cache: 'no-store' }, token);
}

// ---------------------------------------------------------------------------
// Bounties
// ---------------------------------------------------------------------------

export async function fetchBounties(params?: { active?: boolean }): Promise<import('./types').Bounty[]> {
  return apiFetch<import('./types').Bounty[]>('/api/bounties', { params, cache: 'no-store' });
}

export async function adminCreateBounty(token: string, data: import('./types').BountyInput): Promise<import('./types').Bounty> {
  return apiFetch<import('./types').Bounty>('/api/admin/bounties', {
    method: 'POST',
    body: JSON.stringify(data),
  }, token);
}

export async function adminUpdateBounty(token: string, id: string, data: import('./types').BountyInput): Promise<import('./types').Bounty> {
  return apiFetch<import('./types').Bounty>(`/api/admin/bounties/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }, token);
}

export async function adminDeleteBounty(token: string, id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/bounties/${id}`, { method: 'DELETE' }, token);
}

export async function createBountyOffer(data: import('./types').BountyOfferInput): Promise<import('./types').BountyOffer> {
  return apiFetch<import('./types').BountyOffer>('/api/bounties/offers', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminFetchBountyOffers(token: string): Promise<import('./types').BountyOffer[]> {
  return apiFetch<import('./types').BountyOffer[]>('/api/admin/bounties/offers', { cache: 'no-store' }, token);
}

export async function adminUpdateBountyOfferStatus(token: string, id: string, status: string): Promise<import('./types').BountyOffer> {
  return apiFetch<import('./types').BountyOffer>(`/api/admin/bounties/offers/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  }, token);
}

// ---------------------------------------------------------------------------
// Client Requests
// ---------------------------------------------------------------------------

export async function createClientRequest(data: import('./types').ClientRequestInput): Promise<import('./types').ClientRequest> {
  return apiFetch<import('./types').ClientRequest>('/api/client-requests', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminFetchClientRequests(token: string): Promise<import('./types').ClientRequest[]> {
  return apiFetch<import('./types').ClientRequest[]>('/api/admin/client-requests', { cache: 'no-store' }, token);
}

export async function adminUpdateClientRequestStatus(token: string, id: string, status: string): Promise<import('./types').ClientRequest> {
  return apiFetch<import('./types').ClientRequest>(`/api/admin/client-requests/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  }, token);
}

// ---------------------------------------------------------------------------
// Notices (Blog/News)
// ---------------------------------------------------------------------------

export async function fetchNotices(params?: { limit?: number }): Promise<import('./types').Notice[]> {
  return apiFetch<import('./types').Notice[]>('/api/notices', { params, cache: 'no-store' });
}

export async function fetchNoticeBySlug(slug: string): Promise<import('./types').Notice> {
  return apiFetch<import('./types').Notice>(`/api/notices/${slug}`, { cache: 'no-store' });
}

export async function adminFetchNotices(token: string): Promise<import('./types').Notice[]> {
  return apiFetch<import('./types').Notice[]>('/api/admin/notices', { cache: 'no-store' }, token);
}

export async function adminCreateNotice(token: string, data: import('./types').NoticeInput): Promise<import('./types').Notice> {
  return apiFetch<import('./types').Notice>('/api/admin/notices', {
    method: 'POST',
    body: JSON.stringify(data),
  }, token);
}

export async function adminUpdateNotice(token: string, id: string, data: import('./types').NoticeInput): Promise<import('./types').Notice> {
  return apiFetch<import('./types').Notice>(`/api/admin/notices/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }, token);
}

export async function adminDeleteNotice(token: string, id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/notices/${id}`, { method: 'DELETE' }, token);
}

// Newsletter
export async function subscribeToNewsletter(email: string): Promise<void> {
  return apiFetch<void>('/api/newsletter/subscribe', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export async function adminFetchSubscribers(token: string): Promise<NewsletterSubscriber[]> {
  return apiFetch<NewsletterSubscriber[]>('/api/admin/subscribers', {}, token);
}

// CRM - Clients
export async function adminFetchClients(token: string): Promise<CustomerStats[]> {
  return apiFetch<CustomerStats[]>('/api/admin/clients', {}, token);
}

export async function adminFetchClientDetail(token: string, id: string): Promise<CustomerDetail> {
  return apiFetch<CustomerDetail>(`/api/admin/clients/${id}`, {}, token);
}

export async function adminAddCustomerNote(token: string, customerId: string, content: string, orderId?: string): Promise<void> {
  return apiFetch<void>(`/api/admin/clients/${customerId}/notes`, {
    method: 'POST',
    body: JSON.stringify({ content, order_id: orderId }),
  }, token);
}
