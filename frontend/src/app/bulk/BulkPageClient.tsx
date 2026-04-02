'use client';

import { useLanguage } from '@/context/LanguageContext';

export default function BulkPageClient() {
  const { t } = useLanguage();

  const tiers = [
    { key: 'pages.bulk.tiers.common', price: t('pages.bulk.tiers.common.price', '$20.000 COP per 1,000'), label: t('pages.bulk.tiers.common.label', 'BULK COMMONS & UNCOMMONS'), desc: t('pages.bulk.tiers.common.desc', 'Any condition, any set. Sorted or unsorted. We take it all.'), icon: '📦' },
    { key: 'pages.bulk.tiers.rare', price: t('pages.bulk.tiers.rare.price', '$1.000 COP per card'), label: t('pages.bulk.tiers.rare.label', 'BULK RARES & MYTHICS'), desc: t('pages.bulk.tiers.rare.desc', 'NM/LP only. Bulk rares from Standard and below.'), icon: '💎' },
    { key: 'pages.bulk.tiers.junk', price: t('pages.bulk.tiers.junk.price', '$12.000 COP per 100'), label: t('pages.bulk.tiers.junk.label', 'JUNK RARE LOTS'), desc: t('pages.bulk.tiers.junk.desc', 'MP-DMG rares and mythics, or commons/uncommons in poor condition.'), icon: '🗑️' },
    { key: 'pages.bulk.tiers.foil', price: t('pages.bulk.tiers.foil.price', '$40.000 COP per 500'), label: t('pages.bulk.tiers.foil.label', 'FOIL COMMONS & UNCOMMONS'), desc: t('pages.bulk.tiers.foil.desc', 'Any condition. Foil bulk sorted separately.'), icon: '✨' },
  ];

  const accepts = [
    t('pages.bulk.accepts.0', 'Magic: The Gathering (all sets, all formats)'),
    t('pages.bulk.accepts.1', 'Pokémon TCG (English only)'),
    t('pages.bulk.accepts.2', 'Disney Lorcana'),
    t('pages.bulk.accepts.3', 'One Piece TCG'),
    t('pages.bulk.accepts.4', 'Basic lands (we pay $4.000 COP per 200)'),
    t('pages.bulk.accepts.5', 'Tokens and emblems ($0 value but we take them)'),
  ];

  return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      {/* Header */}
      <div className="mb-10 text-center">
        <div className="badge mb-4 inline-block" style={{ background: 'rgba(212,175,55,0.15)', color: 'var(--gold)', border: '1px solid rgba(212,175,55,0.3)' }}>
          {t('pages.bulk.intro.we_buy', 'WE BUY')}
        </div>
        <h1 className="font-display text-7xl mb-4">
          {t('pages.bulk.page.title', 'BRING YOUR BULK').split(' ').map((word, i) => (
            <span key={i} style={word === 'BULK' ? { color: 'var(--gold)' } : {}}>
              {word}{' '}
            </span>
          ))}
        </h1>
        <p className="text-lg max-w-xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
          {t('pages.bulk.page.subtitle', "Got a shoebox of old cards gathering dust? We'll take them off your hands and put cash in your pocket. No appointment needed — just walk in.")}
        </p>
        <div className="gold-line mt-8" />
      </div>

      {/* Pricing tiers */}
      <section className="mb-12">
        <h2 className="font-display text-3xl mb-6">{t('pages.bulk.sections.prices', 'CURRENT BULK PRICES')}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {tiers.map(tier => (
            <div key={tier.key} className="card p-5">
              <div className="text-3xl mb-3">{tier.icon}</div>
              <div className="badge mb-2 inline-block" style={{ background: 'var(--ink-surface)', color: 'var(--kraft-mid)', border: '1px solid var(--ink-border)' }}>
                {tier.label}
              </div>
              <p className="price text-2xl my-2">{tier.price}</p>
              <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>{tier.desc}</p>
            </div>
          ))}
        </div>

        <div style={{ background: 'rgba(212,175,55,0.05)', border: '1px solid rgba(212,175,55,0.2)', borderRadius: 8, padding: '1rem 1.25rem', marginTop: '1rem' }}>
          <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
            {t('pages.bulk.sections.price_disclaimer', '⚠ Prices updated regularly. Large lots (1,000+ cards) may receive bonus offers. Prices are in-store cash — store credit offers up to 25% more.')}
          </p>
        </div>
      </section>

      {/* What we accept */}
      <div className="mb-12">
        <section className="card p-8 bg-surface border-gold" style={{ borderLeft: '4px solid var(--gold)' }}>
          <h2 className="font-display text-4xl mb-6 text-gold-dark">✓ {t('pages.bulk.sections.accept', 'WE GLADLY ACCEPT')}</h2>
          <div className="grid sm:grid-cols-2 gap-4">
            {accepts.map((item, i) => (
              <div key={i} className="flex items-center gap-3 p-3 rounded-sm bg-kraft-light/30 border border-kraft-dark/10">
                <span className="text-nm-color text-xl">⚓</span>
                <span className="text-sm font-semibold text-ink-deep">{item}</span>
              </div>
            ))}
          </div>
        </section>
      </div>

      {/* How it works */}
      <section className="mb-12">
        <h2 className="font-display text-3xl mb-6">{t('pages.bulk.sections.how_it_works', 'HOW IT WORKS')}</h2>
        <div className="flex flex-col gap-4">
          {[
            ['1', t('pages.bulk.how.1.title', 'BRING YOUR CARDS'), t('pages.bulk.how.1.desc', 'Walk in with your bulk. Sorted or unsorted, boxed or bagged — doesn\'t matter.')],
            ['2', t('pages.bulk.how.2.title', 'WE COUNT & GRADE'), t('pages.bulk.how.2.desc', 'We do a quick count of your cards. For large lots we may ask for a day to sort.')],
            ['3', t('pages.bulk.how.3.title', 'GET AN OFFER'), t('pages.bulk.how.3.desc', 'We give you a cash offer on the spot. No pressure to accept.')],
            ['4', t('pages.bulk.how.4.title', 'TAKE THE CASH'), t('pages.bulk.how.4.desc', 'Accept and walk out with cash (or store credit for more). Simple.')],
          ].map(([num, title, desc]) => (
            <div key={num} className="card p-4 flex gap-4 items-start">
              <div className="font-display text-4xl text-gold" style={{ minWidth: 40, lineHeight: 1 }}>{num}</div>
              <div>
                <p className="font-display text-lg mb-1" style={{ color: 'var(--text-primary)' }}>{title}</p>
                <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>{desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <div className="text-center stamp-border rounded-xl p-8">
        <h2 className="font-display text-4xl mb-3">{t('pages.bulk.cta.questions', 'HAVE QUESTIONS?')}</h2>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          {t('pages.bulk.cta.questions_desc', 'Got a large lot or special collection? Reach out before coming in.')}
        </p>
        <a href="mailto:bulk@elbulk.com" className="btn-primary">{t('pages.bulk.cta.email_btn', 'EMAIL US ABOUT YOUR BULK')}</a>
      </div>
    </div>
  );
}
