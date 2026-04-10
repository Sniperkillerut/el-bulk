import { fetchProducts, fetchTCGs } from '@/lib/api';
import SinglesClient from './SinglesClient';

// Enable PPR if possible, or at least dynamic rendering for latest data
export const dynamic = 'force-dynamic';
export const revalidate = 60; // Cache for 60 seconds

export default async function SinglesLandingPage() {
  const [productsRes, tcgsRes] = await Promise.all([
    fetchProducts({ category: 'singles', collection: 'featured', page_size: 12 }),
    fetchTCGs(true)
  ]);

  return <SinglesClient featured={productsRes.products} tcgs={tcgsRes} />;
}
