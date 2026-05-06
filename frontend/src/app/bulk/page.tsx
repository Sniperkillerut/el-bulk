import type { Metadata } from 'next';
import BulkPageClient from './BulkPageClient';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { productId } = await searchParams;
  const productMetadata = await getSharedProductMetadata(productId || null);
  
  if (productMetadata) return productMetadata;

  return {
    title: 'Sell Your Bulk — El Bulk TCG Store',
    description: 'Bring your bulk commons, uncommons, and rares. We pay cash on the spot.',
  };
}

export default function BulkPage(props: PageProps) {
  return <BulkPageClient />;
}
