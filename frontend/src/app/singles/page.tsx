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
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 mb-20">
          {tcgs.map((t_item) => (
            <Link 
              key={t_item.id} 
              href={`/${t_item.id}/singles`}
              className="group relative block aspect-[16/10] overflow-hidden rounded-xl border-2 border-border-main transition-all hover:border-accent-primary hover:shadow-2xl hover:shadow-accent-primary/20"
            >
              {/* Thematic Background Image */}
              <div 
                className="absolute inset-0 bg-cover bg-center transition-transform duration-700 group-hover:scale-110"
                style={{ backgroundImage: `url(${t_item.image_url || '/hero-banner.png'})` }}
              />
              
              {/* Dark Gradient Overlay for Readability */}
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent" />

              {/* Content Overlay */}
              <div className="absolute inset-0 p-8 flex flex-col justify-end items-start">
                {/* ID Tag */}
                <span className="mb-2 px-2 py-0.5 bg-accent-primary/20 backdrop-blur-md border border-accent-primary/30 text-accent-primary font-mono-stack text-[10px] font-bold tracking-widest uppercase rounded-sm">
                  {t_item.id}
                </span>

                <h3 className="font-display text-3xl md:text-4xl text-white group-hover:text-accent-primary transition-colors leading-none mb-3">
                  {t_item.name.toUpperCase()}
                </h3>

                <div className="flex items-center gap-2 group-hover:translate-x-2 transition-transform duration-300">
                  <p className="text-xs font-mono-stack font-bold text-white/70 uppercase tracking-widest">
                    {t('pages.singles.landing.view_all', 'VIEW ALL SINGLES')}
                  </p>
                  <span className="text-accent-primary font-bold">→</span>
                </div>
              </div>

              {/* Glassmorphic border glow effect on hover */}
              <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity bg-gradient-to-br from-accent-primary/10 to-transparent pointer-events-none" />
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
