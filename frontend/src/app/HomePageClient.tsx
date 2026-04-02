'use client';

import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';
import ProductCard from '@/components/ProductCard';
import BountyCard from '@/components/BountyCard';
import NoticeSection from '@/components/NoticeSection';
import HeroSection from '@/components/HeroSection';
import BuyBulkBanner from '@/components/BuyBulkBanner';
import BountiesTitle from '@/components/BountiesTitle';
import { CustomCategory, TCG, Product, Bounty } from '@/lib/types';

interface HomePageClientProps {
  categories: CustomCategory[];
  tcgs: TCG[];
  collections: { category: CustomCategory; products: Product[] }[];
  bounties: Bounty[];
}

export default function HomePageClient({ categories, tcgs, collections, bounties }: HomePageClientProps) {
  const { t } = useLanguage();

  return (
    <div>
      {/* Hero Section - Cardboard Box Aesthetic */}
      <HeroSection />

      {/* Gold divider */}
      <div className="gold-line" />

      {/* TCG Nav strips */}
      <section style={{ background: 'var(--ink-surface)', borderBottom: '1px dashed var(--kraft-dark)', padding: '1rem' }}>
        <div className="centered-container px-4 flex flex-wrap gap-x-6 gap-y-3 justify-center">
          {tcgs.map(t => (
            <Link key={t.id} href={`/${t.id}/singles`}
              className="text-xs sm:text-sm font-display tracking-widest transition-opacity hover:opacity-70 whitespace-nowrap"
              style={{ color: 'var(--text-primary)' }}>
              {t.name.toUpperCase()}
            </Link>
          ))}
          {categories.filter(cat => cat.searchable).map(cat => (
            <Link key={cat.id} href={`/collection/${cat.slug}`}
              className="text-xs sm:text-sm font-display tracking-widest transition-opacity hover:text-gold whitespace-nowrap"
              style={{ color: 'var(--text-muted)' }}>
              {cat.name.toUpperCase()}
            </Link>
          ))}
        </div>
      </section>

      <div className="centered-container px-4 py-8 space-y-16">
        {collections.length === 0 ? (
          <div className="stamp-border rounded-sm p-8 text-center" style={{ color: 'var(--text-muted)' }}>
            <p className="font-display text-2xl mb-2">{t('pages.home.status.empty', 'STORE IS EMPTY')}</p>
            <p className="font-mono-stack text-sm">{t('pages.home.status.empty_desc', 'No collections have been populated yet.')}</p>
          </div>
        ) : (
          collections.map(col => (
            <section key={col.category.id}>
              <div className="flex items-baseline justify-between gap-4 mb-6 border-b-2 border-kraft-dark pb-2">
                <h2 className="font-display text-4xl uppercase" style={{ color: 'var(--ink-deep)' }}>
                  {col.category.name}
                </h2>
                <Link href={`/collection/${col.category.slug}`} className="text-sm font-bold font-mono-stack hover:text-gold transition-colors" style={{ color: 'var(--text-secondary)' }}>
                  {t('pages.home.buttons.view_all', 'VIEW ALL →')}
                </Link>
              </div>
              {col.products.length === 0 ? (
                <div className="text-center p-8 bg-ink-surface border border-dashed border-ink-border rounded-sm">
                  <p className="font-mono-stack text-sm text-text-muted">{t('pages.home.status.no_items', 'No items assigned to this collection yet.')}</p>
                </div>
              ) : (
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                  {col.products.map(p => <ProductCard key={p.id} product={p} />)}
                </div>
              )}
            </section>
          ))
        )}

        {/* Featured Bounties Section */}
        {bounties.length > 0 && (
          <section>
            <div className="flex items-baseline justify-between gap-4 mb-6 border-b-2 border-kraft-dark pb-2">
              <div className="flex items-center gap-3">
                <BountiesTitle />
                <div className="hidden sm:block px-2 py-0.5 bg-hp-color text-white text-[10px] font-bold font-mono-stack rotate-2">
                  {t('pages.home.tags.urgently_needed', 'Urgently Needed')}
                </div>
              </div>
              <Link href="/bounties" className="text-sm font-bold font-mono-stack hover:text-gold transition-colors" style={{ color: 'var(--text-secondary)' }}>
                {t('pages.home.buttons.view_all', 'VIEW ALL →')}
              </Link>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 gap-4">
              {bounties.map(b => <BountyCard key={b.id} bounty={b} />)}
            </div>
          </section>
        )}

        {/* Notices Section */}
        <NoticeSection />
      </div>

      {/* Buy Bulk CTA Banner - Cardboard Package Style */}
      <BuyBulkBanner />
    </div>
  );
}
