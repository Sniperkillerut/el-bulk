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
    title: 'TCG Accessories — El Bulk',
    description: 'Sleeves, binders, deck boxes, playmats and more. Quality accessories for all TCGs.'
  };
}

export default function AccessoriesPage() {
  return (
    <ProductGrid
      tcg="all"
      category="accessories"
      titleKey="pages.accessories.title"
      subtitleKey="pages.accessories.subtitle"
      title="ACCESSORIES"
      subtitle="Sleeves, binders, deck boxes, playmats and more — for all TCGs."
    />
  );
}
