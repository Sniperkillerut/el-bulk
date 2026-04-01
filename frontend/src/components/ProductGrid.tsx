'use client';

import { useState, useEffect, useMemo } from 'react';
import useSWR from 'swr';
import { fetchProducts, fetchCategories } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import { FOIL_LABELS, TREATMENT_LABELS, TCG_LABELS, CustomCategory } from '@/lib/types';

interface FiltersState {
  search: string;
  foil: string[];
  treatment: string[];
  condition: string[];
  collection: string[];
  rarity: string[];
  language: string[];
  color: string[];
  setName: string[];
  inStock: boolean;
}

interface ProductGridProps {
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories' | 'store_exclusives';
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
    color: [],
    setName: [],
    inStock: true
  });
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');
  const [logic, setLogic] = useState<'and' | 'or'>('or');
  const [isMobileFiltersOpen, setIsMobileFiltersOpen] = useState(false);

  useEffect(() => {
    if (isMobileFiltersOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => { document.body.style.overflow = ''; };
  }, [isMobileFiltersOpen]);

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
    set_name: filters.setName.join(',') || undefined,
    in_stock: filters.inStock,
    sort_by: sortBy || undefined,
    sort_dir: sortDir || undefined,
    logic: logic !== 'or' ? logic : undefined,
  }), [tcg, category, page, debouncedSearch, filters, sortBy, sortDir, logic]);

  const { data: res, isLoading: loadingResult } = useSWR(
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

  const toggleFilter = (key: keyof FiltersState, value: string | boolean) => {
    setPage(1);
    setFilters(prev => {
      if (key === 'search') return { ...prev, search: value as string };
      if (key === 'inStock') return { ...prev, inStock: value as boolean };
      const current = prev[key] as string[];
      const next = current.includes(value as string)
        ? current.filter(v => v !== value)
        : [...current, value as string];
      return { ...prev, [key]: next };
    });
  };

  const handleFilterChange = (key: keyof FiltersState, value: string) => {
    toggleFilter(key, value);
  };

  const totalPages = Math.ceil(total / 20);

  const setNameCounts = useMemo(() => {
    const c: Record<string, number> = {};
    facets?.set_name?.forEach(f => { c[f.id] = f.count; });
    return c;
  }, [facets?.set_name]);

  return (
    <div className="centered-container px-4 py-8">
      {/* Header */}
      <div className="mb-6 md:mb-8">
        <p className="text-[10px] sm:text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>
          {TCG_LABELS[tcg] || tcg.toUpperCase()} / {category.toUpperCase()}
        </p>
        <h1 className="font-display text-fluid-h1" style={{ color: 'var(--text-primary)' }}>
          {title}
        </h1>
        {subtitle && <p style={{ color: 'var(--text-secondary)' }} className="mt-2 text-sm md:text-base">{subtitle}</p>}
        <div className="gold-line mt-4" />
      </div>

      {/* Main Layout: Sidebar + Grid */}
      <div className="flex flex-col md:flex-row gap-12 relative">
        {/* Mobile Filters Toggle */}
        <div className="md:hidden">
          <button
            onClick={() => setIsMobileFiltersOpen(true)}
            className="w-full btn-secondary py-3 !flex items-center justify-center border border-kraft-dark group"
          >
            <div className="flex items-center justify-center gap-3">
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="shrink-0 transition-transform group-hover:scale-110"
              >
                <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
              </svg>
              <span className="font-display tracking-wider">FILTERS</span>
            </div>
          </button>
        </div>

        {/* Mobile Overlay */}
        {isMobileFiltersOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/60 md:hidden backdrop-blur-sm"
            onClick={() => setIsMobileFiltersOpen(false)}
          />
        )}

        {/* Sidebar / Mobile Drawer */}
        <aside className={`
          fixed md:sticky md:top-24 top-0 left-0 h-full md:h-[calc(100vh-8rem)] w-[85vw] sm:w-80 md:w-64 z-50 md:z-auto bg-ink-surface md:bg-transparent shadow-2xl md:shadow-none transform transition-transform duration-300 ease-in-out shrink-0 overflow-y-auto custom-scrollbar p-3 md:p-0 md:pr-8
          ${isMobileFiltersOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        `}>
          {/* Mobile Drawer Header */}
          <div className="flex justify-between items-center mb-6 md:hidden border-b border-kraft-dark pb-4">
            <h2 className="font-display text-2xl text-ink-deep">Filters</h2>
            <button onClick={() => setIsMobileFiltersOpen(false)} className="text-2xl text-text-muted hover:text-ink-deep flex items-center justify-center w-8 h-8 rounded-sm border border-kraft-dark bg-kraft-light">
              &times;
            </button>
          </div>

          <div className="flex flex-col gap-3 md:pr-2 pb-8">
            {/* Logic Toggle */}
            <div className="flex flex-col gap-2">
              <p className="text-[10px] font-bold text-text-muted uppercase font-mono-stack">Search Strategy</p>
              <div className="flex p-1 bg-kraft-mid/30 rounded-md border border-kraft-dark/20">
                <button
                  onClick={() => { setLogic('or'); setPage(1); }}
                  className={`flex-1 py-1.5 px-2 text-[10px] font-bold rounded transition-all font-mono-stack ${logic === 'or'
                    ? 'bg-emerald-600 text-white shadow-sm'
                    : 'text-text-muted hover:text-emerald-600'
                    }`}
                >
                  BROAD (OR)
                </button>
                <button
                  onClick={() => { setLogic('and'); setPage(1); }}
                  className={`flex-1 py-1.5 px-2 text-[10px] font-bold rounded transition-all font-mono-stack ${logic === 'and'
                    ? 'bg-orange-600 text-white shadow-sm'
                    : 'text-text-muted hover:text-orange-600'
                    }`}
                >
                  NARROW (AND)
                </button>
              </div>
              <p className="text-[8px] italic text-text-muted/80 leading-tight">
                {logic === 'or'
                  ? "Broadens results: match ANY selected filter."
                  : "Narrows results: match ALL selected filters."}
              </p>
            </div>

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
                {/* Stock - Always at top */}
                <div className="border-b border-ink-border/20 py-2 px-2">
                  <p className="font-display text-xl sm:text-2xl text-ink-deep mb-3 uppercase tracking-tight">Availability</p>
                  <label className="flex items-center justify-between cursor-pointer group">
                    <div className="flex items-center gap-2.5">
                      <input
                        type="checkbox"
                        checked={filters.inStock}
                        onChange={() => toggleFilter('inStock', !filters.inStock)}
                        className="w-4 h-4 border-2 border-kraft-dark rounded-sm checked:bg-gold appearance-none relative checked:after:content-['✓'] checked:after:absolute checked:after:inset-0 checked:after:flex checked:after:items-center checked:after:justify-center checked:after:text-[10px] checked:after:text-white transition-all"
                      />
                      <span className="text-xs font-bold text-text-primary group-hover:text-gold transition-colors">
                        In Stock Only
                      </span>
                    </div>
                    {facets?.in_stock && (
                      <span className="text-[10px] font-bold text-gold opacity-80">
                        ({facets.in_stock})
                      </span>
                    )}
                  </label>
                </div>

                <FilterSection
                  title="Condition"
                  initialOpen={false}
                  items={[
                    { id: 'NM', label: 'NM', color: '#22c55e' },
                    { id: 'LP', label: 'LP', color: '#84cc16' },
                    { id: 'MP', label: 'MP', color: '#eab308' },
                    { id: 'HP', label: 'HP', color: '#f97316' },
                    { id: 'DMG', label: 'DMG', color: '#ef4444' }
                  ]}
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

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Rarity"
                    initialOpen={false}
                    items={[
                      { id: 'Common', label: 'Common', color: '#1a1714' },
                      { id: 'Uncommon', label: 'Uncommon', color: '#707883' },
                      { id: 'Rare', label: 'Rare', color: '#b59119' },
                      { id: 'Mythic', label: 'Mythic', color: '#d14210' },
                      { id: 'Special', label: 'Special', color: '#6e2191' },
                      { id: 'Bonus', label: 'Bonus', color: '#6e2191' }
                    ]}
                    selected={filters.rarity}
                    onToggle={(val) => toggleFilter('rarity', val)}
                    counts={facets?.rarity}
                  />
                )}

                <FilterSection
                  title="Set"
                  items={facets?.set_name || []}
                  selected={filters.setName}
                  onToggle={(val) => toggleFilter('setName', val)}
                  counts={setNameCounts}
                />

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Color"
                    initialOpen={false}
                    items={[
                      { id: 'W', label: 'White', color: '#f8f6d3' },
                      { id: 'U', label: 'Blue', color: '#0e68ab' },
                      { id: 'B', label: 'Black', color: '#150b00' },
                      { id: 'R', label: 'Red', color: '#d3202a' },
                      { id: 'G', label: 'Green', color: '#00733e' },
                      { id: 'C', label: 'Colorless', color: '#90adbb' }
                    ]}
                    selected={filters.color}
                    onToggle={(val) => toggleFilter('color', val)}
                    counts={facets?.color}
                  />
                )}

                {tcg?.toLowerCase() === 'mtg' && (
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
                )}

                <FilterSection
                  title="Collections"
                  items={availableCollections.map(c => ({ id: c.slug, label: c.name }))}
                  selected={filters.collection}
                  onToggle={(val) => toggleFilter('collection', val)}
                  counts={facets?.collection}
                />

                {(filters.search || filters.foil.length > 0 || filters.treatment.length > 0 || filters.condition.length > 0 || filters.collection.length > 0 || filters.rarity.length > 0 || filters.language.length > 0 || filters.color.length > 0 || filters.setName.length > 0) && (
                  <button
                    onClick={() => { setFilters({ search: '', foil: [], treatment: [], condition: [], collection: [], rarity: [], language: [], color: [], setName: [], inStock: true }); setPage(1); }}
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
          {/* Sort & Results bar */}
          <div className="flex items-center justify-between gap-4 mb-4 flex-wrap">
            <p className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>
              {loading ? '...' : `${total} result${total !== 1 ? 's' : ''}`}
            </p>
            <div className="flex items-center gap-2">
              <label className="text-[10px] font-mono-stack font-bold uppercase" style={{ color: 'var(--text-muted)' }}>Sort</label>
              <select
                value={sortBy}
                onChange={e => { setSortBy(e.target.value); setPage(1); }}
                className="text-xs font-mono-stack px-2 py-1.5 rounded-sm border-2 border-kraft-dark cursor-pointer"
                style={{ background: 'var(--kraft-light)', color: 'var(--ink-deep)' }}
                id={`sort-${tcg}-${category}`}
              >
                <option value="created_at">Newest</option>
                <option value="name">Name</option>
                <option value="price">Price</option>
                {(tcg === 'mtg' || tcg === 'all') && <option value="cmc">Mana Cost</option>}
                {(tcg === 'mtg' || tcg === 'all') && <option value="rarity">Rarity</option>}
              </select>
              <button
                onClick={() => { setSortDir(d => d === 'asc' ? 'desc' : 'asc'); setPage(1); }}
                className="flex items-center justify-center w-8 h-8 rounded-sm border-2 border-kraft-dark text-sm font-mono-stack transition-colors hover:bg-kraft-mid"
                style={{ background: 'var(--kraft-light)', color: 'var(--ink-deep)' }}
                title={sortDir === 'asc' ? 'Ascending' : 'Descending'}
                id={`sort-dir-${tcg}-${category}`}
              >
                {sortDir === 'asc' ? '↑' : '↓'}
              </button>
            </div>
          </div>

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
  counts,
  initialOpen = true
}: {
  title: string,
  items: { id: string, label: string, color?: string }[],
  selected: string[],
  onToggle: (id: string) => void,
  counts?: Record<string, number>,
  initialOpen?: boolean
}) {
  const [isOpen, setIsOpen] = useState(initialOpen);

  if (items.length === 0) return null;

  // Hide items with 0 count, but keep currently selected ones visible so they can be uncollapsed
  const visibleItems = items.filter(item => {
    if (!counts) return true; // Show all if counts not loaded yet
    const count = counts[item.id] ?? counts[item.id.toLowerCase()] ?? counts[item.id.toUpperCase()] ?? 0;
    return selected.includes(item.id) || count > 0;
  });

  if (visibleItems.length === 0) return null;

  return (
    <div className="border-b border-ink-border/20 py-2 px-2">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex justify-between items-center group mb-1"
      >
        <span className="font-display text-xl sm:text-2xl text-ink-deep group-hover:text-gold transition-colors uppercase tracking-tight">{title}</span>
        <span className={`text-sm transition-transform duration-300 ${isOpen ? 'rotate-180 text-gold' : 'text-text-muted'}`}>
          {isOpen ? '▲' : '▼'}
        </span>
      </button>

      {isOpen && (
        <div className="mt-3 flex flex-col gap-2 max-h-[280px] overflow-y-auto pr-1">
          {visibleItems.map(item => (
            <label key={item.id} className="flex items-center justify-between cursor-pointer group py-0.5">
              <div className="flex items-center gap-2.5 min-w-0">
                <input
                  type="checkbox"
                  checked={selected.includes(item.id)}
                  onChange={() => onToggle(item.id)}
                  className="w-4 h-4 border-2 border-kraft-dark rounded-sm checked:bg-gold appearance-none relative checked:after:content-['✓'] checked:after:absolute checked:after:inset-0 checked:after:flex checked:after:items-center checked:after:justify-center checked:after:text-[10px] checked:after:text-white transition-all shrink-0"
                />
                <span className="text-[11px] font-bold text-text-secondary group-hover:text-ink-deep transition-colors truncate flex items-center gap-2">
                  {item.color && (
                    <span
                      className="w-2.5 h-2.5 rounded-full border border-black/10 shadow-sm shrink-0"
                      style={{ backgroundColor: item.color }}
                    />
                  )}
                  {item.label}
                </span>
              </div>
              {counts && (
                <span className="text-[10px] font-bold text-gold opacity-80 shrink-0 ml-2">
                  ({counts[item.id] ?? counts[item.id.toLowerCase()] ?? counts[item.id.toUpperCase()] ?? 0})
                </span>
              )}
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
