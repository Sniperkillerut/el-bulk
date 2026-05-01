import ProductGrid from '@/components/ProductGrid';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { productId } = await searchParams;
  const productMetadata = await getSharedProductMetadata(productId || null);
  
  if (productMetadata) return productMetadata;

  return {
    title: 'Store Exclusives — El Bulk',
    description: 'Custom Commander decks, proxy kits, and premium items crafted in-house at El Bulk.'
  };
}

export default function StoreExclusivesPage({ searchParams }: PageProps) {
  return (
    <ProductGrid
      tcg="all"
      category="store_exclusives"
      titleKey="pages.store_exclusives.title"
      subtitleKey="pages.store_exclusives.subtitle"
      title="STORE EXCLUSIVES"
      subtitle="Custom Commander decks, proxy kits, and other premium items crafted in-house."
    />
  );
}
