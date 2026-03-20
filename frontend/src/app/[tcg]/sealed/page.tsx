import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { KNOWN_TCGS, TCG_LABELS } from '@/lib/types';

export function generateStaticParams() {
  return KNOWN_TCGS.map(tcg => ({ tcg }));
}

export default async function SealedPage({ params }: { params: Promise<{ tcg: string }> }) {
  const { tcg } = await params;
  if (!KNOWN_TCGS.includes(tcg)) notFound();

  return (
    <ProductGrid
      tcg={tcg}
      category="sealed"
      title={`${tcg.toUpperCase()} SEALED`}
      subtitle={`Booster boxes, bundles, and sealed product for ${TCG_LABELS[tcg] || tcg}.`}
    />
  );
}
