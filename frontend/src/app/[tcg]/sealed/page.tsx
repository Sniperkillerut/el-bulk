import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { fetchTCGs } from '@/lib/api';
import { Metadata } from 'next';
import { getSharedProductMetadata } from '@/lib/metadata';

interface PageProps {
  params: Promise<{ tcg: string }>;
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

export default async function SealedPage({ params }: PageProps) {
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
      category="sealed"
      titleKey="pages.sealed.title"
      subtitleKey="pages.sealed.subtitle"
      title={`${activeTcg?.name.toUpperCase() || tcg.toUpperCase()} SEALED`}
    />
  );
}
