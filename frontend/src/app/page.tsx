import Link from 'next/link';
import { fetchProducts } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import { TCG_SHORT, KNOWN_TCGS } from '@/lib/types';

export default async function HomePage() {
  let featured = { products: [] as Awaited<ReturnType<typeof fetchProducts>>['products'] };
  try {
    featured = await fetchProducts({ featured: true, page_size: 8 });
  } catch {
    // DB not connected in dev — show empty state gracefully
  }

  return (
    <div>
      {/* Hero */}
      <section style={{
        background: 'linear-gradient(135deg, var(--ink-navy) 0%, #0f1420 50%, var(--ink-deep) 100%)',
        borderBottom: '1px solid var(--ink-border)',
        padding: '5rem 1rem 4rem',
        position: 'relative',
        overflow: 'hidden',
      }}>
        {/* Decorative box lines */}
        <div style={{
          position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
          backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 59px, rgba(46,54,80,0.3) 60px)',
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
          backgroundImage: 'repeating-linear-gradient(90deg, transparent, transparent 59px, rgba(46,54,80,0.15) 60px)',
          pointerEvents: 'none',
        }} />

        <div className="max-w-7xl mx-auto relative">
          <div className="max-w-2xl">
            <div className="badge" style={{ background: 'rgba(212,175,55,0.15)', color: 'var(--gold)', border: '1px solid rgba(212,175,55,0.3)', marginBottom: '1rem', display: 'inline-block' }}>
              YOUR LOCAL TCG SHOP
            </div>
            <h1 className="font-display text-8xl leading-none mb-4" style={{ color: 'var(--text-primary)' }}>
              EL<br/>
              <span style={{ color: 'var(--gold)' }}>BULK</span>
            </h1>
            <p className="text-lg mb-8" style={{ color: 'var(--text-secondary)', maxWidth: 480 }}>
              Singles, sealed, accessories. Magic, Pokémon, Lorcana and more.
              And we pay <strong style={{ color: 'var(--gold)' }}>cash for your bulk.</strong>
            </p>
            <div className="flex flex-wrap gap-3">
              <Link href="/mtg/singles" className="btn-primary">SHOP MTG SINGLES</Link>
              <Link href="/bulk" className="btn-secondary">SELL YOUR BULK →</Link>
            </div>
          </div>
        </div>
      </section>

      {/* Gold divider */}
      <div className="gold-line" />

      {/* TCG Nav strips */}
      <section style={{ background: 'var(--ink-surface)', borderBottom: '1px solid var(--ink-border)', padding: '1rem' }}>
        <div className="max-w-7xl mx-auto flex flex-wrap gap-3 justify-center">
          {KNOWN_TCGS.map(tcg => (
            <Link key={tcg} href={`/${tcg}/singles`}
              className="btn-secondary"
              style={{ fontSize: '0.85rem', padding: '0.4rem 1.2rem' }}>
              {TCG_SHORT[tcg]}
            </Link>
          ))}
        </div>
      </section>

      {/* Featured products */}
      <section className="max-w-7xl mx-auto px-4 py-12">
        <div className="flex items-end gap-4 mb-8">
          <h2 className="font-display text-5xl" style={{ color: 'var(--text-primary)' }}>
            FEATURED <span style={{ color: 'var(--gold)' }}>SINGLES</span>
          </h2>
          <Link href="/mtg/singles" className="text-sm mb-1" style={{ color: 'var(--text-muted)', textDecoration: 'none' }}>
            View all →
          </Link>
        </div>

        {featured.products.length === 0 ? (
          <div className="stamp-border rounded-lg p-12 text-center" style={{ color: 'var(--text-muted)' }}>
            <p className="font-display text-3xl mb-2">COMING SOON</p>
            <p className="text-sm">Products will appear here once inventory is loaded.</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
            {featured.products.map(p => <ProductCard key={p.id} product={p} />)}
          </div>
        )}
      </section>

      {/* Buy Bulk CTA Banner */}
      <section style={{
        background: 'linear-gradient(135deg, #1a2010 0%, var(--ink-navy) 100%)',
        border: '1px solid rgba(212,175,55,0.2)',
        margin: '0 1rem 2rem',
        borderRadius: 12,
        padding: '3rem 2rem',
      }} className="max-w-7xl mx-auto text-center">
        <div className="gold-line mb-6" />
        <h2 className="font-display text-5xl mb-3">
          GOT <span style={{ color: 'var(--gold)' }}>BULK?</span>
        </h2>
        <p className="text-lg mb-6 max-w-xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
          We buy bulk commons and uncommons, bulk rares, and junk rare lots.
          Bring it in, get cash. No appointment needed.
        </p>
        <Link href="/bulk" className="btn-primary" style={{ fontSize: '1.2rem', padding: '0.75rem 2.5rem' }}>
          SEE BULK PRICES
        </Link>
        <div className="gold-line mt-6" />
      </section>
    </div>
  );
}
