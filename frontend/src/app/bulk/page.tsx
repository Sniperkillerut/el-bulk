import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Sell Your Bulk — El Bulk TCG Store',
  description: 'Bring your bulk commons, uncommons, and rares. We pay cash on the spot.',
};

const tiers = [
  { label: 'BULK COMMONS & UNCOMMONS', price: '$5 per 1,000', desc: 'Any condition, any set. Sorted or unsorted. We take it all.', icon: '📦' },
  { label: 'BULK RARES & MYTHICS', price: '$0.25 per card', desc: 'NM/LP only. Bulk rares from Standard and below.', icon: '💎' },
  { label: 'JUNK RARE LOTS', price: '$3 per 100', desc: 'MP-DMG rares and mythics, or commons/uncommons in poor condition.', icon: '🗑️' },
  { label: 'FOIL COMMONS & UNCOMMONS', price: '$10 per 500', desc: 'Any condition. Foil bulk sorted separately.', icon: '✨' },
];

const accepts = [
  'Magic: The Gathering (all sets, all formats)',
  'Pokémon TCG (English only)',
  'Disney Lorcana',
  'One Piece TCG',
  'Basic lands (we pay $1 per 200)',
  'Tokens and emblems ($0 value but we take them)',
];

const doesNotAccept = [
  'Non-English cards (MTG foreign language)',
  'Heavily played or water damaged bulk',
  'World of Warcraft TCG or other discontinued games (ask us first)',
  'Price-listed singles you want buylist prices for (see our buylist)',
];

export default function BulkPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      {/* Header */}
      <div className="mb-10 text-center">
        <div className="badge mb-4 inline-block" style={{ background: 'rgba(212,175,55,0.15)', color: 'var(--gold)', border: '1px solid rgba(212,175,55,0.3)' }}>
          WE BUY
        </div>
        <h1 className="font-display text-7xl mb-4">
          BRING YOUR <span style={{ color: 'var(--gold)' }}>BULK</span>
        </h1>
        <p className="text-lg max-w-xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
          Got a shoebox of old cards gathering dust? We'll take them off your hands
          and put cash in your pocket. No appointment needed — just walk in.
        </p>
        <div className="gold-line mt-8" />
      </div>

      {/* Pricing tiers */}
      <section className="mb-12">
        <h2 className="font-display text-3xl mb-6">CURRENT BULK PRICES</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {tiers.map(tier => (
            <div key={tier.label} className="card p-5">
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
            ⚠ Prices updated regularly. Large lots (1,000+ cards) may receive bonus offers. Prices are in-store cash — store credit offers up to 25% more.
          </p>
        </div>
      </section>

      {/* What we accept */}
      <div className="grid md:grid-cols-2 gap-6 mb-12">
        <section className="card p-6">
          <h2 className="font-display text-2xl mb-4" style={{ color: 'var(--nm-color)' }}>✓ WE ACCEPT</h2>
          <ul className="flex flex-col gap-2">
            {accepts.map(item => (
              <li key={item} className="flex items-start gap-2 text-sm" style={{ color: 'var(--text-secondary)' }}>
                <span style={{ color: 'var(--nm-color)', flexShrink: 0, marginTop: 2 }}>✓</span>
                {item}
              </li>
            ))}
          </ul>
        </section>

        <section className="card p-6">
          <h2 className="font-display text-2xl mb-4" style={{ color: 'var(--hp-color)' }}>✗ WE DON&#39;T ACCEPT</h2>
          <ul className="flex flex-col gap-2">
            {doesNotAccept.map(item => (
              <li key={item} className="flex items-start gap-2 text-sm" style={{ color: 'var(--text-secondary)' }}>
                <span style={{ color: 'var(--hp-color)', flexShrink: 0, marginTop: 2 }}>✗</span>
                {item}
              </li>
            ))}
          </ul>
        </section>
      </div>

      {/* How it works */}
      <section className="mb-12">
        <h2 className="font-display text-3xl mb-6">HOW IT WORKS</h2>
        <div className="flex flex-col gap-4">
          {[
            ['1', 'BRING YOUR CARDS', 'Walk in with your bulk. Sorted or unsorted, boxed or bagged — doesn\'t matter.'],
            ['2', 'WE COUNT & GRADE', 'We do a quick count of your cards. For large lots we may ask for a day to sort.'],
            ['3', 'GET AN OFFER', 'We give you a cash offer on the spot. No pressure to accept.'],
            ['4', 'TAKE THE CASH', 'Accept and walk out with cash (or store credit for more). Simple.'],
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
        <h2 className="font-display text-4xl mb-3">HAVE QUESTIONS?</h2>
        <p className="text-sm mb-4" style={{ color: 'var(--text-secondary)' }}>
          Got a large lot or special collection? Reach out before coming in.
        </p>
        <a href="mailto:bulk@elbulk.com" className="btn-primary">EMAIL US ABOUT YOUR BULK</a>
      </div>
    </div>
  );
}
