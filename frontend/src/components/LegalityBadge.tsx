'use client';

interface LegalityBadgeProps {
  format: string;
  status?: string;
  showFormatName?: boolean;
}

const FORMAT_LABELS: Record<string, string> = {
  commander: 'EDH',
  modern: 'MODERN',
  standard: 'STANDARD',
  pioneer: 'PIONEER',
  legacy: 'LEGACY',
  vintage: 'VINTAGE',
  pauper: 'PAUPER',
};

export default function LegalityBadge({ format, status = 'not_legal', showFormatName = true }: LegalityBadgeProps) {
  const isLegal = status === 'legal';
  const isBanned = status === 'banned';
  const isRestricted = status === 'restricted';

  let bgColor = 'bg-ink-surface/50 border-ink-border text-white/40';
  let label = status.replace('_', ' ').toUpperCase();

  if (isLegal) {
    bgColor = 'bg-green-500/20 border-green-500/40 text-green-400';
    label = 'LEGAL';
  } else if (isBanned) {
    bgColor = 'bg-hp-color/20 border-hp-color/40 text-hp-color';
    label = 'BANNED';
  } else if (isRestricted) {
    bgColor = 'bg-blue-500/20 border-blue-500/40 text-blue-400';
    label = 'RESTRICTED';
  }

  if (status === 'not_legal') {
    label = 'NOT LEGAL';
  }

  return (
    <div className={`inline-flex items-center gap-1.5 px-1.5 py-0.5 rounded-sm border text-[9px] font-mono-stack font-bold tracking-tight ${bgColor}`}>
      {showFormatName && <span className="opacity-70">{FORMAT_LABELS[format.toLowerCase()] || format.toUpperCase()}</span>}
      <span>{label}</span>
    </div>
  );
}
