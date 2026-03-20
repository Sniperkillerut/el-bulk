'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchProduct } from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import Link from 'next/link';

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [added, setAdded] = useState(false);
  const { addItem } = useCart();

  useEffect(() => {
    fetchProduct(id)
      .then(setProduct)
      .catch(() => setError(true))
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
        <div className="skeleton" style={{ height: 400, borderRadius: 8 }} />
        <div className="flex flex-col gap-4">
          <div className="skeleton" style={{ height: 32, width: '60%' }} />
          <div className="skeleton" style={{ height: 20, width: '40%' }} />
          <div className="skeleton" style={{ height: 48, width: '30%' }} />
        </div>
      </div>
    </div>
  );

  if (error || !product) return (
    <div className="max-w-4xl mx-auto px-4 py-16 text-center">
      <h1 className="font-display text-5xl mb-4">CARD NOT FOUND</h1>
      <p style={{ color: 'var(--text-muted)' }} className="mb-6">This item may have been sold or removed.</p>
      <Link href="/" className="btn-secondary">← Back to Home</Link>
    </div>
  );

  const outOfStock = product.stock === 0;
  const conditionClass = product.condition ? `badge-${product.condition.toLowerCase()}` : '';

  return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      {/* Breadcrumb */}
      <nav className="text-xs mb-6 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
        <Link href="/" style={{ color: 'var(--text-muted)', textDecoration: 'none' }}>Home</Link>
        {' / '}
        <Link href={`/${product.tcg}/singles`} style={{ color: 'var(--text-muted)', textDecoration: 'none' }}>
          {product.tcg.toUpperCase()} {product.category}
        </Link>
        {' / '}
        <span style={{ color: 'var(--text-primary)' }}>{product.name}</span>
      </nav>

      <div className="grid md:grid-cols-2 gap-10">
        {/* Image */}
        <div>
          <div className="card overflow-hidden" style={{ aspectRatio: '3/4' }}>
            {product.image_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={product.image_url} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
            ) : (
              <div className="product-img-placeholder h-full" style={{ fontSize: '5rem' }}>🃏</div>
            )}
          </div>
        </div>

        {/* Details */}
        <div className="flex flex-col gap-4">
          <div>
            {product.set_name && (
              <p className="text-xs mb-1 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                {product.set_code ? `[${product.set_code}] ` : ''}{product.set_name}
              </p>
            )}
            <h1 className="font-display text-4xl md:text-5xl" style={{ color: 'var(--text-primary)' }}>
              {product.name}
            </h1>
          </div>

          {/* Badges */}
          <div className="flex flex-wrap gap-2">
            {product.condition && <span className={`badge ${conditionClass}`}>{product.condition}</span>}
            {product.foil_treatment !== 'non_foil' && (
              <span className="badge badge-foil">✦ {FOIL_LABELS[product.foil_treatment]}</span>
            )}
            {product.card_treatment !== 'normal' && (
              <span className="badge" style={{ background: 'rgba(100,130,200,0.12)', color: '#8ba4d0', border: '1px solid rgba(100,130,200,0.25)' }}>
                {TREATMENT_LABELS[product.card_treatment]}
              </span>
            )}
          </div>

          <hr className="divider" />

          {/* Price */}
          <div className="flex items-baseline gap-3">
            <span className="price text-4xl">${product.price.toFixed(2)}</span>
            <span className="text-xs font-mono-stack" style={{ color: outOfStock ? 'var(--hp-color)' : 'var(--nm-color)' }}>
              {outOfStock ? 'OUT OF STOCK' : `${product.stock} in stock`}
            </span>
          </div>

          {/* Description */}
          {product.description && (
            <p className="text-sm leading-relaxed" style={{ color: 'var(--text-secondary)' }}>
              {product.description}
            </p>
          )}

          {/* Actions */}
          <div className="flex gap-3 mt-2">
            <button
              id={`detail-add-to-cart-${product.id}`}
              onClick={handleAddCart}
              disabled={outOfStock}
              className="btn-primary flex-1"
              style={{ textAlign: 'center', opacity: outOfStock ? 0.4 : 1, cursor: outOfStock ? 'not-allowed' : 'pointer' }}
            >
              {added ? '✓ ADDED!' : outOfStock ? 'SOLD OUT' : 'ADD TO CART'}
            </button>
          </div>

          <div style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)', borderRadius: 6, padding: '0.75rem 1rem' }}>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>
              🏪 Complete your purchase in-store or contact us to arrange pickup.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
