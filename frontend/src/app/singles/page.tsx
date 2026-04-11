import { fetchProducts, fetchTCGs } from '@/lib/api';
import SinglesClient from './SinglesClient';

// Enable dynamic rendering via fetch or dynamic functions as per PPR rules


export default async function SinglesLandingPage() {
  const [productsRes, tcgsRes] = await Promise.all([
    fetchProducts({ category: 'singles', collection: 'featured', page_size: 12 }),
    fetchTCGs(true)
  ]);

  return <SinglesClient featured={productsRes.products} tcgs={tcgsRes} />;
}
