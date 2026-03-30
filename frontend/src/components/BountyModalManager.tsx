'use client';

import { Suspense, useEffect, useState } from 'react';
import { Bounty } from '@/lib/types';
import BountyOfferModal from './BountyOfferModal';

export const openBountyModal = (bounty: Bounty) => {
  window.dispatchEvent(new CustomEvent('openBountyModal', { detail: bounty }));
};

function BountyModalManagerInner() {
  const [bounty, setBounty] = useState<Bounty | null>(null);

  useEffect(() => {
    const handler = (e: Event) => {
      const customEvent = e as CustomEvent<Bounty>;
      setBounty(customEvent.detail);
    };

    window.addEventListener('openBountyModal', handler);
    return () => {
      window.removeEventListener('openBountyModal', handler);
    };
  }, []);

  if (!bounty) return null;

  return <BountyOfferModal bounty={bounty} onClose={() => setBounty(null)} />;
}

export default function BountyModalManager() {
  return (
    <Suspense fallback={null}>
      <BountyModalManagerInner />
    </Suspense>
  );
}
