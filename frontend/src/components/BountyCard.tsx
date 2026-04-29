'use client';

import { Bounty } from '@/lib/types';
import CardImage from './CardImage';
import { openBountyModal } from './BountyModalManager';
import CardBadgeList from './cards/CardBadgeList';
import CardInfo from './cards/CardInfo';
import { useLanguage } from '@/context/LanguageContext';

interface BountyCardProps {
  bounty: Bounty;
  delay?: number;
}

export default function BountyCard({ bounty, delay = 0 }: BountyCardProps) {
  const { t } = useLanguage();
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
      data-theme-area="bounty-card"
    >
      <div className="thumb-hover-wrap">
        <CardImage imageUrl={bounty.image_url} name={bounty.name} tcg={bounty.tcg} foilTreatment={bounty.foil_treatment} />
      </div>

      <div className="p-3 flex flex-col flex-1 gap-2">
        <CardBadgeList 
          condition={bounty.condition}
          foil={bounty.foil_treatment}
          treatment={bounty.card_treatment}
        />

        <CardInfo name={bounty.name} setName={bounty.set_name} hoverEffect={false} />
        
        {/* Footer */}
        <div className="flex items-center justify-between mt-auto pt-2 border-t border-border-main" data-theme-area="card-footer">
          <div className="flex flex-col">
            <span className="text-[10px] font-mono-stack uppercase opacity-50 leading-none mb-1 text-text-secondary">{t('pages.common.labels.offer', 'Offer')}</span>
            {bounty.hide_price ? (
              <span className="text-sm font-semibold text-text-secondary">{t('pages.common.labels.ask', 'ASK')}</span>
            ) : (
              <span className="price text-sm leading-none">${bounty.target_price?.toLocaleString('es-CO')}</span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1">
              <span className="text-[10px] font-mono opacity-50 text-text-muted" title={t('pages.common.tooltips.quantity_needed', 'Quantity needed')}>
                ×{bounty.quantity_needed}
              </span>
              {bounty.quantity_needed >= 10 && (
                <span className="flex h-1.5 w-1.5 rounded-full bg-hp-color animate-pulse" title={t('pages.bounties.labels.high_priority', 'High Priority')} />
              )}
            </div>
            <button 
              className="btn-primary text-[0.8rem] px-[0.8rem] py-[0.3rem]"
            >
              {t('pages.common.actions.sell', 'SELL')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
