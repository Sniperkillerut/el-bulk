'use client';

import ProductGrid from '@/components/ProductGrid';
import { useLanguage } from '@/context/LanguageContext';

export default function AccessoriesPage() {
  const { t } = useLanguage();
  return (
    <ProductGrid
      tcg="all"
      category="accessories"
      titleKey="pages.accessories.title"
      subtitleKey="pages.accessories.subtitle"
      title={t('pages.accessories.title', 'ACCESSORIES')}
      subtitle={t('pages.accessories.subtitle', 'Sleeves, binders, deck boxes, playmats and more — for all TCGs.')}
    />
  );
}
