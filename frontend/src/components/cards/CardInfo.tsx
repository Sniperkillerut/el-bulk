'use client';

interface CardInfoProps {
  name: string;
  setName?: string;
  setCode?: string;
  hoverEffect?: boolean;
}

export default function CardInfo({ name, setName, setCode, hoverEffect = true }: CardInfoProps) {
  return (
    <>
      <h3
        className={`text-[11px] font-bold leading-none tracking-tight transition-colors line-clamp-2 ${hoverEffect ? 'hover:text-gold' : ''}`}
        style={{ color: 'var(--text-primary)', fontFamily: 'Inter, sans-serif' }}
      >
        {name}
      </h3>
      
      <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
        {setCode && `[${setCode}] `}{setName || 'Any Edition'}
      </p>
    </>
  );
}
