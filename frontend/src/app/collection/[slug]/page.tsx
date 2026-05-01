import { Metadata } from 'next';
import { fetchProducts, fetchCategories } from '@/lib/api';
import { ProductListResponse, CustomCategory } from '@/lib/types';
import CollectionClient from './CollectionClient';
import { getSharedProductMetadata } from '@/lib/metadata';

type PageParams = Promise<{ slug: string }>;
type SearchParams = Promise<{ [key: string]: string | string[] | undefined }>;

export async function generateMetadata({ params: rawParams, searchParams: rawSearchParams }: { params: PageParams; searchParams: SearchParams }): Promise<Metadata> {
  const params = await rawParams;
  const searchParams = await rawSearchParams;
  const productId = typeof searchParams.productId === 'string' ? searchParams.productId : undefined;

  const productMetadata = await getSharedProductMetadata(productId || null);
  if (productMetadata) return productMetadata;

  try {
    const categories = await fetchCategories();
    const category = categories.find((c: CustomCategory) => c.slug === params.slug);
    
    if (!category) return { title: 'Collection - El Bulk' };

    const title = `${category.name} - El Bulk`;
    const description = `Explore our collection of ${category.name} at El Bulk. Singles, sealed products, and more. Secure shipping from Bogotá.`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        images: ['/og-image.png'],
        type: 'website',
      },
      twitter: {
        card: 'summary_large_image',
        title,
        description,
        images: ['/og-image.png'],
      },
    };
  } catch {
    return { title: 'Collection - El Bulk' };
  }
}

export default async function CollectionPage({ params: rawParams, searchParams: rawSearchParams }: { params: PageParams, searchParams: SearchParams }) {
  const params = await rawParams;
  const searchParams = await rawSearchParams;
  
  const page = parseInt((searchParams.page as string) || '1', 10);
  let categories: CustomCategory[] = [];
  let products: ProductListResponse = { 
    products: [], 
    total: 0, 
    page: 1, 
    page_size: 20,
    facets: {
      condition: {},
      foil: {},
      treatment: {},
      rarity: {},
      language: {},
      color: {},
      collection: {},
      set_name: []
    },
    query_time_ms: 0
  };
  
  try {
    categories = await fetchCategories();
    
    const isBinder = searchParams?.view === 'binder';
    const pageSize = isBinder ? 1000 : 20;
    
    products = await fetchProducts({ 
      page: isBinder ? 1 : page, 
      page_size: pageSize, 
      collection: params.slug 
    });
  } catch (err) {
    console.error(`[CollectionPage] Fetch error:`, err);
  }

  const category = categories.find((c: CustomCategory) => c.slug === params.slug);

  return (
    <CollectionClient 
      params={params}
      searchParams={searchParams}
      category={category}
      products={products}
    />
  );
}
