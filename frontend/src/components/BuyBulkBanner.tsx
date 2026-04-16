'use client';

import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';

export default function BuyBulkBanner() {
  const { t } = useLanguage();

  return (
    <section style={{
      background: 'var(--kraft-base)',
      border: '2px solid var(--kraft-shadow)',
      margin: '2rem auto 4rem',
      borderRadius: 4,
      padding: '2rem 1rem 3rem',
      position: 'relative',
      boxShadow: '4px 6px 15px rgba(0,0,0,0.1), inset 0 0 40px rgba(0,0,0,0.05)',
    }} className="centered-container px-4 text-center box-lid">
      <div className="stamp-border inline-block p-1 bg-surface mb-6 rotate-1">
        <div className="border border-dashed border-kraft-shadow px-4 md:px-6 py-2">
          <h2 className="font-display text-4xl md:text-5xl text-hp-color m-0">
            {t('pages.home.sections.got_bulk', 'GOT BULK?')}
          </h2>
        </div>
      </div>
      <p className="text-lg mb-8 max-w-xl mx-auto font-mono-stack font-bold" style={{ color: 'var(--text-primary)' }}>
        {t('pages.home.sections.bulk_cta_text', 'We buy bulk commons and uncommons, bulk rares, and junk rare lots. Box it up and bring it in, get cash. No appointment needed.')}
      </p>
      <Link href="/bulk" className="btn-primary shadow-md" style={{ fontSize: '1.2rem', padding: '0.75rem 2.5rem' }}>
        {t('pages.home.sections.see_bulk_prices', 'SEE BULK PRICES')}
      </Link>
    </section>
  );
}
