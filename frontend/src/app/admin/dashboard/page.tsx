'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import {
  adminFetchProducts, adminCreateProduct, adminUpdateProduct, adminDeleteProduct
} from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS, KNOWN_TCGS, TCG_SHORT, FoilTreatment, CardTreatment } from '@/lib/types';

interface FormState {
  name: string;
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories';
  set_name: string;
  set_code: string;
  condition: string;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  price: number;
  stock: number;
  description: string;
  featured: boolean;
  image_url: string;
}

const EMPTY_FORM: FormState = {
  name: '', tcg: 'mtg', category: 'singles',
  set_name: '', set_code: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  price: 0, stock: 0, description: '', featured: false, image_url: '',
};

export default function AdminDashboard() {
  const router = useRouter();
  const [token, setToken] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [form, setForm] = useState(EMPTY_FORM);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');

  // Auth check
  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) { router.push('/admin/login'); return; }
    setToken(t);
  }, [router]);

  const loadProducts = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await adminFetchProducts(token, { search, page, page_size: 25 });
      setProducts(res.products);
      setTotal(res.total);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      if (msg.includes('401') || msg.includes('unauthorized')) {
        localStorage.removeItem('el_bulk_admin_token');
        router.push('/admin/login');
      }
    } finally {
      setLoading(false);
    }
  }, [token, search, page, router]);

  useEffect(() => { loadProducts(); }, [loadProducts]);

  const openCreate = () => {
    setEditProduct(null);
    setForm(EMPTY_FORM);
    setFormError('');
    setShowModal(true);
  };

  const openEdit = (p: Product) => {
    setEditProduct(p);
    setForm({
      name: p.name, tcg: p.tcg, category: p.category as typeof EMPTY_FORM['category'],
      set_name: p.set_name || '', set_code: p.set_code || '',
      condition: p.condition || 'NM',
      foil_treatment: p.foil_treatment, card_treatment: p.card_treatment,
      price: p.price, stock: p.stock, description: p.description || '',
      featured: p.featured, image_url: p.image_url || '',
    });
    setFormError('');
    setShowModal(true);
  };

  const handleSave = async () => {
    if (!form.name || !form.tcg || !form.category) {
      setFormError('Name, TCG, and Category are required.');
      return;
    }
    setSaving(true);
    setFormError('');
    try {
      const payload: Partial<Product> = {
        ...form,
        set_name: form.set_name || undefined,
        set_code: form.set_code || undefined,
        condition: (form.condition || undefined) as Product['condition'],
        image_url: form.image_url || undefined,
        description: form.description || undefined,
      };
      if (editProduct) {
        await adminUpdateProduct(token, editProduct.id, payload);
      } else {
        await adminCreateProduct(token, payload);
      }
      setShowModal(false);
      loadProducts();
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Failed to save product.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete "${name}"? This cannot be undone.`)) return;
    try {
      await adminDeleteProduct(token, id);
      loadProducts();
    } catch {
      alert('Failed to delete product.');
    }
  };

  const logout = () => {
    localStorage.removeItem('el_bulk_admin_token');
    router.push('/admin/login');
  };

  const totalPages = Math.ceil(total / 25);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 mb-8">
        <div>
          <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>EL BULK / ADMIN</p>
          <h1 className="font-display text-5xl">PRODUCT MANAGEMENT</h1>
        </div>
        <div className="flex gap-3">
          <button id="admin-create-product" onClick={openCreate} className="btn-primary">+ NEW PRODUCT</button>
          <button onClick={logout} className="btn-secondary" style={{ fontSize: '0.85rem' }}>LOG OUT</button>
        </div>
      </div>

      <div className="gold-line mb-6" />

      {/* Search */}
      <div className="flex gap-3 mb-4">
        <input
          id="admin-search"
          type="search"
          placeholder="Search products..."
          value={search}
          onChange={e => { setSearch(e.target.value); setPage(1); }}
          style={{ maxWidth: 300 }}
        />
        <span className="flex items-center text-sm font-mono-stack" style={{ color: 'var(--text-muted)' }}>
          {total} product{total !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Table */}
      <div className="card overflow-x-auto">
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--ink-border)' }}>
              {['Name', 'TCG', 'Category', 'Set', 'Condition', 'Price', 'Stock', 'Featured', ''].map(h => (
                <th key={h} className="text-left px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              Array.from({ length: 6 }).map((_, i) => (
                <tr key={i} style={{ borderBottom: '1px solid var(--ink-border)' }}>
                  {Array.from({ length: 8 }).map((_, j) => (
                    <td key={j} className="px-4 py-3">
                      <div className="skeleton" style={{ height: 12, width: j === 0 ? 140 : 60 }} />
                    </td>
                  ))}
                </tr>
              ))
            ) : products.length === 0 ? (
              <tr>
                <td colSpan={9} className="text-center py-12 text-sm" style={{ color: 'var(--text-muted)' }}>
                  No products found. Create one to get started.
                </td>
              </tr>
            ) : (
              products.map(p => (
                <tr key={p.id} style={{ borderBottom: '1px solid var(--ink-border)' }}
                  className="transition-colors"
                  onMouseEnter={e => (e.currentTarget.style.background = 'var(--ink-surface)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
                  <td className="px-4 py-3 text-sm font-semibold" style={{ maxWidth: 200 }}>
                    <span className="line-clamp-1">{p.name}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--kraft-mid)', border: '1px solid var(--ink-border)' }}>
                      {TCG_SHORT[p.tcg] || p.tcg}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-secondary)' }}>{p.category}</td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--text-muted)', maxWidth: 120 }}>
                    <span className="line-clamp-1">{p.set_name || '—'}</span>
                  </td>
                  <td className="px-4 py-3">
                    {p.condition ? <span className={`badge badge-${p.condition.toLowerCase()}`}>{p.condition}</span> : <span style={{ color: 'var(--text-muted)' }}>—</span>}
                  </td>
                  <td className="px-4 py-3 price text-sm">${p.price.toFixed(2)}</td>
                  <td className="px-4 py-3 text-sm font-mono-stack"
                    style={{ color: p.stock === 0 ? 'var(--hp-color)' : p.stock < 3 ? 'var(--mp-color)' : 'var(--text-primary)' }}>
                    {p.stock}
                  </td>
                  <td className="px-4 py-3 text-center">
                    {p.featured ? <span className="featured-star text-base">★</span> : <span style={{ color: 'var(--ink-border)' }}>—</span>}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        id={`edit-product-${p.id}`}
                        onClick={() => openEdit(p)}
                        className="btn-secondary"
                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.75rem' }}
                      >Edit</button>
                      <button
                        id={`delete-product-${p.id}`}
                        onClick={() => handleDelete(p.id, p.name)}
                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.75rem', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--hp-color)', borderRadius: 4, cursor: 'pointer' }}
                      >Delete</button>
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

      {/* Product Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-8 px-4"
          style={{ background: 'rgba(0,0,0,0.7)', backdropFilter: 'blur(3px)', overflowY: 'auto' }}>
          <div className="card p-6 w-full max-w-2xl mb-8" style={{ position: 'relative' }}>
            <div className="flex items-center justify-between mb-6">
              <h2 className="font-display text-3xl">{editProduct ? 'EDIT PRODUCT' : 'NEW PRODUCT'}</h2>
              <button onClick={() => setShowModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="sm:col-span-2">
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CARD / PRODUCT NAME *</label>
                <input id="form-name" type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>TCG *</label>
                <select id="form-tcg" value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
                  {KNOWN_TCGS.map(t => <option key={t} value={t}>{TCG_SHORT[t]}</option>)}
                  <option value="accessories">Accessories</option>
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CATEGORY *</label>
                <select id="form-category" value={form.category} onChange={e => setForm(f => ({ ...f, category: e.target.value as typeof EMPTY_FORM['category'] }))}>
                  <option value="singles">Singles</option>
                  <option value="sealed">Sealed</option>
                  <option value="accessories">Accessories</option>
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET NAME</label>
                <input id="form-set-name" type="text" value={form.set_name} onChange={e => setForm(f => ({ ...f, set_name: e.target.value }))} />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET CODE</label>
                <input id="form-set-code" type="text" value={form.set_code} onChange={e => setForm(f => ({ ...f, set_code: e.target.value }))} placeholder="e.g. MH2" />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CONDITION</label>
                <select id="form-condition" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value }))}>
                  {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>FOIL TREATMENT</label>
                <select id="form-foil" value={form.foil_treatment} onChange={e => setForm(f => ({ ...f, foil_treatment: e.target.value as typeof EMPTY_FORM['foil_treatment'] }))}>
                  {Object.entries(FOIL_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CARD TREATMENT / VERSION</label>
                <select id="form-treatment" value={form.card_treatment} onChange={e => setForm(f => ({ ...f, card_treatment: e.target.value as typeof EMPTY_FORM['card_treatment'] }))}>
                  {Object.entries(TREATMENT_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE ($) *</label>
                <input id="form-price" type="number" step="0.01" min="0" value={form.price} onChange={e => setForm(f => ({ ...f, price: parseFloat(e.target.value) || 0 }))} />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>STOCK</label>
                <input id="form-stock" type="number" min="0" value={form.stock} onChange={e => setForm(f => ({ ...f, stock: parseInt(e.target.value) || 0 }))} />
              </div>
              <div className="sm:col-span-2">
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>IMAGE URL</label>
                <input id="form-image" type="text" value={form.image_url} onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))} />
              </div>
              <div className="sm:col-span-2">
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>DESCRIPTION</label>
                <textarea id="form-description" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} rows={3} />
              </div>
              <div className="sm:col-span-2 flex items-center gap-3">
                <input id="form-featured" type="checkbox" checked={form.featured} onChange={e => setForm(f => ({ ...f, featured: e.target.checked }))}
                  style={{ width: 16, height: 16, accentColor: 'var(--gold)', cursor: 'pointer' }} />
                <label htmlFor="form-featured" className="text-sm" style={{ cursor: 'pointer', color: 'var(--text-secondary)' }}>
                  Featured product (shows on homepage)
                </label>
              </div>
            </div>

            {formError && (
              <p className="mt-3 text-sm" style={{ color: 'var(--hp-color)' }}>{formError}</p>
            )}

            <div className="flex gap-3 mt-6">
              <button id="admin-save-product" onClick={handleSave} className="btn-primary flex-1" disabled={saving}>
                {saving ? 'SAVING...' : editProduct ? 'SAVE CHANGES' : 'CREATE PRODUCT'}
              </button>
              <button onClick={() => setShowModal(false)} className="btn-secondary">CANCEL</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
