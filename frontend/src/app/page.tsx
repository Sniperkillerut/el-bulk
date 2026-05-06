import { fetchProducts, fetchCategories, fetchTCGs, fetchBounties } from '@/lib/api';
import HomePageClient from './HomePageClient';
import { CustomCategory } from '@/lib/types';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

type Props = {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}

export async function generateMetadata({ searchParams }: Props): Promise<Metadata> {
  const params = await searchParams;
  const productId = params.productId;

  if (typeof productId === 'string' && productId) {
    const productMetadata = await getSharedProductMetadata(productId);
    if (productMetadata) return productMetadata;
  }

  return {};
}

export default async function HomePage(props: Props) {
  const searchParams = await props.searchParams;
  const productId = searchParams.productId;
  let categories: CustomCategory[] = [];
  let tcgs: import('@/lib/types').TCG[] = [];
  let collections: { category: import('@/lib/types').CustomCategory; products: import('@/lib/types').Product[] }[] = [];
  let bounties: import('@/lib/types').Bounty[] = [];

  try {
    const [catsRes, tcgsRes] = await Promise.all([
      fetchCategories(),
      fetchTCGs(true)
    ]);
    categories = catsRes;
    tcgs = tcgsRes;

    // Fetch top 4 products for each category
    collections = await Promise.all(
      categories.filter(cat => cat.is_active).map(async (cat) => {
        const res = await fetchProducts({ page: 1, page_size: 4, collection: cat.slug });
        return { category: cat, products: res.products };
      })
    );

    // Fetch top active bounties
    const allBounties = await fetchBounties({ active: true });
    bounties = allBounties.slice(0, 8); // Show up to 8 on home page
  } catch {
    // DB not connected in dev — show empty state gracefully
  }

  return (
    <HomePageClient 
      categories={categories}
      tcgs={tcgs}
      collections={collections}
      bounties={bounties}
    />
  );
}
