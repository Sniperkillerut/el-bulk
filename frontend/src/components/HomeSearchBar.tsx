'use client';

import React, { useState, useEffect, useRef } from 'react';
import { fetchProducts } from '@/lib/api';
import { Product, TCG_SHORT, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import CardImage from './CardImage';
import { openProductModal } from './ProductModalManager';
import { useLanguage } from '@/context/LanguageContext';

interface HomeSearchBarProps {
  placeholder?: string;
}

export default function HomeSearchBar({ placeholder }: HomeSearchBarProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [showResults, setShowResults] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const { addItem } = useCart();
  const { t } = useLanguage();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setShowResults(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setLoading(false);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetchProducts({ search: query, page_size: 10 });
        setResults(res.products);
        setShowResults(true);
      } catch (err) {
        console.error('Search failed:', err);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query]);

  return (
    <div className="relative w-full" ref={containerRef}>
      <div className="relative">
        <input
          type="text"
          placeholder={placeholder || t('components.search.placeholder', 'Search for cards, sets, or products...')}
          className="w-full bg-white/40 border border-ink-plum/20 p-4 pr-12 rounded-sm text-sm text-ink-plum focus:outline-none focus:border-ink-plum/40 transition-all placeholder:text-ink-plum/40"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => query.trim() && setShowResults(true)}
        />
        <div className="absolute right-4 top-1/2 -translate-y-1/2 text-ink-plum/40">
          {loading ? (
            <div className="w-5 h-5 border-2 border-ink-plum border-t-transparent rounded-full animate-spin" />
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          )}
        </div>
      </div>

      {showResults && query.trim() && (
        <div className="absolute top-full left-0 right-0 mt-2 z-50 bg-bg-kraft border border-border-plum shadow-2xl rounded-sm animate-in fade-in slide-in-from-top-2 duration-200">
          <div className="max-h-[500px] overflow-y-auto custom-scrollbar">
            {results && results.length > 0 ? (
              <div className="divide-y divide-border-plum/10">
                {results.map((product) => (
                  <div
                    key={product.id}
                    className="p-4 flex items-center gap-5 hover:bg-ink-plum/5 transition-all group cursor-pointer"
                    style={{ overflow: 'visible' }}
                    onClick={() => {
                      openProductModal(product);
                      setShowResults(false);
                    }}
                  >
                    <div className="w-14 flex-shrink-0 thumb-hover-wrap">
                      <CardImage 
                        imageUrl={product.image_url} 
                        name={product.name} 
                        tcg={product.tcg} 
                        foilTreatment={product.foil_treatment} 
                        height={70} 
                        enableHover={true}
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex flex-wrap items-center gap-2 mb-1">
                        <span className="text-[9px] uppercase font-bold px-1.5 py-0.5 rounded border border-border-plum/20 text-ink-plum/60">
                          {TCG_SHORT[product.tcg] || product.tcg}
                        </span>
                        <p className="text-base font-bold truncate text-ink-plum">{product.name}</p>
                      </div>

                      <div className="flex flex-wrap gap-1 mb-1">
                        {product.foil_treatment !== 'non_foil' && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-ink-lavender/10 text-ink-lavender border border-ink-lavender/20 font-bold">
                            {FOIL_LABELS[product.foil_treatment] || product.foil_treatment}
                          </span>
                        )}
                        {product.card_treatment && product.card_treatment !== 'normal' && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent-rose/10 text-accent-rose border border-accent-rose/20 font-bold">
                            {TREATMENT_LABELS[product.card_treatment] || product.card_treatment}
                          </span>
                        )}
                      </div>

                      <p className="text-xs text-ink-plum/60 truncate">
                        {product.set_name || t('components.search.status.no_set', 'No Set')} • <span className={product.stock > 0 ? 'text-ink-plum' : 'text-accent-rose font-bold'}>
                          {product.stock > 0
                            ? t('components.search.status.in_stock', '{count} IN STOCK', { count: product.stock })
                            : t('components.search.status.out_of_stock', 'OUT OF STOCK')}
                        </span>
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <p className="text-lg font-bold text-ink-plum tracking-tighter">${product.price.toLocaleString()} COP</p>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          addItem(product);
                        }}
                        disabled={product.stock <= 0}
                        className="py-1 px-4 text-[10px] font-bold uppercase rounded border border-ink-plum/20 hover:bg-ink-plum hover:text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {product.stock > 0 ? t('components.search.actions.add', '+ ADD') : t('components.search.actions.sold_out', 'SOLD OUT')}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="p-8 text-center text-ink-plum/40 text-sm">
                {t('components.search.status.no_results', 'No products found for "{query}"').replace('{query}', query)}
              </div>
            )}
            {results && results.length > 0 && (
              <div className="p-2 bg-ink-plum/5 text-center border-t border-border-plum/10">
                <p className="text-[10px] font-bold uppercase tracking-widest text-ink-plum/40">{t('components.search.status.top_results', 'Showing top results')}</p>
              </div>
            )}

          </div>
        </div>
      )}
    </div>
  );
}
