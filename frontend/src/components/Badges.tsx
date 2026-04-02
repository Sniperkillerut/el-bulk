import { FOIL_LABELS } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

export function ConditionBadge({ condition }: { condition?: string }) {
  const { t } = useLanguage();
  if (!condition) return null;
  const cls = `badge badge-${condition.toLowerCase()}`;
  return <span className={cls}>{t(`pages.product.condition.${condition.toLowerCase()}`, condition)}</span>;
}

export function FoilBadge({ foil }: { foil: string }) {
  const { t } = useLanguage();
  if (foil === 'non_foil' || !foil) return null;
  const label = FOIL_LABELS[foil as keyof typeof FOIL_LABELS] || foil;
  return <span className="badge badge-foil">✦ {t(`pages.product.finish.${foil}`, label)}</span>;
}
