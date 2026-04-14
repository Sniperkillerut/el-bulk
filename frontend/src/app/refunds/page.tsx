'use client';

import React from 'react';
import { useLanguage } from '@/context/LanguageContext';

export default function RefundsPage() {
  const { t } = useLanguage();

  const sections = [
    {
      id: 'general',
      title: t('pages.refunds.sections.general.title', 'PURCHASE PROTOCOLS'),
      content: t('pages.refunds.sections.general.content', 'All sales at El Bulk are conducted through our secure evaluation process. Because we specialize in collectibles and singles, specific conditions apply to returns and refunds.')
    },
    {
      id: 'grading',
      title: t('pages.refunds.sections.grading.title', 'GRADING DISCREPANCIES'),
      content: t('pages.refunds.sections.grading.content', 'We strive for 100% accuracy in our grading. If you receive a card that you believe is significantly lower in condition than what was advertised, please contact us within 48 hours of receipt for an evaluation.')
    },
    {
      id: 'returns',
      title: t('pages.refunds.sections.returns.title', 'RETURNS & EXCHANGES'),
      content: t('pages.refunds.sections.returns.content', 'Returns are accepted for unopened sealed products within 7 days of delivery or pickup, provided the original seal is intact. Due to market volatility, single cards cannot be returned once the transaction is finalized, except in cases of grading error or damage during shipping.')
    },
    {
      id: 'process',
      title: t('pages.refunds.sections.process.title', 'REFUND PROCESS'),
      content: t('pages.refunds.sections.process.content', 'Approved refunds are processed via the original payment method within 3-5 business days. For local pickup orders, refunds can be issued as store credit or through the original digital payment method.')
    }
  ];

  return (
    <div className="min-h-screen bg-bg-page pt-24 pb-16 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <div className="border-b border-border-main pb-8 mb-12">
          <h1 className="text-4xl sm:text-5xl font-display text-accent-header mb-4 tracking-tight uppercase">
            {t('pages.refunds.title', 'REFUND & RETURN POLICY')}
          </h1>
          <p className="text-sm font-mono text-text-muted italic">
            {t('pages.refunds.legal_intro', 'PROTECTING THE COLLECTIVE INTEGRITY')}
          </p>
        </div>

        <div className="space-y-12">
          {sections.map((section) => (
            <section key={section.id} className="scroll-mt-24">
              <h2 className="text-xl font-display text-accent-main mb-4 tracking-wide uppercase border-l-2 border-accent-main pl-4">
                {section.title}
              </h2>
              <div className="text-text-main leading-relaxed opacity-80 pl-5 whitespace-pre-wrap">
                {section.content}
              </div>
            </section>
          ))}
        </div>

        <div className="mt-16 p-8 bg-hp-color/5 border border-hp-color/20 rounded-lg">
          <p className="text-xs text-hp-color font-mono uppercase tracking-widest mb-2 font-bold">{t('pages.refunds.warning_title', 'SEALED PRODUCT NOTICE')}</p>
          <p className="text-sm text-text-main opacity-80">
            {t('pages.refunds.warning_desc', 'Opening any factory-sealed product immediately voids all return eligibility. No exceptions will be made for "bad pulls" or randomized content.')}
          </p>
        </div>

        <div className="mt-20 pt-8 border-t border-border-main text-center">
          <p className="text-xs font-mono text-text-muted uppercase tracking-widest">
            {t('pages.login.accent.store', 'EL BULK')} {"//"} {t('pages.login.accent.logistics', 'LOGISTICS')}
          </p>
        </div>
      </div>
    </div>
  );
}
