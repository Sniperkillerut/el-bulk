'use client';

import { useState, useEffect, useMemo } from 'react';
import useSWR from 'swr';
import { fetchProducts, fetchCategories } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import ProductCardSkeleton from './skeletons/ProductCardSkeleton';
import { FOIL_LABELS, TREATMENT_LABELS, resolveLabel, TCG_LABELS, CustomCategory } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

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
  isLegendary: string;
  isLand: string;
  isHistoric: string;
  fullArt: string;
  textless: string;
  isBasicLand: string;
  isCreature: string;
  isSorcery: string;
  isInstant: string;
  isArtifact: string;
  isEnchantment: string;
  isPlaneswalker: string;
  isNonBasicLand: string;
  format: string[];
  cardTypes: string[];
}

interface ProductGridProps {
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories' | 'store_exclusives';
  title?: string;
  subtitle?: string;
  titleKey?: string;
  subtitleKey?: string;
}

export default function ProductGrid({ tcg, category, title, subtitle, titleKey, subtitleKey }: ProductGridProps) {
  const { t } = useLanguage();
  const [availableCollections, setAvailableCollections] = useState<CustomCategory[]>([]);
  const [tcgName, setTcgName] = useState(tcg.toUpperCase());
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
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
    inStock: true,
    isLegendary: '',
    isLand: '',
    isHistoric: '',
    fullArt: '',
    textless: '',
    isBasicLand: '',
    isCreature: '',
    isSorcery: '',
    isInstant: '',
    isArtifact: '',
    isEnchantment: '',
    isPlaneswalker: '',
    isNonBasicLand: '',
    format: [],
    cardTypes: []
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
    page_size: pageSize,
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
    is_legendary: filters.isLegendary || undefined,
    is_land: filters.isLand || undefined,
    is_historic: filters.isHistoric || undefined,
    full_art: filters.fullArt || undefined,
    textless: filters.textless || undefined,
    is_basic_land: filters.isBasicLand || undefined,
    is_creature: filters.isCreature || undefined,
    is_sorcery: filters.isSorcery || undefined,
    is_instant: filters.isInstant || undefined,
    is_artifact: filters.isArtifact || undefined,
    is_enchantment: filters.isEnchantment || undefined,
    is_planeswalker: filters.isPlaneswalker || undefined,
    is_non_basic_land: filters.isNonBasicLand || undefined,
    format: filters.format.join(',') || undefined,
    card_types: filters.cardTypes.join(',') || undefined,
  }), [tcg, category, page, pageSize, debouncedSearch, filters, sortBy, sortDir, logic]);

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
  const isRefetching = !!(loadingResult && res); // New data is being fetched but we have old data
  const loading = loadingResult && !res; // Only show main loading on first fetch

  useEffect(() => {
    async function loadCats() {
      try {
        const cats = await fetchCategories();
        // Only show collections that are searchable
        setAvailableCollections((cats || []).filter(c => c.searchable));
      } catch (err) {
        console.error('Failed to load categories for grid:', err);
      }
    }
    async function loadTCGs() {
      if (tcg === 'all' || tcg === 'accessories') return;
      try {
        const { fetchTCGs } = await import('@/lib/api');
        const tcgs = await fetchTCGs(true);
        const active = tcgs.find(it => it.id === tcg);
        if (active) setTcgName(active.name);
      } catch (err) {
        console.error('Failed to load tcg name for grid:', err);
      }
    }
    loadCats();
    loadTCGs();
  }, [tcg]);

  const toggleFilter = (key: keyof FiltersState, value: string | boolean) => {
    setPage(1);
    setFilters(prev => {
      if (key === 'search') return { ...prev, search: value as string };
      if (key === 'inStock') return { ...prev, inStock: value as boolean };
      if (key === 'isLegendary' || key === 'isLand' || key === 'isHistoric' || key === 'fullArt' || key === 'textless' || key === 'isBasicLand' || key === 'isCreature' || key === 'isSorcery' || key === 'isInstant' || key === 'isArtifact' || key === 'isEnchantment' || key === 'isPlaneswalker' || key === 'isNonBasicLand') {
        return { ...prev, [key]: prev[key as keyof FiltersState] === value ? '' : value as string };
      }
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

  const totalPages = Math.ceil(total / pageSize);

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
          {t(`tcg.${tcg}`, TCG_LABELS[tcg] || tcg.toUpperCase())} / {t(`pages.inventory.category.${category}`, category.toUpperCase())}
        </p>
        <h1 className="font-display text-fluid-h1" style={{ color: 'var(--text-main)' }}>
          {titleKey ? t(titleKey, title || '').replace('{tcg}', tcgName) : title}
        </h1>
        {(subtitleKey || subtitle) && (
          <p style={{ color: 'var(--text-secondary)' }} className="mt-2 text-sm md:text-base">
            {subtitleKey ? t(subtitleKey, subtitle || '').replace('{tcg}', tcgName) : subtitle}
          </p>
        )}
        <div className="gold-line mt-4" />
      </div>

      {/* Main Layout: Sidebar + Grid */}
      <div className="flex flex-col sm:flex-row gap-12 relative items-start">
        {/* Mobile Filters Toggle */}
        <div className="sm:hidden">
          <button
            onClick={() => setIsMobileFiltersOpen(true)}
            className="w-full btn-secondary py-3 !flex items-center justify-center border border-border-main group"
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
              <span className="font-display tracking-wider">{t('pages.inventory.grid.filters.title', 'FILTERS')}</span>
            </div>
          </button>
        </div>

        {/* Mobile Overlay */}
        {isMobileFiltersOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/60 sm:hidden backdrop-blur-sm"
            onClick={() => setIsMobileFiltersOpen(false)}
          />
        )}

        {/* Sidebar / Mobile Drawer */}
        <aside className={`
          fixed sm:sticky sm:top-20 top-0 left-0 h-full sm:h-[calc(100vh-100px)] w-[85vw] sm:w-48 md:w-64 z-50 sm:z-auto bg-bg-surface sm:bg-bg-surface/40 sm:backdrop-blur-md sm:border-r sm:border-border-main/30 sm:shadow-[inset_-1px_0_0_0_rgba(255,255,255,0.05),-10px_0_30px_rgba(0,0,0,0.1)] transform transition-transform duration-300 ease-in-out shrink-0 overflow-y-auto custom-scrollbar p-5 sm:p-6 sm:mr-6
          ${isMobileFiltersOpen ? 'translate-x-0' : '-translate-x-full sm:translate-x-0'}
        `}>
          {/* Mobile Drawer Header */}
          <div className="flex justify-between items-center mb-6 sm:hidden border-b border-border-main pb-4">
            <h2 className="font-display text-2xl text-text-main">{t('pages.inventory.grid.filters.title', 'Filters')}</h2>
            <button onClick={() => setIsMobileFiltersOpen(false)} className="text-2xl text-text-muted hover:text-text-main flex items-center justify-center w-8 h-8 rounded-sm border border-border-main bg-bg-page">
              &times;
            </button>
          </div>

          <div className={`flex flex-col gap-3 sm:pr-2 pb-8 transition-opacity duration-300 ${isRefetching ? 'opacity-40' : 'opacity-100'}`}>
            {/* Logic Toggle */}
            <div className="flex flex-col gap-2">
              <p className="text-[10px] font-bold text-text-muted uppercase font-mono-stack">{t('pages.inventory.grid.filters.strategy.title', 'Search Strategy')}</p>
              <div className="flex p-1 bg-black/10 rounded-md border border-border-main/10 shadow-inner">
                <button
                  onClick={() => { setLogic('or'); setPage(1); }}
                  className={`flex-1 py-2 px-2 text-[10px] font-bold rounded transition-all font-mono-stack ${logic === 'or'
                    ? 'bg-emerald-600 text-white shadow-md'
                    : 'text-text-muted hover:text-emerald-500'
                    }`}
                >
                  {t('pages.inventory.grid.filters.strategy.broad', 'BROAD (OR)')}
                </button>
                <button
                  onClick={() => { setLogic('and'); setPage(1); }}
                  className={`flex-1 py-2 px-2 text-[10px] font-bold rounded transition-all font-mono-stack ${logic === 'and'
                    ? 'bg-orange-600 text-white shadow-md'
                    : 'text-text-muted hover:text-orange-500'
                    }`}
                >
                  {t('pages.inventory.grid.filters.strategy.narrow', 'NARROW (AND)')}
                </button>
              </div>
              <p className="text-[8px] italic text-text-muted/80 leading-tight">
                {logic === 'or'
                  ? t('pages.inventory.grid.filters.strategy.broad_desc', "Broadens within each filter group (e.g., Foil OR Etched). Filters across groups still narrow results.")
                  : t('pages.inventory.grid.filters.strategy.narrow_desc', "Narrows results: requires ALL selected colors, collections, and formats.")}
              </p>
            </div>

            {/* Search (Mobile/Desktop sidebar) */}
            <div>
              <p className="text-[10px] font-bold text-text-muted uppercase mb-2 font-mono-stack">{t('pages.inventory.grid.filters.keywords', 'Keywords')}</p>
              <input
                type="search"
                placeholder={t('pages.inventory.grid.filters.search_placeholder', 'Search cards...')}
                value={filters.search}
                onChange={e => handleFilterChange('search', e.target.value)}
                id={`search-${tcg}-${category}`}
              />
            </div>

            {category === 'singles' && (
              <div className="flex flex-col gap-1">
                {/* Stock - Always at top */}
                <div className="border-b border-border-main/20 py-2 px-2">
                  <p className="font-display text-xl sm:text-2xl text-text-main mb-3 uppercase tracking-tight">{t('pages.inventory.grid.filters.availability', 'Availability')}</p>
                  <label className="flex items-center justify-between cursor-pointer group">
                    <div className="flex items-center gap-2.5">
                      <input
                        type="checkbox"
                        checked={filters.inStock}
                        onChange={() => toggleFilter('inStock', !filters.inStock)}
                        className="w-4 h-4 border-2 border-border-main rounded-sm checked:bg-accent-primary appearance-none relative checked:after:content-['✓'] checked:after:absolute checked:after:inset-0 checked:after:flex checked:after:items-center checked:after:justify-center checked:after:text-[10px] checked:after:text-white transition-all"
                      />
                      <span className="text-xs font-bold text-text-main group-hover:text-accent-primary transition-colors">
                        {t('pages.inventory.grid.filters.in_stock', 'In Stock Only')}
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
                  isRefetching={isRefetching}
                />

                <FilterSection
                  title="Finish"
                  items={Object.entries(FOIL_LABELS).map(([id, label]) => ({ id, label: t(`pages.product.finish.${id}`, label) }))}
                  selected={filters.foil}
                  onToggle={(val) => toggleFilter('foil', val)}
                  counts={facets?.foil}
                  isRefetching={isRefetching}
                />

                <FilterSection
                  title="Version"
                  items={Object.entries(TREATMENT_LABELS).map(([id, label]) => ({ id, label: resolveLabel(id, TREATMENT_LABELS) }))}
                  selected={filters.treatment}
                  onToggle={(val) => toggleFilter('treatment', val)}
                  counts={facets?.treatment}
                  isRefetching={isRefetching}
                />

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Rarity"
                    initialOpen={false}
                    items={[
                      { id: 'Common', label: t('pages.inventory.grid.sort.rarity.common', 'Common'), color: '#1a1714' },
                      { id: 'Uncommon', label: t('pages.inventory.grid.sort.rarity.uncommon', 'Uncommon'), color: '#707883' },
                      { id: 'Rare', label: t('pages.inventory.grid.sort.rarity.rare', 'Rare'), color: '#b59119' },
                      { id: 'Mythic', label: t('pages.inventory.grid.sort.rarity.mythic', 'Mythic'), color: '#d14210' },
                      { id: 'Special', label: t('pages.inventory.grid.sort.rarity.special', 'Special'), color: '#6e2191' },
                      { id: 'Bonus', label: t('pages.inventory.grid.sort.rarity.bonus', 'Bonus'), color: '#6e2191' }
                    ]}
                    selected={filters.rarity}
                    onToggle={(val) => toggleFilter('rarity', val)}
                    counts={facets?.rarity}
                    isRefetching={isRefetching}
                  />
                )}

                <FilterSection
                  title="Set"
                  items={facets?.set_name || []}
                  selected={filters.setName}
                  onToggle={(val) => toggleFilter('setName', val)}
                  counts={setNameCounts}
                  isRefetching={isRefetching}
                />

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Color"
                    initialOpen={false}
                    items={[
                      { id: 'W', label: t('pages.inventory.grid.sort.color.white', 'White'), color: '#f8f6d3', iconClass: 'ms ms-w ms-cost ms-shadow text-[1rem]' },
                      { id: 'U', label: t('pages.inventory.grid.sort.color.blue', 'Blue'), color: '#0e68ab', iconClass: 'ms ms-u ms-cost ms-shadow text-[1rem]' },
                      { id: 'B', label: t('pages.inventory.grid.sort.color.black', 'Black'), color: '#150b00', iconClass: 'ms ms-b ms-cost ms-shadow text-[1rem]' },
                      { id: 'R', label: t('pages.inventory.grid.sort.color.red', 'Red'), color: '#d3202a', iconClass: 'ms ms-r ms-cost ms-shadow text-[1rem]' },
                      { id: 'G', label: t('pages.inventory.grid.sort.color.green', 'Green'), color: '#00733e', iconClass: 'ms ms-g ms-cost ms-shadow text-[1rem]' },
                      { id: 'C', label: t('pages.inventory.grid.sort.color.colorless', 'Colorless'), color: '#90adbb', iconClass: 'ms ms-c ms-cost ms-shadow text-[1rem]' }
                    ]}
                    selected={filters.color}
                    onToggle={(val) => toggleFilter('color', val)}
                    counts={facets?.color}
                    isRefetching={isRefetching}
                  />
                )}

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Language"
                    items={[
                      { id: 'en', label: t('pages.inventory.grid.sort.language.en', 'English') },
                      { id: 'es', label: t('pages.inventory.grid.sort.language.es', 'Spanish') },
                      { id: 'jp', label: t('pages.inventory.grid.sort.language.jp', 'Japanese') },
                      { id: 'it', label: t('pages.inventory.grid.sort.language.it', 'Italian') },
                      { id: 'fr', label: t('pages.inventory.grid.sort.language.fr', 'French') },
                      { id: 'de', label: t('pages.inventory.grid.sort.language.de', 'German') },
                      { id: 'pt', label: t('pages.inventory.grid.sort.language.pt', 'Portuguese') },
                      { id: 'ru', label: t('pages.inventory.grid.sort.language.ru', 'Russian') },
                      { id: 'kr', label: t('pages.inventory.grid.sort.language.kr', 'Korean') },
                      { id: 'zhs', label: t('pages.inventory.grid.sort.language.zhs', 'CH Simpl.') },
                      { id: 'zht', label: t('pages.inventory.grid.sort.language.zht', 'CH Trad.') }
                    ]}
                    selected={filters.language}
                    onToggle={(val) => toggleFilter('language', val)}
                    counts={facets?.language}
                    isRefetching={isRefetching}
                  />
                )}

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Properties"
                    initialOpen={false}
                    items={[
                      { id: 'true', label: t('grid.filters.legendary', 'Legendary') },
                      { id: 'historic', label: t('grid.filters.historic', 'Historic') },
                      { id: 'land', label: t('grid.filters.land', 'Land') },
                      { id: 'basicLand', label: t('grid.filters.basicLand', 'Basic Land') },
                      { id: 'nonBasicLand', label: t('grid.filters.nonBasicLand', 'Non-Basic Land') },
                      { id: 'fullArt', label: t('grid.filters.fullArt', 'Full Art') },
                      { id: 'textless', label: t('grid.filters.textless', 'Textless') },
                      { id: 'creature', label: t('grid.filters.creature', 'Creature') },
                      { id: 'sorcery', label: t('grid.filters.sorcery', 'Sorcery') },
                      { id: 'instant', label: t('grid.filters.instant', 'Instant') },
                      { id: 'artifact', label: t('grid.filters.artifact', 'Artifact') },
                      { id: 'enchantment', label: t('grid.filters.enchantment', 'Enchantment') },
                      { id: 'planeswalker', label: t('grid.filters.planeswalker', 'Planeswalker') }
                    ]}
                    selected={[
                      filters.isLegendary === 'true' ? 'true' : '',
                      filters.isHistoric === 'true' ? 'historic' : '',
                      filters.isLand === 'true' ? 'land' : '',
                      filters.isBasicLand === 'true' ? 'basicLand' : '',
                      filters.isNonBasicLand === 'true' ? 'nonBasicLand' : '',
                      filters.fullArt === 'true' ? 'fullArt' : '',
                      filters.textless === 'true' ? 'textless' : '',
                      filters.isCreature === 'true' ? 'creature' : '',
                      filters.isSorcery === 'true' ? 'sorcery' : '',
                      filters.isInstant === 'true' ? 'instant' : '',
                      filters.isArtifact === 'true' ? 'artifact' : '',
                      filters.isEnchantment === 'true' ? 'enchantment' : '',
                      filters.isPlaneswalker === 'true' ? 'planeswalker' : ''
                    ].filter(Boolean)}
                    onToggle={(id) => {
                      if (id === 'true') toggleFilter('isLegendary', 'true');
                      if (id === 'historic') toggleFilter('isHistoric', 'true');
                      if (id === 'land') toggleFilter('isLand', 'true');
                      if (id === 'basicLand') toggleFilter('isBasicLand', 'true');
                      if (id === 'nonBasicLand') toggleFilter('isNonBasicLand', 'true');
                      if (id === 'fullArt') toggleFilter('fullArt', 'true');
                      if (id === 'textless') toggleFilter('textless', 'true');
                      if (id === 'creature') toggleFilter('isCreature', 'true');
                      if (id === 'sorcery') toggleFilter('isSorcery', 'true');
                      if (id === 'instant') toggleFilter('isInstant', 'true');
                      if (id === 'artifact') toggleFilter('isArtifact', 'true');
                      if (id === 'enchantment') toggleFilter('isEnchantment', 'true');
                      if (id === 'planeswalker') toggleFilter('isPlaneswalker', 'true');
                    }}
                    counts={{
                      'true': facets?.is_legendary?.['true'] || 0,
                      'historic': facets?.is_historic?.['true'] || 0,
                      'land': facets?.is_land?.['true'] || 0,
                      'basicLand': facets?.is_basic_land?.['true'] || 0,
                      'nonBasicLand': facets?.is_non_basic_land?.['true'] || 0,
                      'fullArt': facets?.full_art?.['true'] || 0,
                      'textless': facets?.textless?.['true'] || 0,
                      'creature': facets?.is_creature?.['true'] || 0,
                      'sorcery': facets?.is_sorcery?.['true'] || 0,
                      'instant': facets?.is_instant?.['true'] || 0,
                      'artifact': facets?.is_artifact?.['true'] || 0,
                      'enchantment': facets?.is_enchantment?.['true'] || 0,
                      'planeswalker': facets?.is_planeswalker?.['true'] || 0
                    }}
                    isRefetching={isRefetching}
                  />
                )}

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Card Type"
                    initialOpen={false}
                    items={[
                      { id: 'Creature', label: t('grid.filters.types.creature', 'Creature') },
                      { id: 'Instant', label: t('grid.filters.types.instant', 'Instant') },
                      { id: 'Sorcery', label: t('grid.filters.types.sorcery', 'Sorcery') },
                      { id: 'Artifact', label: t('grid.filters.types.artifact', 'Artifact') },
                      { id: 'Enchantment', label: t('grid.filters.types.enchantment', 'Enchantment') },
                      { id: 'Planeswalker', label: t('grid.filters.types.planeswalker', 'Planeswalker') },
                      { id: 'Land', label: t('grid.filters.types.land', 'Land') },
                      { id: 'Battle', label: t('grid.filters.types.battle', 'Battle') },
                      { id: 'Tribal', label: t('grid.filters.types.tribal', 'Tribal') }
                    ]}
                    selected={filters.cardTypes}
                    onToggle={(val) => toggleFilter('cardTypes', val)}
                    counts={facets?.card_types}
                    isRefetching={isRefetching}
                  />
                )}

                {tcg?.toLowerCase() === 'mtg' && (
                  <FilterSection
                    title="Legality"
                    initialOpen={false}
                    items={[
                      { id: 'commander', label: 'Commander' },
                      { id: 'modern', label: 'Modern' },
                      { id: 'standard', label: 'Standard' },
                      { id: 'legacy', label: 'Legacy' },
                      { id: 'vintage', label: 'Vintage' },
                      { id: 'pauper', label: 'Pauper' },
                      { id: 'pioneer', label: 'Pioneer' }
                    ]}
                    selected={filters.format}
                    onToggle={(val) => toggleFilter('format', val)}
                    counts={facets?.format}
                    isRefetching={isRefetching}
                  />
                )}

                <FilterSection
                  title="Collections"
                  items={availableCollections.map(c => ({ id: c.slug, label: c.name }))}
                  selected={filters.collection}
                  onToggle={(val) => toggleFilter('collection', val)}
                  counts={facets?.collection}
                  isRefetching={isRefetching}
                />

                {(filters.search || filters.foil.length > 0 || filters.treatment.length > 0 || filters.condition.length > 0 || filters.collection.length > 0 || filters.rarity.length > 0 || filters.language.length > 0 || filters.color.length > 0 || filters.setName.length > 0 || filters.isLegendary || filters.isLand || filters.isBasicLand || filters.isNonBasicLand || filters.isHistoric || filters.fullArt || filters.textless || filters.isCreature || filters.isSorcery || filters.isInstant || filters.isArtifact || filters.isEnchantment || filters.isPlaneswalker || filters.format.length > 0 || filters.cardTypes.length > 0) && (
                  <button
                    onClick={() => { setFilters({ search: '', foil: [], treatment: [], condition: [], collection: [], rarity: [], language: [], color: [], setName: [], inStock: true, isLegendary: '', isLand: '', isBasicLand: '', isNonBasicLand: '', isHistoric: '', fullArt: '', textless: '', isCreature: '', isSorcery: '', isInstant: '', isArtifact: '', isEnchantment: '', isPlaneswalker: '', format: [], cardTypes: [] }); setPage(1); }}
                    className="btn-secondary w-full mt-4"
                    style={{ fontSize: '0.85rem', padding: '0.4rem' }}
                  >
                    {t('pages.inventory.grid.filters.clear', 'Clear All Filters ×')}
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
              {loading ? '...' : `${total} ${total === 1 
                                  ? t('pages.inventory.grid.status.result', 'result') 
                                  : t('pages.inventory.grid.status.results', 'results')}`}
              {isRefetching && <span className="ml-2 animate-pulse">↻</span>}
            </p>
            <div className="flex items-center gap-2">
              <label className="text-[10px] font-mono-stack font-bold uppercase" style={{ color: 'var(--text-muted)' }}>{t('pages.inventory.grid.sort.label', 'Sort')}</label>
              <select
                value={sortBy}
                onChange={e => { setSortBy(e.target.value); setPage(1); }}
                className="text-xs font-mono-stack px-2 py-1.5 rounded-sm border-2 border-border-main cursor-pointer"
                style={{ background: 'var(--bg-page)', color: 'var(--text-main)' }}
                id={`sort-${tcg}-${category}`}
              >
                <option value="created_at">{t('pages.inventory.grid.sort.newest', 'Newest')}</option>
                <option value="name">{t('pages.inventory.grid.sort.name', 'Name')}</option>
                <option value="price">{t('pages.inventory.grid.sort.price', 'Price')}</option>
                {(tcg === 'mtg' || tcg === 'all') && <option value="cmc">{t('pages.inventory.grid.sort.cmc', 'Mana Cost')}</option>}
                {(tcg === 'mtg' || tcg === 'all') && <option value="rarity">{t('pages.inventory.grid.sort.rarity', 'Rarity')}</option>}
              </select>
              <button
                onClick={() => { setSortDir(d => d === 'asc' ? 'desc' : 'asc'); setPage(1); }}
                className="flex items-center justify-center w-8 h-8 rounded-sm border-2 border-border-main text-sm font-mono-stack transition-colors hover:bg-bg-page/50"
                style={{ background: 'var(--bg-page)', color: 'var(--text-main)' }}
                title={sortDir === 'asc' ? t('pages.common.status.ascending', 'Ascending') : t('pages.common.status.descending', 'Descending')}
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
                <div key={i} style={{ animationDelay: `${i * 0.05}s` }} className="animate-fade-up">
                  <ProductCardSkeleton />
                </div>
              ))}
            </div>
          ) : (!products || products.length === 0) ? (
            <div className="stamp-border rounded-lg p-16 text-center" style={{ color: 'var(--text-muted)' }}>
              <p className="font-display text-3xl mb-2">{t('pages.inventory.grid.status.no_results', 'NO RESULTS FOUND')}</p>
              <p className="text-sm">{t('pages.inventory.grid.status.no_results_desc', 'Try clearing your filters or check back later.')}</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {products.map((p, i) => (
                <div key={p.id} style={{ animationDelay: `${i * 0.05}s` }} className="animate-fade-up">
                  <ProductCard product={p} />
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex flex-col items-center gap-4 mt-8">
              <div className="flex justify-center gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="btn-secondary"
                  style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === 1 ? 0.4 : 1 }}
                >{t('pages.inventory.grid.pagination.prev', '← Prev')}</button>
                <span className="flex items-center px-3 text-sm font-mono-stack" style={{ color: 'var(--text-secondary)' }}>
                  {page} / {totalPages}
                </span>
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="btn-secondary"
                  style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === totalPages ? 0.4 : 1 }}
                >{t('pages.inventory.grid.pagination.next', 'Next →')}</button>
              </div>
              <div className="flex items-center gap-2">
                <label className="text-[10px] font-mono-stack font-bold uppercase" style={{ color: 'var(--text-muted)' }}>{t('pages.inventory.grid.pagination.per_page', 'Items per page:')}</label>
                <select
                  value={pageSize}
                  onChange={e => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  className="text-xs font-mono-stack px-2 py-1 rounded-sm border-2 border-border-main cursor-pointer"
                  style={{ background: 'var(--bg-page)', color: 'var(--text-main)' }}
                >
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </div>
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
  initialOpen = true,
  isRefetching = false
}: {
  title: string,
  items: { id: string, label: string, color?: string, iconClass?: string }[],
  selected: string[],
  onToggle: (id: string) => void,
  counts?: Record<string, number>,
  initialOpen?: boolean,
  isRefetching?: boolean
}) {
  const { t } = useLanguage();
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
    <div className="border-b border-border-main/10 py-3 px-3 mb-2 bg-black/5 rounded-lg shadow-inner">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex justify-between items-center group mb-1"
      >
        <span className="font-display text-xl sm:text-2xl text-text-main group-hover:text-accent-primary transition-colors uppercase tracking-tight">{t(`grid.filters.${title.toLowerCase()}`, title)}</span>
        <span className={`text-sm transition-transform duration-300 ${isOpen ? 'rotate-180 text-accent-primary' : 'text-text-muted'}`}>
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
                  className="w-4 h-4 border-2 border-border-main rounded-sm checked:bg-accent-primary appearance-none relative checked:after:content-['✓'] checked:after:absolute checked:after:inset-0 checked:after:flex checked:after:items-center checked:after:justify-center checked:after:text-[10px] checked:after:text-white transition-all shrink-0"
                />
                <span className="text-[11px] font-bold text-text-secondary group-hover:text-text-main transition-colors truncate flex items-center gap-2">
                  {item.iconClass ? (
                    <i className={`${item.iconClass} shrink-0`} />
                  ) : item.color && (
                    <span
                      className="w-2.5 h-2.5 rounded-full border border-black/10 shadow-sm shrink-0"
                      style={{ backgroundColor: item.color }}
                    />
                  )}
                  {item.label}
                </span>
              </div>
              {counts && (
                <span className={`text-[10px] font-bold text-gold shrink-0 ml-2 ${isRefetching ? 'animate-pulse opacity-40' : 'opacity-80'}`}>
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
