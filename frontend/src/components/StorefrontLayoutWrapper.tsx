'use client';

import { useState, useEffect } from 'react';
import IsClient from './IsClient';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import Navbar from './Navbar';
import ProductModalManager from './ProductModalManager';
import BountyModalManager from './BountyModalManager';
import CookieBanner from './CookieBanner';
import { useLanguage } from '@/context/LanguageContext';

export default function StorefrontLayoutWrapper({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { t } = useLanguage();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const isAdmin = pathname?.startsWith('/admin');

  // Defer the layout decision until after hydration to avoid mismatch
  if (mounted && isAdmin) {
    return <>{children}</>;
  }

  return (
    <>
      <Navbar />
      <ProductModalManager />
      <BountyModalManager />
      <CookieBanner />
      <main id="main-content" data-theme-area="main" className="min-h-[calc(100vh-64px)]">
        {children}
      </main>
      <footer
        id="main-footer"
        data-theme-area="footer"
        className="bg-bg-header border-t border-border-main py-8 px-4"
      >
        <div className="w-full px-4 sm:px-6 lg:px-8 flex flex-col md:flex-row items-center justify-between gap-4">
          <div>
            <span className="font-display text-xl text-accent-header">EL BULK</span>
            <span className="text-xs ml-2 text-text-muted font-mono">{t('pages.nav.main.tcg_store', 'TCG STORE')}</span>
          </div>
          <p className="text-xs text-center text-text-muted">
            {t('pages.layout.footer.slogan', 'We buy bulk. We sell singles. We love cardboard.')}
          </p>
          <p className="text-xs text-text-muted font-mono flex flex-wrap justify-center gap-x-3 gap-y-1">
            <span>© <IsClient fallback="2026">{new Date().getFullYear()}</IsClient> El Bulk</span>
            <span className="opacity-30">•</span>
            <Link href="/about" className="hover:text-gold transition-colors underline decoration-border-main underline-offset-4">{t('pages.about.title_short', 'About')}</Link>
            <span className="opacity-30">•</span>
            <Link href="/shipping" className="hover:text-gold transition-colors underline decoration-border-main underline-offset-4">{t('pages.shipping.title_short', 'Shipping')}</Link>
            <span className="opacity-30">•</span>
            <Link href="/refunds" className="hover:text-gold transition-colors underline decoration-border-main underline-offset-4">{t('pages.refunds.title_short', 'Refunds')}</Link>
            <span className="opacity-30">•</span>
            <Link href="/terms" className="hover:text-gold transition-colors underline decoration-border-main underline-offset-4">{t('pages.terms.title_short', 'Terms')}</Link>
            <span className="opacity-30">•</span>
            <Link href="/privacy" className="hover:text-gold transition-colors underline decoration-border-main underline-offset-4">{t('pages.privacy.title_short', 'Privacy')}</Link>
          </p>
        </div>
      </footer>
    </>
  );
}
