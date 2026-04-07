'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import {
  adminFetchTCGs, adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminDeleteProduct, adminSyncSets
} from '@/lib/api';
import { Product, StoredIn, CustomCategory, TCG } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import ProductEditModal from '@/components/admin/ProductEditModal';
import CSVImportModal from '@/components/admin/CSVImportModal';
import StorageManagerModal from '@/components/admin/modals/StorageManagerModal';
import CategoryManagerModal from '@/components/admin/modals/CategoryManagerModal';
import ProductTable from '@/components/admin/dashboard/ProductTable';
import { useAdminProducts } from '@/hooks/useAdminProducts';
import { useLanguage } from '@/context/LanguageContext';

export default function AdminDashboard() {
  const { t } = useLanguage();
  const { token, settings, logout } = useAdmin();
  const [tcgs, setTCGs] = useState<TCG[]>([]);
  const [storageLocations, setStorageLocations] = useState<StoredIn[]>([]);
  const [categories, setCategories] = useState<CustomCategory[]>([]);

  // Custom Hook for Product Data Orchestration
  const {
    products, loading, total, page, pageSize,
    search, setSearch, tcgFilter, setTcgFilter,
    categoryFilter, setCategoryFilter,
    storageFilter, setStorageFilter, sortKey, sortDir,
    queryTime,
    setPage, handleSort, refresh: refreshProducts
  } = useAdminProducts();

  // Modal States
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showStorageModal, setShowStorageModal] = useState(false);
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);

  const loadStaticData = useCallback(async () => {
    try {
      const [tcgData, storageData, catData] = await Promise.all([
        adminFetchTCGs(), adminFetchStorage(), adminFetchCategories()
      ]);
      setTCGs(tcgData || []);
      setStorageLocations(storageData || []);
      setCategories(catData || []);
    } catch (err) {
      if (err instanceof Error && err.message?.includes('401')) logout();
    }
  }, [logout]);

  // Initial Load
  useEffect(() => {
    if (token) {
      // Use a small delay to avoid "cascading render" warnings when multiple state updates 
      // are triggered by the same dependency (token) across different hooks/effects.
      const timer = setTimeout(() => {
        void loadStaticData();
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [token, loadStaticData]);

  // CRUD Actions
  const handleSaveProduct = () => { setShowEditModal(false); setEditingProduct(null); refreshProducts(); };
  const handleSaveAndNew = () => { refreshProducts(); setEditingProduct(null); };

  const handleDeleteProduct = async (id: string, name: string) => {
    if (!confirm(t('pages.admin.inventory.confirm_delete', 'Are you sure you want to delete {name}?', { name }))) return;
    try {
      await adminDeleteProduct(id);
      refreshProducts();
    } catch { 
      alert(t('pages.admin.inventory.error_delete', 'Failed to delete product.')); 
    }
  };

  // Shared Handlers
  const handleCreateStorage = async (name: string) => {
    await adminCreateStorage(name);
    setStorageLocations(await adminFetchStorage() || []);
  };
  const handleUpdateStorage = async (id: string, name: string) => {
    await adminUpdateStorage(id, name);
    setStorageLocations(await adminFetchStorage() || []);
  };
  const handleDeleteStorage = async (id: string, name: string, count: number) => {
    if (count > 0 && !confirm(t('pages.admin.inventory.confirm_delete_storage_items', 'Location "{name}" contains {count} items. Deleting it will clear these assignments. Continue?', { name, count }))) return;
    else if (!confirm(t('pages.admin.inventory.confirm_delete_storage', 'Delete storage location "{name}"?', { name }))) return;
    await adminDeleteStorage(id);
    setStorageLocations(await adminFetchStorage() || []);
  };

  const handleCreateCategory = async (name: string, data?: Partial<CustomCategory>) => {
    await adminCreateCategory(name, data?.slug, data?.is_active, data?.show_badge, data?.searchable, data?.bg_color, data?.text_color, data?.icon);
    setCategories(await adminFetchCategories() || []);
  };
  const handleUpdateCategory = async (id: string, name: string, data?: Partial<CustomCategory>) => {
    await adminUpdateCategory(id, name, data?.slug, data?.is_active, data?.show_badge, data?.searchable, data?.bg_color, data?.text_color, data?.icon);
    setCategories(await adminFetchCategories() || []);
  };
  const handleDeleteCategory = async (id: string, name: string) => {
    if (!confirm(t('pages.admin.inventory.confirm_delete_category', 'Delete collection "{name}"?', { name }))) return;
    await adminDeleteCategory(id);
    setCategories(await adminFetchCategories() || []);
  };

  const handleSyncSets = async () => {
    if (!token) return;
    setIsSyncing(true);
    try {
      const res = await adminSyncSets();
      alert(t('pages.admin.inventory.sync_success', 'Successfully synced {count} sets!', { count: res.count }));
      // Reload to refresh settings (last sync date)
      window.location.reload();
    } catch {
      alert(t('pages.admin.inventory.sync_error', 'Failed to sync sets.'));
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader
        title={t('pages.admin.inventory.title', 'INVENTORY MANAGEMENT')}
        subtitle={t('pages.admin.inventory.subtitle', 'Store Dashboard // Operations Active')}
        actions={
          <>
            <Link href="/admin/accounting/valuation" className="btn-secondary px-3 py-1.5 text-[10px] flex items-center gap-2">
              <span>📊</span> VALUATION
            </Link>
            <Link href="/admin/inventory/low-stock" className="btn-secondary px-3 py-1.5 text-[10px] flex items-center gap-2 border-hp-color/30 text-hp-color">
              <span>⚠️</span> LOW STOCK
            </Link>
            <div className="flex flex-col items-center">
              <button
                onClick={handleSyncSets}
                disabled={isSyncing}
                title={t('pages.admin.inventory.sync_sets_tooltip', 'Sync set metadata for date ordering on client filters')}
                className="btn-secondary px-4 py-1.5 text-[10px] flex items-center gap-2 mb-0.5"
              >
                <span>{isSyncing ? '⌛' : '🔄'}</span> {isSyncing ? t('pages.admin.inventory.syncing', 'SYNCING...') : t('pages.admin.inventory.sync_sets_btn', 'SYNC SETS')}
              </button>
              {settings?.last_set_sync && (
                <span className="text-[9px] font-mono-stack text-text-muted opacity-60">
                  {t('pages.admin.inventory.last_sync', 'LAST SYNC: {date}', { date: new Date(settings.last_set_sync).toLocaleString() })}
                </span>
              )}
            </div>
            <button onClick={() => setShowImportModal(true)} className="btn-secondary px-6 flex items-center gap-2">
              <span>📥</span> {t('pages.admin.inventory.import_csv_btn', 'IMPORT CSV')}
            </button>
            <button onClick={() => { setEditingProduct(null); setShowEditModal(true); }} className="btn-primary px-8 flex items-center gap-2 shadow-lg shadow-gold/20">
              <span className="text-xl">+</span> {t('pages.admin.inventory.add_product_btn', 'ADD NEW PRODUCT')}
            </button>
          </>
        }
      />

      {/* Filters and Search Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 mb-4 flex-shrink-0">
        <div className="sm:col-span-2 lg:col-span-3 card p-4 bg-white/40 backdrop-blur shadow-sm border-kraft-dark/20 flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-[200px]">
            <label className="text-[10px] font-mono-stack mb-1.5 block uppercase font-bold text-text-muted">{t('pages.admin.inventory.search_label', 'Product Search')}</label>
            <div className="relative">
              <input type="text" placeholder={t('pages.admin.inventory.search_placeholder', 'Search by name, set, code...')} value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full" />
              <span className="absolute right-3 top-1/2 -translate-y-1/2 opacity-30">🔍</span>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-4 w-full sm:w-auto">
            <div className="flex-1 min-w-[120px]">
              <label className="text-[10px] font-mono-stack mb-1.5 block uppercase font-bold text-text-muted">{t('pages.admin.inventory.tcg_filter_label', 'TCG Filter')}</label>
              <select value={tcgFilter} onChange={e => { setTcgFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full">
                <option value="">{t('pages.common.labels.all_tcgs', 'ALL TCGS')}</option>
                {tcgs.map(t_item => <option key={t_item.id} value={t_item.id}>{t_item.name}</option>)}
              </select>
            </div>
            <div className="flex-1 min-w-[120px]">
              <label className="text-[10px] font-mono-stack mb-1.5 block uppercase font-bold text-text-muted">{t('pages.admin.inventory.category_label', 'Category')}</label>
              <select value={categoryFilter} onChange={e => { setCategoryFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full">
                <option value="">{t('pages.common.labels.all_categories', 'ALL CATEGORIES')}</option>
                <option value="singles">{t('pages.common.categories.singles', 'SINGLES')}</option>
                <option value="sealed">{t('pages.common.categories.sealed', 'SEALED')}</option>
                <option value="accessories">{t('pages.common.categories.accessories', 'ACCESSORIES')}</option>
                <option value="store_exclusives">{t('pages.common.categories.store_exclusives', 'STORE EXCLUSIVES')}</option>
              </select>
            </div>
            <div className="flex-1 min-w-[120px]">
              <label className="text-[10px] font-mono-stack mb-1.5 block uppercase font-bold text-text-muted">{t('pages.admin.inventory.storage_label', 'Physical Location')}</label>
              <select value={storageFilter} onChange={e => { setStorageFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full">
                <option value="">{t('pages.common.labels.all_locations', 'ALL LOCATIONS')}</option>
                {storageLocations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>
            <div className="flex gap-2 items-end">
              <button onClick={() => setShowStorageModal(true)} title={t('pages.admin.inventory.manage_locations_tooltip', 'Manage Locations')} className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center shrink-0">📦</button>
              <button onClick={() => setShowCategoryModal(true)} title={t('pages.admin.inventory.manage_collections_tooltip', 'Manage Collections')} className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center shrink-0">🔖</button>
            </div>
          </div>
        </div>

        <div className="card p-4 text-ink-deep flex flex-col justify-center border-none shadow-xl shadow-gold/20 relative overflow-hidden group">
          <div className="text-[10px] font-mono-stack uppercase font-bold opacity-60 mb-1">{t('pages.admin.inventory.count_label', 'INVENTORY COUNT')}</div>
          <div className="text-3xl sm:text-4xl font-display leading-none">{total.toLocaleString()}</div>
          <div className="mt-2 border-t border-ink-deep/10 pt-2 flex justify-between items-center">
            <span className="text-[10px] font-mono-stack opacity-60">{t('pages.admin.inventory.response_time_label', 'RESPONSE TIME')}</span>
            <span className="font-mono-stack text-[10px] font-bold">~{queryTime}ms</span>
          </div>
        </div>
      </div>


      {/* Product Table Area - Now Flexible and Scrollable */}
      <div className="flex-1 min-h-0 card border-kraft-dark/20 shadow-sm bg-white overflow-hidden flex flex-col">
        <div className="flex-1 overflow-auto">
          <ProductTable
            products={products}
            sortKey={sortKey}
            sortDir={sortDir}
            onSort={handleSort}
            onEdit={(p) => { setEditingProduct(p); setShowEditModal(true); }}
            onDelete={handleDeleteProduct}
            loading={loading}
          />
        </div>
      </div>

      {/* Pagination Footer - Fixed at Bottom */}
      <footer className="flex flex-col sm:flex-row justify-between items-center mt-4 mb-2 gap-4 px-0 flex-shrink-0">
        <div className="text-[10px] sm:text-xs font-mono-stack text-text-muted font-bold order-2 sm:order-1">
          {t('pages.common.pagination.showing', 'SHOWING {start} - {end} OF {total} ENTRIES', {
            start: ((page - 1) * pageSize) + 1,
            end: Math.min(page * pageSize, total),
            total: total
          })}
        </div>
        <div className="flex gap-2 order-1 sm:order-2 w-full sm:w-auto justify-center">
          <button disabled={page === 1} onClick={() => setPage(page - 1)} className="btn-secondary py-1 px-4 text-xs font-bold disabled:opacity-30">{t('pages.common.pagination.prev', '← PREV')}</button>
          <div className="flex items-center px-4 font-mono-stack text-xs font-bold bg-white rounded border border-kraft-dark/20 h-8">
            {page} / {Math.max(1, Math.ceil(total / pageSize))}
          </div>
          <button disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(page + 1)} className="btn-secondary py-1 px-4 text-xs font-bold disabled:opacity-30">{t('pages.common.pagination.next', 'NEXT →')}</button>
        </div>
      </footer>

      {/* Modals Layer */}
      {showEditModal && (
        <ProductEditModal
          editProduct={editingProduct}
          storageLocations={storageLocations}
          categories={categories}
          tcgs={tcgs}
          settings={settings!}
          onClose={() => { setShowEditModal(false); setEditingProduct(null); }}
          onSaved={handleSaveProduct}
          onSaveAndNew={handleSaveAndNew}
        />
      )}

      {showImportModal && (
        <CSVImportModal
          storageLocations={storageLocations}
          categories={categories}
          onClose={() => setShowImportModal(false)}
          onImported={() => { setShowImportModal(false); refreshProducts(); }}
        />
      )}

      {showStorageModal && (
        <StorageManagerModal
          storageLocations={storageLocations}
          onCreate={handleCreateStorage}
          onUpdate={handleUpdateStorage}
          onDelete={handleDeleteStorage}
          onClose={() => setShowStorageModal(false)}
        />
      )}

      {showCategoryModal && (
        <CategoryManagerModal
          categories={categories}
          onCreate={handleCreateCategory}
          onUpdate={handleUpdateCategory}
          onDelete={handleDeleteCategory}
          onClose={() => setShowCategoryModal(false)}
        />
      )}
    </div>
  );
}
