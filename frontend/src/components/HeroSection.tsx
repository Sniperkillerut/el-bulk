'use client';

import Link from 'next/link';
import HomeSearchBar from '@/components/HomeSearchBar';
import { useLanguage } from '@/context/LanguageContext';

export default function HeroSection() {
  const { t } = useLanguage();

  return (
    <section className="relative w-full overflow-hidden bg-ink-deep box-lid" style={{ minHeight: '300px' }}>
      {/* Background Image - Absolute Positioning & Cover */}
      <div
        className="absolute inset-0 z-0 bg-cover bg-center bg-no-repeat"
        style={{
          backgroundImage: 'url("/hero-banner.png")',
          backgroundSize: 'cover',
          backgroundRepeat: 'no-repeat',
          filter: 'brightness(0.4) contrast(1.1)'
        }}
      />

      {/* Stronger Gradient Overlay for Text Readability */}
      <div className="absolute inset-0 z-10 bg-gradient-to-t from-ink-deep via-ink-deep/40 to-transparent md:bg-gradient-to-r md:from-ink-deep/90 md:via-ink-deep/60 md:to-transparent" />

      {/* Subtle Cardboard Texture Overlay */}
      <div className="absolute inset-0 z-20 opacity-5 pointer-events-none" style={{
        backgroundImage: 'url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'400\' height=\'400\'%3E%3Cfilter id=\'noise\'%3E%3CfeTurbulence type=\'fractalNoise\' baseFrequency=\'0.8\' numOctaves=\'3\' stitchTiles=\'stitch\'/%3E%3CfeColorMatrix type=\'saturate\' values=\'0\'/%3E%3C/filter%3E%3Crect width=\'400\' height=\'400\' filter=\'url(%23noise)\' opacity=\'1\'/%3E%3C/svg%3E")',
        backgroundRepeat: 'repeat'
      }} />

      <div className="centered-container relative z-30 h-full flex flex-col justify-center items-center lg:items-start px-6 py-10 md:py-16">
        <div className="max-w-3xl animate-fade-up flex flex-col items-center lg:items-start text-center lg:text-left">
          {/* Brand Tag */}
          <div className="inline-flex items-center gap-2 px-3 py-1.5 bg-ink-deep text-[#FFD700] font-mono-stack text-[11px] font-bold tracking-widest mb-8 rotate-[-1deg] shadow-xl border border-gold">
            {t('pages.home.hero.subtitle', 'YOUR LOCAL TCG SHOP')} • EST. 2024
          </div>

          {/* Main Slogan - High Impact & Shielded for Readability */}
          <div className="mb-8">
            <h1 className="font-display text-fluid-h1 leading-[0.82] drop-shadow-[0_4px_12px_rgba(0,0,0,0.5)]" style={{ color: '#FFD700' }}>
              {t('pages.home.hero.title_refined', 'EL BULK / THE CARDS THEY OVERLOOKED.')}
            </h1>
          </div>

          {/* Supportive Text - High Contrast */}
          <p className="text-xl md:text-2xl text-white/95 mb-12 max-w-2xl font-medium leading-relaxed drop-shadow-[0_2px_4px_rgba(0,0,0,0.4)]">
            {t('pages.home.hero.description_refined', "The essential pieces others treat as trash. We curate the common, the uncommon, and the impossible-to-find cards that complete your strategy.")}
          </p>

          {/* Functional Search Bar Area */}
          <div className="mb-12 max-w-xl">
            <div className="p-1 bg-white/10 backdrop-blur-md rounded-md shadow-2xl border border-white/20">
              <HomeSearchBar placeholder={t('pages.home.hero.search_placeholder', 'Find the cards everyone else missed...')} />
            </div>
          </div>

          {/* CTAs */}
          <div className="flex flex-wrap gap-5">
            <Link href="/singles" className="btn-primary flex items-center justify-center min-w-[220px] py-4 text-xl border-2 border-transparent transition-all hover:scale-105 active:scale-95 shadow-xl">
              {t('pages.home.hero.browse_singles', 'BROWSE SINGLES')}
            </Link>
            <Link href="/bulk" className="btn-secondary flex items-center justify-center min-w-[220px] py-4 text-xl bg-transparent border-2 border-gold hover:bg-gold hover:text-ink-deep transition-all shadow-xl font-bold" style={{ color: '#FFD700' }}>
              {t('pages.home.hero.sell_bulk', 'SELL YOUR BULK →')}
            </Link>
          </div>
        </div>
      </div>

      {/* Decorative Box Detail */}
      <div className="absolute bottom-0 left-0 right-0 h-1.5 bg-gradient-to-r from-gold via-gold-dark to-gold z-10 opacity-80" />
    </section>
  );
}
