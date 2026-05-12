import { ConditionBadge, FoilBadge } from '../Badges';
import { TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

interface CardBadgeListProps {
  condition?: string;
  foil?: string;
  treatment?: string;
  textless?: boolean;
  fullArt?: boolean;
}

export default function CardBadgeList({ 
  condition, 
  foil, 
  treatment, 
  textless, 
  fullArt
}: CardBadgeListProps) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-wrap gap-1" data-theme-area="card-badges">
      {condition && <ConditionBadge condition={condition} />}
      {foil && <FoilBadge foil={foil} />}
      
      {treatment && treatment !== 'normal' && (
        <span className="badge opacity-80" style={{ background: 'var(--bg-header)', color: 'white', borderColor: 'var(--bg-page-dark)' }}>
          {resolveLabel(treatment, TREATMENT_LABELS)}
        </span>
      )}
      
      {textless && (
        <span className="badge" style={{ background: 'var(--status-hp)', color: 'white', opacity: 0.15, border: '1px solid var(--status-hp)' }}>
          {resolveLabel('textless', TREATMENT_LABELS)}
        </span>
      )}
      
      {fullArt && treatment !== 'full_art' && (
        <span className="badge" style={{ background: 'var(--status-nm)', color: 'white', opacity: 0.15, border: '1px solid var(--status-nm)' }}>
          {resolveLabel('full_art', TREATMENT_LABELS)}
        </span>
      )}
      
      {/* Categories moved to floating position on card image */}
    </div>
  );
}
