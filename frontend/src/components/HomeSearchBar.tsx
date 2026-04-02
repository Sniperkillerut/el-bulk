'use client';

import React, { useState, useEffect, useRef } from 'react';
import { fetchProducts } from '@/lib/api';
import { Product, TCG_SHORT, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import CardImage from './CardImage';
import { openProductModal } from './ProductModalManager';
import { useLanguage } from '@/context/LanguageContext';

export default function HomeSearchBar() {
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
          placeholder={t('components.search.placeholder', 'Search for cards, sets, or products...')}
          className="w-full bg-white border-2 border-kraft-shadow p-4 pr-12 rounded-sm font-mono-stack text-sm focus:outline-none focus:border-gold transition-colors"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => query.trim() && setShowResults(true)}
        />
        <div className="absolute right-4 top-1/2 -translate-y-1/2 text-kraft-dark">
          {loading ? (
            <div className="w-5 h-5 border-2 border-gold border-t-transparent rounded-full animate-spin" />
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          )}
        </div>
      </div>

      {showResults && query.trim() && (
        <div className="absolute top-full left-0 right-0 mt-2 z-50 bg-surface border-2 border-kraft-dark shadow-2xl rounded-sm animate-in fade-in slide-in-from-top-2 duration-200" style={{ width: 'min(640px, 95vw)', left: '0' }}>
          <div className="max-h-[500px] overflow-y-auto">
            {results && results.length > 0 ? (
              <div className="divide-y divide-kraft-light">
                {results.map((product) => (
                  <div 
                    key={product.id} 
                    className="p-4 flex items-center gap-5 hover:bg-kraft-light/30 transition-colors group cursor-pointer" 
                    style={{ overflow: 'visible' }}
                    onClick={() => {
                      openProductModal(product);
                      setShowResults(false);
                    }}
                  >
                    <div className="w-14 flex-shrink-0 thumb-hover-wrap">
                      <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} height={70} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex flex-wrap items-center gap-2 mb-1">
                        <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded" style={{ background: 'var(--ink-surface)', color: 'var(--kraft-mid)', border: '1px solid var(--ink-border)' }}>
                          {TCG_SHORT[product.tcg] || product.tcg}
                        </span>
                        <p className="text-base font-bold truncate" style={{ color: 'var(--ink-deep)' }}>{product.name}</p>
                      </div>
                      
                      <div className="flex flex-wrap gap-1 mb-1">
                        {product.foil_treatment !== 'non_foil' && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-gold/10 text-gold-dark border border-gold/20 font-bold">
                            {FOIL_LABELS[product.foil_treatment] || product.foil_treatment}
                          </span>
                        )}
                        {product.card_treatment && product.card_treatment !== 'normal' && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-hp-color/10 text-hp-color border border-hp-color/20 font-bold">
                            {TREATMENT_LABELS[product.card_treatment] || product.card_treatment}
                          </span>
                        )}
                        {product.price_cop_override ? (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-kraft-mid/10 text-kraft-dark border border-kraft-dark/20 font-bold">{t('components.search.status.manual_price', 'MANUAL PRICE')}</span>
                        ) : null}
                      </div>

                      <p className="text-xs text-text-muted truncate">
                        {product.set_name || t('components.search.status.no_set', 'No Set')} • <span className={product.stock > 0 ? 'text-text-primary' : 'text-hp-color font-bold'}>
                          {product.stock > 0 
                            ? t('components.search.status.in_stock', '{count} IN STOCK', { count: product.stock }) 
                            : t('components.search.status.out_of_stock', 'OUT OF STOCK')}
                        </span>
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <p className="text-lg font-display text-gold-dark tracking-tighter">${product.price.toLocaleString()} COP</p>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          addItem(product);
                        }}
                        disabled={product.stock <= 0}
                        className="btn-primary py-1 px-4 text-xs whitespace-nowrap shadow-sm hover:scale-105 transition-transform disabled:opacity-50 disabled:cursor-not-allowed disabled:scale-100"
                        style={{ background: product.stock > 0 ? 'var(--ink-deep)' : 'var(--hp-color)', borderColor: 'transparent' }}
                      >
                        {product.stock > 0 ? t('components.search.actions.add', '+ ADD') : t('components.search.actions.sold_out', 'SOLD OUT')}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="p-8 text-center text-text-muted font-mono-stack text-sm">
                {t('components.search.status.no_results', 'No products found for "{query}"').replace('{query}', query)}
              </div>
            )}
            {results && results.length > 0 && (
               <div className="p-2 bg-kraft-light/20 text-center border-t border-kraft-light">
                  <p className="text-[10px] font-mono-stack text-text-muted italic">{t('components.search.status.top_results', 'Showing top results')}</p>
               </div>
            )}

          </div>
        </div>
      )}
    </div>
  );
}
