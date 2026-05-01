import { fetchProducts, fetchTCGs } from '@/lib/api';
import SinglesClient from './SinglesClient';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  try {
    const { productId } = await searchParams;
    
    if (productId) {
      const productMetadata = await getSharedProductMetadata(productId);
      if (productMetadata) return productMetadata;
    }

    return {
      title: 'Singles Collection — El Bulk',
      description: 'Browse our massive selection of TCG singles. MTG, Pokémon, Lorcana and more.'
    };
  } catch (error) {
    console.error('[Metadata] Error in generateMetadata (Root Singles):', error);
    return {
      title: 'Singles Collection — El Bulk',
      description: 'Browse our massive selection of TCG singles. MTG, Pokémon, Lorcana and more.'
    };
  }
}

// Enable dynamic rendering via fetch or dynamic functions as per PPR rules


export default async function SinglesLandingPage({ searchParams }: PageProps) {
  const [productsRes, tcgsRes] = await Promise.all([
    fetchProducts({ category: 'singles', collection: 'featured', page_size: 12 }),
    fetchTCGs(true)
  ]);

  return <SinglesClient featured={productsRes.products} tcgs={tcgsRes} />;
}
