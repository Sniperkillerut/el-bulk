'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { fetchProduct } from '@/lib/api';
import { Product } from '@/lib/types';
import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';
import ProductDetails from '@/components/ProductDetails';

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const { t } = useLanguage();

  const [prevId, setPrevId] = useState(id);

  // Derived state to reset loading when ID changes, avoiding cascading render warning in useEffect
  if (id !== prevId) {
    setPrevId(id);
    setLoading(true);
    setProduct(null);
    setError(false);
  }

  useEffect(() => {
    if (!id) return;

    fetchProduct(id)
      .then(p => {
        setProduct(p);
        setError(false);
      })
      .catch(() => {
        // Error is already logged to the server via the API client's logAndThrow.
        // We set local state to show the UI without triggering a red console error.
        setError(true);
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return (
    <div className="max-w-5xl mx-auto px-4 py-12">
      <div className="grid md:grid-cols-2 gap-8">
        <div className="skeleton h-[500px] rounded-[2px]" />
        <div className="flex flex-col gap-4">
          <div className="skeleton h-10 w-[60%]" />
          <div className="skeleton h-6 w-[40%]" />
          <div className="skeleton h-16 w-[30%]" />
          <div className="skeleton h-32 w-full" />
        </div>
      </div>
    </div>
  );

  if (error || !product) return (
    <div className="max-w-4xl mx-auto px-4 py-16 text-center stamp-border mt-12 bg-surface p-12">
      <div role="heading" aria-level={1} className="font-display text-3xl mb-4 text-hp-color uppercase">{t('pages.product.details.not_found', 'ITEM NOT FOUND')}</div>
      <p style={{ color: 'var(--text-muted)' }} className="mb-6 font-mono-stack">{t('pages.product.details.not_found_desc', 'This item may have been sold or removed.')}</p>
      <Link href="/" className="btn-secondary">← {t('pages.common.buttons.back_to_home', 'Back to Shoebox')}</Link>
    </div>
  );

  return (
    <div className="max-w-6xl mx-auto px-4 py-12">
      {/* Breadcrumb */}
      <nav className="text-xs mb-6 font-mono-stack" style={{ color: 'var(--text-muted)' }}>
        <Link href="/" className="hover:text-text-primary transition-colors" style={{ textDecoration: 'none' }}>Home</Link>
        {' / '}
        <Link href={`/${product.tcg}/${product.category}`} className="hover:text-text-primary transition-colors uppercase" style={{ textDecoration: 'none' }}>
          {product.tcg} {product.category}
        </Link>
        {' / '}
        <span style={{ color: 'var(--text-primary)' }}>{product.name}</span>
      </nav>

      <div className="cardbox overflow-hidden bg-surface shadow-xl">
        <ProductDetails product={product} idPrefix="detail" />
      </div>
    </div>
  );
}
