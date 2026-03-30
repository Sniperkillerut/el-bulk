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
    <div className="flex flex-wrap gap-1">
      {condition && <ConditionBadge condition={condition as any} />}
      {foil && <FoilBadge foil={foil} />}
      
      {treatment && treatment !== 'normal' && TREATMENT_LABELS[treatment] && (
        <span className="badge" style={{ background: 'rgba(100,130,200,0.12)', color: '#8ba4d0', border: '1px solid rgba(100,130,200,0.25)' }}>
          {TREATMENT_LABELS[treatment]}
        </span>
      )}
      
      {textless && (
        <span className="badge" style={{ background: 'rgba(248,113,113,0.15)', color: 'var(--hp-color)', border: '1px solid rgba(248,113,113,0.3)' }}>
          TEXTLESS
        </span>
      )}
      
      {fullArt && treatment !== 'full_art' && (
        <span className="badge" style={{ background: 'rgba(120,180,120,0.15)', color: 'var(--nm-color)', border: '1px solid rgba(120,180,120,0.3)' }}>
          FULL ART
        </span>
      )}
      
      {categories?.map(c => (
        <span key={c.id} className="badge" style={{ background: 'var(--gold)', color: 'var(--ink-deep)', border: '1px solid rgba(212,175,55,0.3)' }}>
          {c.name}
        </span>
      ))}
    </div>
  );
}
