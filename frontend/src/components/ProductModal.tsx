'use client';

import { useEffect, useState } from 'react';
import { fetchProduct } from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import Link from 'next/link';

interface ProductModalProps {
  productId: string;
  initialProduct?: Product;
  onClose: () => void;
}

export default function ProductModal({ productId, initialProduct, onClose }: ProductModalProps) {
  const [product, setProduct] = useState<Product | null>(initialProduct || null);
  const [loading, setLoading] = useState(!initialProduct);
  const [error, setError] = useState(false);
  const [added, setAdded] = useState(false);
  const { addItem } = useCart();

  // Prevent scrolling on body when modal is open
  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, []);


  useEffect(() => {
    if (initialProduct) {
      setProduct(initialProduct);
      setLoading(false);
      return;
    }

    if (!productId) return;

    setLoading(true);
    fetchProduct(productId)
      .then(p => {
      setProduct(p);
      setError(false);
      })
      .catch((err) => {
        console.error("fetchProduct failed:", err);
        setError(true);
      })
      .finally(() => setLoading(false));
  }, [productId, initialProduct]);

  const handleAddCart = () => {
    if (!product) return;
    addItem(product);
    setAdded(true);
    setTimeout(() => setAdded(false), 1500);
  };

  // Close on Escape sequence
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center px-4 py-8"
      style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}
      onClick={onClose}
    >
      <div 
        className="w-full max-w-5xl max-h-[90vh] overflow-y-auto"
        onClick={e => e.stopPropagation()} // Prevent close when clicking inside
      >
        <div className="flex justify-end mb-2">
          <button onClick={onClose} className="text-white hover:text-gold text-3xl leading-none">×</button>
        </div>

        {loading ? (
          <div className="card no-tilt bg-surface p-12">
            <div className="grid md:grid-cols-2 gap-8">
              <div className="skeleton" style={{ height: 400, borderRadius: 2 }} />
              <div className="flex flex-col gap-4">
                <div className="skeleton" style={{ height: 32, width: '60%' }} />
                <div className="skeleton" style={{ height: 20, width: '40%' }} />
                <div className="skeleton" style={{ height: 48, width: '30%' }} />
              </div>
            </div>
          </div>
        ) : error || !product ? (
          <div className="card no-tilt p-16 text-center stamp-border bg-surface">
            <h1 className="font-display text-5xl mb-4 text-hp-color">ITEM NOT FOUND</h1>
            <p style={{ color: 'var(--text-muted)' }} className="mb-6 font-mono-stack">This item may have been sold or removed.</p>
            <button onClick={onClose} className="btn-secondary">Close</button>
          </div>
        ) : (
          <div className="card no-tilt bg-surface">
            <div className="grid md:grid-cols-2 gap-0 overflow-hidden">
              {/* Image Section */}
              <div className="p-8" style={{ background: 'var(--kraft-light)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="cardbox overflow-hidden shadow-md w-full max-w-sm" style={{ aspectRatio: '3/4', background: 'var(--ink-surface)', padding: '0.5rem' }}>
                  <div className="w-full h-full border border-ink-border relative" style={{ boxShadow: 'inset 0 0 10px rgba(0,0,0,0.05)' }}>
                    {product.image_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={product.image_url} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'contain', padding: '1rem' }} />
                    ) : (
                      <div className="flex items-center justify-center h-full w-full text-5xl opacity-20">🃏</div>
                    )}
                  </div>
                </div>
              </div>

              {/* Details Section */}
              <div className="p-8 flex flex-col h-full bg-surface" style={{ borderLeft: '4px solid var(--kraft-dark)' }}>
                <div>
                  <nav className="text-[10px] mb-2 font-mono-stack uppercase" style={{ color: 'var(--text-muted)' }}>
                    {product.tcg} / {product.category}
                  </nav>
                  
                  {product.set_name && (
                    <p className="text-xs mb-1 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                      {product.set_code ? `[${product.set_code}] ` : ''}{product.set_name}
                    </p>
                  )}
                  <h1 className="font-display text-4xl md:text-5xl" style={{ color: 'var(--ink-deep)', lineHeight: 1 }}>
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

                <div className="flex flex-wrap gap-2 mt-4 block">
                  <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--text-muted)', border: '1px solid var(--kraft-dark)' }}>
                    {product.language?.toUpperCase() || 'EN'}
                  </span>
                  {product.condition && <span className={`badge badge-${product.condition.toLowerCase()} border`}>{product.condition}</span>}
                  {product.promo_type && product.promo_type !== 'none' && (
                    <span className="badge" style={{ background: 'var(--hp-color)', color: '#fff', border: 'none' }}>
                      {resolveLabel(product.promo_type, {})}
                    </span>
                  )}
                  {product.foil_treatment !== 'non_foil' && FOIL_LABELS[product.foil_treatment] && (
                    <span className="badge badge-foil">✦ {FOIL_LABELS[product.foil_treatment]}</span>
                  )}
                  {product.card_treatment !== 'normal' && TREATMENT_LABELS[product.card_treatment] && (
                    <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--text-muted)', borderColor: 'var(--kraft-dark)' }}>
                      {TREATMENT_LABELS[product.card_treatment]}
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

                <hr className="divider w-full my-6" />

                {/* Price */}
                <div className="flex items-baseline gap-3 mb-6">
                  <span className="price text-5xl tracking-tighter">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
                  <span className="text-xs font-mono-stack font-bold px-2 py-1 rounded-sm" style={{ background: product.stock === 0 ? 'var(--hp-color)' : 'var(--nm-color)', color: '#fff' }}>
                    {product.stock === 0 ? 'OUT OF STOCK' : `${product.stock} IN STOCK`}
                  </span>
                </div>

                {/* Description / Rules Text */}
                <div className="flex-1 min-h-[100px]">
                  {(product.oracle_text || product.description) ? (
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

                {/* Actions */}
                <div className="mt-8">
                  <button
                    id={`modal-add-to-cart-${product.id}`}
                    onClick={handleAddCart}
                    disabled={product.stock === 0}
                    className="btn-primary w-full text-lg py-4 shadow-sm"
                    style={{ textAlign: 'center', opacity: product.stock === 0 ? 0.4 : 1, cursor: product.stock === 0 ? 'not-allowed' : 'pointer' }}
                  >
                    {added ? '✓ ADDED TO CART' : product.stock === 0 ? 'SOLD OUT' : 'ADD TO CART'}
                  </button>
                  <p className="text-[10px] text-center mt-3 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                    <Link href={`/product/${product.id}`} className="hover:text-gold" onClick={onClose}>View full page →</Link>
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
