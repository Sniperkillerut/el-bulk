'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Product, FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from './CardImage';
import DeckContents from './DeckContents';
import { HotBadge, NewBadge, CONDITION_MAP } from './Badges';
import SetIcon from './SetIcon';
import LegalityBadge from './LegalityBadge';
import ManaText from './ManaText';
import { filterPromoTags } from '@/lib/mtg-logic';

interface ProductDetailsProps {
  product: Product;
  idPrefix: string;
  showViewFullPage?: boolean;
  onClose?: () => void; // Used by modal to close when "View full page" is clicked
  className?: string;
}

export default function ProductDetails({ product, idPrefix, showViewFullPage, onClose, className = '' }: ProductDetailsProps) {
  const [added, setAdded] = useState(false);
  const { addItem } = useCart();
  const { t } = useLanguage();

  const handleAddCart = () => {
    addItem(product);
    setAdded(true);
    setTimeout(() => setAdded(false), 1500);
  };

  const renderManaIcons = (identity?: string) => {
    if (!identity || identity === 'C') return <i className="ms ms-c ms-cost ms-shadow text-[1.1rem]" />;
    const colors = identity.split(',').map(c => c.trim().toLowerCase());
    return (
      <div className="flex gap-1 items-center justify-center">
        {colors.map(c => (
          <i key={c} className={`ms ms-${c} ms-cost ms-shadow text-[1.1rem]`} />
        ))}
      </div>
    );
  };

  const outOfStock = product.stock === 0;
  const conditionClass = product.condition ? `badge-${product.condition.toLowerCase()}` : '';

  return (
    <div className={`grid md:grid-cols-2 gap-0 min-h-0 ${className}`}>
      {/* Image Section */}
      <div className="p-8 bg-bg-page flex items-center justify-center md:h-full md:sticky md:top-0">
        <div className="cardbox overflow-hidden shadow-md w-full max-w-sm aspect-[3/4] bg-bg-surface p-2">
          <div className="w-full h-full border border-border-main relative shadow-[inset_0_0_10px_rgba(0,0,0,0.05)]">
            <CardImage 
              imageUrl={product.image_url} 
              name={product.name} 
              tcg={product.tcg}
              foilTreatment={product.foil_treatment}
              enableModal={true}
            />
          </div>
        </div>
      </div>

      {/* Details Section */}
      <div className="p-8 flex flex-col h-full bg-bg-surface border-l-4 border-border-main md:overflow-y-auto custom-scrollbar">
        <div>
          <nav className="text-[10px] mb-2 font-mono-stack uppercase text-text-muted">
            {t(`tcg.${product.tcg}`, product.tcg.toUpperCase())} / {t(`pages.inventory.category.${product.category}`, product.category.toUpperCase())}
          </nav>
          
          {product.set_name && (
            <div className="flex items-center gap-2 mb-1">
              {product.set_code && <SetIcon setCode={product.set_code} rarity={product.rarity} size="sm" />}
              <p className="text-xs font-mono-stack text-text-muted">
                {product.set_name}
              </p>
            </div>
          )}
          <h1 className="font-display text-3xl md:text-4xl text-text-main leading-none flex items-center gap-2">
            {product.name}
            {product.is_hot && <HotBadge />}
            {product.is_new && <NewBadge />}
          </h1>
          {product.type_line && (
            <p className="text-xs mt-2 font-mono-stack text-text-secondary font-bold">
              {product.type_line}
            </p>
          )}
          {product.artist && (
            <p className="text-[10px] mt-1 font-mono-stack text-text-muted">
              {t('pages.common.labels.art_by', 'Art by')} {product.artist} {product.collector_number ? `(#${product.collector_number})` : ''}
            </p>
          )}
        </div>

        {/* Badges */}
        <div className="flex flex-wrap gap-2 mt-4">
          <span className="badge bg-bg-surface text-text-muted border border-border-main">
            {t(`pages.inventory.grid.sort.language.${product.language}`, product.language?.toUpperCase() || 'EN')}
          </span>
          {product.condition && (
            <span className={`badge ${conditionClass} border !px-1.5 !py-0 !text-[10px] !font-black !leading-tight`}>
              {CONDITION_MAP[product.condition.toLowerCase()] || product.condition}
            </span>
          )}
          {filterPromoTags(product.promo_type, product.foil_treatment, product.card_treatment).map(t => (
            <span key={t} className="badge bg-ink-surface text-text-secondary border border-kraft-dark uppercase">
              {resolveLabel(t, {})}
            </span>
          ))}
          {product.foil_treatment !== 'non_foil' && (
            <span className="badge badge-foil">✦ {t(`pages.product.finish.${product.foil_treatment}`, FOIL_LABELS[product.foil_treatment] || product.foil_treatment)}</span>
          )}
          {product.card_treatment !== 'normal' && (
            <span className="badge bg-ink-surface text-text-secondary border border-kraft-dark">
              {t(`pages.product.version.${product.card_treatment}`, TREATMENT_LABELS[product.card_treatment] || product.card_treatment)}
            </span>
          )}
          {product.textless && (
            <span className="badge bg-status-hp/10 text-status-hp border border-status-hp">
              {t('pages.product.details.textless', 'TEXTLESS')}
            </span>
          )}
          {product.full_art && product.card_treatment !== 'full_art' && (
            <span className="badge bg-status-nm/10 text-status-nm border border-status-nm">
              {t('pages.product.details.full_art', 'FULL ART')}
            </span>
          )}
        </div>

        {product.tcg === 'mtg' && product.category === 'singles' && (
          <>
            <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 py-3 border-t border-dashed border-border-main">
              <div className="text-center">
                <p className="text-[10px] font-bold text-text-muted uppercase mb-1">{t('pages.common.labels.identity', 'Identity')}</p>
                {renderManaIcons(product.color_identity)}
              </div>
              <div className="text-center border-l md:border-l border-dashed border-border-main px-2">
                <p className="text-[10px] font-bold text-text-muted uppercase">{t('pages.common.labels.rarity', 'Rarity')}</p>
                <p className="text-sm font-mono-stack capitalize">
                  {t(`pages.inventory.grid.sort.rarity.${product.rarity?.toLowerCase() || 'common'}`, product.rarity || 'Common')}
                </p>
              </div>
              <div className="text-center border-l border-dashed border-border-main px-2">
                <p className="text-[10px] font-bold text-text-muted uppercase">{t('pages.common.labels.art_var', 'Art Var.')}</p>
                <p className="text-sm font-mono-stack truncate">
                  {product.art_variation ? t(`pages.product.art_variation.${product.art_variation.toLowerCase().replace(' ', '_')}`, product.art_variation) : t('pages.common.status.normal', 'Normal')}
                </p>
              </div>
              <div className="text-center border-l border-dashed border-border-main px-2">
                <p className="text-[10px] font-bold text-text-muted uppercase">{t('pages.common.labels.cmc', 'CMC')}</p>
                <p className="text-sm font-mono-stack">{product.cmc ?? 0}</p>
              </div>
            </div>

            {product.legalities && Object.values(product.legalities).some(status => status !== 'not_legal') && (
              <div className="mt-2 py-3 border-t border-b border-dashed border-border-main">
                <p className="text-[10px] font-bold text-text-muted uppercase mb-2 tracking-widest">{t('pages.product.details.legalities', 'FORMAT LEGALITY')}</p>
                <div className="flex flex-wrap gap-1.5">
                  {Object.entries(product.legalities).map(([fmt, status]) => (
                    status !== 'not_legal' && (
                      <LegalityBadge key={fmt} format={fmt} status={status as string} />
                    )
                  ))}
                </div>
              </div>
            )}
          </>
        )}

        <hr className="divider w-full my-6 border-border-main" />

        {/* Price */}
        <div className="flex items-baseline gap-3 mb-6 flex-wrap">
          <span className="price text-5xl tracking-tighter text-text-main leading-none font-display">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} {t('pages.common.currency.cop', 'COP')}</span>
          <div className="flex flex-col gap-1">
            <span className={`text-xs font-mono-stack font-bold px-2 py-1 rounded-sm w-fit text-white ${outOfStock ? 'bg-status-hp' : 'bg-status-nm'}`}>
              {outOfStock 
                ? t('pages.common.status.out_of_stock', 'OUT OF STOCK') 
                : t('pages.product.status.in_stock', '{count} IN STOCK').replace('{count}', product.stock.toString())}
            </span>
            {(product.cart_count ?? 0) > 0 && (
              <span className="text-[10px] font-mono tracking-wider text-gold opacity-90">
                ● {(product.cart_count ?? 0) === 1 
                  ? t('pages.product.cart_users_has', '{count} OTHER USER HAS THIS IN THEIR CART').replace('{count}', (product.cart_count ?? 0).toString())
                  : t('pages.product.cart_users_have', '{count} OTHER USERS HAVE THIS IN THEIR CART').replace('{count}', (product.cart_count ?? 0).toString())}
              </span>
            )}
          </div>
        </div>

        {/* Description / Rules Text */}
        <div className="mb-4 flex-shrink-0">
          {(product.oracle_text || product.description) ? (
            <div className="flex flex-col gap-4">
              {product.oracle_text && (
                <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm bg-bg-page/40 text-text-main border border-dashed border-border-main">
                  <ManaText text={product.oracle_text} />
                </div>
              )}
              {!product.oracle_text && product.description && (
                 <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm bg-bg-page/40 text-text-secondary border border-dashed border-border-main">
                   <ManaText text={product.description} />
                 </div>
              )}
            </div>
          ) : (
            <div className="text-sm italic text-text-muted">
              {t('pages.product.details.no_info', 'No additional information available.')}
            </div>
          )}
        </div>

        {/* Deck Cards Grid */}
        {product.deck_cards && product.deck_cards.length > 0 && (
          <DeckContents cards={product.deck_cards} tcg={product.tcg} className="border-t border-kraft-dark pt-6 mt-6" />
        )}

        {/* Actions */}
        <div className="mt-8">
          <button
            id={`${idPrefix}-add-to-cart-${product.id}`}
            onClick={handleAddCart}
            disabled={outOfStock}
            className={`btn-primary w-full text-base py-3.5 md:text-lg md:py-4 shadow-sm transition-all ${outOfStock ? 'opacity-40 cursor-not-allowed' : 'opacity-100 cursor-pointer'}`}
          >
            {added 
              ? `✓ ${t('pages.common.buttons.added_to_cart', 'ADDED TO CART')}` 
              : outOfStock 
                ? t('pages.common.status.sold_out', 'SOLD OUT') 
                : t('pages.common.buttons.add_to_cart', 'ADD TO CART')}
          </button>
          
          {showViewFullPage && (
            <p className="text-[10px] text-center mt-3 font-mono-stack text-text-muted">
              <Link href={`/product/${product.id}`} className="hover:text-gold transition-colors" onClick={onClose}>
                {t('pages.product.details.view_full_page', 'View full page')} →
              </Link>
            </p>
          )}
        </div>

        {!showViewFullPage && (
          <div className="mt-4 text-center">
            <p className="text-xs font-mono-stack text-text-muted">
              🏪 {t('pages.product.details.store_notice', 'Complete purchase in-store or verify availability at counter.')}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
