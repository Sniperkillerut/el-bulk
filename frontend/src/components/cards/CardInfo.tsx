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
        className={`text-sm font-semibold leading-snug transition-colors line-clamp-2 ${hoverEffect ? 'hover:text-gold' : ''}`}
        style={{ color: 'var(--text-primary)' }}
      >
        {name}
      </h3>
      
      <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
        {setCode && `[${setCode}] `}{setName || 'Any Edition'}
      </p>
    </>
  );
}
