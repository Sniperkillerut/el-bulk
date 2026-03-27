'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { adminFetchTCGs, getAdminSettings, updateAdminSettings,
  adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminDeleteProduct
} from '@/lib/api';
import { Product, Settings, StoredIn, CustomCategory, TCG } from '@/lib/types';
import AdminSidebar from '@/components/admin/dashboard/AdminSidebar';
import ProductEditModal from '@/components/admin/ProductEditModal';
import CSVImportModal from '@/components/admin/CSVImportModal';
import SettingsModal from '@/components/admin/modals/SettingsModal';
import StorageManagerModal from '@/components/admin/modals/StorageManagerModal';
import CategoryManagerModal from '@/components/admin/modals/CategoryManagerModal';
import ProductTable from '@/components/admin/dashboard/ProductTable';
import { useAdminProducts } from '@/hooks/useAdminProducts';

export default function AdminDashboard() {
  const router = useRouter();
  const [token, setToken] = useState<string>('');
  const [tcgs, setTCGs] = useState<TCG[]>([]);
  const [settings, setSettings] = useState<Settings>();
  const [storageLocations, setStorageLocations] = useState<StoredIn[]>([]);
  const [categories, setCategories] = useState<CustomCategory[]>([]);
  
  // Custom Hook for Product Data Orchestration
  const { 
    products, loading, total, page, pageSize, 
    search, setSearch, tcgFilter, setTcgFilter, 
    storageFilter, setStorageFilter, sortKey, sortDir,
    setPage, handleSort, refresh: refreshProducts 
  } = useAdminProducts(token);

  // Modal States
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [showStorageModal, setShowStorageModal] = useState(false);
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [savingSettings, setSavingSettings] = useState(false);

  // Initial Load
  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    
    if (!t) {
      console.warn('[Dashboard] No token found, redirecting to login.');
      router.push('/admin/login');
      return;
    }
    
    setToken(t);
    loadStaticData(t);
  }, [router]);

  const loadStaticData = async (t: string) => {
    try {
      const [tcgData, settingsData, storageData, catData] = await Promise.all([
        adminFetchTCGs(t), getAdminSettings(t), adminFetchStorage(t), adminFetchCategories(t)
      ]);
      setTCGs(tcgData || []);
      setSettings(settingsData);
      setStorageLocations(storageData || []);
      setCategories(catData || []);
    } catch (err: any) {
      console.error('[Dashboard] Failed to load dashboard data:', err);
      // If we get a 401, the token might be truly expired
      if (err.message?.includes('401') || err.message?.includes('Unauthorized')) {
        console.warn('[Dashboard] Unauthorized static data fetch, clearing token.');
        localStorage.removeItem('el_bulk_admin_token');
        router.push('/admin/login');
      }
    }
  };

  // CRUD Actions
  const handleSaveProduct = () => { setShowEditModal(false); setEditingProduct(null); refreshProducts(); };
  
  const handleSaveAndNew = () => {
    refreshProducts();
    setEditingProduct(null);
    // Modal stays open, form resets (handled inside ProductEditModal)
  };

  const handleDeleteProduct = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete ${name}?`)) return;
    try {
      await adminDeleteProduct(token, id);
      refreshProducts();
    } catch { alert('Failed to delete product.'); }
  };

  const handleUpdateSettings = async (newSettings: Settings) => {
    setSavingSettings(true);
    try {
      await updateAdminSettings(token, newSettings);
      setSettings(newSettings);
      setShowSettingsModal(false);
    } catch { alert('Failed to update settings.'); }
    finally { setSavingSettings(false); }
  };

  // Storage Handlers
  const handleCreateStorage = async (name: string) => {
    await adminCreateStorage(token, name);
    const data = await adminFetchStorage(token);
    setStorageLocations(data || []);
  };
  const handleUpdateStorage = async (id: string, name: string) => {
    await adminUpdateStorage(token, id, name);
    const data = await adminFetchStorage(token);
    setStorageLocations(data || []);
  };
  const handleDeleteStorage = async (id: string, name: string, count: number) => {
    if (count > 0 && !confirm(`Location "${name}" contains ${count} items. Deleting it will clear these assignments. Continue?`)) return;
    else if (!confirm(`Delete storage location "${name}"?`)) return;
    await adminDeleteStorage(token, id);
    const data = await adminFetchStorage(token);
    setStorageLocations(data || []);
  };

  // Category Handlers
  const handleCreateCategory = async (name: string) => {
    await adminCreateCategory(token, name);
    const data = await adminFetchCategories(token);
    setCategories(data || []);
  };
  const handleUpdateCategory = async (id: string, name: string) => {
    await adminUpdateCategory(token, id, name);
    const data = await adminFetchCategories(token);
    setCategories(data || []);
  };
  const handleDeleteCategory = async (id: string, name: string) => {
    if (!confirm(`Delete collection "${name}"?`)) return;
    await adminDeleteCategory(token, id);
    const data = await adminFetchCategories(token);
    setCategories(data || []);
  };

  return (
    <div className="flex min-h-screen bg-kraft-paper">
      <AdminSidebar />
      <main className="flex-1 p-8">
        <div className="flex justify-between items-start mb-10">
          <div className="space-y-1">
            <h1 className="font-display text-5xl tracking-tighter text-ink-deep m-0">INVENTORY LOGISTICS</h1>
            <p className="font-mono-stack text-xs text-text-muted opacity-60">ADMIN_DASHBOARD_V2.4 // SESSION_ACTIVE</p>
          </div>
          <div className="flex gap-4">
             <button onClick={() => setShowImportModal(true)} className="btn-secondary px-6 flex items-center gap-2">
                <span>📥</span> IMPORT CSV
             </button>
             <button onClick={() => { setEditingProduct(null); setShowEditModal(true); }} className="btn-primary px-8 flex items-center gap-2">
                <span className="text-xl">+</span> ADD NEW PRODUCT
             </button>
          </div>
        </div>

        {/* Filters and Stats */}
        <div className="grid grid-cols-1 xl:grid-cols-4 gap-6 mb-8">
          <div className="xl:col-span-3 card p-6 bg-ink-surface/40 flex flex-wrap gap-6 items-end">
            <div className="flex-1 min-w-[240px]">
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Scanner Search</label>
              <div className="relative">
                <input type="text" placeholder="Search by name, set, code..." value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 opacity-30">🔍</span>
              </div>
            </div>
            <div style={{ width: '160px' }}>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">TCG Filter</label>
              <select value={tcgFilter} onChange={e => { setTcgFilter(e.target.value); setPage(1); }}>
                <option value="">ALL SYSTEMS</option>
                {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
            <div style={{ width: '180px' }}>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Physical Location</label>
              <select value={storageFilter} onChange={e => { setStorageFilter(e.target.value); setPage(1); }}>
                <option value="">ALL WAREHOUSE</option>
                {storageLocations.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>
            <div className="flex gap-2">
               <button onClick={() => setShowStorageModal(true)} title="Manage Locations" className="w-10 h-10 border border-ink-border rounded hover:bg-ink-surface transition-colors flex items-center justify-center">📦</button>
               <button onClick={() => setShowCategoryModal(true)} title="Manage Collections" className="w-10 h-10 border border-ink-border rounded hover:bg-ink-surface transition-colors flex items-center justify-center">🔖</button>
               <button onClick={() => setShowSettingsModal(true)} title="Global Settings" className="w-10 h-10 border border-ink-border rounded hover:bg-ink-surface transition-colors flex items-center justify-center">⚙️</button>
            </div>
          </div>

          <div className="card p-6 bg-gold text-ink-deep flex flex-col justify-center border-none shadow-xl">
             <div className="text-[10px] font-mono-stack uppercase opacity-60 mb-1">TOTAL_SKU_RECORDS</div>
             <div className="text-4xl font-display leading-none">{total.toLocaleString()}</div>
             <div className="mt-4 pt-4 border-t border-ink-deep/10 flex justify-between items-center">
                <span className="text-[10px] font-mono-stack opacity-60">QUERY_SPEED</span>
                <span className="font-mono-stack text-[10px] font-bold">~14ms</span>
             </div>
          </div>
        </div>

        {/* Product Table */}
        <ProductTable 
          products={products}
          sortKey={sortKey}
          sortDir={sortDir}
          onSort={handleSort}
          onEdit={(p) => { setEditingProduct(p); setShowEditModal(true); }}
          onDelete={handleDeleteProduct}
          loading={loading}
        />

        {/* Pagination */}
        <div className="mt-8 flex justify-between items-center bg-white/40 p-3 rounded border border-ink-border/30">
          <div className="text-xs font-mono-stack text-text-muted">
            SHOWING <span className="text-ink-deep font-bold">{((page-1)*pageSize)+1} - {Math.min(page*pageSize, total)}</span> OF {total} ENTRIES
          </div>
          <div className="flex gap-2">
            <button disabled={page === 1} onClick={() => setPage(page-1)} className="btn-secondary py-1 px-4 text-xs font-bold' disabled:opacity-30">← PREV</button>
            <div className="flex items-center px-4 font-mono-stack text-xs font-bold bg-ink-surface rounded border border-ink-border">
              PAGE {page} / {Math.max(1, Math.ceil(total / pageSize))}
            </div>
            <button disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(page+1)} className="btn-secondary py-1 px-4 text-xs font-bold' disabled:opacity-30">NEXT →</button>
          </div>
        </div>
      </main>

      {/* Modals */}
      {showEditModal && (
        <ProductEditModal 
          editProduct={editingProduct}
          token={token}
          storageLocations={storageLocations}
          categories={categories}
          tcgs={tcgs}
          settings={settings}
          storageFilter={storageFilter}
          onClose={() => { setShowEditModal(false); setEditingProduct(null); }}
          onSaved={handleSaveProduct}
          onSaveAndNew={handleSaveAndNew}
        />
      )}

      {showImportModal && (
        <CSVImportModal 
          token={token}
          storageLocations={storageLocations}
          categories={categories}
          onClose={() => setShowImportModal(false)}
          onImported={() => { setShowImportModal(false); refreshProducts(); }}
        />
      )}

      {showSettingsModal && settings && (
        <SettingsModal 
          settings={settings}
          onSave={handleUpdateSettings}
          onClose={() => setShowSettingsModal(false)}
          saving={savingSettings}
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
