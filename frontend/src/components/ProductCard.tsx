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
    <div className="card flex flex-col overflow-hidden animate-fade-up" data-theme-area="product-card">
      {/* Image area */}
      <a href={href} onClick={handleOpenModal} className="thumb-hover-wrap no-underline">
        <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} foilTreatment={product.foil_treatment} />
      </a>

      <div className="flex flex-col flex-1 gap-2" style={{ padding: 'var(--padding-card)' }}>
        <CardBadgeList 
          condition={product.condition}
          foil={product.foil_treatment}
          treatment={product.card_treatment}
          textless={product.textless}
          fullArt={product.full_art}
          categories={product.categories}
        />

        <a href={href} onClick={handleOpenModal} className="no-underline">
          <CardInfo name={product.name} setName={product.set_name} setCode={product.set_code} />
        </a>

        {/* Footer */}
        <div className="mt-auto pt-2 flex flex-col gap-2 border-t border-border-main" data-theme-area="card-footer">
          {displayCartCount > 0 && (
            <div className="flex items-center gap-1.5 text-[10px] font-mono tracking-wider mb-0.5 text-text-secondary opacity-80">
              <span className="text-accent-primary">●</span>
              {displayCartCount} {displayCartCount === 1 ? 'OTHER USER HAS' : 'OTHER USERS HAVE'} THIS IN THEIR CART
            </div>
          )}
          
          <div className="flex items-center justify-between">
            <span className="price text-base">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
            <div className="flex items-center gap-2">
              <span className="text-xs text-text-muted font-mono">
                {outOfStock ? '—' : `×${product.stock}`}
              </span>
              <button
                id={`add-to-cart-${product.id}`}
                onClick={() => !outOfStock && addItem(product)}
                disabled={outOfStock}
                className="btn-primary text-[0.8rem] px-[0.8rem] py-[0.3rem]"
                style={{ opacity: outOfStock ? 0.4 : 1, cursor: outOfStock ? 'not-allowed' : 'pointer' }}
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
