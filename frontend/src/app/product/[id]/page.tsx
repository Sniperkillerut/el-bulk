'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchProduct } from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import Link from 'next/link';
import DeckContents from '@/components/DeckContents';

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [added, setAdded] = useState(false);
  const { addItem } = useCart();

  useEffect(() => {
    if (!id) return;

    setLoading(true);
    fetchProduct(id)
      .then(p => {
        setProduct(p);
        setError(false);
      })
      .catch(() => {
        // Error is already logged to the server via the API client's logAndThrow.
        // We set local state to show the UI without triggering a red console error.
        setError(true);
      })
      .finally(() => setLoading(false));
  }, [id]);

  const handleAddCart = () => {
    if (!product) return;
    addItem(product);
    setAdded(true);
    setTimeout(() => setAdded(false), 1500);
  };

  if (loading) return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      <div className="grid md:grid-cols-2 gap-8">
        <div className="skeleton" style={{ height: 400, borderRadius: 2 }} />
        <div className="flex flex-col gap-4">
          <div className="skeleton" style={{ height: 32, width: '60%' }} />
          <div className="skeleton" style={{ height: 20, width: '40%' }} />
          <div className="skeleton" style={{ height: 48, width: '30%' }} />
        </div>
      </div>
    </div>
  );

  if (error || !product) return (
    <div className="max-w-4xl mx-auto px-4 py-16 text-center stamp-border mt-12 bg-surface p-12">
      <h1 className="font-display text-5xl mb-4 text-hp-color">ITEM NOT FOUND</h1>
      <p style={{ color: 'var(--text-muted)' }} className="mb-6 font-mono-stack">This item may have been sold or removed.</p>
      <Link href="/" className="btn-secondary">← Back to Shoebox</Link>
    </div>
  );

  const outOfStock = product.stock === 0;
  const conditionClass = product.condition ? `badge-${product.condition.toLowerCase()}` : '';
  const displayDescription = product.oracle_text || product.description;

  return (
    <div className="max-w-5xl mx-auto px-4 py-12">
      {/* Breadcrumb */}
      <nav className="text-xs mb-6 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
        <Link href="/" className="hover:text-text-primary transition-colors" style={{ textDecoration: 'none' }}>Home</Link>
        {' / '}
        <Link href={`/${product.tcg}/${product.category}`} className="hover:text-text-primary transition-colors uppercase" style={{ textDecoration: 'none' }}>
          {product.tcg} {product.category}
        </Link>
        {' / '}
        <span style={{ color: 'var(--text-primary)' }}>{product.name}</span>
      </nav>

      <div className="grid md:grid-cols-2 gap-10">
        {/* Image */}
        <div>
          <div className="cardbox overflow-hidden" style={{ aspectRatio: '3/4', padding: '0.5rem', background: 'var(--kraft-light)' }}>
            <div className="w-full h-full bg-ink-card border border-ink-border relative" style={{ boxShadow: 'inset 0 0 10px rgba(0,0,0,0.05)' }}>
              {product.image_url ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={product.image_url} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'contain', padding: '1rem' }} />
              ) : (
                <div className="product-img-placeholder h-full w-full" style={{ fontSize: '5rem', border: 'none', margin: 0 }}>🃏</div>
              )}
            </div>
          </div>
        </div>

        {/* Details label (acting like a physical label on a box) */}
        <div className="flex flex-col">
          <div className="cardbox p-8 flex flex-col h-full bg-surface" style={{ borderLeft: '4px solid var(--kraft-dark)' }}>
            <div>
              {product.set_name && (
                <p className="text-xs mb-1 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                  {product.set_code ? `[${product.set_code}] ` : ''}{product.set_name}
                </p>
              )}
              <h1 className="font-display text-4xl md:text-5xl" style={{ color: 'var(--ink-deep)' }}>
                {product.name}
              </h1>
              {product.type_line && (
                <p className="text-xs mt-2 font-mono-stack" style={{ color: 'var(--text-secondary)', fontWeight: 'bold' }}>
                  {product.type_line}
                </p>
              )}
              {product.artist && (
                <p className="text-[10px] mt-1 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                  Art by {product.artist} {product.collector_number ? `(#${product.collector_number})` : ''}
                </p>
              )}
            </div>

            {/* Badges */}
            <div className="flex flex-wrap gap-2 mt-3 block">
              <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--text-muted)', border: '1px solid var(--kraft-dark)' }}>
                {product.language?.toUpperCase() || 'EN'}
              </span>
              {product.condition && <span className={`badge ${conditionClass} border`}>{product.condition}</span>}
              {product.promo_type && product.promo_type !== 'none' && (
                <span className="badge" style={{ background: 'var(--hp-color)', color: '#fff', border: 'none' }}>
                  {resolveLabel(product.promo_type, {})}
                </span>
              )}
              {product.foil_treatment !== 'non_foil' && FOIL_LABELS[product.foil_treatment] && (
                <span className="badge badge-foil">✦ {FOIL_LABELS[product.foil_treatment]}</span>
              )}
              {product.card_treatment !== 'normal' && TREATMENT_LABELS[product.card_treatment] && (
                <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--text-secondary)', border: '1px solid var(--kraft-dark)' }}>
                  {TREATMENT_LABELS[product.card_treatment]}
                </span>
              )}
              {product.textless && (
                <span className="badge" style={{ background: 'rgba(248,113,113,0.1)', color: 'var(--hp-color)', border: '1px solid var(--hp-color)' }}>
                  TEXTLESS
                </span>
              )}
              {product.full_art && product.card_treatment !== 'full_art' && (
                <span className="badge" style={{ background: 'rgba(120,180,120,0.1)', color: 'var(--nm-color)', border: '1px solid var(--nm-color)' }}>
                  FULL ART
                </span>
              )}
            </div>
            
            {product.tcg === 'mtg' && product.category === 'singles' && (
              <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 py-3 border-t border-b border-dashed border-kraft-dark">
                <div className="text-center">
                  <p className="text-[10px] font-bold text-text-muted uppercase">Identity</p>
                  <p className="text-sm font-mono-stack">{product.color_identity || 'C'}</p>
                </div>
                <div className="text-center border-l md:border-l border-dashed border-kraft-dark px-2">
                  <p className="text-[10px] font-bold text-text-muted uppercase">Rarity</p>
                  <p className="text-sm font-mono-stack capitalize">{product.rarity || 'Common'}</p>
                </div>
                <div className="text-center border-l border-dashed border-kraft-dark px-2">
                  <p className="text-[10px] font-bold text-text-muted uppercase">Art Var.</p>
                  <p className="text-sm font-mono-stack truncate">{product.art_variation || 'Normal'}</p>
                </div>
                <div className="text-center border-l border-dashed border-kraft-dark px-2">
                  <p className="text-[10px] font-bold text-text-muted uppercase">CMC</p>
                  <p className="text-sm font-mono-stack">{product.cmc ?? 0}</p>
                </div>
              </div>
            )}

            <hr className="divider w-full" />

            {/* Price */}
            <div className="flex items-baseline gap-3">
              <span className="price text-5xl tracking-tighter">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
              <span className="text-xs font-mono-stack font-bold px-2 py-1 rounded-sm" style={{ background: outOfStock ? 'var(--hp-color)' : 'var(--nm-color)', color: '#fff' }}>
                {outOfStock ? 'OUT OF STOCK' : `${product.stock} IN STOCK`}
              </span>
            </div>

            {/* Description / Rules Text */}
            <div className="mt-6 flex-1">
              {displayDescription ? (
                <div className="flex flex-col gap-4">
                  {product.oracle_text && (
                    <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm" style={{ background: 'rgba(230, 218, 195, 0.4)', color: 'var(--ink-deep)', border: '1px dashed var(--kraft-dark)' }}>
                      {product.oracle_text}
                    </div>
                  )}
                  {!product.oracle_text && product.description && (
                     <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm" style={{ background: 'rgba(230, 218, 195, 0.4)', color: 'var(--text-secondary)', border: '1px dashed var(--kraft-dark)' }}>
                       {product.description}
                     </div>
                  )}
                </div>
              ) : (
                <div className="text-sm italic" style={{ color: 'var(--text-muted)' }}>
                  No additional information available.
                </div>
              )}
            </div>

            {/* Deck Cards Grid for Store Exclusives */}
            {product.category === 'store_exclusives' && product.deck_cards && product.deck_cards.length > 0 && (
               <DeckContents cards={product.deck_cards} tcg={product.tcg} className="border-t border-kraft-dark pt-6 mt-6" />
            )}

            {/* Actions */}
            <div className="flex gap-3 mt-8">
              <button
                id={`detail-add-to-cart-${product.id}`}
                onClick={handleAddCart}
                disabled={outOfStock}
                className="btn-primary w-full text-lg py-4 shadow-sm"
                style={{ textAlign: 'center', opacity: outOfStock ? 0.4 : 1, cursor: outOfStock ? 'not-allowed' : 'pointer' }}
              >
                {added ? '✓ ADDED TO CART' : outOfStock ? 'SOLD OUT' : 'ADD TO CART'}
              </button>
            </div>

            <div className="mt-4 text-center">
              <p className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                🏪 Complete purchase in-store or verify availability at counter.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
