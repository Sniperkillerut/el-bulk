'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { fetchProducts, fetchTCGs } from '@/lib/api';
import { Product, TCG } from '@/lib/types';
import ProductCard from '@/components/ProductCard';
import { useLanguage } from '@/context/LanguageContext';

export default function SinglesLandingPage() {
  const [featured, setFeatured] = useState<Product[]>([]);
  const [tcgs, setTcgs] = useState<TCG[]>([]);
  const [loading, setLoading] = useState(true);
  const { t } = useLanguage();

  useEffect(() => {
    async function loadData() {
      try {
        const [productsRes, tcgsRes] = await Promise.all([
          fetchProducts({ category: 'singles', collection: 'featured', page_size: 12 }),
          fetchTCGs(true)
        ]);
        setFeatured(productsRes.products);
        setTcgs(tcgsRes);
      } catch (err) {
        console.error('Failed to fetch data for singles landing:', err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, []);

  return (
    <div className="min-h-screen pb-20">
      {/* Header Section */}
      <section className="bg-kraft-mid border-b-4 border-kraft-dark py-10 md:py-12 px-4 relative overflow-hidden box-lid">
        <div className="centered-container relative z-10 text-center px-4">
          <div className="badge inline-flex mb-4" style={{ background: 'var(--kraft-light)', color: 'var(--ink-deep)', transform: 'rotate(-1deg)' }}>
            {t('pages.singles.landing.category', 'CATEGORY // SINGLES')}
          </div>
          <h1 className="font-display text-5xl sm:text-6xl md:text-7xl mb-4" style={{ color: 'var(--ink-deep)' }}>
            {t('pages.singles.landing.title.main', 'INDIVIDUAL')} <span style={{ color: 'var(--gold-dark)' }}>{t('pages.singles.landing.title.accent', 'CARDS')}</span>
          </h1>
          <p className="text-base md:text-lg max-w-2xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
            {t('pages.singles.landing.desc', 'Browse our collection of hundreds of singles across your favorite TCGs. Pick your game to see the full inventory.')}
          </p>
        </div>
      </section>

      <div className="centered-container px-4 mt-12">
        {/* TCG Selection Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-20">
          {tcgs.map((t_item) => (
            <Link 
              key={t_item.id} 
              href={`/${t_item.id}/singles`}
              className="group relative block transition-transform hover:-translate-y-1"
            >
              <div className="card h-full p-8 flex flex-col items-center justify-center text-center gap-4 border-2 border-kraft-dark transition-shadow hover:shadow-xl relative overflow-hidden">
                <div className="w-16 h-16 rounded-full flex items-center justify-center font-display text-2xl mb-2" 
                  style={{ background: 'var(--ink-surface)', border: '2px solid var(--gold)', color: 'var(--gold)' }}>
                  {t_item.name[0]}
                </div>
                <div>
                  <h3 className="font-display text-2xl group-hover:text-gold-dark transition-colors">{t_item.name}</h3>
                  <p className="text-xs font-mono-stack mt-1" style={{ color: 'var(--text-muted)' }}>{t('pages.singles.landing.view_all', 'VIEW ALL SINGLES →')}</p>
                </div>
                
                <div className="absolute -bottom-4 -right-4 opacity-5 rotate-12 group-hover:opacity-10 transition-opacity">
                   <h2 className="text-8xl font-display">{t_item.name.substring(0, 3).toUpperCase()}</h2>
                </div>
              </div>
            </Link>
          ))}
        </div>

        {/* Featured across all TCGs */}
        <section>
          <div className="flex items-center gap-4 mb-8">
            <h2 className="font-display text-4xl whitespace-nowrap" style={{ color: 'var(--ink-deep)' }}>
              {t('pages.singles.landing.featured.main', 'FEATURED')} <span style={{ color: 'var(--text-muted)' }}>{t('pages.singles.landing.featured.accent', 'SINGLES')}</span>
            </h2>
            <div className="h-[2px] w-full bg-kraft-dark" />
          </div>

          {loading ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4 animate-pulse">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="bg-kraft-light h-64 rounded-sm" />
              ))}
            </div>
          ) : featured.length === 0 ? (
            <div className="stamp-border rounded-sm p-12 text-center" style={{ color: 'var(--text-muted)' }}>
              <p className="font-display text-2xl mb-2">{t('pages.singles.landing.no_featured', 'NO FEATURED SINGLES FOUND')}</p>
              <p className="text-sm">{t('pages.inventory.grid.status.no_results_desc', 'Try selecting a specific TCG above.')}</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {featured.map(p => <ProductCard key={p.id} product={p} />)}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
