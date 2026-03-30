'use client';

import { useState } from 'react';
import { Bounty, TREATMENT_LABELS } from '@/lib/types';
import ClientRequestModal from '@/components/ClientRequestModal';
import BountyCard from '@/components/BountyCard';

export default function PublicBountiesClient({ initialBounties }: { initialBounties: Bounty[] }) {
  const [bounties, setBounties] = useState<Bounty[]>(initialBounties);
  const [showModal, setShowModal] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');

  const handleSuccess = () => {
    setShowModal(false);
    setSuccessMsg('Your request was submitted successfully! We will contact you if we locate the card.');
    setTimeout(() => setSuccessMsg(''), 5000);
  };

  return (
    <>
      <div className="flex flex-col md:flex-row justify-between items-center mb-8 gap-6 animate-in fade-in duration-1000 delay-150 fill-mode-both">
        <p className="font-mono-stack text-sm tracking-widest text-text-muted">
          SHOWING {bounties.length} ENTRIES
        </p>
        <button onClick={() => setShowModal(true)} className="btn-primary py-3 px-8 xl:w-80 shadow-gold/20 hover:shadow-gold/40 border-gold/40 bg-ink-surface text-ink-deep font-display tracking-tight text-xl transition-all duration-300">
          REQUEST A CARD
        </button>
      </div>

      {successMsg && (
        <div className="mb-8 p-4 bg-emerald-100 text-emerald-800 border-l-4 border-emerald-500 font-bold text-center animate-in slide-in-from-top fade-in duration-300">
          ✓ {successMsg}
        </div>
      )}

      {bounties.length === 0 ? (
        <div className="py-24 text-center border-2 border-dashed border-ink-border/30 rounded-xl bg-white/40">
           <h3 className="font-display text-3xl text-ink-deep mb-2">NO ACTIVE BOUNTIES</h3>
           <p className="text-text-muted">We are not currently explicitly searching for any specific cards.<br/>But you can still send us a request!</p>
           <button onClick={() => setShowModal(true)} className="mt-6 btn-secondary px-8 font-bold">SUBMIT REQUEST</button>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {bounties.map((b, idx) => (
            <BountyCard key={b.id} bounty={b} delay={idx * 50} />
          ))}
        </div>
      )}

      {showModal && <ClientRequestModal onClose={() => setShowModal(false)} onSuccess={handleSuccess} />}
    </>
  );
}
