import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { fetchTCGs } from '@/lib/api';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  params: Promise<{ tcg: string }>;
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ params, searchParams }: PageProps): Promise<Metadata> {
  try {
    const { productId } = await searchParams;
    const { tcg } = await params;
    
    if (productId) {
      const productMetadata = await getSharedProductMetadata(productId);
      if (productMetadata) return productMetadata;
    }

    return {
      title: `${tcg.toUpperCase()} Singles — El Bulk`,
      description: `Browse our collection of ${tcg.toUpperCase()} singles at El Bulk Bogotá. Secure shipping and evaluation.`
    };
  } catch (error) {
    console.error('[Metadata] Error in generateMetadata (Singles):', error);
    return {
      title: 'Singles Collection — El Bulk',
      description: 'Browse our massive selection of TCG singles. MTG, Pokémon, Lorcana and more.'
    };
  }
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
