'use client';

import React from 'react';
import { useLanguage } from '@/context/LanguageContext';

export default function ShippingPage() {
  const { t } = useLanguage();

  return (
    <div className="min-h-screen bg-bg-page pt-24 pb-16 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <div className="border-b border-border-main pb-8 mb-12">
          <h1 className="text-4xl sm:text-5xl font-display text-accent-header mb-4 tracking-tight uppercase">
            {t('pages.shipping.title', 'SHIPPING & LOGISTICS')}
          </h1>
          <p className="text-sm font-mono text-text-muted italic">
            {t('pages.shipping.subtitle', 'GLOBAL INVENTORY DEPLOYMENT PROTOCOLS')}
          </p>
        </div>

        <div className="space-y-16">
          {/* Dispatch Section */}
          <section className="scroll-mt-24">
            <h2 className="text-xl font-display text-accent-main mb-6 tracking-wide uppercase flex items-center gap-3">
              <span className="w-8 h-px bg-accent-main"></span>
              {t('pages.shipping.sections.dispatch.title', 'DISPATCH TIMELINES')}
            </h2>
            <div className="card p-8 bg-surface border border-border-main/20">
              <p className="text-text-main leading-relaxed opacity-90 mb-4">
                {t('pages.shipping.sections.dispatch.content', 'All orders are processed and dispatched within 24-48 business hours of payment confirmation. Orders placed on weekends or national holidays in Colombia will be processed the next business day.')}
              </p>
              <div className="inline-block p-2 border border-dashed border-accent-primary/40 rounded mt-2">
                <span className="text-[10px] font-mono text-accent-primary uppercase font-bold tracking-widest">
                  {t('pages.shipping.notice', 'PRIORITY DEPLOYMENT ACTIVE')}
                </span>
              </div>
            </div>
          </section>

          {/* Local Pickup Section */}
          <section className="scroll-mt-24">
            <h2 className="text-xl font-display text-accent-main mb-6 tracking-wide uppercase flex items-center gap-3">
              <span className="w-8 h-px bg-accent-main"></span>
              {t('pages.shipping.sections.pickup.title', 'LOCAL PICKUP (BOGOTÁ)')}
            </h2>
            <div className="card p-8 bg-ink-surface/30 border border-border-main/20">
              <p className="text-text-main leading-relaxed opacity-90 mb-4">
                {t('pages.shipping.sections.pickup.content', 'Local pickup is available at our flagship location in Bogotá. Please wait for an "Order Ready for Pickup" notification (via Email or WhatsApp) before visiting.')}
              </p>
              <ul className="text-sm text-text-muted space-y-2 list-disc pl-5 mt-4">
               <li>{t('pages.shipping.pickup.p1', 'Free of charge for all orders.')}</li>
               <li>{t('pages.shipping.pickup.p2', 'Valid ID required for pickup.')}</li>
               <li>{t('pages.shipping.pickup.p3', 'Pickup hours: Monday to Friday (11 AM - 7 PM), Saturdays (10 AM - 4 PM).')}</li>
              </ul>
            </div>
          </section>

          {/* Delivery Rates Section */}
          <section className="scroll-mt-24">
            <h2 className="text-xl font-display text-accent-main mb-6 tracking-wide uppercase flex items-center gap-3">
              <span className="w-8 h-px bg-accent-main"></span>
              {t('pages.shipping.sections.rates.title', 'DELIVERY RATES & COVERAGE')}
            </h2>
            <div className="grid md:grid-cols-2 gap-6">
              <div className="p-6 bg-surface border-t-4 border-accent-primary rounded shadow-sm">
                <h3 className="font-display text-lg mb-2 uppercase">{t('pages.shipping.rates.nat.title', 'NATIONAL SHIPPING')}</h3>
                <p className="text-xs text-text-muted mb-4">{t('pages.shipping.rates.nat.desc', 'Standard delivery across Colombia via certified carriers.')}</p>
                <p className="font-bold text-lg text-accent-primary">{t('pages.shipping.rates.nat.price', 'FLAT RATE: $15,000 COP')}</p>
              </div>
              <div className="p-6 bg-surface border-t-4 border-ink-deep rounded shadow-sm">
                <h3 className="font-display text-lg mb-2 uppercase">{t('pages.shipping.rates.intl.title', 'INTERNATIONAL')}</h3>
                <p className="text-xs text-text-muted mb-4">{t('pages.shipping.rates.intl.desc', 'Global shipping available upon request.')}</p>
                <p className="font-bold text-sm text-ink-deep italic">{t('pages.shipping.rates.intl.price', 'CONTACT FOR QUOTE')}</p>
              </div>
            </div>
          </section>
        </div>

        {/* Packing Standard */}
        <div className="mt-20 p-10 bg-kraft-paper border-4 border-kraft-dark shadow-xl rotate-1">
          <h2 className="text-2xl font-display text-ink-deep mb-4 uppercase">{t('pages.shipping.packing.title', 'OUR PACKING STANDARD')}</h2>
          <p className="text-sm text-ink-deep leading-relaxed">
            {t('pages.shipping.packing.content', 'We know how important condition is. All singles are shipped in sleeves and "toploaders" or rigid semi-rigid holders, secured within waterproof bubble mailers. Every package is sealed with tamper-evident protocol.')}
          </p>
        </div>

        <div className="mt-24 pt-8 border-t border-border-main text-center opacity-60">
          <p className="text-xs font-mono text-text-muted uppercase tracking-widest">
            {t('pages.login.accent.store', 'EL BULK')} {"//"} {t('pages.login.accent.logistics', 'LOGISTICS')}
          </p>
        </div>
      </div>
    </div>
  );
}
