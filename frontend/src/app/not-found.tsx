'use client';

export const dynamic = 'force-static';

import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';

export default function NotFound() {
  const { t } = useLanguage();

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center p-6 text-center">
      {/* Glitchy 404 Effect */}
      <div className="relative mb-8">
        <h1 className="font-display text-[120px] leading-none opacity-10 select-none">404</h1>
        <div className="absolute inset-0 flex items-center justify-center">
          <h2 className="font-display text-5xl tracking-tighter text-gold uppercase">
            {t('pages.404.title', 'LINK SEVERED')}
          </h2>
        </div>
      </div>

      <div className="max-w-md mx-auto space-y-6">
        <div className="h-px bg-gradient-to-r from-transparent via-gold/30 to-transparent w-full" />
        
        <p className="font-mono-stack text-sm text-text-muted leading-relaxed uppercase tracking-widest">
          {t('pages.404.message', 'The requested coordinates lead to a void in the archive. Error Code: [PAGE_NOT_FOUND].')}
        </p>

        <div className="pt-8">
          <Link 
            href="/" 
            className="btn-primary px-8 py-3 text-sm font-bold tracking-[0.2em] uppercase transition-all hover:scale-105 active:scale-95"
          >
            {t('pages.404.back_home', '← Return to Command Center')}
          </Link>
        </div>
      </div>

      {/* Decorative corners */}
      <div className="fixed top-24 left-1/4 w-8 h-8 border-t-2 border-l-2 border-gold/10 pointer-events-none" />
      <div className="fixed bottom-24 right-1/4 w-8 h-8 border-b-2 border-r-2 border-gold/10 pointer-events-none" />
    </div>
  );
}
