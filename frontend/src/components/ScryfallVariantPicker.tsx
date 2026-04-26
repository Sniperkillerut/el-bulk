'use client';
import { useState, useEffect } from 'react';
import { useLanguage } from '@/context/LanguageContext';

interface ScryfallPrint {
  id: string;
  set_name: string;
  set: string;
  collector_number: string;
  image_uris?: { small: string; normal: string };
  card_faces?: { image_uris?: { small: string } }[];
  foil: boolean;
  nonfoil: boolean;
  finishes: string[];
}

interface Props {
  cardName: string;
  onSelect: (print: ScryfallPrint | null) => void; // null = "Any Version"
  selectedId?: string;
}

export default function ScryfallVariantPicker({ cardName, onSelect, selectedId }: Props) {
  const { t } = useLanguage();
  const [prints, setPrints] = useState<ScryfallPrint[]>([]);
  const [loading, setLoading] = useState(false);
  const [anyVersion, setAnyVersion] = useState(true);

  useEffect(() => {
    if (!cardName || cardName.length < 3) {
      setPrints([]);
      return;
    }
    const timer = setTimeout(() => {
      setLoading(true);
      fetch(`https://api.scryfall.com/cards/search?q=!"${encodeURIComponent(cardName)}"&unique=prints&order=released`)
        .then(r => r.json())
        .then(data => setPrints(data.data || []))
        .catch(() => setPrints([]))
        .finally(() => setLoading(false));
    }, 500);

    return () => clearTimeout(timer);
  }, [cardName]);

  const getImage = (p: ScryfallPrint) =>
    p.image_uris?.small ?? p.card_faces?.[0]?.image_uris?.small ?? '/placeholder-card.png';

  const toggleAnyVersion = () => {
    const newVal = !anyVersion;
    setAnyVersion(newVal);
    if (newVal) {
      onSelect(null);
    }
  };

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-top-2 duration-300">
      {/* Any Version Toggle */}
      <div 
        onClick={toggleAnyVersion}
        className={`p-4 rounded-xl border-2 transition-all cursor-pointer group ${
          anyVersion 
            ? 'border-gold bg-gold/5 shadow-md shadow-gold/10' 
            : 'border-kraft-dark/20 bg-kraft-light/30 hover:border-gold/40'
        }`}
      >
        <div className="flex items-center gap-4">
          <div className={`w-10 h-6 rounded-full transition-colors relative flex-shrink-0 ${anyVersion ? 'bg-gold' : 'bg-kraft-dark/40'}`}>
            <div className={`w-4 h-4 bg-white rounded-full absolute top-1 transition-all shadow ${anyVersion ? 'left-5' : 'left-1'}`} />
          </div>
          <div className="flex-1">
            <span className="text-xs font-mono-stack uppercase font-black text-ink-deep tracking-wider block">
              {t('components.variant_picker.any_version', 'Any Version')}
            </span>
            <span className="text-[10px] text-text-muted font-medium">
              {t('components.variant_picker.any_version_desc', "I'll take whatever edition or print you find first")}
            </span>
          </div>
          {anyVersion && <span className="text-gold">✨</span>}
        </div>
      </div>

      {/* Variant Grid */}
      <div className={`transition-all duration-500 overflow-hidden ${anyVersion ? 'max-h-0 opacity-0' : 'max-h-[500px] opacity-100'}`}>
        <div className="flex items-center justify-between mb-2">
          <h4 className="text-[10px] font-mono-stack uppercase text-text-muted font-bold tracking-widest">
            {t('components.variant_picker.specific_version', 'Or pick a specific version:')}
          </h4>
          {loading && <span className="text-[8px] animate-pulse text-gold font-bold">SEARCHING SCryfall...</span>}
        </div>
        
        <div className="grid grid-cols-3 sm:grid-cols-4 gap-3 max-h-64 overflow-y-auto pr-2 no-scrollbar">
          {prints.map(p => (
            <button
              key={p.id}
              type="button"
              onClick={() => onSelect(p)}
              className={`group relative rounded-lg overflow-hidden border-2 transition-all hover:scale-105 active:scale-95 ${
                selectedId === p.id 
                  ? 'border-gold shadow-lg shadow-gold/30 ring-2 ring-gold/20' 
                  : 'border-transparent hover:border-gold/40 bg-kraft-dark/10'
              }`}
              title={`${p.set_name} #${p.collector_number}`}
            >
              <img src={getImage(p)} alt={p.set_name} className="w-full aspect-[2.5/3.5] object-cover" loading="lazy" />
              <div className="absolute inset-0 bg-gradient-to-t from-ink-deep/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex flex-col justify-end p-1.5">
                <p className="text-[7px] font-black text-white uppercase tracking-tighter leading-tight truncate">{p.set_name}</p>
                <p className="text-[6px] font-mono-stack text-gold truncate">#{p.collector_number}</p>
              </div>
              {selectedId === p.id && (
                <div className="absolute top-1 right-1 w-4 h-4 bg-gold rounded-full flex items-center justify-center text-[10px] shadow-sm">✓</div>
              )}
            </button>
          ))}
          {!loading && prints.length === 0 && cardName.length >= 3 && (
            <div className="col-span-full py-8 text-center bg-kraft-paper/30 rounded border border-dashed border-kraft-dark/20">
              <p className="text-[10px] text-text-muted uppercase font-bold tracking-widest">{t('components.variant_picker.no_prints', 'No prints found for this name')}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
