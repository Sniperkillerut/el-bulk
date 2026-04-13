'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import {
  adminFetchTCGs, adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminDeleteProduct, adminSyncSets, adminBulkUpdateSource
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
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [isBulkSyncing, setIsBulkSyncing] = useState(false);

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

  const handleBulkUpdateSource = async (source: 'tcgplayer' | 'cardkingdom') => {
    if (selectedIds.length === 0) return;
    setIsBulkSyncing(true);
    try {
      const res = await adminBulkUpdateSource(selectedIds, source);
      alert(t('pages.admin.inventory.bulk_sync_success', 'Updated {count} products and triggered price refresh.', { count: res.count }));
      setSelectedIds([]);
      refreshProducts();
    } catch {
      alert(t('pages.admin.inventory.bulk_sync_error', 'Failed to update price source.'));
    } finally {
      setIsBulkSyncing(false);
    }
  };

  const handleSelect = (id: string, selected: boolean) => {
    setSelectedIds(prev => selected ? [...prev, id] : prev.filter(i => i !== id));
  };

  const handleSelectAll = (selected: boolean) => {
    setSelectedIds(selected ? products.map(p => p.id) : []);
  };

  return (
    <div className="flex-1 flex flex-col p-1.5 lg:p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader
        customMargin="mb-1"
        title={t('pages.admin.inventory.title', 'INVENTORY MANAGEMENT')}
        subtitle={`${total.toLocaleString()} ${t('pages.admin.inventory.items', 'ITEMS')} // ~${queryTime}ms // ${t('pages.admin.inventory.subtitle', 'OPS ACTIVE')}`}
        actions={
          <div className="flex flex-wrap items-center gap-1.5">
            {/* Utility Operations Group */}
            <div className="flex items-center gap-1 bg-ink-surface/50 border border-ink-border/20 p-0.5 rounded-md">
              <Link 
                href="/admin/accounting/valuation" 
                className="btn-secondary !px-2 !py-1 !text-[10px] flex items-center gap-1 whitespace-nowrap !leading-none min-h-[30px]"
              >
                <span>📊</span> <span className="hidden md:inline">{t('pages.admin.dashboard.valuation', 'VALUATION')}</span>
              </Link>
              <Link 
                href="/admin/inventory/low-stock" 
                className="btn-secondary !px-2 !py-1 !text-[10px] flex items-center gap-1 border-hp-color/30 text-hp-color whitespace-nowrap !leading-none min-h-[30px]"
              >
                <span>⚠️</span> <span className="hidden md:inline">{t('pages.admin.dashboard.low_stock', 'LOW STOCK')}</span>
              </Link>
              <button
                onClick={handleSyncSets}
                disabled={isSyncing}
                className="btn-secondary !px-2 !py-1 !text-[10px] flex items-center gap-1 border-l border-ink-border/10 !pl-2 ml-1 whitespace-nowrap !leading-none min-h-[30px]"
              >
                <span>{isSyncing ? '⌛' : '🔄'}</span> 
                <span className="hidden md:inline">{isSyncing ? t('pages.admin.dashboard.syncing', 'SYNCING...') : t('pages.admin.dashboard.sync', 'SYNC')}</span>
                {settings?.last_set_sync && (
                  <span className="opacity-40 ml-1 font-mono-stack text-[9px]">
                    [{new Date(settings.last_set_sync).toLocaleDateString(undefined, { day: '2-digit', month: '2-digit' })}]
                  </span>
                )}
              </button>
              <button 
                onClick={() => setShowImportModal(true)} 
                className="btn-secondary !px-2 !py-1 !text-[10px] flex items-center gap-1 border-l border-ink-border/10 whitespace-nowrap ml-1 !leading-none min-h-[30px]"
              >
                <span>📥</span> <span className="hidden md:inline">{t('pages.admin.dashboard.import_csv', 'IMPORT CSV')}</span>
              </button>
            </div>

            {/* Search and Main Action */}
            <div className="flex items-center gap-1.5 flex-1 lg:flex-none">
              <div className="relative flex-1 sm:flex-none sm:w-56">
                <input 
                  type="text" 
                  placeholder={t('pages.admin.inventory.search_placeholder', 'Search by name, set, code...')} 
                  value={search} 
                  onChange={e => { setSearch(e.target.value); setPage(1); }} 
                  className="bg-white border-ink-border/20 w-full !py-1.5 pl-3 pr-8 text-xs min-h-[38px] h-auto shadow-sm focus:border-gold transition-all !leading-none" 
                />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 opacity-30 text-xs">🔍</span>
              </div>
              <button 
                onClick={() => { setEditingProduct(null); setShowEditModal(true); }} 
                className="btn-primary !px-4 min-h-[38px] h-auto text-xs font-bold flex items-center justify-center gap-2 whitespace-nowrap shadow-md shadow-gold/10 !leading-none"
              >
                <span className="text-base leading-none translate-y-[-1px]">+</span> <span className="hidden sm:inline leading-none">{t('pages.admin.dashboard.add_product', 'PRODUCT')}</span>
              </button>
            </div>
          </div>
        }
      />

      {/* Bulk Action Toolbar - Appears when items are selected */}
      {selectedIds.length > 0 && (
        <div className="mb-2 p-2 bg-gold/10 border border-gold/30 rounded-lg flex items-center justify-between animate-in fade-in slide-in-from-top-1 duration-200">
          <div className="flex items-center gap-3">
            <span className="text-xs font-bold text-ink-deep">
              {t('pages.admin.inventory.selected_count', '{count} items selected', { count: selectedIds.length })}
            </span>
            <div className="h-4 w-px bg-gold/30" />
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-bold uppercase text-text-muted opacity-70">
                {t('pages.admin.inventory.bulk_sync_to', 'SYNC SOURCE TO:')}
              </span>
              <button 
                onClick={() => handleBulkUpdateSource('tcgplayer')}
                disabled={isBulkSyncing}
                className="btn-secondary !py-1 !px-2 !text-[10px] bg-white border-gold/20 hover:border-gold hover:text-gold transition-all"
              >
                TCGPLAYER
              </button>
              <button 
                onClick={() => handleBulkUpdateSource('cardkingdom')}
                disabled={isBulkSyncing}
                className="btn-secondary !py-1 !px-2 !text-[10px] bg-white border-gold/20 hover:border-gold hover:text-gold transition-all"
              >
                CARDKINGDOM
              </button>
            </div>
          </div>
          <button 
            onClick={() => setSelectedIds([])}
            className="text-[10px] font-bold text-hp-color hover:underline"
          >
            {t('pages.common.actions.clear_selection', 'Clear Selection')}
          </button>
        </div>
      )}

      {/* High-Density Filter Bar Area */}
      <div className="mb-2 flex-shrink-0 card p-1.5 bg-white/40 backdrop-blur shadow-sm border-ink-border/10">
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-2">
          <div className="flex flex-col sm:flex-row sm:items-center gap-1">
            <label className="text-[10px] font-mono-stack hidden sm:block uppercase font-bold text-text-muted shrink-0 mr-1">{t('pages.admin.inventory.tcg_filter_label', 'TCG')}</label>
            <select value={tcgFilter} onChange={e => { setTcgFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full px-2 text-xs min-h-[38px] h-auto !py-1 !leading-none">
              <option value="">{t('pages.common.labels.all_tcgs', 'ALL TCGS')}</option>
              {tcgs.map(t_item => <option key={t_item.id} value={t_item.id}>{t_item.name}</option>)}
            </select>
          </div>
          <div className="flex flex-col sm:flex-row sm:items-center gap-1">
            <label className="text-[10px] font-mono-stack hidden sm:block uppercase font-bold text-text-muted shrink-0 mr-1">{t('pages.admin.inventory.category_label', 'CAT')}</label>
            <select value={categoryFilter} onChange={e => { setCategoryFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full px-2 text-xs min-h-[38px] h-auto !py-1 !leading-none">
              <option value="">{t('pages.common.labels.all_categories', 'ALL CATEGORIES')}</option>
              <option value="singles">{t('pages.common.categories.singles', 'SINGLES')}</option>
              <option value="sealed">{t('pages.common.categories.sealed', 'SEALED')}</option>
              <option value="accessories">{t('pages.common.categories.accessories', 'ACCESSORIES')}</option>
              <option value="store_exclusives">{t('pages.common.categories.store_exclusives', 'STORE EXCLUSIVES')}</option>
            </select>
          </div>
          <div className="col-span-1 md:col-span-1 flex flex-col sm:flex-row sm:items-center gap-1">
            <label className="text-[10px] font-mono-stack hidden sm:block uppercase font-bold text-text-muted shrink-0 mr-1">{t('pages.admin.inventory.storage_label', 'LOC')}</label>
            <select value={storageFilter} onChange={e => { setStorageFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30 w-full px-2 text-xs min-h-[38px] h-auto !py-1 !leading-none">
              <option value="">{t('pages.common.labels.all_locations', 'ALL LOCATIONS')}</option>
              {storageLocations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </div>
          <div className="flex gap-2 items-center justify-end">
            <button onClick={() => setShowStorageModal(true)} title={t('pages.admin.inventory.manage_locations_tooltip', 'Manage Locations')} className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center shrink-0">📦</button>
            <button onClick={() => setShowCategoryModal(true)} title={t('pages.admin.inventory.manage_collections_tooltip', 'Manage Collections')} className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center shrink-0">🔖</button>
          </div>
        </div>
      </div>

      {/* Product Table Area - Now Flexible and Scrollable */}
      <div className="flex-1 min-h-0 card border-kraft-dark/20 shadow-sm bg-white overflow-hidden flex flex-col">
        <div className="flex-1 overflow-auto">
          <ProductTable
            products={products}
            selectedIds={selectedIds}
            onSelect={handleSelect}
            onSelectAll={handleSelectAll}
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
      <footer className="flex flex-col sm:flex-row justify-between items-center mt-1.5 mb-1 gap-1 px-0 flex-shrink-0">
        <div className="text-[9px] sm:text-xs font-mono-stack text-text-muted font-bold order-2 sm:order-1 opacity-80">
          {t('pages.common.pagination.showing', 'SHOWING {start} - {end} OF {total} ENTRIES', {
            start: ((page - 1) * pageSize) + 1,
            end: Math.min(page * pageSize, total),
            total: total
          })}
        </div>
        <div className="flex gap-1.5 order-1 sm:order-2 w-full sm:w-auto justify-center">
          <button disabled={page === 1} onClick={() => setPage(page - 1)} className="btn-secondary !py-1 !px-2 sm:!px-4 !text-[10px] sm:!text-xs font-bold disabled:opacity-30 flex items-center gap-1 min-h-[32px]">
            <span>←</span> <span className="hidden sm:inline">{t('pages.admin.dashboard.prev', 'PREV')}</span>
          </button>
          <div className="flex items-center px-3 font-mono-stack text-[10px] sm:text-xs font-bold bg-white rounded border border-kraft-dark/20 h-7 sm:h-8">
            {page} / {Math.max(1, Math.ceil(total / pageSize))}
          </div>
          <button disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(page + 1)} className="btn-secondary !py-1 !px-2 sm:!px-4 !text-[10px] sm:!text-xs font-bold disabled:opacity-30 flex items-center gap-1 min-h-[32px]">
            <span className="hidden sm:inline">{t('pages.admin.dashboard.next', 'NEXT')}</span> <span>→</span>
          </button>
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
