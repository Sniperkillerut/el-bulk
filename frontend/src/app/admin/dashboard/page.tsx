'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  adminFetchTCGs, adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminDeleteProduct
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

export default function AdminDashboard() {
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
  } = useAdminProducts(token || '');

  // Modal States
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showStorageModal, setShowStorageModal] = useState(false);
  const [showCategoryModal, setShowCategoryModal] = useState(false);

  const loadStaticData = useCallback(async (t: string) => {
    try {
      const [tcgData, storageData, catData] = await Promise.all([
        adminFetchTCGs(t), adminFetchStorage(t), adminFetchCategories(t)
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
        void loadStaticData(token);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [token, loadStaticData]);

  // CRUD Actions
  const handleSaveProduct = () => { setShowEditModal(false); setEditingProduct(null); refreshProducts(); };
  const handleSaveAndNew = () => { refreshProducts(); setEditingProduct(null); };

  const handleDeleteProduct = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete ${name}?`)) return;
    try {
      await adminDeleteProduct(token!, id);
      refreshProducts();
    } catch { alert('Failed to delete product.'); }
  };

  // Shared Handlers
  const handleCreateStorage = async (name: string) => {
    await adminCreateStorage(token!, name);
    setStorageLocations(await adminFetchStorage(token!) || []);
  };
  const handleUpdateStorage = async (id: string, name: string) => {
    await adminUpdateStorage(token!, id, name);
    setStorageLocations(await adminFetchStorage(token!) || []);
  };
  const handleDeleteStorage = async (id: string, name: string, count: number) => {
    if (count > 0 && !confirm(`Location "${name}" contains ${count} items. Deleting it will clear these assignments. Continue?`)) return;
    else if (!confirm(`Delete storage location "${name}"?`)) return;
    await adminDeleteStorage(token!, id);
    setStorageLocations(await adminFetchStorage(token!) || []);
  };

  const handleCreateCategory = async (name: string) => {
    await adminCreateCategory(token!, name);
    setCategories(await adminFetchCategories(token!) || []);
  };
  const handleUpdateCategory = async (id: string, name: string) => {
    await adminUpdateCategory(token!, id, name);
    setCategories(await adminFetchCategories(token!) || []);
  };
  const handleDeleteCategory = async (id: string, name: string) => {
    if (!confirm(`Delete collection "${name}"?`)) return;
    await adminDeleteCategory(token!, id);
    setCategories(await adminFetchCategories(token!) || []);
  };

  return (
    <div className="flex-1 flex flex-col p-8 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title="INVENTORY MANAGEMENT" 
        subtitle="Store Dashboard // Operations Active"
        actions={
          <>
            <button onClick={() => setShowImportModal(true)} className="btn-secondary px-6 flex items-center gap-2">
              <span>📥</span> IMPORT CSV
            </button>
            <button onClick={() => { setEditingProduct(null); setShowEditModal(true); }} className="btn-primary px-8 flex items-center gap-2 shadow-lg shadow-gold/20">
              <span className="text-xl">+</span> ADD NEW PRODUCT
            </button>
          </>
        }
      />

      {/* Filters and Search Bar */}
      <div className="grid grid-cols-1 xl:grid-cols-4 gap-6 mb-8 flex-shrink-0">
        <div className="xl:col-span-3 card p-6 bg-white/40 backdrop-blur shadow-sm border-kraft-dark/20 flex flex-wrap gap-6 items-end">
          <div className="flex-1 min-w-[240px]">
            <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Product Search</label>
            <div className="relative">
              <input type="text" placeholder="Search by name, set, code..." value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30" />
              <span className="absolute right-3 top-1/2 -translate-y-1/2 opacity-30">🔍</span>
            </div>
          </div>
          <div style={{ width: '130px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">TCG Filter</label>
            <select value={tcgFilter} onChange={e => { setTcgFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30">
              <option value="">ALL TCGS</option>
              {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
          </div>
          <div style={{ width: '150px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Category</label>
            <select value={categoryFilter} onChange={e => { setCategoryFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30">
              <option value="">ALL CATEGORIES</option>
              <option value="singles">SINGLES</option>
              <option value="sealed">SEALED</option>
              <option value="accessories">ACCESSORIES</option>
              <option value="store_exclusives">STORE EXCLUSIVES</option>
            </select>
          </div>
          <div style={{ width: '160px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase font-bold text-text-muted">Physical Location</label>
            <select value={storageFilter} onChange={e => { setStorageFilter(e.target.value); setPage(1); }} className="bg-white border-kraft-dark/30">
              <option value="">ALL LOCATIONS</option>
              {storageLocations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </div>
          <div className="flex gap-2">
            <button onClick={() => setShowStorageModal(true)} title="Manage Locations" className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center">📦</button>
            <button onClick={() => setShowCategoryModal(true)} title="Manage Collections" className="w-10 h-10 border border-kraft-dark/30 rounded bg-white hover:bg-kraft-light transition-colors flex items-center justify-center">🔖</button>
          </div>
        </div>

        <div className="card p-6 bg-gold text-ink-deep flex flex-col justify-center border-none shadow-xl shadow-gold/20 relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-white/10 -rotate-45 translate-x-12 -translate-y-12"></div>
          <div className="text-[10px] font-mono-stack uppercase font-bold opacity-60 mb-1">INVENTORY COUNT</div>
          <div className="text-4xl font-display leading-none">{total.toLocaleString()}</div>
          <div className="mt-4 pt-4 border-t border-ink-deep/10 flex justify-between items-center">
            <span className="text-[10px] font-mono-stack opacity-60">RESPONSE TIME</span>
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
      <footer className="mt-6 flex justify-between items-center bg-white/40 p-4 rounded border border-kraft-dark/20 backdrop-blur-sm flex-shrink-0">
        <div className="text-xs font-mono-stack text-text-muted font-bold">
          SHOWING <span className="text-ink-deep">{((page - 1) * pageSize) + 1} - {Math.min(page * pageSize, total)}</span> OF {total} ENTRIES
        </div>
        <div className="flex gap-2">
          <button disabled={page === 1} onClick={() => setPage(page - 1)} className="btn-secondary py-1 px-4 text-xs font-bold disabled:opacity-30">← PREV</button>
          <div className="flex items-center px-4 font-mono-stack text-xs font-bold bg-white rounded border border-kraft-dark/20">
            PAGE {page} / {Math.max(1, Math.ceil(total / pageSize))}
          </div>
          <button disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(page + 1)} className="btn-secondary py-1 px-4 text-xs font-bold disabled:opacity-30">NEXT →</button>
        </div>
      </footer>

      {/* Modals Layer */}
      {showEditModal && (
        <ProductEditModal
          editProduct={editingProduct}
          token={token!}
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
          token={token!}
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
