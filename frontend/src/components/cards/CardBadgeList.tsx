import { ConditionBadge, FoilBadge } from '../Badges';
import { TREATMENT_LABELS } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

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
  const { t } = useLanguage();

  return (
    <div className="flex flex-wrap gap-1" data-theme-area="card-badges">
      {condition && <ConditionBadge condition={condition} />}
      {foil && <FoilBadge foil={foil} />}
      
      {treatment && treatment !== 'normal' && (
        <span className="badge opacity-80" style={{ background: 'var(--bg-header)', color: 'white', borderColor: 'var(--bg-page-dark)' }}>
          {t(`pages.product.version.${treatment}`, TREATMENT_LABELS[treatment] || treatment)}
        </span>
      )}
      
      {textless && (
        <span className="badge" style={{ background: 'var(--status-hp)', color: 'white', opacity: 0.15, border: '1px solid var(--status-hp)' }}>
          {t('pages.product.version.textless', 'TEXTLESS')}
        </span>
      )}
      
      {fullArt && treatment !== 'full_art' && (
        <span className="badge" style={{ background: 'var(--status-nm)', color: 'white', opacity: 0.15, border: '1px solid var(--status-nm)' }}>
          {t('pages.product.version.full_art', 'FULL ART')}
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
