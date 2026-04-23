'use client';

import Link from 'next/link';
import ProductCard from '@/components/ProductCard';
import { useLanguage } from '@/context/LanguageContext';
import { ProductListResponse, CustomCategory, Product } from '@/lib/types';

interface CollectionClientProps {
  params: { slug: string };
  searchParams: { [key: string]: string | string[] | undefined };
  category?: CustomCategory;
  products: ProductListResponse;
}

export default function CollectionClient({ 
  params, 
  searchParams, 
  category, 
  products 
}: CollectionClientProps) {
  const { t } = useLanguage();
  const page = parseInt((searchParams.page as string) || '1', 10);
  const totalPages = Math.ceil(products.total / 20);

  if (!category) {
    return (
      <div className="centered-container px-4 py-20 text-center">
        <h1 className="font-display text-5xl mb-4">
          {t('pages.collection.not_found', 'COLLECTION NOT FOUND')}
        </h1>
        <p className="font-mono-stack text-text-muted">
          {t('pages.collection.not_found_desc', 'The requested collection does not exist or has been removed.')}
        </p>
        <Link href="/" className="btn-secondary inline-block mt-8">
          {t('pages.common.buttons.back_home', '← BACK HOME')}
        </Link>
      </div>
    );
  }

  return (
    <div className="centered-container px-4 py-8">
      {/* Dynamic Header */}
      <div className="flex flex-col sm:flex-row items-baseline justify-between border-b-2 border-kraft-dark pb-4 mb-8">
        <div>
          <div className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>
            <Link href="/" className="hover:text-gold transition-colors">{t('pages.common.labels.home', 'HOME')}</Link> / {t('pages.common.labels.collection', 'COLLECTION')}
          </div>
          <h1 className="font-display text-5xl uppercase mt-2" style={{ color: 'var(--ink-deep)' }}>
            {category ? category.name : params.slug}
          </h1>
        </div>
        <span className="text-sm font-mono-stack text-text-secondary">
          {products.total} {products.total === 1 
            ? t('pages.common.labels.product', 'PRODUCT') 
            : t('pages.common.labels.products', 'PRODUCTS')}
        </span>
      </div>

      {products.products.length === 0 ? (
        <div className="stamp-border rounded-sm p-12 text-center" style={{ color: 'var(--text-muted)' }}>
          <p className="font-display text-3xl mb-4">{t('pages.collection.empty', 'EMPTY COLLECTION')}</p>
          <p className="font-mono-stack">{t('pages.collection.empty_desc', 'No products found in this collection.')}</p>
          <Link href="/" className="btn-secondary inline-block mt-8">
            {t('pages.common.buttons.back_home', '← BACK HOME')}
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
          {products.products.map((p: Product) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-4 mt-12 mb-8">
          <Link
            href={`/collection/${params.slug}?page=${Math.max(1, page - 1)}`}
            className={`btn-secondary ${page <= 1 ? 'opacity-50 pointer-events-none' : ''}`}
            style={{ padding: '0.5rem 1rem' }}
          >
            {t('pages.common.pagination.prev', '← PREV')}
          </Link>
          <span className="flex items-center text-sm font-mono-stack" style={{ color: 'var(--text-muted)' }}>
            {page} / {totalPages}
          </span>
          <Link
            href={`/collection/${params.slug}?page=${Math.min(totalPages, page + 1)}`}
            className={`btn-secondary ${page >= totalPages ? 'opacity-50 pointer-events-none' : ''}`}
            style={{ padding: '0.5rem 1rem' }}
          >
            {t('pages.common.pagination.next', 'NEXT →')}
          </Link>
        </div>
      )}
    </div>
  );
}
