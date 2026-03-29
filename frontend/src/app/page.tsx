import Link from 'next/link';
import { fetchProducts, fetchCategories, fetchTCGs } from '@/lib/api';
import ProductCard from '@/components/ProductCard';
import { TCG_SHORT, CustomCategory } from '@/lib/types';
import HomeSearchBar from '@/components/HomeSearchBar';

export default async function HomePage() {
  let categories: CustomCategory[] = [];
  let tcgs: import('@/lib/types').TCG[] = [];
  let collections: { category: CustomCategory; products: import('@/lib/types').Product[] }[] = [];
  
  try {
    const [catsRes, tcgsRes] = await Promise.all([
      fetchCategories(),
      fetchTCGs(true)
    ]);
    categories = catsRes;
    tcgs = tcgsRes;
    
    // Fetch top 4 products for each category
    collections = await Promise.all(
      categories.filter(cat => cat.is_active).map(async (cat) => {
        const res = await fetchProducts({ page: 1, page_size: 4, collection: cat.slug });
        return { category: cat, products: res.products };
      })
    );
  } catch {
    // DB not connected in dev — show empty state gracefully
  }

  return (
    <div>
      {/* Hero Section - Cardboard Box Aesthetic */}
      <section style={{
        background: 'var(--kraft-mid)',
        borderBottom: '4px solid var(--kraft-dark)',
        padding: '3rem 1rem 3rem',
        position: 'relative',
        minHeight: 'min-content'
      }} className="box-lid">
        <div style={{
          position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
          backgroundImage: 'linear-gradient(rgba(139, 121, 92, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(139, 121, 92, 0.05) 1px, transparent 1px)',
          backgroundSize: '20px 20px',
          pointerEvents: 'none',
        }} />
        <div className="tape-stripe absolute top-4 left-0" />
        <div className="tape-stripe absolute bottom-4 right-0" style={{ transform: 'rotate(180deg)' }} />

        <div className="centered-container relative mt-4 md:mt-6 px-4">
          <div className="max-w-2xl bg-surface p-6 md:p-8 rounded-sm shadow-sm" style={{ border: '2px solid var(--kraft-shadow)', position: 'relative' }}>
            <div className="absolute top-0 right-10 w-16 h-8 bg-kraft-light hidden sm:block" style={{ transform: 'translateY(-50%)', border: '1px solid var(--kraft-shadow)' }} />
            <div className="absolute top-0 right-12 w-12 h-8 bg-kraft-mid" style={{ transform: 'translateY(-50%) rotate(5deg)', border: '1px solid var(--kraft-shadow)' }} />
            
            <div className="badge flex items-center justify-center inline-flex" style={{ background: 'var(--kraft-light)', color: 'var(--hp-color)', borderColor: 'var(--hp-color)', marginBottom: '1.5rem', borderWidth: '2px', transform: 'rotate(-2deg)' }}>
              STORE_01 // YOUR LOCAL TCG SHOP
            </div>
            <h1 className="font-display text-5xl sm:text-7xl md:text-8xl leading-none mb-4" style={{ color: 'var(--ink-deep)' }}>
              EL <span style={{ color: 'var(--gold-dark)' }}>BULK</span>
            </h1>
            <p className="text-base md:text-lg mb-8" style={{ color: 'var(--text-secondary)', maxWidth: 480 }}>
              The shoebox where we keep all the good stuff. 
              Singles, sealed product, and accessories. 
              And we pay <strong style={{ color: 'var(--gold-dark)' }}>cash for your bulk.</strong>
            </p>

            <div className="mb-8 max-w-lg">
              <HomeSearchBar />
            </div>

            <div className="responsive-stack gap-3">
              {tcgs.length > 0 ? (
                <Link href={`/${tcgs[0].id}/singles`} className="btn-primary text-center">SHOP {tcgs[0].name.toUpperCase()}</Link>
              ) : (
                <Link href="/singles" className="btn-primary text-center">BROWSE SINGLES</Link>
              )}
              <Link href="/bulk" className="btn-secondary text-center">SELL YOUR BULK →</Link>
            </div>
          </div>
        </div>
      </section>

      {/* Gold divider */}
      <div className="gold-line" />

      {/* TCG Nav strips */}
      <section style={{ background: 'var(--ink-surface)', borderBottom: '1px dashed var(--kraft-dark)', padding: '1rem' }}>
        <div className="centered-container px-4 flex flex-wrap gap-x-6 gap-y-3 justify-center">
          {tcgs.map(t => (
            <Link key={t.id} href={`/${t.id}/singles`}
              className="text-xs sm:text-sm font-display tracking-widest transition-opacity hover:opacity-70 whitespace-nowrap"
              style={{ color: 'var(--text-primary)' }}>
              {t.name.toUpperCase()}
            </Link>
          ))}
          {categories.filter(cat => cat.searchable).map(cat => (
            <Link key={cat.id} href={`/collection/${cat.slug}`}
              className="text-xs sm:text-sm font-display tracking-widest transition-opacity hover:text-gold whitespace-nowrap"
              style={{ color: 'var(--text-muted)' }}>
              {cat.name.toUpperCase()}
            </Link>
          ))}
        </div>
      </section>

      <div className="centered-container px-4 py-8 space-y-16">
        {collections.length === 0 ? (
           <div className="stamp-border rounded-sm p-8 text-center" style={{ color: 'var(--text-muted)' }}>
             <p className="font-display text-2xl mb-2">STORE IS EMPTY</p>
             <p className="font-mono-stack text-sm">No collections have been populated yet.</p>
           </div>
        ) : (
          collections.map(col => (
              <section key={col.category.id}>
                <div className="flex items-baseline justify-between gap-4 mb-6 border-b-2 border-kraft-dark pb-2">
                  <h2 className="font-display text-4xl uppercase" style={{ color: 'var(--ink-deep)' }}>
                    {col.category.name}
                  </h2>
                  <Link href={`/collection/${col.category.slug}`} className="text-sm font-bold font-mono-stack hover:text-gold transition-colors" style={{ color: 'var(--text-secondary)' }}>
                    VIEW ALL →
                  </Link>
                </div>
                {col.products.length === 0 ? (
                  <div className="text-center p-8 bg-ink-surface border border-dashed border-ink-border rounded-sm">
                    <p className="font-mono-stack text-sm text-text-muted">No items assigned to this collection yet.</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                    {col.products.map(p => <ProductCard key={p.id} product={p} />)}
                  </div>
                )}
              </section>
          ))
        )}
      </div>

      {/* Buy Bulk CTA Banner - Cardboard Package Style */}
      <section style={{
        background: 'var(--kraft-base)',
        border: '2px solid var(--kraft-shadow)',
        margin: '2rem 1rem 4rem',
        borderRadius: 4,
        padding: '2rem 1rem 3rem',
        position: 'relative',
        boxShadow: '4px 6px 15px rgba(0,0,0,0.1), inset 0 0 40px rgba(0,0,0,0.05)',
      }} className="centered-container px-4 text-center box-lid">
        <div className="stamp-border inline-block p-1 bg-surface mb-6 rotate-1">
          <div className="border border-dashed border-kraft-shadow px-4 md:px-6 py-2">
            <h2 className="font-display text-4xl md:text-5xl text-hp-color m-0">
              GOT BULK?
            </h2>
          </div>
        </div>
        <p className="text-lg mb-8 max-w-xl mx-auto font-mono-stack font-bold" style={{ color: 'var(--text-primary)' }}>
          We buy bulk commons and uncommons, bulk rares, and junk rare lots.
          Box it up and bring it in, get cash. No appointment needed.
        </p>
        <Link href="/bulk" className="btn-primary shadow-md" style={{ fontSize: '1.2rem', padding: '0.75rem 2.5rem' }}>
          SEE BULK PRICES
        </Link>
      </section>
    </div>
  );
}
