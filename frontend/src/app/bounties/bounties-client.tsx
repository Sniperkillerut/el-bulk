'use client';

import { useState } from 'react';
import { Bounty } from '@/lib/types';
import ClientRequestModal from '@/components/ClientRequestModal';
import BountyOfferModal from '@/components/BountyOfferModal';

export default function PublicBountiesClient({ initialBounties }: { initialBounties: Bounty[] }) {
  const [bounties, setBounties] = useState<Bounty[]>(initialBounties);
  const [showModal, setShowModal] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [offerBounty, setOfferBounty] = useState<Bounty | null>(null);

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
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 lg:gap-8">
          {bounties.map((b, idx) => (
            <div key={b.id} className="card bg-white/80 p-0 overflow-hidden flex flex-col group hover:-translate-y-2 transition-all duration-500 shadow-xl hover:shadow-gold/20 group animate-in fade-in slide-in-from-bottom flex" style={{ animationDelay: `${idx * 50}ms`, animationFillMode: 'both' }}>
              <div className="aspect-[4/3] w-full bg-ink-border/5 p-4 flex items-center justify-center relative">
                 {b.image_url ? (
                    <img src={b.image_url} alt={b.name} className="w-full h-full object-contain filter group-hover:brightness-110 transition-all duration-700 max-h-56" />
                 ) : (
                    <div className="h-full w-full flex items-center justify-center font-mono-stack text-xs text-text-muted">NO IMAGE</div>
                 )}
                 <div className="absolute inset-0 bg-gradient-to-t from-ink-deep/20 to-transparent"></div>
                 <span className="absolute bottom-2 right-2 bg-gold text-ink-deep font-mono-stack text-[9px] px-2 py-1 uppercase tracking-widest font-bold shadow-md transform rotate-2">{b.tcg}</span>
              </div>
              <div className="p-5 flex-1 flex flex-col">
                 <h3 className="font-display text-2xl leading-tight mb-1 text-ink-deep group-hover:text-gold transition-colors">{b.name}</h3>
                 <p className="text-xs text-text-muted mb-4 tracking-widest uppercase truncate">{b.set_name || 'Any Edition'}</p>
                 
                 <div className="grid grid-cols-2 gap-y-3 gap-x-2 text-[10px] font-mono-stack mt-auto border-t border-ink-border/20 pt-4">
                    <div>
                      <span className="opacity-50 block mb-0.5">CONDITION</span>
                      <strong className="text-ink-deep">{b.condition || 'ANY'}</strong>
                    </div>
                    <div>
                      <span className="opacity-50 block mb-0.5">FOIL</span>
                      <strong className="text-ink-deep">{b.foil_treatment.replace(/_/g, ' ')}</strong>
                    </div>
                    <div>
                      <span className="opacity-50 block mb-0.5">WANTED</span>
                      <strong className="text-ink-deep">{b.quantity_needed}</strong>
                    </div>
                    <div>
                      <span className="opacity-50 block mb-0.5">OFFER PRICE</span>
                      {b.hide_price ? (
                         <strong className="text-text-secondary cursor-help" title="Contact Us to determine a price">CONTACT US</strong>
                      ) : (
                         <strong className="text-gold-dark font-mono bg-gold/10 px-1 py-0.5 rounded shadow-sm">
                           ${b.target_price?.toLocaleString('es-CO')} COP
                         </strong>
                      )}
                    </div>
                 </div>
                 <button 
                   onClick={() => setOfferBounty(b)}
                   className="mt-4 w-full bg-gold/10 hover:bg-gold text-gold hover:text-ink-deep font-bold font-mono tracking-widest text-[10px] py-2 transition-all border border-gold/40 rounded-sm uppercase"
                 >
                   HAVE THIS? SELL IT
                 </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showModal && <ClientRequestModal onClose={() => setShowModal(false)} onSuccess={handleSuccess} />}
      {offerBounty && <BountyOfferModal bounty={offerBounty} onClose={() => setOfferBounty(null)} />}
    </>
  );
}
