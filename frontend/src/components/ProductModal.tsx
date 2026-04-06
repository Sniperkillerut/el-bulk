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
              <div className="skeleton h-20 w-full" />
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
        <div className="card bg-surface overflow-hidden shadow-2xl">
          <ProductDetails 
            product={product} 
            idPrefix="modal" 
            showViewFullPage={true} 
            onClose={onClose} 
          />
        </div>
      )}
    </Modal>
  );
}
