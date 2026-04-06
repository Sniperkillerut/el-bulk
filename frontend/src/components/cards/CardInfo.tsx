'use client';

import SetIcon from '../SetIcon';

interface CardInfoProps {
  name: string;
  setName?: string;
  setCode?: string;
  rarity?: string;
  hoverEffect?: boolean;
}

export default function CardInfo({ name, setName, setCode, rarity, hoverEffect = true }: CardInfoProps) {
  return (
    <>
      <h3
        className={`text-[11px] font-bold leading-none tracking-tight transition-colors line-clamp-2 ${hoverEffect ? 'hover:text-gold' : ''}`}
        style={{ color: 'var(--text-primary)', fontFamily: 'Inter, sans-serif' }}
      >
        {name}
      </h3>
      
      <p className="text-xs flex items-center gap-2 mt-0.5" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
        {setCode && <SetIcon setCode={setCode} rarity={rarity} size="xs" />}
        <span className="truncate">{setName || 'Any Edition'}</span>
      </p>
    </>
  );
}
