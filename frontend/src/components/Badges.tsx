import { FOIL_LABELS } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

export const CONDITION_MAP: Record<string, string> = {
  'nm': 'NM',
  'lp': 'LP',
  'mp': 'MP',
  'hp': 'HP',
  'dmg': 'DM',
};

export function ConditionBadge({ condition }: { condition?: string }) {
  if (!condition) return null;
  const key = condition.toLowerCase();
  const label = CONDITION_MAP[key] || condition;
  const cls = `badge badge-${key} !px-1.5 !py-0 !text-[10px] !font-black !leading-tight`;
  return <span className={cls}>{label}</span>;
}

export function FoilBadge({ foil }: { foil: string }) {
  const { t } = useLanguage();
  if (foil === 'non_foil' || !foil) return null;
  const label = FOIL_LABELS[foil as keyof typeof FOIL_LABELS] || foil;
  return <span className="badge badge-foil">✦ {t(`pages.product.finish.${foil}`, label)}</span>;
}

export function HotBadge() {
  const { t } = useLanguage();
  return (
    <span className="badge" style={{ 
      background: 'linear-gradient(45deg, #ff4d4d, #f0932b)', 
      color: '#fff', 
      fontWeight: 'bold',
      boxShadow: '0 0 10px rgba(255, 77, 77, 0.4)',
      border: 'none'
    }}>
      🔥 {t('pages.product.badges.hot', 'HOT')}
    </span>
  );
}

export function NewBadge() {
  const { t } = useLanguage();
  return (
    <span className="badge" style={{ 
      background: 'linear-gradient(45deg, #2ecc71, #27ae60)', 
      color: '#fff', 
      fontWeight: 'bold',
      boxShadow: '0 0 10px rgba(46, 204, 113, 0.4)',
      border: 'none'
    }}>
      🆕 {t('pages.product.badges.new', 'NEW')}
    </span>
  );
}

