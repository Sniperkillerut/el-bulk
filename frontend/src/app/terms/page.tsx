'use client';

import React from 'react';
import { useLanguage } from '@/context/LanguageContext';

export default function TermsPage() {
  const { t } = useLanguage();

  const sections = [
    'rules_of_engagement',
    'condition_grading',
    'inventory_availability',
    'order_fulfillment',
    'pricing_inventory',
    'bulk_submissions',
    'user_conduct',
    'limits_of_liability'
  ];

  return (
    <div className="min-h-screen bg-bg-page pt-24 pb-16 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <div className="border-b border-border-main pb-8 mb-12">
          <h1 className="text-4xl sm:text-5xl font-display text-accent-header mb-4 tracking-tight uppercase">
            {t('pages.terms.title', 'Terms of Engagement')}
          </h1>
          <p className="text-sm font-mono text-text-muted">
            {t('pages.terms.last_updated', 'Last updated: April 12, 2026')}
          </p>
        </div>

        {/* Intro */}
        <div className="prose prose-invert prose-emerald max-w-none">
          <p className="text-lg text-text-main leading-relaxed mb-12 italic opacity-90">
            {t('pages.terms.intro', 'By accessing El Bulk, you agree to these protocols. We curate the common, the uncommon, and the essential pieces of your TCG strategy.')}
          </p>

          {/* Sections */}
          <div className="space-y-12">
            {sections.map((section) => (
              <section key={section} className="scroll-mt-24">
                <h2 className="text-xl font-display text-accent-main mb-4 tracking-wide uppercase border-l-2 border-accent-main pl-4">
                  {t(`pages.terms.sections.${section}.title`)}
                </h2>
                <div className="text-text-main leading-relaxed opacity-80 pl-5">
                  {t(`pages.terms.sections.${section}.content`)}
                </div>
              </section>
            ))}
          </div>
        </div>

        {/* Footer Accent */}
        <div className="mt-20 pt-8 border-t border-border-main text-center">
          <p className="text-xs font-mono text-text-muted uppercase tracking-widest">
            {t('pages.login.accent.store', 'EL BULK')} {"//"} {t('pages.login.accent.logistics', 'LOGISTICS')}
          </p>
        </div>
      </div>
    </div>
  );
}
