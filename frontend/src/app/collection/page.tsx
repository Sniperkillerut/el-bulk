'use client';

import { useUser } from '@/context/UserContext';
import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';

export default function CollectionPage() {
  const { user } = useUser();
  const { t } = useLanguage();

  return (
    <div className="min-h-screen bg-kraft-paper flex flex-col items-center justify-center p-12 overflow-hidden relative">
      {/* Decorative Accents */}
      <div className="absolute top-12 left-12 font-display text-[80px] opacity-[0.03] select-none pointer-events-none">
        {t('pages.collection.vault.bg_text_1', 'MY VAULT')}
      </div>
      <div className="absolute bottom-12 right-12 font-display text-[80px] opacity-[0.03] select-none pointer-events-none rotate-180">
        {t('pages.collection.vault.bg_text_2', 'COLLECTION')}
      </div>

      <div className="cardbox max-w-2xl w-full p-12 backdrop-blur-md bg-white/40 border border-white/20 shadow-2xl text-center relative">
        <div className="mb-8">
          <h1 className="font-display text-5xl tracking-tighter text-ink-deep mb-4 uppercase">
            {t('pages.collection.vault.title_part1', 'THE')} <span className="text-gold">{t('pages.collection.vault.title_part2', 'VAULT')}</span>
          </h1>
          <div className="h-px bg-gold/30 w-24 mx-auto mb-6"></div>
          <p className="font-mono-stack text-[11px] uppercase tracking-[0.3em] text-text-muted opacity-80">
            {t('pages.collection.vault.subject_prefix', 'Secure Digital Archive // Subject:')} {user?.first_name || t('pages.collection.vault.default_subject', 'COLLECTOR')}
          </p>
        </div>

        <div className="py-20 border-y border-kraft-dark/10 my-8">
          <div className="text-6xl mb-6 opacity-20">🗃️</div>
          <h2 className="font-mono-stack text-sm font-bold text-ink-deep mb-2">{t('pages.collection.vault.scanning', 'SCANNING INVENTORY...')}</h2>
          <p className="text-sm text-text-muted max-w-sm mx-auto">
            {t('pages.collection.vault.desc', 'Your personal TCG archive is currently being indexed. Soon you will be able to manage your physical collection and track market valuations here.')}
          </p>
        </div>

        <Link 
          href="/" 
          className="inline-block font-mono-stack text-[10px] font-bold text-gold hover:text-hp-color transition-colors tracking-widest uppercase"
        >
          ← {t('pages.collection.vault.return', 'Return to Command Center')}
        </Link>
      </div>
    </div>
  );
}
