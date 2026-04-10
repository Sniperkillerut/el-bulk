'use client';

import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';
import ProductDetails from '@/components/ProductDetails';
import { Product } from '@/lib/types';

interface ProductDetailClientProps {
  product: Product | null;
  error?: boolean;
}

export default function ProductDetailClient({ product, error }: ProductDetailClientProps) {
  const { t } = useLanguage();

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
        <Link href="/" className="hover:text-text-primary transition-colors" style={{ textDecoration: 'none' }}>{t('pages.common.breadcrumb.home', 'Home')}</Link>
        {' / '}
        <Link href={`/${product.tcg}/${product.category}`} className="hover:text-text-primary transition-colors uppercase" style={{ textDecoration: 'none' }}>
          {product.tcg} {product.category}
        </Link>
        {' / '}
        <span style={{ color: 'var(--text-primary)' }}>{product.name}</span>
      </nav>

      <div className="cardbox overflow-hidden bg-surface shadow-xl">
        {/* Step 3.1: View Transition Name for the detail image matches the card */}
        <div style={{ viewTransitionName: `card-image-${product.id}` } as React.CSSProperties}>
           <ProductDetails product={product} idPrefix="detail" />
        </div>
      </div>
    </div>
  );
}
