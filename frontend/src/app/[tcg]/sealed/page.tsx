import { notFound } from 'next/navigation';
import ProductGrid from '@/components/ProductGrid';
import { fetchTCGs } from '@/lib/api';

export default async function SealedPage({ params }: { params: Promise<{ tcg: string }> }) {
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
      title={`${activeTcg?.name.toUpperCase() || tcg.toUpperCase()} SEALED`}
      subtitle={`Booster boxes, bundles, and sealed product for ${activeTcg?.name || tcg}.`}
    />
  );
}
