import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';
import SealedLandingClient from './SealedLandingClient';

interface PageProps {
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { productId } = await searchParams;
  const productMetadata = await getSharedProductMetadata(productId || null);
  
  if (productMetadata) return productMetadata;

  return {
    title: 'Sealed Products — El Bulk',
    description: 'Browse our selection of booster boxes, packs, and special editions. Secure shipping from Bogotá.'
  };
}

export default function SealedLandingPage() {
  return <SealedLandingClient />;
}
