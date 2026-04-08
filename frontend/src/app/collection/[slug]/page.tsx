import { Metadata } from 'next';
import { fetchProducts, fetchCategories } from '@/lib/api';
import CollectionClient from './CollectionClient';

export async function generateMetadata({ params: rawParams }: any): Promise<Metadata> {
  const params = await rawParams;
  try {
    const category = (await fetchCategories()).find((c: any) => c.slug === params.slug);
    return {
      title: category ? `${category.name} - El Bulk` : 'Collection - El Bulk',
    };
  } catch {
    return { title: 'Collection - El Bulk' };
  }
}

export default async function CollectionPage({ params: rawParams, searchParams: rawSearchParams }: any) {
  const params = await rawParams;
  const searchParams = await rawSearchParams;
  
  const page = parseInt((searchParams.page as string) || '1', 10);
  let categories: any[] = [];
  let products = { products: [] as any[], total: 0, page: 1, page_size: 20 };
  
  try {
    categories = await fetchCategories();
    products = await fetchProducts({ page, page_size: 20, collection: params.slug });
  } catch { }

  const category = categories.find((c: any) => c.slug === params.slug);

  return (
    <CollectionClient 
      params={params}
      searchParams={searchParams}
      category={category}
      products={products}
    />
  );
}
