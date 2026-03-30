'use client';

import { Bounty, TREATMENT_LABELS } from '@/lib/types';
import CardImage from './CardImage';
import { ConditionBadge, FoilBadge } from './Badges';
import { openBountyModal } from './BountyModalManager';

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
        {/* Badges row */}
        <div className="flex flex-wrap gap-1">
          <ConditionBadge condition={bounty.condition} />
          <FoilBadge foil={bounty.foil_treatment} />
          {bounty.card_treatment && bounty.card_treatment !== 'normal' && TREATMENT_LABELS[bounty.card_treatment] && (
            <span className="badge" style={{ background: 'rgba(100,130,200,0.12)', color: '#8ba4d0', border: '1px solid rgba(100,130,200,0.25)' }}>
              {TREATMENT_LABELS[bounty.card_treatment]}
            </span>
          )}
        </div>

        {/* Name */}
        <h3 className="text-sm font-semibold leading-snug group-hover:text-gold transition-colors line-clamp-2"
          style={{ color: 'var(--text-primary)' }}>
          {bounty.name}
        </h3>

        {/* Set */}
        <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
          {bounty.set_name || 'Any Edition'}
        </p>
        
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
