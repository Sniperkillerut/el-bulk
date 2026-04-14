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

  // Reduce log noise for common unauthorized checks (e.g. /api/auth/me) 
  if (res.status === 401) {
    remoteLogger.info(`API Auth [401]: ${errorMessage}`, { url: String(res.url) });
  } else {
    remoteLogger.error(`API Error [${res.status}]: ${errorMessage}`, {
      url: String(res.url),
      status: Number(res.status),
      statusText: String(res.statusText || res.status),
    });
  }

  const error = new Error(errorMessage);
  (error as Error & { _remoteLogged?: boolean })._remoteLogged = true;
  throw error;
}

export interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined | null>;
}

export async function apiFetch<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
  const { params, headers: customHeaders, ...rest } = options;

  const base = API_BASE || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost');
  const url = new URL(`${base}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`);
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
  headers.set('X-Requested-With', 'XMLHttpRequest'); // CSRF protection

  const fullUrl = url.toString();
  remoteLogger.trace(`API Request: ${rest.method || 'GET'} ${fullUrl}`, { params, headers: Object.fromEntries(headers.entries()) });

  const res = await fetch(fullUrl, {
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

export async function bulkSearchDeck(list: string): Promise<import('./types').BulkSearchResponse> {
  return apiFetch<import('./types').BulkSearchResponse>('/api/products/search-deck', {
    method: 'POST',
    body: JSON.stringify({ list }),
  });
}

export async function fetchSettings(): Promise<import('./types').Settings> {
  return apiFetch<import('./types').Settings>('/api/settings', { cache: 'no-store' });
}

export async function fetchProduct(id: string): Promise<Product> {
  return apiFetch<Product>(`/api/products/${id}`, { cache: 'default' });
}

// Metadata caching (in-memory session)
const metadataCache = new Map<string, { data: unknown; timestamp: number }>();
const CACHE_TTL = 300000; // 5 minutes

function getCached<T>(key: string): T | null {
  if (isServer) return null;
  const entry = metadataCache.get(key);
  if (entry && Date.now() - entry.timestamp < CACHE_TTL) return entry.data as T;
  return null;
}

function setCached(key: string, data: unknown) {
  if (isServer) return;
  metadataCache.set(key, { data, timestamp: Date.now() });
}

export async function fetchCategories(): Promise<import('./types').CustomCategory[]> {
  const cached = getCached<import('./types').CustomCategory[]>('categories');
  if (cached) return cached;

  try {
    const data = await apiFetch<import('./types').CustomCategory[]>('/api/categories', { cache: 'default' });
    const categories = data || [];
    setCached('categories', categories);
    return categories;
  } catch {
    return [];
  }
}

export async function fetchTCGs(activeOnly: boolean = true, options: FetchOptions = {}): Promise<import('./types').TCG[]> {
  const key = `tcgs_${activeOnly}`;
  const cached = getCached<import('./types').TCG[]>(key);
  if (cached) return cached;

  try {
    const data = await apiFetch<import('./types').TCG[] | { tcgs: import('./types').TCG[] }>('/api/tcgs', { 
      params: { active_only: activeOnly }, 
      ...options,
      cache: options.cache || 'default' 
    });
    const tcgs = Array.isArray(data) ? data : (data.tcgs || []);
    setCached(key, tcgs);
    return tcgs;
  } catch {
    return [];
  }
}

export async function fetchPublicSettings(options: FetchOptions = {}): Promise<import('./types').Settings> {
  const cached = getCached<import('./types').Settings>('settings');
  if (cached) return cached;

  const data = await apiFetch<import('./types').Settings>('/api/settings', { 
    ...options,
    cache: options.cache || 'default' 
  });
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

export async function userUpdateMe(data: Partial<import('./types').UserProfile>): Promise<import('./types').UserProfile> {
  return apiFetch<import('./types').UserProfile>('/api/auth/me', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function userLogout(): Promise<void> {
  await apiFetch('/api/auth/logout', { method: 'POST' });
}

export async function userFetchOrders(): Promise<import('./types').UserOrder[]> {
  return apiFetch<import('./types').UserOrder[]>('/api/orders/me');
}

export async function userFetchOrderDetail(id: string): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/orders/me/${id}`);
}

export async function userCancelOrder(id: string): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/orders/me/${id}/cancel`, {
    method: 'POST'
  });
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

export async function adminFetchProducts(filters: ProductFilters = {}): Promise<ProductListResponse> {
  return apiFetch<ProductListResponse>('/api/admin/products', { params: filters, cache: 'no-store' });
}

export async function adminCreateProduct(data: Partial<Product>): Promise<Product> {
  return apiFetch<Product>('/api/admin/products', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminUpdateProduct(id: string, data: Partial<Product>): Promise<Product> {
  return apiFetch<Product>(`/api/admin/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function adminFetchLowStock(threshold: number = 5): Promise<Product[]> {
  return apiFetch<Product[]>('/api/admin/products/low-stock', { params: { threshold }, cache: 'no-store' });
}

export async function adminDeleteProduct(id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/products/${id}`, { method: 'DELETE' });
}

export async function adminBulkCreateProducts(req: import('./types').BulkCreateRequest): Promise<{ message: string; count: number }> {
  return apiFetch<{ message: string; count: number }>('/api/admin/products/bulk', {
    method: 'POST',
    body: JSON.stringify(req),
  });
}

export async function adminBulkUpdateSource(ids: string[], source: import('./types').PriceSource): Promise<{ count: number }> {
  return apiFetch<{ count: number }>('/api/admin/products/bulk-source', {
    method: 'PUT',
    body: JSON.stringify({ ids, source }),
  });
}

export async function adminFetchLogLevel(): Promise<{ level: string }> {
  return apiFetch<{ level: string }>('/api/admin/logs/level', { cache: 'no-store' });
}

export async function adminUpdateLogLevel(level: string): Promise<{ message: string; level: string }> {
  return apiFetch<{ message: string; level: string }>('/api/admin/logs/level', {
    method: 'PUT',
    body: JSON.stringify({ level }),
  });
}

export async function adminFetchAuditLogs(filters: { action?: string; resource_type?: string; page?: number; page_size?: number } = {}): Promise<import('./types').AuditLogListResponse> {
  return apiFetch<import('./types').AuditLogListResponse>('/api/admin/audit-logs', { params: filters, cache: 'no-store' });
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

  // Added fields for multi-stage enrichment
  foil_treatment: import('./types').FoilTreatment;
  card_treatment: import('./types').CardTreatment;
  promo_type?: string;
  scryfall_id?: string;
}

export async function adminFetchExternalPrice(name: string, set: string, setName: string, cn: string, foil: string, treatment: string, source: import('./types').PriceSource): Promise<{ price: number; currency: string }> {
  return apiFetch<{ price: number; currency: string }>('/api/admin/lookup/external/prices', {
    params: { name, set, set_name: setName, collector: cn, foil, treatment, source },
  });
}

export async function lookupMTGCard(
  name: string,
  set?: string,
  cn?: string,
  foil?: string,
  sid?: string,
): Promise<CardLookupResult> {
  return apiFetch<CardLookupResult>('/api/admin/lookup/mtg', {
    params: { name, set, cn, foil, sid },
  });
}

export async function adminBatchLookupMTG(
  identifiers: { name?: string; set?: string; cn?: string; scryfall_id?: string }[]
): Promise<CardLookupResult[]> {
  return apiFetch<CardLookupResult[]>('/api/admin/lookup/mtg/batch', {
    method: 'POST',
    body: JSON.stringify({ identifiers }),
  });
}

export async function lookupPokemonCard(
  name: string,
  set?: string,
): Promise<CardLookupResult> {
  return apiFetch<CardLookupResult>('/api/admin/lookup/pokemon', {
    params: { name, set },
  });
}

// ---------------------------------------------------------------------------
// Admin: exchange rate settings
// ---------------------------------------------------------------------------

export async function getAdminSettings(): Promise<import('./types').Settings> {
  const cached = getCached<import('./types').Settings>('admin_settings');
  if (cached) return cached;

  const data = await apiFetch<import('./types').Settings>('/api/admin/settings', { cache: 'default' });
  setCached('admin_settings', data);
  return data;
}

export async function updateAdminSettings(
  settings: Partial<import('./types').Settings>,
): Promise<import('./types').Settings> {
  return apiFetch<import('./types').Settings>('/api/admin/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  });
}

// ---------------------------------------------------------------------------
// Admin: price refresh
// ---------------------------------------------------------------------------

export async function triggerPriceRefresh(): Promise<{ updated: number; errors: number }> {
  return apiFetch<{ updated: number; errors: number }>('/api/admin/prices/refresh', {
    method: 'POST',
  });
}

// adminFetchStats removed from here to avoid duplication

// ---------------------------------------------------------------------------
// Admin: Storage Locations
// ---------------------------------------------------------------------------

export async function adminFetchStorage(): Promise<import('./types').StoredIn[]> {
  const cached = getCached<import('./types').StoredIn[]>('admin_storage');
  if (cached) return cached;

  const data = await apiFetch<import('./types').StoredIn[]>('/api/admin/storage', { cache: 'default' });
  setCached('admin_storage', data);
  return data;
}

export async function adminCreateStorage(name: string): Promise<import('./types').StoredIn> {
  const data = await apiFetch<import('./types').StoredIn>('/api/admin/storage', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
  metadataCache.delete('admin_storage');
  return data;
}

export async function adminUpdateStorage(id: string, name: string): Promise<import('./types').StoredIn> {
  const data = await apiFetch<import('./types').StoredIn>(`/api/admin/storage/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
  metadataCache.delete('admin_storage');
  return data;
}

export async function adminDeleteStorage(id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/storage/${id}`, { method: 'DELETE' });
  metadataCache.delete('admin_storage');
}

// ---------------------------------------------------------------------------
// Admin: Custom Categories
// ---------------------------------------------------------------------------

export async function adminFetchCategories(): Promise<import('./types').CustomCategory[]> {
  const cached = getCached<import('./types').CustomCategory[]>('admin_categories');
  if (cached) return cached;

  const data = await apiFetch<import('./types').CustomCategory[]>('/api/admin/categories', { cache: 'default' });
  setCached('admin_categories', data);
  return data;
}

export async function adminCreateCategory(
  name: string,
  slug?: string,
  is_active: boolean = true,
  show_badge: boolean = true,
  searchable: boolean = true,
  bg_color?: string,
  text_color?: string,
  icon?: string
): Promise<import('./types').CustomCategory> {
  const data = await apiFetch<import('./types').CustomCategory>('/api/admin/categories', {
    method: 'POST',
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable, bg_color, text_color, icon }),
  });
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
  return data;
}

export async function adminUpdateCategory(
  id: string,
  name: string,
  slug?: string,
  is_active?: boolean,
  show_badge?: boolean,
  searchable?: boolean,
  bg_color?: string,
  text_color?: string,
  icon?: string
): Promise<import('./types').CustomCategory> {
  const data = await apiFetch<import('./types').CustomCategory>(`/api/admin/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, slug, is_active, show_badge, searchable, bg_color, text_color, icon }),
  });
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
  return data;
}

export async function adminDeleteCategory(id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/categories/${id}`, { method: 'DELETE' });
  metadataCache.delete('admin_categories');
  metadataCache.delete('categories');
}

// ---------------------------------------------------------------------------
// Admin: TCG Management
// ---------------------------------------------------------------------------

export async function adminFetchTCGs(): Promise<import('./types').TCG[]> {
  const cached = getCached<import('./types').TCG[]>('admin_tcgs');
  if (cached) return cached;

  const data = await apiFetch<import('./types').TCG[]>('/api/admin/tcgs', { cache: 'default' });
  setCached('admin_tcgs', data);
  return data;
}

export async function adminCreateTCG(id: string, name: string): Promise<import('./types').TCG> {
  const data = await apiFetch<import('./types').TCG>('/api/admin/tcgs', {
    method: 'POST',
    body: JSON.stringify({ id, name }),
  });
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
  return data;
}

export async function adminUpdateTCG(id: string, name: string, is_active: boolean): Promise<import('./types').TCG> {
  const data = await apiFetch<import('./types').TCG>(`/api/admin/tcgs/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, is_active }),
  });
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
  return data;
}

export async function adminDeleteTCG(id: string): Promise<void> {
  await apiFetch<void>(`/api/admin/tcgs/${id}`, { method: 'DELETE' });
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('tcgs_true');
  metadataCache.delete('tcgs_false');
}

export async function adminSyncSets(): Promise<{ count: number; last_sync: string }> {
  const data = await apiFetch<{ count: number; last_sync: string }>('/api/admin/tcgs/sync-sets', {
    method: 'POST',
  });
  metadataCache.delete('admin_tcgs');
  metadataCache.delete('settings');
  metadataCache.delete('admin_settings');
  return data;
}

// ---------------------------------------------------------------------------
// Admin: Product Storage Management
// ---------------------------------------------------------------------------

export async function adminUpdateProductStorage(productId: string, updates: import('./types').ProductStorageInput[]): Promise<import('./types').StorageLocation[]> {
  return apiFetch<import('./types').StorageLocation[]>(`/api/admin/products/${productId}/storage`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  });
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

export async function adminFetchOrders(filters: { status?: string; search?: string; page?: number; page_size?: number } = {}): Promise<import('./types').OrderListResponse> {
  return apiFetch<import('./types').OrderListResponse>('/api/admin/orders', { params: filters, cache: 'no-store' });
}

export async function adminFetchOrderDetail(id: string): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}`, { cache: 'no-store' });
}

export async function adminUpdateOrder(id: string, data: { 
  status?: string; 
  payment_method?: string;
  shipping_cop?: number;
  tracking_number?: string; 
  tracking_url?: string;
  items?: { id: string; quantity: number }[];
  added_items?: { product_id: string; quantity: number; unit_price_cop: number }[];
  deleted_ids?: string[];
}): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function adminConfirmOrder(id: string, decrements: { product_id: string; stored_in_id: string; quantity: number }[]): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}/confirm`, {
    method: 'POST',
    body: JSON.stringify({ decrements }),
  });
}

export async function adminRestoreOrderStock(id: string, increments: { product_id: string; stored_in_id: string; quantity: number }[]): Promise<import('./types').OrderDetail> {
  return apiFetch<import('./types').OrderDetail>(`/api/admin/orders/${id}/restore`, {
    method: 'POST',
    body: JSON.stringify({ increments }),
  });
}

export async function adminFetchStats(): Promise<AdminStats> {
  return apiFetch<AdminStats>('/api/admin/stats', { cache: 'no-store' });
}

// ---------------------------------------------------------------------------
// Bounties
// ---------------------------------------------------------------------------

export async function fetchBounties(params?: { active?: boolean }): Promise<import('./types').Bounty[]> {
  try {
    const data = await apiFetch<import('./types').Bounty[]>('/api/bounties', { params, cache: 'no-store' });
    return data || [];
  } catch {
    return [];
  }
}

export async function adminCreateBounty(data: import('./types').BountyInput): Promise<import('./types').Bounty> {
  return apiFetch<import('./types').Bounty>('/api/admin/bounties', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminUpdateBounty(id: string, data: import('./types').BountyInput): Promise<import('./types').Bounty> {
  return apiFetch<import('./types').Bounty>(`/api/admin/bounties/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function adminDeleteBounty(id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/bounties/${id}`, { method: 'DELETE' });
}

export async function createBountyOffer(data: import('./types').BountyOfferInput): Promise<import('./types').BountyOffer> {
  return apiFetch<import('./types').BountyOffer>('/api/bounties/offers', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminFetchBountyOffers(): Promise<import('./types').BountyOffer[]> {
  return apiFetch<import('./types').BountyOffer[]>('/api/admin/bounties/offers', { cache: 'no-store' });
}

export async function adminUpdateBountyOfferStatus(id: string, status: string): Promise<import('./types').BountyOffer> {
  return apiFetch<import('./types').BountyOffer>(`/api/admin/bounties/offers/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  });
}

export async function userFetchBountyOffers(): Promise<import('./types').BountyOffer[]> {
  return apiFetch<import('./types').BountyOffer[]>('/api/bounties/offers/me');
}

export async function userCancelBountyOffer(id: string): Promise<void> {
  return apiFetch<void>(`/api/bounties/offers/me/${id}`, {
    method: 'DELETE'
  });
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

export async function userFetchClientRequests(): Promise<import('./types').ClientRequest[]> {
  return apiFetch<import('./types').ClientRequest[]>('/api/client-requests/me');
}

export async function userCancelClientRequest(id: string): Promise<void> {
  return apiFetch<void>(`/api/client-requests/me/${id}`, {
    method: 'DELETE'
  });
}

export async function adminFetchClientRequests(): Promise<import('./types').ClientRequest[]> {
  return apiFetch<import('./types').ClientRequest[]>('/api/admin/client-requests', { cache: 'no-store' });
}

export async function adminUpdateClientRequestStatus(id: string, status: string): Promise<import('./types').ClientRequest> {
  return apiFetch<import('./types').ClientRequest>(`/api/admin/client-requests/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  });
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

export async function adminFetchNotices(): Promise<import('./types').Notice[]> {
  return apiFetch<import('./types').Notice[]>('/api/admin/notices', { cache: 'no-store' });
}

export async function adminCreateNotice(data: import('./types').NoticeInput): Promise<import('./types').Notice> {
  return apiFetch<import('./types').Notice>('/api/admin/notices', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function adminUpdateNotice(id: string, data: import('./types').NoticeInput): Promise<import('./types').Notice> {
  return apiFetch<import('./types').Notice>(`/api/admin/notices/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function adminDeleteNotice(id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/notices/${id}`, { method: 'DELETE' });
}

// Newsletter
export async function subscribeToNewsletter(email: string): Promise<void> {
  return apiFetch<void>('/api/newsletter/subscribe', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export async function adminFetchSubscribers(): Promise<NewsletterSubscriber[]> {
  return apiFetch<NewsletterSubscriber[]>('/api/admin/subscribers', {});
}

// CRM - Clients
export async function adminFetchClients(): Promise<CustomerStats[]> {
  try {
    const data = await apiFetch<CustomerStats[]>('/api/admin/clients', {});
    return data || [];
  } catch {
    return [];
  }
}

export async function adminFetchClientDetail(id: string): Promise<CustomerDetail> {
  return apiFetch<CustomerDetail>(`/api/admin/clients/${id}`, {});
}

export async function adminFetchInventoryValuation(): Promise<import('./types').InventoryValuation> {
  return apiFetch<import('./types').InventoryValuation>('/api/admin/accounting/valuation', { cache: 'no-store' });
}

export async function adminAddCustomerNote(customerId: string, content: string, orderId?: string): Promise<void> {
  return apiFetch<void>(`/api/admin/clients/${customerId}/notes`, {
    method: 'POST',
    body: JSON.stringify({ content, order_id: orderId }),
  });
}

export async function adminFetchAccountingExportURL(filters: { start_date?: string; end_date?: string }): Promise<string> {
  const base = API_BASE || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost');
  const url = new URL(`${base}/api/admin/accounting/export`);
  if (filters.start_date) url.searchParams.set('start_date', filters.start_date);
  if (filters.end_date) url.searchParams.set('end_date', filters.end_date);
  return url.toString();
}

export async function adminDownloadAccountingCSV(filters: { start_date?: string; end_date?: string }): Promise<void> {
  const url = await adminFetchAccountingExportURL(filters);
  const response = await fetch(url, { credentials: 'include' });
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to download accounting export');
  }
  const blob = await response.blob();
  const downloadUrl = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = downloadUrl;

  let filename = `accounting_export_${new Date().toISOString().split('T')[0]}.csv`;
  if (filters.start_date && filters.end_date) {
    filename = `accounting_${filters.start_date.split('T')[0]}_to_${filters.end_date.split('T')[0]}.csv`;
  } else if (filters.start_date) {
    filename = `accounting_from_${filters.start_date.split('T')[0]}.csv`;
  } else if (filters.end_date) {
    filename = `accounting_to_${filters.end_date.split('T')[0]}.csv`;
  }

  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  link.remove();
}

// ---------------------------------------------------------------------------
// Translations
// ---------------------------------------------------------------------------

export async function fetchTranslations(locale?: string): Promise<Record<string, string> | Record<string, Record<string, string>>> {
  return apiFetch<Record<string, Record<string, string>>>('/api/translations', { params: { locale }, cache: 'no-store' });
}

export async function adminFetchTranslations(): Promise<import('./types').Translation[]> {
  return apiFetch<import('./types').Translation[]>('/api/admin/translations', { cache: 'no-store' });
}

export async function adminUpdateTranslation(data: { key: string; locale: string; value: string }): Promise<void> {
  return apiFetch<void>('/api/admin/translations', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function adminDeleteTranslation(key: string, locale: string): Promise<void> {
  return apiFetch<void>(`/api/admin/translations/${key}`, {
    params: { locale },
    method: 'DELETE'
  });
}

export async function adminDeleteLocale(locale: string): Promise<void> {
  return apiFetch<void>(`/api/admin/translations/locales/${locale}`, {
    method: 'DELETE'
  });
}
