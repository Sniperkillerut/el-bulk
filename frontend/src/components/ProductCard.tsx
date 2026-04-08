'use client';

import { usePathname, useSearchParams } from 'next/navigation';
import { Product } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import CardImage from './CardImage';
import CardInfo from './cards/CardInfo';
import CardBadgeList from './cards/CardBadgeList';
import { openProductModal } from './ProductModalManager';
import { useLanguage } from '@/context/LanguageContext';
import { CategoryIcon } from './CategoryIcon';
import { HotBadge, NewBadge } from './Badges';


interface ProductCardProps {
  product: Product;
}

export default function ProductCard({ product }: ProductCardProps) {
  const { addItem } = useCart();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { t } = useLanguage();
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
      <a href={href} onClick={handleOpenModal} className="thumb-hover-wrap no-underline relative block">
        <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} foilTreatment={product.foil_treatment} />
        
        {/* Product Badges (Hot/New) */}
        <div className="absolute top-2 right-2 z-10 flex flex-col gap-1">
          {product.is_hot && <HotBadge />}
          {product.is_new && <NewBadge />}
        </div>

        {/* Floating Categories */}
        {product.categories && product.categories.length > 0 && (
          <div className="absolute top-2 left-2 z-10 flex flex-col gap-1 pointer-events-none">
            {product.categories.map(c => (
              <span key={c.id} className="badge shadow-lg backdrop-blur-md" 
                    style={{ 
                      background: c.bg_color ? `${c.bg_color}E6` : 'rgba(var(--accent-primary-rgb, 184, 134, 11), 0.85)', 
                      color: c.text_color || 'var(--text-on-accent, #fff)', 
                      borderColor: 'rgba(255,255,255,0.2)',
                      fontSize: '0.6rem',
                      padding: '0.15rem 0.5rem',
                      letterSpacing: '0.05em',
                      fontWeight: 'bold',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.25rem'
                    }}>
                <CategoryIcon icon={c.icon} />
                {c.name.toUpperCase()}
              </span>
            ))}
          </div>
        )}
      </a>

      <div className="flex flex-col flex-1 gap-2" style={{ padding: 'var(--padding-card)' }}>
        <CardBadgeList 
          condition={product.condition}
          foil={product.foil_treatment}
          treatment={product.card_treatment}
          textless={product.textless}
          fullArt={product.full_art}
        />

        <a href={href} onClick={handleOpenModal} className="no-underline">
          <CardInfo name={product.name} setName={product.set_name} setCode={product.set_code} rarity={product.rarity} />
        </a>

        {/* Footer */}
        <div className="mt-auto pt-2 flex flex-col gap-2 border-t border-border-main" data-theme-area="card-footer">
          {displayCartCount > 0 && (
            <div className="flex items-center gap-1.5 text-[10px] font-mono tracking-wider mb-0.5 text-text-secondary opacity-80">
              <span className="text-accent-primary">●</span>
              {displayCartCount === 1 
                ? t('pages.product.cart_users_has', '{count} OTHER USER HAS THIS IN THEIR CART', { count: displayCartCount })
                : t('pages.product.cart_users_have', '{count} OTHER USERS HAVE THIS IN THEIR CART', { count: displayCartCount })}
            </div>
          )}
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <span className="price text-base flex flex-col sm:block">
              <span className="leading-tight">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}</span>
              <span className="text-[10px] sm:text-xs text-text-muted sm:ml-1 align-baseline uppercase font-mono-stack">{t('pages.common.currency.cop', 'COP')}</span>
            </span>
            <div className="flex items-center justify-between sm:justify-end gap-2 w-full sm:w-auto">
              <span className="text-xs text-text-muted font-mono sm:hidden">
                {outOfStock ? '—' : `×${product.stock}`}
              </span>
              <button
                id={`add-to-cart-${product.id}`}
                onClick={() => !outOfStock && addItem(product)}
                disabled={outOfStock}
                className="btn-primary text-[0.8rem] px-4 py-2 sm:px-[0.8rem] sm:py-[0.3rem] flex-1 sm:flex-initial"
                style={{ opacity: outOfStock ? 0.4 : 1, cursor: outOfStock ? 'not-allowed' : 'pointer' }}
              >
                {outOfStock ? t('pages.common.status.sold_out', 'SOLD OUT') : t('pages.common.buttons.add', 'ADD')}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
