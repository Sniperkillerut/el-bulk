import { Metadata } from 'next';
import { fetchNotices } from '@/lib/api';
import { getSharedProductMetadata } from '@/lib/metadata';
import NoticesClient from './NoticesClient';

interface PageProps {
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { productId } = await searchParams;
  const productMetadata = await getSharedProductMetadata(productId || null);
  
  if (productMetadata) return productMetadata;

  return {
    title: 'News & Updates - El Bulk TCG',
    description: 'Latest news, card spoilers, and reviews from El Bulk.',
  };
}

export default async function NoticesPage(props: PageProps) {
  const notices = await fetchNotices().catch(() => []);

  return <NoticesClient initialNotices={notices} />;
}
