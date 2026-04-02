'use client';

import { 
  getTreatmentOptions, getArtOptions, getPromoOptions, getFoilOptions, ArtOption 
} from '@/lib/mtg-logic';
import { ScryfallCard, FoilTreatment, CardTreatment } from '@/lib/types';

export interface MTGVariantSelectorProps {
  tcg: string;
  setCode: string;
  setName?: string;
  cardTreatment: CardTreatment;
  collectorNumber: string;
  promoType?: string;
  foilTreatment: FoilTreatment;
  prints: ScryfallCard[];
  onTreatmentChange: (t: CardTreatment) => void;
  onArtChange: (a: string) => void;
  onPromoChange: (p: string) => void;
  onFoilChange: (f: FoilTreatment) => void;
}

export default function MTGVariantSelector({
  tcg, setCode, setName, cardTreatment, collectorNumber, promoType, foilTreatment, prints,
  onTreatmentChange, onArtChange, onPromoChange, onFoilChange
}: MTGVariantSelectorProps) {
  // Filter waterfalls
  const treatments = getTreatmentOptions(prints, setCode);
  const arts = getArtOptions(prints, setCode, cardTreatment);
  const promos = getPromoOptions(prints, setCode, cardTreatment, collectorNumber);
  const foils = getFoilOptions(prints, setCode, cardTreatment, collectorNumber, promoType || '');

  // If no prints loaded, show message
  if (prints.length === 0 && tcg === 'mtg') {
    return (
      <div className="p-8 text-center bg-ink-surface/50 border border-dashed border-ink-border rounded text-text-muted">
        <p className="text-sm font-mono-stack">POPULATE SCRYFALL DATA TO ENABLE VARIANT SELECTION</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-8 animate-in fade-in slide-in-from-bottom-2 duration-300">
      <div className="space-y-6">
        {/* Treatment Selection */}
        <div>
          <label className="text-[10px] font-mono-stack mb-2 block uppercase text-text-muted letter-spacing-widest">1. Card Treatment</label>
          <div className="grid grid-cols-2 gap-2">
            {treatments.map((t: CardTreatment) => (
              <button
                key={t}
                onClick={(e) => { e.preventDefault(); onTreatmentChange(t); }}
                className={`text-left px-3 py-2 rounded-sm border text-xs font-mono-stack transition-all ${
                  cardTreatment === t 
                    ? 'bg-gold/10 border-gold text-gold shadow-[0_0_10px_rgba(212,175,55,0.2)]' 
                    : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                }`}
              >
                {t.replace(/_/g, ' ').toUpperCase()}
              </button>
            ))}
          </div>
        </div>

        {/* Art Selection */}
        <div>
          <label className="text-[10px] font-mono-stack mb-2 block uppercase text-text-muted">2. Art / Variation (Collector #)</label>
          <div className="flex flex-wrap gap-2">
            {arts.length > 0 ? (arts as ArtOption[]).map((a: ArtOption) => (
              <button
                key={a.cn}
                onClick={(e) => { e.preventDefault(); onArtChange(a.cn); }}
                className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack flex flex-col items-start min-w-[100px] transition-all ${
                  collectorNumber === a.cn 
                    ? 'bg-gold text-black border-gold shadow-[0_0_8px_rgba(212,175,55,0.3)]' 
                    : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                }`}
              >
                <span className="font-bold">#{a.cn}</span>
                <span className="opacity-70 truncate w-full text-left font-sans">{a.artist}</span>
              </button>
            )) : <span className="text-[10px] italic text-text-muted">Standard Art Only</span>}
          </div>
        </div>

        {/* Promo Selection */}
        <div>
          <label className="text-[10px] font-mono-stack mb-2 block uppercase text-text-muted">3. Promo Version</label>
          <div className="flex flex-wrap gap-2">
            {promos.length > 0 ? (
              promos.length === 1 && (!promoType || promoType === 'none' || promoType === '') ? (
                <div className="px-3 py-1 rounded-sm border border-ink-border bg-ink-surface/50 text-[10px] font-mono-stack text-text-muted uppercase">
                  Standard Version
                </div>
              ) : promos.length > 0 ? (
                <>
                  <button
                    onClick={(e) => { e.preventDefault(); onPromoChange('none'); }}
                    className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack transition-all ${
                      (!promoType || promoType === 'none' || promoType === '') ? 'bg-gold text-black border-gold shadow-[0_0_8px_rgba(212,175,55,0.3)]' : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                    }`}
                  >
                    STANDARD
                  </button>
                  {promos.map((p: string) => (
                    <button
                      key={p}
                      onClick={(e) => { e.preventDefault(); onPromoChange(p); }}
                      className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack transition-all ${
                        (promoType === p || (promoType || '').split(',').includes(p)) ? 'bg-gold text-black border-gold shadow-[0_0_8px_rgba(212,175,55,0.3)]' : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                      }`}
                    >
                      {p.toUpperCase()}
                    </button>
                  ))}
                </>
              ) : (
                <div className="px-3 py-1 rounded-sm border border-ink-border bg-ink-surface/50 text-[10px] font-mono-stack text-text-muted uppercase">
                  Standard Version
                </div>
              )
            ) : (
              <div className="px-3 py-1 rounded-sm border border-ink-border bg-ink-surface/50 text-[10px] font-mono-stack text-text-muted uppercase">
                Standard Version
              </div>
            )}
          </div>
        </div>

        {/* Foil Selection */}
        <div>
          <label className="text-[10px] font-mono-stack mb-2 block uppercase text-text-muted">4. Foil Treatment</label>
          <div className="flex flex-wrap gap-2">
            {foils.map((f: FoilTreatment) => (
              <button
                key={f}
                onClick={(e) => { e.preventDefault(); onFoilChange(f); }}
                className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack ${
                  foilTreatment === f ? 'bg-gold text-black border-gold' : 'bg-ink-surface border-ink-border text-text-muted'
                }`}
              >
                {f.replace(/_/g, ' ').toUpperCase()}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="space-y-4 bg-ink-surface/30 p-4 rounded border border-ink-border h-fit">
        <div>
          <h4 className="text-[10px] font-mono-stack uppercase text-gold mb-3">Identity Summary</h4>
          <div className="space-y-2 text-xs">
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Set:</span>
              <div className="text-right">
                <span className="font-bold">{setCode.toUpperCase()}</span>
                {setName && <div className="text-[10px] opacity-70 truncate max-w-[150px]">{setName}</div>}
              </div>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Collector #:</span>
              <span>{collectorNumber}</span>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Treatment:</span>
              <span>{cardTreatment.replace(/_/g, ' ').toUpperCase()}</span>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Foil:</span>
              <span className={foilTreatment !== 'non_foil' ? 'text-gold italic' : ''}>
                {foilTreatment.replace(/_/g, ' ').toUpperCase()}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
