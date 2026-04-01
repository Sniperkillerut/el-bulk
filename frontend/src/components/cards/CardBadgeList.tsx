'use client';

import { ConditionBadge, FoilBadge } from '../Badges';
import { TREATMENT_LABELS } from '@/lib/types';

interface CardBadgeListProps {
  condition?: string;
  foil?: string;
  treatment?: string;
  textless?: boolean;
  fullArt?: boolean;
  categories?: { id: string; name: string }[];
}

export default function CardBadgeList({ 
  condition, 
  foil, 
  treatment, 
  textless, 
  fullArt, 
  categories 
}: CardBadgeListProps) {
  return (
    <div className="flex flex-wrap gap-1" data-theme-area="card-badges">
      {condition && <ConditionBadge condition={condition} />}
      {foil && <FoilBadge foil={foil} />}
      
      {treatment && treatment !== 'normal' && TREATMENT_LABELS[treatment] && (
        <span className="badge opacity-80" style={{ background: 'var(--bg-header)', color: 'white', borderColor: 'var(--bg-page-dark)' }}>
          {TREATMENT_LABELS[treatment]}
        </span>
      )}
      
      {textless && (
        <span className="badge" style={{ background: 'var(--status-hp)', color: 'white', opacity: 0.15, border: '1px solid var(--status-hp)' }}>
          TEXTLESS
        </span>
      )}
      
      {fullArt && treatment !== 'full_art' && (
        <span className="badge" style={{ background: 'var(--status-nm)', color: 'white', opacity: 0.15, border: '1px solid var(--status-nm)' }}>
          FULL ART
        </span>
      )}
      
      {categories?.map(c => (
        <span key={c.id} className="badge" style={{ background: 'var(--accent-primary)', color: 'var(--text-on-accent)', borderColor: 'var(--accent-primary-hover)' }}>
          {c.name}
        </span>
      ))}
    </div>
  );
}
