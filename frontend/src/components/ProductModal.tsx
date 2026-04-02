'use client';

import { useEffect, useState } from 'react';
import { fetchProduct } from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import Link from 'next/link';
import Modal from './ui/Modal';
import CardImage from './CardImage';
import DeckContents from './DeckContents';
import { useLanguage } from '@/context/LanguageContext';

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
  const { t } = useLanguage();
  
  const [prevId, setPrevId] = useState(productId);
  const [prevInitialProduct, setPrevInitialProduct] = useState(initialProduct);

  // Sync: Reset state if productId or initialProduct changes (Derived state pattern)
  if (productId !== prevId || initialProduct !== prevInitialProduct) {
    setPrevId(productId);
    setPrevInitialProduct(initialProduct);
    if (initialProduct) {
      setProduct(initialProduct);
      setLoading(false);
      setError(false);
    } else {
      setProduct(null);
      setLoading(true);
      setError(false);
    }
  }

  useEffect(() => {
    if (initialProduct || !productId) return;

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

  return (
    <Modal 
      isOpen={true} 
      onClose={onClose} 
      maxWidth="max-w-5xl" 
      showHeader={false}
      containerClassName="bg-transparent border-none shadow-none overflow-visible"
    >
      <div className="flex justify-end mb-2">
        <button onClick={onClose} className="text-white hover:text-gold text-3xl leading-none transition-colors">×</button>
      </div>

      {loading ? (
        <div className="card bg-surface p-12">
          <div className="grid md:grid-cols-2 gap-8">
            <div className="skeleton h-[400px] rounded-[2px]" />
            <div className="flex flex-col gap-4">
              <div className="skeleton h-8 w-[60%]" />
              <div className="skeleton h-5 w-[40%]" />
              <div className="skeleton h-12 w-[30%]" />
            </div>
          </div>
        </div>
      ) : error || !product ? (
        <div className="card p-16 text-center stamp-border bg-surface">
          <h1 className="font-display text-3xl mb-4 text-hp-color uppercase">{t('pages.product.details.not_found', 'ITEM NOT FOUND')}</h1>
          <p className="text-text-muted mb-6 font-mono-stack">{t('pages.product.details.not_found_desc', 'This item may have been sold or removed.')}</p>
          <button onClick={onClose} className="btn-secondary px-8">{t('pages.common.buttons.close', 'Close')}</button>
        </div>
      ) : (
        <div className="card bg-surface overflow-hidden">
          <div className="grid md:grid-cols-2 gap-0">
            {/* Image Section */}
            <div className="p-8 bg-bg-page flex items-center justify-center">
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
            <div className="p-8 flex flex-col h-full bg-bg-surface border-l-4 border-border-main">
              <div>
                <nav className="text-[10px] mb-2 font-mono-stack uppercase text-text-muted">
                  {t(`tcg.${product.tcg}`, product.tcg.toUpperCase())} / {t(`pages.inventory.category.${product.category}`, product.category.toUpperCase())}
                </nav>
                
                {product.set_name && (
                  <p className="text-xs mb-1 font-mono-stack text-text-muted">
                    {product.set_code ? `[${product.set_code}] ` : ''}{product.set_name}
                  </p>
                )}
                <h1 className="font-display text-3xl md:text-4xl text-text-main leading-none">
                  {product.name}
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

              <div className="flex flex-wrap gap-2 mt-4">
                <span className="badge bg-bg-surface text-text-muted border border-border-main">
                  {t(`pages.inventory.grid.sort.language.${product.language}`, product.language?.toUpperCase() || 'EN')}
                </span>
                {product.condition && (
                  <span className={`badge badge-${product.condition.toLowerCase()} border`}>
                    {t(`pages.product.condition.${product.condition.toLowerCase()}`, product.condition)}
                  </span>
                )}
                {product.promo_type && product.promo_type !== 'none' && (
                  <span className="badge bg-status-hp text-white border-none">
                    {resolveLabel(product.promo_type, {})}
                  </span>
                )}
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
                <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 py-3 border-t border-b border-dashed border-border-main">
                  <div className="text-center">
                    <p className="text-[10px] font-bold text-text-muted uppercase">{t('pages.common.labels.identity', 'Identity')}</p>
                    <p className="text-sm font-mono-stack">{product.color_identity || 'C'}</p>
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
              )}

              <hr className="divider w-full my-6 border-border-main" />

              {/* Price */}
              <div className="flex items-baseline gap-3 mb-6 flex-wrap">
                <span className="price text-5xl tracking-tighter text-text-main leading-none font-display">${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
                <div className="flex flex-col gap-1">
                  <span className={`text-xs font-mono-stack font-bold px-2 py-1 rounded-sm w-fit text-white ${product.stock === 0 ? 'bg-status-hp' : 'bg-status-nm'}`}>
                    {product.stock === 0 
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
              <div className="flex-1 min-h-[100px]">
                {(product.oracle_text || product.description) ? (
                  <div className="flex flex-col gap-4">
                    {product.oracle_text && (
                      <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm bg-bg-page/40 text-text-main border border-dashed border-border-main">
                        {product.oracle_text}
                      </div>
                    )}
                    {!product.oracle_text && product.description && (
                       <div className="text-sm leading-relaxed whitespace-pre-wrap font-mono-stack p-4 rounded-sm bg-bg-page/40 text-text-secondary border border-dashed border-border-main">
                         {product.description}
                       </div>
                    )}
                  </div>
                ) : (
                  <div className="text-sm italic text-text-muted">
                    {t('pages.product.details.no_info', 'No additional information available.')}
                  </div>
                )}
              </div>

              {/* Deck Cards Grid for Store Exclusives */}
              {product.category === 'store_exclusives' && product.deck_cards && product.deck_cards.length > 0 && (
                <DeckContents cards={product.deck_cards} tcg={product.tcg} className="border-t border-kraft-dark pt-6" />
              )}

              {/* Actions */}
              <div className="mt-8">
                <button
                  id={`modal-add-to-cart-${product.id}`}
                  onClick={handleAddCart}
                  disabled={product.stock === 0}
                  className={`btn-primary w-full text-lg py-4 shadow-sm transition-all ${product.stock === 0 ? 'opacity-40 cursor-not-allowed' : 'opacity-100 cursor-pointer'}`}
                >
                  {added 
                    ? `✓ ${t('pages.common.buttons.added_to_cart', 'ADDED TO CART')}` 
                    : product.stock === 0 
                      ? t('pages.common.status.sold_out', 'SOLD OUT') 
                      : t('pages.common.buttons.add_to_cart', 'ADD TO CART')}
                </button>
                <p className="text-[10px] text-center mt-3 font-mono-stack text-text-muted">
                  <Link href={`/product/${product.id}`} className="hover:text-gold transition-colors" onClick={onClose}>
                    {t('pages.product.details.view_full_page', 'View full page')} →
                  </Link>
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </Modal>
  );
}
