import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { KNOWN_TCGS, TCG_LABELS } from '@/lib/types';

export function generateStaticParams() {
  return KNOWN_TCGS.map(tcg => ({ tcg }));
}

export default async function SinglesPage({ params }: { params: Promise<{ tcg: string }> }) {
  const { tcg } = await params;
  if (!KNOWN_TCGS.includes(tcg)) notFound();

  return (
    <ProductGrid
      tcg={tcg}
      category="singles"
      title={`${tcg.toUpperCase()} SINGLES`}
      subtitle={`Browse individual ${TCG_LABELS[tcg] || tcg} cards by condition, treatment, and foil finish.`}
    />
  );
}
