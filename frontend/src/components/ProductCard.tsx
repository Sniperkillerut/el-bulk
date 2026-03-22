'use client';

import Link from 'next/link';
import { Product, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import CardImage from './CardImage';

interface ProductCardProps {
  product: Product;
}

function ConditionBadge({ condition }: { condition?: string }) {
  if (!condition) return null;
  const cls = `badge badge-${condition.toLowerCase()}`;
  return <span className={cls}>{condition}</span>;
}

function FoilBadge({ foil }: { foil: string }) {
  if (foil === 'non_foil') return null;
  return <span className="badge badge-foil">✦ {FOIL_LABELS[foil as keyof typeof FOIL_LABELS]}</span>;
}

export default function ProductCard({ product }: ProductCardProps) {
  const { addItem } = useCart();
  const outOfStock = product.stock === 0;

  return (
    <div className="card flex flex-col overflow-hidden animate-fade-up">
      {/* Image area */}
      <Link href={`/product/${product.id}`} style={{ textDecoration: 'none' }}>
        <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} height={160} />
      </Link>

      <div className="p-3 flex flex-col flex-1 gap-2">
        {/* Badges row */}
        <div className="flex flex-wrap gap-1">
          <ConditionBadge condition={product.condition} />
          <FoilBadge foil={product.foil_treatment} />
          {product.card_treatment !== 'normal' && (
            <span className="badge" style={{ background: 'rgba(100,130,200,0.12)', color: '#8ba4d0', border: '1px solid rgba(100,130,200,0.25)' }}>
              {TREATMENT_LABELS[product.card_treatment]}
            </span>
          )}
          {product.featured && <span className="featured-star">★</span>}
        </div>

        {/* Name */}
        <Link href={`/product/${product.id}`} style={{ textDecoration: 'none' }}>
          <h3 className="text-sm font-semibold leading-snug hover:text-gold transition-colors line-clamp-2"
            style={{ color: 'var(--text-primary)' }}>
            {product.name}
          </h3>
        </Link>

        {/* Set */}
        {product.set_name && (
          <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
            {product.set_code ? `[${product.set_code}]` : ''} {product.set_name}
          </p>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between mt-auto pt-2" style={{ borderTop: '1px solid var(--ink-border)' }}>
          <span className="price text-base">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
          <div className="flex items-center gap-2">
            <span className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono' }}>
              {outOfStock ? '—' : `×${product.stock}`}
            </span>
            <button
              id={`add-to-cart-${product.id}`}
              onClick={() => !outOfStock && addItem(product)}
              disabled={outOfStock}
              className="btn-primary"
              style={{ fontSize: '0.8rem', padding: '0.3rem 0.8rem', opacity: outOfStock ? 0.4 : 1, cursor: outOfStock ? 'not-allowed' : 'pointer' }}
            >
              {outOfStock ? 'SOLD OUT' : 'ADD'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
