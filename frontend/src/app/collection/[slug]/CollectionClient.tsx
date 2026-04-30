'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import ProductCard from '@/components/ProductCard';
import BinderView from '@/components/collection/BinderView';
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
  searchParams: initialSearchParams, 
  category, 
  products 
}: CollectionClientProps) {
  const { t } = useLanguage();
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const isBinderView = searchParams.get('view') === 'binder';
  
  const setView = (view: 'grid' | 'binder') => {
    const params = new URLSearchParams(searchParams.toString());
    params.set('view', view);
    router.push(`${pathname}?${params.toString()}`);
  };
  const page = parseInt((initialSearchParams.page as string) || '1', 10);
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
      <div className="flex flex-col sm:flex-row items-baseline justify-between border-b-2 border-border-main/20 pb-4 mb-8">
        <div>
          <div className="text-xs font-mono-stack text-text-muted">
            <Link href="/" className="hover:text-accent-primary transition-colors">{t('pages.common.labels.home', 'HOME')}</Link> / {t('pages.common.labels.collection', 'COLLECTION')}
          </div>
          <h1 className="font-display text-5xl uppercase mt-2 text-text-main">
            {category ? category.name : params.slug}
          </h1>
        </div>
        
        <div className="flex items-center gap-6 mt-4 sm:mt-0">
          {/* View Toggle */}
          <div className="flex items-center bg-bg-surface p-1 rounded-lg border border-border-main">
            <button 
              onClick={() => setView('grid')}
              className={`px-4 py-1.5 rounded-md text-[10px] font-mono tracking-widest uppercase transition-all ${!isBinderView ? 'bg-accent-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}
            >
              {t('pages.collection.views.grid', 'GRID')}
            </button>
            <button 
              onClick={() => setView('binder')}
              className={`px-4 py-1.5 rounded-md text-[10px] font-mono tracking-widest uppercase transition-all ${isBinderView ? 'bg-accent-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}
            >
              📖 {t('pages.collection.views.binder', 'BINDER')}
            </button>
          </div>

          <span className="text-xs font-mono text-text-muted uppercase tracking-tighter">
            {products.total} {products.total === 1 
              ? t('pages.common.labels.product', 'PRODUCT') 
              : t('pages.common.labels.products', 'PRODUCTS')}
          </span>
        </div>
      </div>

      {products.products.length === 0 ? (
        <div className="card p-24 text-center">
          <p className="font-display text-3xl mb-4 text-text-main">{t('pages.collection.empty', 'EMPTY COLLECTION')}</p>
          <p className="font-mono text-text-muted uppercase tracking-widest">{t('pages.collection.empty_desc', 'No products found in this collection.')}</p>
          <Link href="/" className="btn-secondary inline-block mt-8">
            {t('pages.common.buttons.back_home', '← BACK HOME')}
          </Link>
        </div>
      ) : isBinderView ? (
        <BinderView products={products.products} setName={category.name} />
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
          {products.products.map((p: Product) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {!isBinderView && totalPages > 1 && (
        <div className="flex justify-center gap-4 mt-12 mb-8">
          <Link
            href={`/collection/${params.slug}?page=${Math.max(1, page - 1)}`}
            className={`btn-secondary ${page <= 1 ? 'opacity-50 pointer-events-none' : ''}`}
            style={{ padding: '0.5rem 1rem' }}
          >
            {t('pages.common.pagination.prev', '← PREV')}
          </Link>
          <span className="flex items-center text-sm font-mono text-text-muted">
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
