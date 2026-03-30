'use client';

import { FOIL_LABELS } from '@/lib/types';

export function ConditionBadge({ condition }: { condition?: string }) {
  if (!condition) return null;
  const cls = `badge badge-${condition.toLowerCase()}`;
  return <span className={cls}>{condition}</span>;
}

export function FoilBadge({ foil }: { foil: string }) {
  if (foil === 'non_foil') return null;
  const label = FOIL_LABELS[foil as keyof typeof FOIL_LABELS];
  if (!label) return null;
  return <span className="badge badge-foil">✦ {label}</span>;
}
