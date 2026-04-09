import { useState, useEffect, useRef } from 'react';
import { ScryfallCard } from '@/lib/types';
import { getScryfallImage, getInitialFoilTreatment } from '@/lib/mtg-logic';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from '@/components/CardImage';

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
  onCardSelect: (card: ScryfallCard) => void;
  onSetSearchChange: (newSet: string) => void;
}

export default function ScryfallPopulate({
  name, setCode, collectorNumber, setName, scryfallPrints, lookingUp,
  onNameChange, onSetCodeChange, onCollectorNumberChange, onPopulate, onCardSelect, onSetSearchChange
}: ScryfallPopulateProps) {
  const { t } = useLanguage();
  const [suggestions, setSuggestions] = useState<ScryfallCard[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [searching, setSearching] = useState(false);
  const [hoverCard, setHoverCard] = useState<ScryfallCard | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Handle outside clicks to close suggestions
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowSuggestions(false);
        setHoverCard(null);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  // Debounced fuzzy search
  useEffect(() => {
    if (!name.trim() || name.length < 3) {
      setSuggestions([]);
      return;
    }

    const timer = setTimeout(async () => {
      setSearching(true);
      try {
        const q = encodeURIComponent(`${name} game:paper unique:cards`);
        const res = await fetch(`https://api.scryfall.com/cards/search?q=${q}`);
        if (!res.ok) throw new Error('Not found');
        const data = await res.json();
        setSuggestions(data.data?.slice(0, 8) || []);
        setShowSuggestions(true);
      } catch {
        setSuggestions([]);
      } finally {
        setSearching(false);
      }
    }, 400);

    return () => clearTimeout(timer);
  }, [name]);

  const handleSelectCard = (card: ScryfallCard) => {
    onCardSelect(card);
    setShowSuggestions(false);
    setHoverCard(null);
  };

  return (
    <div className="relative" 
      ref={containerRef}>
      
      <div className="flex items-end gap-2 flex-wrap sm:flex-nowrap">
        <div style={{ width: '90px' }}>
          <label className="text-[10px] font-mono-stack mb-1 block uppercase opacity-60">{t('components.admin.product_modal.variant.set_code_label', 'Set')}</label>
          {scryfallPrints.length > 0 ? (
            <select value={setCode} onChange={e => onSetSearchChange(e.target.value)} className="font-bold bg-white/50 border-white/20">
              {Array.from(new Map(scryfallPrints.filter(c => !!c).map(c => [c.set, c.set_name])).entries()).map(([code, name]) => (
                <option key={code} value={code}>[{code.toUpperCase()}] {name}</option>
              ))}
            </select>
          ) : (
            <input 
              type="text" value={setCode} 
              onChange={e => onSetCodeChange(e.target.value.toUpperCase())} 
              placeholder="MH2" className="text-center font-bold uppercase py-2 bg-white/50 border-white/20" 
            />
          )}
        </div>

        <div style={{ width: '70px' }}>
          <label className="text-[10px] font-mono-stack mb-1 block uppercase opacity-60">CN</label>
          <input 
            type="text" value={collectorNumber} 
            onChange={e => onCollectorNumberChange(e.target.value)} 
            placeholder="CN" className="text-center font-bold py-2 bg-white/50 border-white/20" 
          />
        </div>

        <div className="flex-1 min-w-[120px] relative">
          <div className="flex justify-between items-end mb-1">
            <label className="text-[10px] font-mono-stack uppercase opacity-60">{t('components.admin.product_modal.scryfall.search_hint', 'Card Name (Search Scryfall)')}</label>
            {searching && <span className="text-[9px] font-mono-stack animate-pulse text-gold">{t('components.admin.product_modal.scryfall.looking_up', 'Searching...')}</span>}
            {setName && <span className="text-[10px] font-mono-stack truncate text-gold max-w-[250px]">{setName}</span>}
          </div>
          <input 
            type="text" value={name} 
            onChange={e => onNameChange(e.target.value)} 
            className="font-bold py-2 px-3 text-lg bg-white shadow-inner focus:border-gold transition-all"
            placeholder="e.g. Black Lotus" 
          />

          {/* Suggestions Dropdown */}
          {showSuggestions && suggestions.length > 0 && (
            <div className="absolute top-full left-0 right-0 mt-1 z-[60] bg-white shadow-2xl border border-kraft-dark divide-y divide-kraft-light animate-in fade-in slide-in-from-top-1 duration-200">
              {suggestions.map((card) => (
                <div key={card.id}
                  onClick={() => handleSelectCard(card)}
                  onMouseEnter={() => setHoverCard(card)}
                  onMouseLeave={() => setHoverCard(null)}
                  className="flex items-center gap-3 p-2 hover:bg-gold/10 cursor-pointer group transition-colors">
                  <div className="w-8 h-8 flex-shrink-0 bg-ink-border/10 overflow-hidden relative">
                    <CardImage imageUrl={getScryfallImage(card)} name={card.name} tcg="mtg" foilTreatment={getInitialFoilTreatment(card)} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-bold truncate group-hover:text-gold transition-colors">{card.name}</div>
                    <div className="text-[9px] font-mono-stack text-text-muted">[{card.set.toUpperCase()}] {card.set_name}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <button type="button" onClick={onPopulate}
          disabled={lookingUp || (!name.trim() && (!setCode.trim() || !collectorNumber.trim()))}
          className="btn-primary px-8 h-[46px] text-sm tracking-widest font-bold shadow-lg"
          style={{ opacity: lookingUp ? 0.7 : 1 }}>
          {lookingUp ? t('components.admin.product_modal.scryfall.looking_up', '⌛ PROCESSING...') : scryfallPrints.length > 0 ? '✓ RE-SYNC' : t('components.admin.product_modal.scryfall.populate_btn', '📥 POPULATE')}
        </button>
      </div>

      {/* Hover Preview Overlay */}
      {hoverCard && (
        <div className="fixed z-[100] pointer-events-none animate-in fade-in zoom-in duration-200"
          style={{ 
            left: containerRef.current ? containerRef.current.offsetLeft + containerRef.current.offsetWidth + 20 : '50%',
            top: '50%', transform: 'translateY(-50%)'
          }}>
          <div className="card p-1 shadow-[0_0_50px_rgba(0,0,0,0.5)] border-gold/50 bg-ink-surface" style={{ width: '280px' }}>
             <CardImage imageUrl={getScryfallImage(hoverCard)} name={hoverCard.name} tcg="mtg" foilTreatment={getInitialFoilTreatment(hoverCard)} enableHover={false} />
             <div className="p-3 text-center">
                <p className="text-[10px] font-mono-stack text-gold">{hoverCard.type_line}</p>
                <p className="text-[9px] font-mono-stack text-text-muted mt-1 italic">{hoverCard.artist}</p>
             </div>
          </div>
        </div>
      )}
    </div>
  );
}
