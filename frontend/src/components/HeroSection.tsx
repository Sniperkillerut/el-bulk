'use client';

import Link from 'next/link';
import HomeSearchBar from '@/components/HomeSearchBar';
import { useLanguage } from '@/context/LanguageContext';

export default function HeroSection() {
  const { t } = useLanguage();

  return (
    <section style={{
      background: 'var(--kraft-mid)',
      borderBottom: '4px solid var(--kraft-dark)',
      padding: '3rem 1rem 3rem',
      position: 'relative',
      minHeight: 'min-content'
    }} className="box-lid">
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
        backgroundImage: 'linear-gradient(rgba(139, 121, 92, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(139, 121, 92, 0.05) 1px, transparent 1px)',
        backgroundSize: '20px 20px',
        pointerEvents: 'none',
      }} />
      <div className="tape-stripe absolute top-4 left-0" />
      <div className="tape-stripe absolute bottom-4 right-0" style={{ transform: 'rotate(180deg)' }} />

      <div className="centered-container relative mt-4 md:mt-6 px-4">
        <div className="max-w-2xl bg-surface p-6 md:p-8 rounded-sm shadow-sm" style={{ border: '2px solid var(--kraft-shadow)', position: 'relative' }}>
          <div className="absolute top-0 right-10 w-16 h-8 bg-kraft-light hidden sm:block" style={{ transform: 'translateY(-50%)', border: '1px solid var(--kraft-shadow)' }} />
          <div className="absolute top-0 right-12 w-12 h-8 bg-kraft-mid" style={{ transform: 'translateY(-50%) rotate(5deg)', border: '1px solid var(--kraft-shadow)' }} />

          <div className="badge flex items-center justify-center inline-flex" style={{ background: 'var(--kraft-light)', color: 'var(--hp-color)', borderColor: 'var(--hp-color)', marginBottom: '1.5rem', borderWidth: '2px', transform: 'rotate(-2deg)' }}>
            STORE_01 // {t('pages.home.hero.subtitle', 'YOUR LOCAL TCG SHOP')}
          </div>
          <h1 className="font-display text-5xl sm:text-7xl md:text-8xl leading-none mb-4 text-fluid-h1" style={{ color: 'var(--ink-deep)' }}>
            EL <span style={{ color: 'var(--gold-dark)' }}>BULK</span>
          </h1>
          <p className="text-base md:text-lg mb-8" style={{ color: 'var(--text-secondary)', maxWidth: 480 }}>
            {t('pages.home.hero.description', 'The shoebox where we keep all the good stuff. Singles, sealed product, and accessories.')}
          </p>

          <div className="mb-8 max-w-lg">
            <HomeSearchBar />
          </div>

          <div className="responsive-stack gap-3">
            <Link href="/singles" className="btn-primary text-center">{t('pages.home.hero.browse_singles', 'BROWSE SINGLES')}</Link>
            <Link href="/bulk" className="btn-secondary text-center">{t('pages.home.hero.sell_bulk', 'SELL YOUR BULK →')}</Link>
          </div>
        </div>
      </div>
    </section>
  );
}
