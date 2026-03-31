'use client';

import { usePathname, useSearchParams } from 'next/navigation';
import { Product } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import CardImage from './CardImage';
import { openProductModal } from './ProductModalManager';
import CardBadgeList from './cards/CardBadgeList';
import CardInfo from './cards/CardInfo';

interface ProductCardProps {
  product: Product;
}

export default function ProductCard({ product }: ProductCardProps) {
  const { addItem } = useCart();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const outOfStock = product.stock === 0;

  const displayCartCount = product.cart_count || 0;

  let href = `/product/${product.id}`;
  if (pathname && searchParams) {
    const params = new URLSearchParams(searchParams.toString());
    params.set('productId', product.id);
    href = `${pathname}?${params.toString()}`;
  }

  const handleOpenModal = (e: React.MouseEvent) => {
    e.preventDefault();
    openProductModal(product);
  };

  return (
    <div className="card flex flex-col overflow-hidden animate-fade-up">
      {/* Image area */}
      <a href={href} onClick={handleOpenModal} style={{ textDecoration: 'none' }} className="thumb-hover-wrap">
        <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} foilTreatment={product.foil_treatment} />
      </a>

      <div className="p-3 flex flex-col flex-1 gap-2">
        <CardBadgeList 
          condition={product.condition}
          foil={product.foil_treatment}
          treatment={product.card_treatment}
          textless={product.textless}
          fullArt={product.full_art}
          categories={product.categories}
        />

        <a href={href} onClick={handleOpenModal} style={{ textDecoration: 'none' }}>
          <CardInfo name={product.name} setName={product.set_name} setCode={product.set_code} />
        </a>

        {/* Footer */}
        <div className="mt-auto pt-2 flex flex-col gap-2" style={{ borderTop: '1px solid var(--ink-border)' }}>
          {displayCartCount > 0 && (
            <div className="flex items-center gap-1.5 text-[10px] font-mono tracking-wider mb-0.5" style={{ color: 'var(--text-secondary)', opacity: 0.8 }}>
              <span style={{ color: 'var(--gold)' }}>●</span>
              {displayCartCount} {displayCartCount === 1 ? 'OTHER USER HAS' : 'OTHER USERS HAVE'} THIS IN THEIR CART
            </div>
          )}
          
          <div className="flex items-center justify-between">
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
    </div>
  );
}
