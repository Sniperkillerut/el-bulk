'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import useSWR from 'swr';
import { fetchProducts, fetchCategories } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import { Product, FOIL_LABELS, TREATMENT_LABELS, TCG_LABELS, CustomCategory } from '@/lib/types';

interface FiltersState {
  search: string;
  foil: string[];
  treatment: string[];
  condition: string[];
  collection: string[];
  rarity: string[];
  language: string[];
  color: string[];
}

interface ProductGridProps {
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories';
  title: string;
  subtitle?: string;
}

export default function ProductGrid({ tcg, category, title, subtitle }: ProductGridProps) {
  const [availableCollections, setAvailableCollections] = useState<CustomCategory[]>([]);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState<FiltersState>({ 
    search: '', 
    foil: [], 
    treatment: [], 
    condition: [], 
    collection: [], 
    rarity: [], 
    language: [], 
    color: [] 
  });
  const [debouncedSearch, setDebouncedSearch] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(filters.search);
    }, 300);
    return () => clearTimeout(timer);
  }, [filters.search]);

  const fetcherArgs = useMemo(() => ({
    tcg: tcg !== 'all' ? tcg : undefined,
    category,
    page,
    page_size: 20,
    search: debouncedSearch || undefined,
    foil: filters.foil.join(',') || undefined,
    treatment: filters.treatment.join(',') || undefined,
    condition: filters.condition.join(',') || undefined,
    collection: filters.collection.join(',') || undefined,
    rarity: filters.rarity.join(',') || undefined,
    language: filters.language.join(',') || undefined,
    color: filters.color.join(',') || undefined,
  }), [tcg, category, page, debouncedSearch, filters]);

  const { data: res, error, isLoading: loadingResult } = useSWR(
    ['/api/products', fetcherArgs],
    ([, args]) => fetchProducts(args),
    { 
      keepPreviousData: true,
      revalidateOnFocus: false,
    }
  );

  const products = res?.products || [];
  const total = res?.total || 0;
  const facets = res?.facets || null;
  const loading = loadingResult && !res; // Only show main loading on first fetch

  useEffect(() => {
    async function loadCats() {
      try {
        const cats = await fetchCategories();
        // Only show collections that are searchable
        setAvailableCollections(cats.filter(c => c.searchable));
      } catch (err) {
        console.error('Failed to load categories for grid:', err);
      }
    }
    loadCats();
  }, []);

  const toggleFilter = (key: keyof FiltersState, value: string) => {
    setPage(1);
    setFilters(prev => {
      if (key === 'search') return { ...prev, search: value };
      const current = prev[key] as string[];
      const next = current.includes(value) 
        ? current.filter(v => v !== value)
        : [...current, value];
      return { ...prev, [key]: next };
    });
  };

  const handleFilterChange = (key: keyof FiltersState, value: string) => {
     toggleFilter(key, value);
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <div className="centered-container px-4 py-8">
      {/* Header */}
      <div className="mb-6 md:mb-8">
        <p className="text-[10px] sm:text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>
          {TCG_LABELS[tcg] || tcg.toUpperCase()} / {category.toUpperCase()}
        </p>
        <h1 className="font-display text-4xl sm:text-5xl md:text-6xl" style={{ color: 'var(--text-primary)' }}>
          {title}
        </h1>
        {subtitle && <p style={{ color: 'var(--text-secondary)' }} className="mt-2 text-sm md:text-base">{subtitle}</p>}
        <div className="gold-line mt-4" />
      </div>

      {/* Main Layout: Sidebar + Grid */}
      <div className="flex flex-col md:flex-row gap-8">
        {/* Sidebar */}
        <aside className="w-full md:w-64 shrink-0">
          <div className="sticky top-24 flex flex-col gap-6">
            {/* Search (Mobile/Desktop sidebar) */}
            <div>
              <p className="text-[10px] font-bold text-text-muted uppercase mb-2 font-mono-stack">Keywords</p>
              <input
                type="search"
                placeholder="Search cards..."
                value={filters.search}
                onChange={e => handleFilterChange('search', e.target.value)}
                id={`search-${tcg}-${category}`}
              />
            </div>

            {category === 'singles' && (
              <div className="flex flex-col gap-1">
                <FilterSection 
                  title="Condition" 
                  items={['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => ({ id: c, label: c }))} 
                  selected={filters.condition}
                  onToggle={(val) => toggleFilter('condition', val)}
                  counts={facets?.condition}
                />
                
                <FilterSection 
                  title="Finish" 
                  items={Object.entries(FOIL_LABELS).map(([id, label]) => ({ id, label }))} 
                  selected={filters.foil}
                  onToggle={(val) => toggleFilter('foil', val)}
                  counts={facets?.foil}
                />

                <FilterSection 
                  title="Version" 
                  items={Object.entries(TREATMENT_LABELS).map(([id, label]) => ({ id, label }))} 
                  selected={filters.treatment}
                  onToggle={(val) => toggleFilter('treatment', val)}
                  counts={facets?.treatment}
                />

                {tcg === 'mtg' && (
                  <>
                    <FilterSection 
                      title="Rarity" 
                      items={['Common', 'Uncommon', 'Rare', 'Mythic', 'Special', 'Bonus'].map(r => ({ id: r, label: r }))} 
                      selected={filters.rarity}
                      onToggle={(val) => toggleFilter('rarity', val)}
                      counts={facets?.rarity}
                    />
                    
                    <FilterSection 
                      title="Color" 
                      items={[
                        { id: 'W', label: 'White' },
                        { id: 'U', label: 'Blue' },
                        { id: 'B', label: 'Black' },
                        { id: 'R', label: 'Red' },
                        { id: 'G', label: 'Green' },
                        { id: 'C', label: 'Colorless' }
                      ]} 
                      selected={filters.color}
                      onToggle={(val) => toggleFilter('color', val)}
                      counts={facets?.color}
                    />

                    <FilterSection 
                      title="Language" 
                      items={[
                        { id: 'en', label: 'English' },
                        { id: 'es', label: 'Spanish' },
                        { id: 'jp', label: 'Japanese' },
                        { id: 'it', label: 'Italian' },
                        { id: 'fr', label: 'French' },
                        { id: 'de', label: 'German' },
                        { id: 'pt', label: 'Portuguese' },
                        { id: 'ru', label: 'Russian' },
                        { id: 'kr', label: 'Korean' },
                        { id: 'zhs', label: 'CH Simpl.' },
                        { id: 'zht', label: 'CH Trad.' }
                      ]} 
                      selected={filters.language}
                      onToggle={(val) => toggleFilter('language', val)}
                      counts={facets?.language}
                    />
                  </>
                )}

                <FilterSection 
                  title="Collections" 
                  items={availableCollections.map(c => ({ id: c.slug, label: c.name }))} 
                  selected={filters.collection}
                  onToggle={(val) => toggleFilter('collection', val)}
                  counts={facets?.collection}
                />

                {(filters.search || filters.foil.length > 0 || filters.treatment.length > 0 || filters.condition.length > 0 || filters.collection.length > 0 || filters.rarity.length > 0 || filters.language.length > 0 || filters.color.length > 0) && (
                  <button
                    onClick={() => { setFilters({ search: '', foil: [], treatment: [], condition: [], collection: [], rarity: [], language: [], color: [] }); setPage(1); }}
                    className="btn-secondary w-full mt-4"
                    style={{ fontSize: '0.85rem', padding: '0.4rem' }}
                  >
                    Clear All Filters ×
                  </button>
                )}
              </div>
            )}
          </div>
        </aside>

        {/* Content Area */}
        <div className="flex-1">
          {/* Results count */}
          <p className="text-xs mb-4 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
            {loading ? '...' : `${total} result${total !== 1 ? 's' : ''}`}
          </p>

          {/* Grid */}
          {loading ? (
            <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="card p-3 flex flex-col gap-2">
                  <div className="skeleton" style={{ height: 160 }} />
                  <div className="skeleton" style={{ height: 14, width: '80%' }} />
                  <div className="skeleton" style={{ height: 12, width: '50%' }} />
                </div>
              ))}
            </div>
          ) : (!products || products.length === 0) ? (
            <div className="stamp-border rounded-lg p-16 text-center" style={{ color: 'var(--text-muted)' }}>
              <p className="font-display text-3xl mb-2">NO RESULTS</p>
              <p className="text-sm">Try clearing your filters or check back later.</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 animate-fade-up">
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
      </div>
    </div>
  );
}

// Sidebar Components
function FilterSection({ 
  title, 
  items, 
  selected, 
  onToggle, 
  counts 
}: { 
  title: string, 
  items: { id: string, label: string }[], 
  selected: string[], 
  onToggle: (id: string) => void, 
  counts?: Record<string, number> 
}) {
  const [isOpen, setIsOpen] = useState(true);
  
  if (items.length === 0) return null;

  // Hide items with 0 count, but keep currently selected ones visible so they can be uncollapsed
  const visibleItems = items.filter(item => {
    if (!counts) return true; // Show all if counts not loaded yet
    const count = counts[item.id] ?? counts[item.id.toLowerCase()] ?? counts[item.id.toUpperCase()] ?? 0;
    return selected.includes(item.id) || count > 0;
  });

  if (visibleItems.length === 0) return null;

  return (
    <div className="border-b border-dashed border-kraft-dark py-3">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex justify-between items-center group"
      >
        <span className="font-display text-xl sm:text-2xl text-ink-deep group-hover:text-gold transition-colors">{title}</span>
        <span className="text-lg font-mono-stack text-text-muted">{isOpen ? '×' : '+'}</span>
      </button>
      
      {isOpen && (
        <div className="mt-2 flex flex-col gap-1.5 max-h-[250px] overflow-y-auto">
          {visibleItems.map(item => (
            <label key={item.id} className="flex items-center gap-2 cursor-pointer group">
              <input 
                type="checkbox" 
                checked={selected.includes(item.id)}
                onChange={() => onToggle(item.id)}
                className="w-4 h-4 border-2 border-kraft-dark rounded-sm checked:bg-ink-deep appearance-none relative checked:after:content-['✓'] checked:after:absolute checked:after:inset-0 checked:after:flex checked:after:items-center checked:after:justify-center checked:after:text-[10px] checked:after:text-white"
              />
              <span className="text-xs font-mono-stack text-text-secondary group-hover:text-ink-deep transition-colors truncate">
                {item.label}
                {counts && (
                  <span className="ml-1 opacity-60 text-[10px]">
                    ({counts[item.id] ?? counts[item.id.toLowerCase()] ?? counts[item.id.toUpperCase()] ?? 0})
                  </span>
                )}
              </span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
