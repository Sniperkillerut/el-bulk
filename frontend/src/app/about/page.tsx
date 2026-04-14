'use client';

import React from 'react';
import { useLanguage } from '@/context/LanguageContext';
import Link from 'next/link';

export default function AboutPage() {
  const { t } = useLanguage();

  return (
    <div className="min-h-screen bg-bg-page pt-24 pb-16 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        {/* Header Section */}
        <div className="text-center mb-16">
          <div className="inline-block px-3 py-1 bg-accent-primary/10 rounded-full mb-6">
            <span className="font-mono text-[10px] text-accent-primary font-bold tracking-[0.2em] uppercase">
              {t('pages.about.subtitle', 'ESTABLISHED IN BOGOTÁ // 2019')}
            </span>
          </div>
          <h1 className="text-5xl sm:text-7xl font-display text-accent-header mb-6 tracking-tight uppercase">
            {t('pages.about.title', 'THE EL BULK PHILOSOPHY')}
          </h1>
          <div className="w-24 h-1 bg-accent-main mx-auto mb-8" />
        </div>

        {/* Origin Story */}
        <div className="grid md:grid-cols-2 gap-12 items-center mb-24">
          <div className="card bg-ink-surface/30 p-8 border border-border-main/20">
            <h2 className="text-2xl font-display text-accent-main mb-4 uppercase tracking-wide">
              {t('pages.about.origin.title', 'BORN FROM THE BOX')}
            </h2>
            <p className="text-text-main leading-relaxed opacity-90 mb-4">
              {t('pages.about.origin.p1', 'El Bulk started in 2019 as a small collective of players tired of the complexity of finding the pieces they needed for their casual decks. We began by sorting through literal "shoeboxes" of bulk cards to find the hidden gems.')}
            </p>
            <p className="text-text-main leading-relaxed opacity-90">
              {t('pages.about.origin.p2', 'Today, we are Bogotá\'s premier destination for competitive and casual TCG players, specializing in Magic: The Gathering, Pokémon, Lorcana, and One Piece.')}
            </p>
          </div>
          <div className="relative">
            <div className="aspect-[4/5] bg-kraft-paper rounded-lg shadow-2xl rotate-2 flex items-center justify-center p-8 border-4 border-kraft-dark">
              <div className="text-center">
                 <span className="text-6xl mb-4 block">📦</span>
                 <p className="font-mono text-xs text-text-muted uppercase tracking-widest mt-4">
                   {t('pages.about.origin.stamp', 'SECURE_INVENTORY_MATRIX_V4')}
                 </p>
              </div>
            </div>
            <div className="absolute -bottom-6 -left-6 w-32 h-32 bg-accent-primary/20 rounded-full blur-3xl -z-10 animate-pulse"></div>
          </div>
        </div>

        {/* Mission Pillars */}
        <div className="grid sm:grid-cols-3 gap-8 mb-24">
          {[
            { icon: '🔍', title: t('pages.about.pillars.p1.title', 'CURATION'), desc: t('pages.about.pillars.p1.desc', 'We personally inspect every card to ensure accurate grading and authenticity.') },
            { icon: '🤝', title: t('pages.about.pillars.p2.title', 'COMMUNITY'), desc: t('pages.about.pillars.p2.desc', 'Building a safe space for players to grow, trade, and enjoy the hobby.') },
            { icon: '🚀', title: t('pages.about.pillars.p3.title', 'EFFICIENCY'), desc: t('pages.about.pillars.p3.desc', 'Get the cards you need, when you need them, without the hassle.') }
          ].map((pillar, i) => (
            <div key={i} className="text-center p-6 border border-border-main/10 rounded-xl hover:bg-bg-surface/40 transition-colors">
              <div className="text-4xl mb-4">{pillar.icon}</div>
              <h3 className="font-display text-xl text-text-main mb-2 uppercase">{pillar.title}</h3>
              <p className="text-xs text-text-muted leading-relaxed">{pillar.desc}</p>
            </div>
          ))}
        </div>

        {/* Call to Action */}
        <div className="card bg-accent-primary/5 p-12 text-center border border-accent-primary/20">
          <h2 className="text-3xl font-display text-accent-header mb-4 uppercase">
            {t('pages.about.cta.title', 'JOIN THE COLLECTIVE')}
          </h2>
          <p className="text-text-muted mb-8 max-w-lg mx-auto">
            {t('pages.about.cta.desc', 'Whether you are looking for that last piece for your deck or have a stack of bulk to sell, we are here for you.')}
          </p>
          <div className="flex flex-wrap justify-center gap-4">
            <Link href="/" className="btn-primary px-8 py-3">
              {t('pages.about.cta.browse', 'BROWSE ARMORY')}
            </Link>
            <Link href="/contact" className="btn-secondary px-8 py-3 border-border-main/30">
              {t('pages.about.cta.contact', 'GET IN TOUCH')}
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
