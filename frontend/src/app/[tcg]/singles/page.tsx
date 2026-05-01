import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { fetchTCGs } from '@/lib/api';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  params: Promise<{ tcg: string }>;
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { productId } = await searchParams;
  const productMetadata = await getSharedProductMetadata(productId || null);
  
  if (productMetadata) return productMetadata;

  return {
    title: 'Singles Collection — El Bulk',
    description: 'Browse our massive selection of TCG singles. MTG, Pokémon, Lorcana and more.'
  };
}

export default async function SinglesPage({ params }: PageProps) {
  const { tcg } = await params;
  
  // Verify TCG is active
  const tcgs = await fetchTCGs(true);
  const activeTcg = tcgs.find(t => t.id === tcg);
  
  if (!activeTcg && tcg !== 'accessories') {
    notFound();
  }

  return (
    <ProductGrid
      tcg={tcg}
      category="singles"
      titleKey="pages.singles.title"
      subtitleKey="pages.singles.subtitle"
      title={`${activeTcg?.name.toUpperCase() || tcg.toUpperCase()} SINGLES`}
    />
  );
}
