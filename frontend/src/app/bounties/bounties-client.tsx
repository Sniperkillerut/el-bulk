'use client';

import { useState } from 'react';
import { Bounty, TREATMENT_LABELS } from '@/lib/types';
import ClientRequestModal from '@/components/ClientRequestModal';
import BountyOfferModal from '@/components/BountyOfferModal';
import CardImage from '@/components/CardImage';
import { ConditionBadge, FoilBadge } from '@/components/Badges';

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
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {bounties.map((b, idx) => (
            <div 
              key={b.id} 
              onClick={() => setOfferBounty(b)}
              className="card flex flex-col overflow-hidden animate-fade-up group cursor-pointer" 
              style={{ animationDelay: `${idx * 50}ms`, animationFillMode: 'both' }}
            >
              <div className="thumb-hover-wrap">
                <CardImage imageUrl={b.image_url} name={b.name} tcg={b.tcg} />
              </div>

              <div className="p-3 flex flex-col flex-1 gap-2">
                {/* Badges row */}
                <div className="flex flex-wrap gap-1">
                  <ConditionBadge condition={b.condition} />
                  <FoilBadge foil={b.foil_treatment} />
                  {b.card_treatment && b.card_treatment !== 'normal' && TREATMENT_LABELS[b.card_treatment] && (
                    <span className="badge" style={{ background: 'rgba(100,130,200,0.12)', color: '#8ba4d0', border: '1px solid rgba(100,130,200,0.25)' }}>
                      {TREATMENT_LABELS[b.card_treatment]}
                    </span>
                  )}
                </div>

                {/* Name */}
                <h3 className="text-sm font-semibold leading-snug group-hover:text-gold transition-colors line-clamp-2"
                  style={{ color: 'var(--text-primary)' }}>
                  {b.name}
                </h3>

                {/* Set */}
                <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
                  {b.set_name || 'Any Edition'}
                </p>
                
                {/* Footer */}
                <div className="flex items-center justify-between mt-auto pt-2" style={{ borderTop: '1px solid var(--ink-border)' }}>
                  <div className="flex flex-col">
                    <span className="text-[10px] font-mono-stack uppercase opacity-50 leading-none mb-1">Offer</span>
                    {b.hide_price ? (
                      <span className="text-sm font-semibold" style={{ color: 'var(--text-secondary)' }}>ASK</span>
                    ) : (
                      <span className="price text-sm leading-none">${b.target_price?.toLocaleString('es-CO')}</span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <span className="text-[10px] font-mono opacity-50" title="Quantity needed">
                      ×{b.quantity_needed}
                    </span>
                    <button 
                      className="btn-primary"
                      style={{ fontSize: '0.8rem', padding: '0.3rem 0.8rem' }}
                    >
                      SELL
                    </button>
                  </div>
                </div>
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
