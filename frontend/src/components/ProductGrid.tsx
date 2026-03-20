'use client';

import { useState, useEffect, useCallback } from 'react';
import { fetchProducts } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import { Product, FOIL_LABELS, TREATMENT_LABELS, TCG_LABELS, FoilTreatment, CardTreatment } from '@/lib/types';

interface FiltersState {
  search: string;
  foil: string;
  treatment: string;
  condition: string;
}

interface ProductGridProps {
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories';
  title: string;
  subtitle?: string;
}

export default function ProductGrid({ tcg, category, title, subtitle }: ProductGridProps) {
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<FiltersState>({ search: '', foil: '', treatment: '', condition: '' });

  const load = useCallback(async (p: number, f: FiltersState) => {
    setLoading(true);
    try {
      const res = await fetchProducts({
        tcg: tcg !== 'all' ? tcg : undefined,
        category,
        page: p,
        page_size: 20,
        search: f.search || undefined,
        foil: f.foil || undefined,
        treatment: f.treatment || undefined,
        condition: f.condition || undefined,
      });
      setProducts(res.products);
      setTotal(res.total);
    } catch {
      setProducts([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [tcg, category]);

  useEffect(() => { load(page, filters); }, [page, filters, load]);

  const handleFilterChange = (key: keyof FiltersState, value: string) => {
    setPage(1);
    setFilters(prev => ({ ...prev, [key]: value }));
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>
          {TCG_LABELS[tcg] || tcg.toUpperCase()} / {category.toUpperCase()}
        </p>
        <h1 className="font-display text-6xl" style={{ color: 'var(--text-primary)' }}>
          {title}
        </h1>
        {subtitle && <p style={{ color: 'var(--text-secondary)' }} className="mt-2">{subtitle}</p>}
        <div className="gold-line mt-4" />
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-6">
        <input
          type="search"
          placeholder="Search cards..."
          value={filters.search}
          onChange={e => handleFilterChange('search', e.target.value)}
          style={{ maxWidth: 240 }}
          id={`search-${tcg}-${category}`}
        />

        {category === 'singles' && (
          <>
            <select value={filters.condition} onChange={e => handleFilterChange('condition', e.target.value)} style={{ maxWidth: 140 }}>
              <option value="">All Conditions</option>
              {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
            </select>

            <select value={filters.foil} onChange={e => handleFilterChange('foil', e.target.value)} style={{ maxWidth: 160 }}>
              <option value="">All Treatments</option>
              {Object.entries(FOIL_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
            </select>

            <select value={filters.treatment} onChange={e => handleFilterChange('treatment', e.target.value)} style={{ maxWidth: 180 }}>
              <option value="">All Versions</option>
              {Object.entries(TREATMENT_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
            </select>
          </>
        )}

        {(filters.search || filters.foil || filters.treatment || filters.condition) && (
          <button
            onClick={() => { setFilters({ search: '', foil: '', treatment: '', condition: '' }); setPage(1); }}
            className="btn-secondary"
            style={{ fontSize: '0.85rem', padding: '0.5rem 1rem' }}
          >
            Clear Filters ×
          </button>
        )}
      </div>

      {/* Results count */}
      <p className="text-xs mb-4 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
        {loading ? '...' : `${total} result${total !== 1 ? 's' : ''}`}
      </p>

      {/* Grid */}
      {loading ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="card p-3 flex flex-col gap-2">
              <div className="skeleton" style={{ height: 160 }} />
              <div className="skeleton" style={{ height: 14, width: '80%' }} />
              <div className="skeleton" style={{ height: 12, width: '50%' }} />
            </div>
          ))}
        </div>
      ) : products.length === 0 ? (
        <div className="stamp-border rounded-lg p-16 text-center" style={{ color: 'var(--text-muted)' }}>
          <p className="font-display text-3xl mb-2">NO RESULTS</p>
          <p className="text-sm">Try clearing your filters or check back later.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          {products.map(p => <ProductCard key={p.id} product={p} />)}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-8">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
            className="btn-secondary"
            style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === 1 ? 0.4 : 1 }}
          >← Prev</button>
          <span className="flex items-center px-3 text-sm font-mono-stack" style={{ color: 'var(--text-secondary)' }}>
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="btn-secondary"
            style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === totalPages ? 0.4 : 1 }}
          >Next →</button>
        </div>
      )}
    </div>
  );
}
