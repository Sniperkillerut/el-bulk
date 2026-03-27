'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import useSWR, { useSWRConfig } from 'swr';
import { useRouter } from 'next/navigation';
import {
  adminFetchProducts, adminDeleteProduct,
  getAdminSettings, updateAdminSettings,
  adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminFetchTCGs, adminCreateTCG, adminUpdateTCG, adminDeleteTCG
} from '@/lib/api';
import { Product, KNOWN_TCGS, TCG_SHORT, Settings, StoredIn, StorageLocation, CustomCategory } from '@/lib/types';
import OrdersPanel from '@/components/admin/OrdersPanel';
import TCGManager from '@/components/admin/TCGManager';
import CardImage from '@/components/CardImage';
import CSVImportModal from '@/components/admin/CSVImportModal';
import ProductEditModal from '@/components/admin/ProductEditModal';



export default function AdminDashboard() {
  const router = useRouter();
  const { mutate: globalMutate } = useSWRConfig();
  const [token, setToken] = useState('');
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [storageFilter, setStorageFilter] = useState('');
  const [adminSortBy, setAdminSortBy] = useState('created_at');
  const [adminSortDir, setAdminSortDir] = useState<'asc' | 'desc'>('desc');
  const [deleteConfirm, setDeleteConfirm] = useState<{ id: string; name: string } | null>(null);



  // Modal states
  const [showModal, setShowModal] = useState(false);
  const [editProduct, setEditProduct] = useState<Product | null>(null);

  // Storage Locations Global Modal
  const [showStorageModal, setShowStorageModal] = useState(false);
  // TCG Management Modal
  const [showTCGModal, setShowTCGModal] = useState(false);

  // Storage Locations Modal States
  const [newStorageName, setNewStorageName] = useState('');
  const [editingStorageId, setEditingStorageId] = useState<string | null>(null);
  const [editingStorageName, setEditingStorageName] = useState('');

  // Category Management Modal States
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newCategoryIsActive, setNewCategoryIsActive] = useState(true);
  const [newCategoryShowBadge, setNewCategoryShowBadge] = useState(true);
  const [newCategorySearchable, setNewCategorySearchable] = useState(true);
  const [editingCategoryId, setEditingCategoryId] = useState<string | null>(null);
  const [editingCategoryName, setEditingCategoryName] = useState('');
  const [editingCategoryIsActive, setEditingCategoryIsActive] = useState(true);
  const [editingCategoryShowBadge, setEditingCategoryShowBadge] = useState(true);
  const [editingCategorySearchable, setEditingCategorySearchable] = useState(true);
  


  // Settings states
  const [showSettings, setShowSettings] = useState(false);
  const [showOrders, setShowOrders] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [editingSettings, setEditingSettings] = useState<Settings>({ 
    usd_to_cop_rate: 4200, 
    eur_to_cop_rate: 4600,
    contact_address: '',
    contact_phone: '',
    contact_email: '',
    contact_instagram: '',
    contact_hours: ''
  });
  const [savingSettings, setSavingSettings] = useState(false);

  // Auth check
  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) { router.push('/admin/login'); return; }
    setToken(t);
  }, [router]);

  const productKey = useMemo(() => 
    token ? ['/api/admin/products/all'] : null,
    [token]
  );

  const { data: productRes, error: productError, isLoading: productsLoading, mutate: mutateProducts } = useSWR(
    productKey,
    () => adminFetchProducts(token, { page_size: 5000 }),
    { keepPreviousData: true, revalidateOnFocus: false }
  );

  const allProducts = useMemo(() => productRes?.products || [], [productRes]);

  const filteredAndSortedProducts = useMemo(() => {
    let result = [...allProducts];

    // 1. Filter by Storage
    if (storageFilter) {
      result = result.filter(p => p.stored_in?.some(s => s.stored_in_id === storageFilter));
    }

    // 2. Filter by Search
    if (search) {
      const s = search.toLowerCase();
      result = result.filter(p => 
        p.name.toLowerCase().includes(s) ||
        (p.set_name?.toLowerCase().includes(s)) ||
        (p.set_code?.toLowerCase().includes(s)) ||
        (p.artist?.toLowerCase().includes(s)) ||
        (p.collector_number?.toLowerCase().includes(s)) ||
        (p.oracle_text?.toLowerCase().includes(s)) ||
        (p.type_line?.toLowerCase().includes(s)) ||
        (p.promo_type?.toLowerCase().includes(s))
      );
    }

    // 3. Sort
    result.sort((a, b) => {
      let valA: any, valB: any;
      switch (adminSortBy) {
        case 'name': valA = a.name; valB = b.name; break;
        case 'tcg': valA = a.tcg; valB = b.tcg; break;
        case 'category': valA = a.category; valB = b.category; break;
        case 'set_name': valA = a.set_name || ''; valB = b.set_name || ''; break;
        case 'condition': valA = a.condition || ''; valB = b.condition || ''; break;
        case 'stock': valA = a.stock; valB = b.stock; break;
        case 'price': valA = a.price || 0; valB = b.price || 0; break;
        case 'created_at': valA = a.created_at || ''; valB = b.created_at || ''; break;
        default: valA = a.created_at || ''; valB = b.created_at || '';
      }

      if (valA < valB) return adminSortDir === 'asc' ? -1 : 1;
      if (valA > valB) return adminSortDir === 'asc' ? 1 : -1;
      return 0;
    });

    return result;
  }, [allProducts, search, storageFilter, adminSortBy, adminSortDir]);

  const total = filteredAndSortedProducts.length;
  const products = useMemo(() => {
    const start = (page - 1) * 25;
    return filteredAndSortedProducts.slice(start, start + 25);
  }, [filteredAndSortedProducts, page]);

  const loading = productsLoading && !productRes;

  const { data: settings, mutate: mutateSettings } = useSWR(
    token ? '/api/admin/settings' : null,
    () => getAdminSettings(token)
  );

  const { data: storageLocations = [] } = useSWR(
    token ? '/api/admin/storage' : null,
    () => adminFetchStorage(token)
  );

  const { data: categories = [] } = useSWR(
    token ? '/api/admin/categories' : null,
    () => adminFetchCategories(token)
  );

  const { data: tcgs = [] } = useSWR(
    token ? '/api/admin/tcgs' : null,
    () => adminFetchTCGs(token)
  );



  const openCreate = () => {
    setEditProduct(null);
    setShowModal(true);
  };

  const openEdit = (p: Product) => {
    setEditProduct(p);
    setShowModal(true);
  };

  const openSettings = () => {
    if (settings) setEditingSettings(settings);
    setShowSettings(true);
  };

  // ── Product modal save/close callbacks ──
  const handleModalClose = () => setShowModal(false);
  const handleModalSaved = () => { setShowModal(false); mutateProducts(); };
  const handleSaveAndNew = (lastForm: { tcg: string; category: string; condition: string; storageIds: string[] }) => {
    mutateProducts();
    setEditProduct(null);
    // Modal will re-open with new form since showModal stays true
  };



  const handleCreateStorage = async () => {
    if (!newStorageName.trim()) return;
    try {
      await adminCreateStorage(token, newStorageName);
      setNewStorageName('');
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleUpdateStorage = async (id: string) => {
    if (!editingStorageName.trim()) return;
    try {
      await adminUpdateStorage(token, id, editingStorageName);
      setEditingStorageId(null);
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleDeleteStorage = async (id: string, name: string, count: number = 0) => {
    let msg = `Delete location "${name}"?`;
    if (count > 0) {
      msg = `WARNING: "${name}" currently holds ${count} items!\n\nDeleting this location will instantly and permanently erase these items from your global stock.\n\nAre you sure you want to delete it?`;
    }
    if (!confirm(msg)) return;
    try {
      await adminDeleteStorage(token, id);
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleCreateCategory = async () => {
    if (!newCategoryName.trim()) return;
    try {
      await adminCreateCategory(token, newCategoryName, undefined, newCategoryIsActive, newCategoryShowBadge, newCategorySearchable);
      setNewCategoryName('');
      setNewCategoryIsActive(true);
      setNewCategoryShowBadge(true);
      setNewCategorySearchable(true);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const handleUpdateCategory = async (id: string, slug: string) => {
    if (!editingCategoryName.trim()) return;
    try {
      await adminUpdateCategory(token, id, editingCategoryName, slug, editingCategoryIsActive, editingCategoryShowBadge, editingCategorySearchable);
      setEditingCategoryId(null);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const handleDeleteCategory = async (id: string, name: string) => {
    if (!confirm(`Delete custom category "${name}"?\nThis won't delete products, only remove the grouping.`)) return;
    try {
      await adminDeleteCategory(token, id);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const handleSaveSettings = async () => {
    if (!token) return;
    setSavingSettings(true);
    try {
      const updated = await updateAdminSettings(token, editingSettings);
      mutateSettings();
      setShowSettings(false);
      mutateProducts();
    } catch (e) {
      alert('Failed to save settings: ' + (e instanceof Error ? e.message : 'Unknown error'));
    } finally {
      setSavingSettings(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    setDeleteConfirm({ id, name });
  };

  const confirmDelete = async () => {
    if (!deleteConfirm) return;
    try {
      await adminDeleteProduct(token, deleteConfirm.id);
      mutateProducts();
    } catch {
      alert('Failed to delete product.');
    } finally {
      setDeleteConfirm(null);
    }
  };

  const toggleAdminSort = (col: string) => {
    if (adminSortBy === col) {
      setAdminSortDir(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setAdminSortBy(col);
      setAdminSortDir(col === 'name' ? 'asc' : 'desc');
    }
    setPage(1);
  };

  const logout = () => {
    localStorage.removeItem('el_bulk_admin_token');
    router.push('/admin/login');
  };

  const totalPages = Math.ceil(total / 25);

  return (
    <div className="centered-container px-4 py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <p className="text-[10px] sm:text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>EL BULK / ADMIN</p>
          <h1 className="font-display text-4xl sm:text-5xl">PRODUCT MANAGEMENT</h1>
        </div>
        <div className="flex flex-wrap gap-2 sm:gap-3 w-full sm:w-auto">
          <button onClick={() => setShowOrders(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--nm-color)', color: 'var(--nm-color)' }}>📋 ÓRDENES</button>
          <button onClick={() => setShowCategoryModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--gold)', color: 'var(--gold)' }}>📋 COLLECTIONS</button>
          <button onClick={() => setShowTCGModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--kraft-dark)', color: 'var(--kraft-dark)' }}>🃏 TCG REGISTRY</button>
          <button onClick={() => setShowStorageModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">📦 STORAGE</button>
          <button id="admin-settings" onClick={openSettings} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">⚙ SETTINGS</button>
          <button id="admin-create-product" onClick={openCreate} className="btn-primary flex-1 sm:flex-none text-[10px] sm:text-[1.1rem] px-3 sm:px-6 py-2 sm:py-2.4">+ NEW PRODUCT</button>
          <button onClick={() => setShowImportModal(true)} className="btn-primary flex-1 sm:flex-none text-[10px] sm:text-[1rem] px-3 sm:px-6 py-2 sm:py-2.4" style={{ background: 'var(--nm-color)' }}>📂 IMPORT CSV</button>
          <button onClick={logout} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">LOG OUT</button>
        </div>
      </div>

      <div className="gold-line mb-6" />

      {/* Search & Filters */}
      <div className="flex flex-wrap gap-3 mb-4">
        <input
          id="admin-search"
          type="search"
          placeholder="Search products..."
          value={search}
          onChange={e => { setSearch(e.target.value); setPage(1); }}
          style={{ maxWidth: 300, flex: 1 }}
        />
        <select 
          value={storageFilter} 
          onChange={e => { setStorageFilter(e.target.value); setPage(1); }} 
          className="px-3 py-2 border border-kraft-dark bg-white" 
          style={{ fontSize: '0.9rem', flex: 1, maxWidth: 200, color: storageFilter ? 'var(--text-primary)' : 'var(--text-muted)' }}
        >
          <option value="">All Storage Locations</option>
          {storageLocations.map(l => (
            <option key={l.id} value={l.id}>{l.name}</option>
          ))}
        </select>
        <span className="flex items-center text-sm font-mono-stack ml-auto" style={{ color: 'var(--text-muted)' }}>
          {total} product{total !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Table */}
      <div className="card no-tilt overflow-x-auto">
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--ink-border)' }}>
              <th
                className="text-left px-2 py-3 text-xs font-mono-stack cursor-pointer select-none"
                style={{ color: 'var(--text-muted)', whiteSpace: 'nowrap', width: 40, transition: 'all 0.15s', borderTopWidth: '2px', borderTopStyle: 'solid', borderTopColor: 'transparent', borderBottomWidth: '2px', borderBottomStyle: 'solid', borderBottomColor: 'transparent' }}
                title="Reset sorting"
                onClick={() => { setAdminSortBy('created_at'); setAdminSortDir('desc'); setPage(1); }}
                onMouseEnter={e => {
                  if (adminSortBy !== 'created_at') {
                    e.currentTarget.style.background = 'var(--ink-surface)';
                    e.currentTarget.style.borderTopColor = 'var(--kraft-dark)';
                    e.currentTarget.style.borderBottomColor = 'var(--kraft-dark)';
                  }
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.background = 'transparent';
                  e.currentTarget.style.borderTopColor = 'transparent';
                  e.currentTarget.style.borderBottomColor = 'transparent';
                }}
              >{adminSortBy !== 'created_at' ? '↺' : ''}</th>
              {[
                { label: 'Name', key: 'name' },
                { label: 'TCG', key: 'tcg' },
                { label: 'Category', key: 'category' },
                { label: 'Set', key: 'set_name' },
                { label: 'Condition', key: 'condition' },
                { label: 'Stored In', key: '' },
                { label: 'Final Price', key: 'price' },
                { label: 'Stock', key: 'stock' },
                { label: 'Collections', key: '' },
              ].map((col, i) => {
                const isActive = col.key && adminSortBy === col.key;
                const isSortable = !!col.key;
                return (
                  <th
                    key={`${col.label}-${i}`}
                    className={`text-left px-4 py-3 text-xs font-mono-stack ${isSortable ? 'cursor-pointer select-none' : ''}`}
                    style={{
                      color: isActive ? 'var(--ink-deep)' : 'var(--text-muted)',
                      whiteSpace: 'nowrap',
                      transition: 'all 0.15s',
                      borderTopWidth: '2px', borderTopStyle: 'solid', borderTopColor: 'transparent',
                      borderBottomWidth: '2px', borderBottomStyle: 'solid', borderBottomColor: 'transparent',
                    }}
                    onClick={isSortable ? () => toggleAdminSort(col.key) : undefined}
                    onMouseEnter={isSortable ? e => {
                      e.currentTarget.style.background = 'var(--ink-surface)';
                      e.currentTarget.style.borderTopColor = 'var(--kraft-dark)';
                      e.currentTarget.style.borderBottomColor = 'var(--kraft-dark)';
                      e.currentTarget.style.color = 'var(--ink-deep)';
                    } : undefined}
                    onMouseLeave={isSortable ? e => {
                      e.currentTarget.style.background = 'transparent';
                      e.currentTarget.style.borderTopColor = 'transparent';
                      e.currentTarget.style.borderBottomColor = 'transparent';
                      e.currentTarget.style.color = isActive ? 'var(--ink-deep)' : 'var(--text-muted)';
                    } : undefined}
                  >
                    {col.label}
                    {isActive && (
                      <span className="ml-1 text-[10px]">{adminSortDir === 'asc' ? '▲' : '▼'}</span>
                    )}
                    {isSortable && !isActive && (
                      <span className="ml-1 text-[10px]" style={{ opacity: 0 }}>▲</span>
                    )}
                  </th>
                );
              })}
              <th className="text-left px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-muted)', whiteSpace: 'nowrap' }}></th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              Array.from({ length: 6 }).map((_, i) => (
                <tr key={i} style={{ borderBottom: '1px solid var(--ink-border)' }}>
                  {Array.from({ length: 11 }).map((_, j) => (
                    <td key={j} className="px-4 py-3">
                      <div className="skeleton" style={{ height: j === 0 ? 32 : 12, width: j === 0 ? 24 : j === 1 ? 140 : 60 }} />
                    </td>
                  ))}
                </tr>
              ))
            ) : products.length === 0 ? (
              <tr>
                <td colSpan={11} className="text-center py-12 text-sm" style={{ color: 'var(--text-muted)' }}>
                  No products found. Create one to get started.
                </td>
              </tr>
            ) : (
              products.map(p => (
                <tr key={p.id} style={{ borderBottom: '1px solid var(--ink-border)' }}
                  className="transition-colors"
                  onMouseEnter={e => (e.currentTarget.style.background = 'var(--ink-surface)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
                  <td key="thumb" className="px-2 py-2" style={{ width: 40, overflow: 'visible' }}>
                    <CardImage imageUrl={p.image_url} name={p.name} tcg={p.tcg} height={40} enableHover={true} enableModal={true} />
                  </td>
                  <td key="name" className="px-4 py-3 text-sm font-semibold" style={{ maxWidth: 200 }}>
                    <span className="line-clamp-1">{p.name}</span>
                  </td>
                  <td key="tcg" className="px-4 py-3">
                    <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--kraft-mid)', border: '1px solid var(--ink-border)' }}>
                      {TCG_SHORT[p.tcg] || p.tcg}
                    </span>
                  </td>
                  <td key="type" className="px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-secondary)' }}>{p.category}</td>
                  <td key="set" className="px-4 py-3 text-xs" style={{ color: 'var(--text-muted)', maxWidth: 120 }}>
                    <span className="line-clamp-1">{p.set_name || '—'}</span>
                  </td>
                  <td key="cond" className="px-4 py-3">
                    {p.condition ? <span className={`badge badge-${p.condition.toLowerCase()}`}>{p.condition}</span> : <span style={{ color: 'var(--text-muted)' }}>—</span>}
                  </td>
                  <td key="storage" className="px-4 py-3 text-xs font-mono-stack" style={{ maxWidth: 150 }}>
                    {p.stored_in && p.stored_in.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {p.stored_in.map((s, idx) => (
                          <span key={s.stored_in_id || `loc-${idx}`} className="badge" style={{ background: 'var(--kraft-light)', color: 'var(--text-secondary)', border: '1px solid var(--kraft-dark)', padding: '0.1rem 0.3rem', fontSize: '0.65rem' }}>
                            {s.name}: {s.quantity}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <span className="text-text-muted italic">—</span>
                    )}
                  </td>
                  <td key="price" className="px-4 py-3 price text-sm" title={`Computed from: ${p.price_source}`}>
                    ${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
                  </td>
                  <td key="stock" className="px-4 py-3 text-sm font-mono-stack"
                    style={{ color: p.stock === 0 ? 'var(--hp-color)' : p.stock < 3 ? 'var(--mp-color)' : 'var(--text-primary)' }}>
                    {p.stock}
                  </td>
                  <td key="collections" className="px-4 py-3 text-center">
                    {p.categories && p.categories.length > 0 ? (
                      <div className="flex flex-wrap justify-center gap-1 max-w-[120px]">
                        {p.categories.map(c => (
                          <span key={c.id} className="text-[9px] px-1.5 py-0.5 rounded" style={{ background: 'var(--gold)', color: 'var(--ink-deep)' }} title={c.name}>{c.name}</span>
                        ))}
                      </div>
                    ) : (
                      <span style={{ color: 'var(--ink-border)' }}>—</span>
                    )}
                  </td>
                  <td key="actions" className="px-4 py-3">
                    <div className="flex gap-2 items-center">
                      <button
                        id={`edit-product-${p.id}`}
                        onClick={() => openEdit(p)}
                        className="btn-secondary"
                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.75rem' }}
                      >Edit</button>
                      <button
                        id={`delete-product-${p.id}`}
                        onClick={() => handleDelete(p.id, p.name)}
                        title="Delete product"
                        style={{ width: 28, height: 28, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--hp-color)', borderRadius: 4, cursor: 'pointer', fontSize: '0.85rem', transition: 'all 0.15s' }}
                        onMouseEnter={e => { e.currentTarget.style.background = 'rgba(248,113,113,0.25)'; }}
                        onMouseLeave={e => { e.currentTarget.style.background = 'rgba(248,113,113,0.1)'; }}
                      >🗑</button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-6">
          <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary" style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === 1 ? 0.4 : 1 }}>← Prev</button>
          <span className="flex items-center px-3 text-sm font-mono-stack" style={{ color: 'var(--text-secondary)' }}>{page} / {totalPages}</span>
          <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page === totalPages} className="btn-secondary" style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === totalPages ? 0.4 : 1 }}>Next →</button>
        </div>
      )}

      {/* Settings Modal */}
      {showSettings && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-4xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)', position: 'relative' }}>
             {/* Decorative Corner */}
             <div className="absolute top-0 right-0 w-16 h-16 pointer-events-none opacity-20" style={{ borderTop: '8px solid var(--gold)', borderRight: '8px solid var(--gold)' }} />
            
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">GLOBAL SETTINGS</h2>
              <div className="px-3 py-1 bg-nm-color text-white text-xs font-mono-stack rounded shadow-sm">SYSTEM_CONFIG_V2</div>
            </div>

            <div className="grid md:grid-cols-2 gap-10">
              {/* Rates */}
              <div className="space-y-6">
                <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
                  <span className="text-2xl">📈</span>
                  <h4 className="text-lg font-display text-ink-deep m-0">EXCHANGE RATES</h4>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="cardbox p-4 bg-kraft-light/30">
                    <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>USD TO COP (TCG)</label>
                    <input type="number" className="font-bold text-lg" value={editingSettings.usd_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} />
                  </div>
                  <div className="cardbox p-4 bg-kraft-light/30">
                    <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>EUR TO COP (MCK)</label>
                    <input type="number" className="font-bold text-lg" value={editingSettings.eur_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} />
                  </div>
                </div>
                <p className="text-[10px] font-mono-stack text-text-muted mt-2">
                  * These rates are used to compute final COP prices from external sources.
                </p>
              </div>

              {/* Contact Info */}
              <div className="space-y-6">
                <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
                  <span className="text-2xl">📦</span>
                  <h4 className="text-lg font-display text-ink-deep m-0">STORE IDENTITY</h4>
                </div>
                
                <div className="space-y-4">
                  <div>
                    <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>PHYSICAL ADDRESS</label>
                    <input type="text" className="bg-white" value={editingSettings.contact_address} onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>WHATSAPP</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_phone} onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>INSTAGRAM</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_instagram} onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>STORE EMAIL</label>
                      <input type="email" className="bg-white" value={editingSettings.contact_email} onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>BUSINESS HOURS</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_hours} onChange={e => setEditingSettings({ ...editingSettings, contact_hours: e.target.value })} />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-4 mt-12 bg-kraft-light/20 p-4 -m-8 mt-8 border-t border-kraft-dark">
              <button onClick={handleSaveSettings} className="btn-primary flex-1 shadow-md" disabled={savingSettings}>
                {savingSettings ? 'SYNCING...' : 'SAVE ENTIRE DB CONFIG →'}
              </button>
              <button onClick={() => setShowSettings(false)} className="btn-secondary px-10">DISCARD</button>
            </div>
          </div>
        </div>
      )}

      {/* Storage Locations Modal */}
      {showStorageModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">STORAGE LOCATIONS</h2>
              <button onClick={() => setShowStorageModal(false)} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <div className="flex gap-2 mb-6">
              <input type="text" placeholder="New Location Name (e.g. Binder A)" value={newStorageName} onChange={e => setNewStorageName(e.target.value)} className="flex-1 bg-white" />
              <button onClick={handleCreateStorage} className="btn-primary px-6">ADD</button>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto pr-2">
              {storageLocations.map(loc => (
                <div key={loc.id} className="flex items-center justify-between p-3 border border-kraft-dark bg-kraft-light/10">
                  {editingStorageId === loc.id ? (
                    <div className="flex gap-2 flex-1 mr-4">
                      <input type="text" value={editingStorageName} onChange={e => setEditingStorageName(e.target.value)} className="flex-1 py-1 bg-white" />
                      <button onClick={() => handleUpdateStorage(loc.id)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                      <button onClick={() => setEditingStorageId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center gap-3">
                        <span className="font-semibold text-lg">{loc.name}</span>
                        <span className="text-xs font-mono-stack text-text-muted bg-kraft-light px-2 py-0.5 rounded border border-kraft-dark">
                          {loc.item_count || 0} items
                        </span>
                      </div>
                      <div className="flex gap-2">
                        <button onClick={() => { setEditingStorageId(loc.id); setEditingStorageName(loc.name); }} className="btn-secondary px-3 py-1 text-xs">EDIT</button>
                        <button onClick={() => handleDeleteStorage(loc.id, loc.name, loc.item_count || 0)} className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" style={{ borderRadius: 4 }}>DELETE</button>
                      </div>
                    </>
                  )}
                </div>
              ))}
              {storageLocations.length === 0 && <p className="text-center text-text-muted py-8">No storage locations configured.</p>}
            </div>
          </div>
        </div>
      )}

      {/* Category Management Modal */}
      {showCategoryModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--gold)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">CUSTOM COLLECTIONS</h2>
              <button onClick={() => setShowCategoryModal(false)} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <div className="flex flex-col gap-3 mb-6 p-4 bg-kraft-light/10 border border-kraft-dark rounded">
              <div className="flex gap-2">
                <input type="text" placeholder="New Collection Name (e.g. Staples)" value={newCategoryName} onChange={e => setNewCategoryName(e.target.value)} className="flex-1 bg-white" />
                <button onClick={handleCreateCategory} className="btn-primary px-6 border-gold text-gold">ADD</button>
              </div>
              <div className="flex flex-wrap gap-x-4 gap-y-2">
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="HOME SECTION: Shows the collection as a dedicated row on the landing page.">
                  <input type="checkbox" checked={newCategoryIsActive} onChange={e => setNewCategoryIsActive(e.target.checked)} />
                  HOME SECTION
                </label>
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="NAVBAR/SEARCH: Includes the collection in navigation links and global search filters.">
                  <input type="checkbox" checked={newCategorySearchable} onChange={e => setNewCategorySearchable(e.target.checked)} />
                  NAVBAR/SEARCH
                </label>
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="SHOW BADGE: Displays the collection name as a tag on product cards.">
                  <input type="checkbox" checked={newCategoryShowBadge} onChange={e => setNewCategoryShowBadge(e.target.checked)} />
                  SHOW BADGE
                </label>
              </div>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto pr-2">
              {categories.map(cat => (
                <div key={cat.id} className={`flex flex-col sm:flex-row items-start sm:items-center justify-between p-3 border ${cat.is_active ? 'border-kraft-dark bg-kraft-light/10' : 'border-ink-border bg-ink-surface'} gap-2`}>
                  {editingCategoryId === cat.id ? (
                    <div className="flex flex-col gap-3 flex-1 w-full">
                      <div className="flex gap-2">
                        <input type="text" value={editingCategoryName} onChange={e => setEditingCategoryName(e.target.value)} className="flex-1 py-1 bg-white" />
                        <button onClick={() => handleUpdateCategory(cat.id, cat.slug)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                        <button onClick={() => setEditingCategoryId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                      </div>
                      <div className="flex flex-wrap gap-x-4 gap-y-2">
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="HOME SECTION: Shows the collection as a dedicated row on the landing page.">
                          <input type="checkbox" checked={editingCategoryIsActive} onChange={e => setEditingCategoryIsActive(e.target.checked)} />
                          IS ACTIVE (HOME)
                        </label>
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="NAVBAR/SEARCH: Includes the collection in navigation links and global search filters.">
                          <input type="checkbox" checked={editingCategorySearchable} onChange={e => setEditingCategorySearchable(e.target.checked)} />
                          SEARCHABLE
                        </label>
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="SHOW BADGE: Displays the collection name as a tag on product cards.">
                          <input type="checkbox" checked={editingCategoryShowBadge} onChange={e => setEditingCategoryShowBadge(e.target.checked)} />
                          SHOW BADGE
                        </label>
                      </div>
                    </div>
                  ) : (
                    <>
                      <div className="flex flex-col gap-1">
                        <div className="flex items-center gap-3">
                          <span className="font-semibold text-lg">{cat.name}</span>
                          <div className="flex gap-1.5 items-center">
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.is_active ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.is_active ? 'var(--gold)' : 'transparent',
                                color: cat.is_active ? 'var(--ink-deep)' : 'var(--text-muted)',
                                borderColor: cat.is_active ? 'var(--gold)' : 'var(--ink-border)',
                                opacity: cat.is_active ? 1 : 0.4
                              }}
                            >HOME</span>
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.searchable ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.searchable ? 'var(--hp-color)' : 'transparent',
                                color: cat.searchable ? 'var(--ink-surface)' : 'var(--text-muted)',
                                borderColor: cat.searchable ? 'var(--hp-color)' : 'var(--ink-border)',
                                opacity: cat.searchable ? 1 : 0.4
                              }}
                            >SEARCH</span>
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.show_badge ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.show_badge ? 'var(--kraft-dark)' : 'transparent',
                                color: cat.show_badge ? 'var(--ink-surface)' : 'var(--text-muted)',
                                borderColor: cat.show_badge ? 'var(--kraft-dark)' : 'var(--ink-border)',
                                opacity: cat.show_badge ? 1 : 0.4
                              }}
                            >BADGE</span>
                          </div>
                          <span className="text-xs font-mono-stack text-text-muted border border-ink-border px-2 py-0.5 rounded">
                            {cat.item_count || 0} items
                          </span>
                        </div>
                        <span className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>/collection/{cat.slug}</span>
                      </div>
                      <div className="flex gap-2">
                        <button onClick={() => { 
                          setEditingCategoryId(cat.id); 
                          setEditingCategoryName(cat.name); 
                          setEditingCategoryIsActive(cat.is_active);
                          setEditingCategoryShowBadge(cat.show_badge);
                          setEditingCategorySearchable(cat.searchable);
                        }} className="btn-secondary px-3 py-1 text-xs">EDIT</button>
                        <button onClick={() => handleDeleteCategory(cat.id, cat.name)} className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" style={{ borderRadius: 4 }}>DELETE</button>
                      </div>
                    </>
                  )}
                </div>
              ))}
              {categories.length === 0 && <p className="text-center text-text-muted py-8">No collections created.</p>}
            </div>

            <div className="mt-8 pt-6 border-t border-kraft-dark/30">
              <p className="font-mono-stack text-[10px] uppercase text-hp-color font-bold mb-3 tracking-widest flex items-center gap-2">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                VISIBILITY GUIDE
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-[9px] font-mono-stack leading-relaxed opacity-70">
                <div>
                  <span className="text-gold font-bold">HOME SECTION:</span> Shows the collection as a dedicated row on the landing page.
                </div>
                <div>
                  <span className="text-hp-color font-bold">NAVBAR/SEARCH:</span> Includes the collection in navigation links and global search filters.
                </div>
                <div>
                  <span className="text-kraft-dark font-bold">SHOW BADGE:</span> Displays the collection name as a tag on product cards.
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Product Modal */}

      {/* Product Edit Modal */}
      {showModal && (
        <ProductEditModal
          editProduct={editProduct}
          token={token}
          storageLocations={storageLocations}
          categories={categories}
          tcgs={tcgs}
          settings={settings}
          storageFilter={storageFilter}
          onClose={handleModalClose}
          onSaved={handleModalSaved}
          onSaveAndNew={handleSaveAndNew}
        />
      )}

      {/* TCG Registry Modal */}
      {showTCGModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface', border: '4px solid var(--kraft-dark)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">TCG MANAGEMENT</h2>
              <button onClick={() => { setShowTCGModal(false); globalMutate('/api/admin/tcgs'); }} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <TCGManager token={token} />
          </div>
        </div>
      )}

      {/* Orders Panel */}
      {showOrders && (
        <OrdersPanel token={token} onClose={() => { setShowOrders(false); mutateProducts(); }} />
      )}

      {showImportModal && (
        <CSVImportModal 
          token={token} 
          storageLocations={storageLocations} 
          categories={categories}
          onClose={() => setShowImportModal(false)} 
          onImported={() => mutateProducts()} 
        />
      )}
      {/* Delete Confirmation Dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }} onClick={() => setDeleteConfirm(null)}>
          <div className="card no-tilt p-6 max-w-sm w-full mx-4" onClick={e => e.stopPropagation()} style={{ background: 'var(--kraft-light)' }}>
            <div className="flex items-center gap-3 mb-4">
              <span style={{ fontSize: '1.5rem' }}>🗑️</span>
              <h3 className="font-display text-xl">DELETE PRODUCT</h3>
            </div>
            <p className="text-sm mb-1" style={{ color: 'var(--text-secondary)' }}>
              Are you sure you want to delete:
            </p>
            <p className="text-sm font-semibold mb-4" style={{ color: 'var(--ink-deep)' }}>
              &quot;{deleteConfirm.name}&quot;
            </p>
            <p className="text-xs mb-6" style={{ color: 'var(--hp-color)' }}>
              This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeleteConfirm(null)}
                className="btn-secondary"
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem' }}
              >Cancel</button>
              <button
                onClick={confirmDelete}
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem', background: 'var(--hp-color)', color: 'white', border: 'none', borderRadius: 4, cursor: 'pointer', fontWeight: 600 }}
              >Delete</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
