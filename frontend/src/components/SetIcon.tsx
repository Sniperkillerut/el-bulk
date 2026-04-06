'use client';

interface SetIconProps {
  setCode?: string;
  className?: string;
  rarity?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
}

const RARITY_MAP: Record<string, string> = {
  common: 'ss-common',
  uncommon: 'ss-uncommon',
  rare: 'ss-rare',
  mythic: 'ss-mythic',
  special: 'ss-special',
  bonus: 'ss-bonus',
};

const SIZE_MAP: Record<string, string> = {
  xs: 'text-[10px]',
  sm: 'text-xs',
  md: 'text-base',
  lg: 'text-xl',
  xl: 'text-2xl',
};

export default function SetIcon({ setCode, className = '', rarity, size = 'md' }: SetIconProps) {
  if (!setCode) return null;

  const rClass = rarity ? RARITY_MAP[rarity.toLowerCase()] || '' : '';
  const sClass = SIZE_MAP[size] || SIZE_MAP.md;
  const code = setCode.toLowerCase();

  return (
    <i 
      className={`ss ss-${code} ${rClass} ${sClass} ${className} ss-fw ss-grad`} 
      title={setCode.toUpperCase()}
    />
  );
}
