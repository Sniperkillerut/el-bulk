'use client';

import { Bounty } from '@/lib/types';
import CardImage from './CardImage';
import { openBountyModal } from './BountyModalManager';
import CardBadgeList from './cards/CardBadgeList';
import CardInfo from './cards/CardInfo';

interface BountyCardProps {
  bounty: Bounty;
  delay?: number;
}

export default function BountyCard({ bounty, delay = 0 }: BountyCardProps) {
  const handleOpenOffer = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    openBountyModal(bounty);
  };

  return (
    <div 
      onClick={handleOpenOffer}
      className="card flex flex-col overflow-hidden animate-fade-up group cursor-pointer" 
      style={{ animationDelay: `${delay}ms`, animationFillMode: 'both' }}
    >
      <div className="thumb-hover-wrap">
        <CardImage imageUrl={bounty.image_url} name={bounty.name} tcg={bounty.tcg} />
      </div>

      <div className="p-3 flex flex-col flex-1 gap-2">
        <CardBadgeList 
          condition={bounty.condition}
          foil={bounty.foil_treatment}
          treatment={bounty.card_treatment}
        />

        <CardInfo name={bounty.name} setName={bounty.set_name} hoverEffect={false} />
        
        {/* Footer */}
        <div className="flex items-center justify-between mt-auto pt-2" style={{ borderTop: '1px solid var(--ink-border)' }}>
          <div className="flex flex-col">
            <span className="text-[10px] font-mono-stack uppercase opacity-50 leading-none mb-1">Offer</span>
            {bounty.hide_price ? (
              <span className="text-sm font-semibold" style={{ color: 'var(--text-secondary)' }}>ASK</span>
            ) : (
              <span className="price text-sm leading-none">${bounty.target_price?.toLocaleString('es-CO')}</span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[10px] font-mono opacity-50" title="Quantity needed">
              ×{bounty.quantity_needed}
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
  );
}
