'use client';

import { useEffect, useState } from 'react';
import { fetchProduct } from '@/lib/api';
import { Product } from '@/lib/types';
import Modal from './ui/Modal';
import { useLanguage } from '@/context/LanguageContext';
import ProductDetails from './ProductDetails';

interface ProductModalProps {
  productId: string;
  initialProduct?: Product;
  onClose: () => void;
}

export default function ProductModal({ productId, initialProduct, onClose }: ProductModalProps) {
  const [product, setProduct] = useState<Product | null>(initialProduct || null);
  const [loading, setLoading] = useState(!initialProduct);
  const [error, setError] = useState(false);
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

  return (
    <Modal 
      isOpen={true} 
      onClose={onClose} 
      maxWidth="max-w-5xl" 
      showHeader={true}
      title={
        <div className="flex items-center gap-3">
          <span>{product?.name || t('pages.common.status.loading', 'Loading...')}</span>
          {product?.scryfall_id && (
            <a 
              href={`https://scryfall.com/card/${product.scryfall_id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-[9px] lowercase tracking-tight opacity-20 hover:opacity-100 hover:text-gold transition-all cursor-pointer no-underline"
              title={t('pages.product.details.visit_scryfall', 'Visit Scryfall Page')}
              onClick={(e) => e.stopPropagation()}
            >
              scryfall id: {product.scryfall_id}
            </a>
          )}
        </div>
      }
      containerClassName="bg-bg-surface border border-border-main shadow-2xl overflow-hidden"
      bodyClassName="md:overflow-hidden"
    >
      <div className="h-full">
        {loading ? (
          <div className="p-8">
            <div className="grid md:grid-cols-2 gap-8">
              <div className="skeleton aspect-[3/4] w-full rounded-sm" />
              <div className="flex flex-col gap-6">
                <div className="space-y-2">
                  <div className="skeleton h-4 w-24" />
                  <div className="skeleton h-10 w-[80%]" />
                  <div className="skeleton h-4 w-48" />
                </div>
                <div className="grid grid-cols-4 gap-4">
                  <div className="skeleton h-12" />
                  <div className="skeleton h-12" />
                  <div className="skeleton h-12" />
                  <div className="skeleton h-12" />
                </div>
                <div className="skeleton h-16 w-[40%]" />
                <div className="skeleton h-32 w-full" />
                <div className="skeleton h-14 w-full mt-auto" />
              </div>
            </div>
          </div>
        ) : error || !product ? (
          <div className="p-16 text-center stamp-border">
            <h1 className="font-display text-fluid-h1 mb-4 text-hp-color uppercase">{t('pages.product.details.not_found', 'ITEM NOT FOUND')}</h1>
            <p className="text-text-muted mb-6 font-mono-stack">{t('pages.product.details.not_found_desc', 'This item may have been sold or removed.')}</p>
            <button onClick={onClose} className="btn-secondary px-8">{t('pages.common.buttons.close', 'Close')}</button>
          </div>
        ) : (
          <ProductDetails 
            product={product} 
            idPrefix="modal" 
            showViewFullPage={true} 
            onClose={onClose} 
            className="md:h-[min(800px,calc(100vh-10rem))]"
          />
        )}
      </div>
    </Modal>
  );
}
