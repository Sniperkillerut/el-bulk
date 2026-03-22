'use client';

import { Suspense, useEffect, useState } from 'react';
import ProductModal from './ProductModal';
import { Product } from '@/lib/types';

export const openProductModal = (productOrId: string | Product) => {
  window.dispatchEvent(new CustomEvent('openProductModal', { detail: productOrId }));
  const id = typeof productOrId === 'string' ? productOrId : productOrId.id;
  const url = new URL(window.location.href);
  url.searchParams.set('productId', id);
  window.history.pushState({}, '', url.toString());
};

function ProductModalManagerInner() {
  const [modalData, setModalData] = useState<{ id: string, product?: Product } | null>(null);

  useEffect(() => {
    // Check initial load
    const params = new URLSearchParams(window.location.search);
    if (params.has('productId')) {
      setModalData({ id: params.get('productId')! });
    }

    const handler = (e: Event) => {
      const customEvent = e as CustomEvent<string | Product>;
      const detail = customEvent.detail;
      if (typeof detail === 'string') {
        setModalData({ id: detail });
      } else {
        setModalData({ id: detail.id, product: detail });
      }
    };

    const popstateHandler = () => {
      const p = new URLSearchParams(window.location.search);
      if (p.has('productId')) {
        setModalData({ id: p.get('productId')! });
      } else {
        setModalData(null);
      }
    };

    window.addEventListener('openProductModal', handler);
    window.addEventListener('popstate', popstateHandler);
    return () => {
      window.removeEventListener('openProductModal', handler);
      window.removeEventListener('popstate', popstateHandler);
    };
  }, []);

  if (!modalData) return null;

  const handleClose = () => {
    setModalData(null);
    const url = new URL(window.location.href);
    url.searchParams.delete('productId');
    window.history.pushState({}, '', url.toString());
  };

  return <ProductModal productId={modalData.id} initialProduct={modalData.product} onClose={handleClose} />;
}

export default function ProductModalManager() {
  return (
    <Suspense fallback={null}>
      <ProductModalManagerInner />
    </Suspense>
  );
}
