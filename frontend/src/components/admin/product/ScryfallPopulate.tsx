import { ScryfallCard } from '@/lib/types';

interface ScryfallPopulateProps {
  name: string;
  setCode: string;
  collectorNumber: string;
  setName: string;
  scryfallPrints: ScryfallCard[];
  lookingUp: boolean;
  onNameChange: (val: string) => void;
  onSetCodeChange: (val: string) => void;
  onCollectorNumberChange: (val: string) => void;
  onPopulate: () => void;
  onSetSearchChange: (newSet: string) => void;
}

export default function ScryfallPopulate({
  name,
  setCode,
  collectorNumber,
  setName,
  scryfallPrints,
  lookingUp,
  onNameChange,
  onSetCodeChange,
  onCollectorNumberChange,
  onPopulate,
  onSetSearchChange
}: ScryfallPopulateProps) {
  return (
    <div className="mx-4 md:mx-6 mt-4 p-4 rounded-sm" style={{ background: 'var(--kraft-light)', border: '2px dashed var(--kraft-dark)' }}>
      <div className="flex items-end gap-3 flex-wrap sm:flex-nowrap">
        <div style={{ width: '90px' }}>
          <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET</label>
          {scryfallPrints.length > 0 ? (
            <select value={setCode} onChange={e => onSetSearchChange(e.target.value)} className="font-bold" style={{ fontSize: '0.85rem' }}>
              {Array.from(new Map(scryfallPrints.filter(c => !!c).map(c => [c.set, c.set_name])).entries()).map(([code, name]) => (
                <option key={code} value={code}>[{code.toUpperCase()}] {name}</option>
              ))}
            </select>
          ) : (
            <input 
              type="text" 
              value={setCode} 
              onChange={e => onSetCodeChange(e.target.value.toUpperCase())} 
              placeholder="MH2" 
              className="text-center font-bold uppercase" 
              style={{ fontSize: '0.85rem' }} 
            />
          )}
        </div>
        <div style={{ width: '70px' }}>
          <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}># CN</label>
          <input 
            type="text" 
            value={collectorNumber} 
            onChange={e => onCollectorNumberChange(e.target.value)} 
            placeholder="123" 
            className="text-center font-bold" 
            style={{ fontSize: '0.85rem' }} 
          />
        </div>
        <div className="flex-1 min-w-[200px]">
          <div className="flex justify-between items-end mb-1">
            <label className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>CARD NAME *</label>
            {setName && <span className="text-[10px] font-mono-stack truncate" style={{ color: 'var(--gold)', maxWidth: '350px' }}>{setName}</span>}
          </div>
          <input 
            type="text" 
            value={name} 
            onChange={e => onNameChange(e.target.value)} 
            style={{ fontSize: '1rem', fontWeight: 600 }} 
            placeholder="e.g. Lightning Bolt" 
          />
        </div>
        <button type="button" onClick={onPopulate}
          disabled={lookingUp || (!name.trim() && (!setCode.trim() || !collectorNumber.trim()))}
          className="btn-primary px-5 transition-all"
          style={{ height: '42px', fontSize: '0.85rem', whiteSpace: 'nowrap', opacity: lookingUp ? 0.7 : 1 }}>
          {lookingUp ? '⏳ LOOKING UP...' : scryfallPrints.length > 0 ? '✓ RE-POPULATE' : '📥 POPULATE'}
        </button>
      </div>
    </div>
  );
}
