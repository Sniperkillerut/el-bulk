import { FormState } from '../types';
import { 
  getTreatmentOptions, getArtOptions, getPromoOptions, getFoilOptions, ArtOption 
} from '@/lib/mtg-logic';
import { ScryfallCard, FoilTreatment, CardTreatment } from '@/lib/types';

interface VariantTabProps {
  form: FormState;
  prints: ScryfallCard[];
  isMTGSingles: boolean;
  onUpdate: (update: Partial<FormState>) => void;
  onTreatmentChange: (t: CardTreatment) => void;
  onArtChange: (a: string) => void;
  onPromoChange: (p: string) => void;
  onFoilChange: (f: FoilTreatment) => void;
}

export default function VariantTab({
  form, prints, isMTGSingles, onUpdate,
  onTreatmentChange, onArtChange, onPromoChange, onFoilChange
}: VariantTabProps) {
  // Filter waterfalls
  const treatments = getTreatmentOptions(prints, form.set_code);
  const arts = getArtOptions(prints, form.set_code, form.card_treatment);
  const promos = getPromoOptions(prints, form.set_code, form.card_treatment, form.collector_number);
  const foils = getFoilOptions(prints, form.set_code, form.card_treatment, form.collector_number, form.promo_type);

  // If no prints loaded, show message
  if (prints.length === 0 && form.tcg === 'mtg') {
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
                onClick={() => onTreatmentChange(t)}
                className={`text-left px-3 py-2 rounded-sm border text-xs font-mono-stack transition-all ${
                  form.card_treatment === t 
                    ? 'bg-gold/10 border-gold text-gold shadow-[0_0_10px_rgba(212,175,55,0.2)]' 
                    : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                }`}
              >
                {t.replace('_', ' ').toUpperCase()}
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
                onClick={() => onArtChange(a.cn)}
                className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack flex flex-col items-start min-w-[100px] transition-all ${
                  form.collector_number === a.cn 
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
              // If there are multiple promo options (e.g. Standard vs Promo Pack), show buttons
              // Otherwise, just show a badge/label if it's purely informational
              promos.length === 1 && (!form.promo_type || form.promo_type === 'none' || form.promo_type === '') ? (
                <div className="px-3 py-1 rounded-sm border border-ink-border bg-ink-surface/50 text-[10px] font-mono-stack text-text-muted uppercase">
                  Standard Version
                </div>
              ) : promos.length > 0 ? (
                <>
                  <button
                    onClick={() => onPromoChange('none')}
                    className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack transition-all ${
                      (!form.promo_type || form.promo_type === 'none' || form.promo_type === '') ? 'bg-gold text-black border-gold shadow-[0_0_8px_rgba(212,175,55,0.3)]' : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
                    }`}
                  >
                    STANDARD
                  </button>
                  {promos.map((p: string) => (
                    <button
                      key={p}
                      onClick={() => onPromoChange(p)}
                      className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack transition-all ${
                        (form.promo_type === p || (form.promo_type || '').split(',').includes(p)) ? 'bg-gold text-black border-gold shadow-[0_0_8px_rgba(212,175,55,0.3)]' : 'bg-ink-surface border-ink-border text-text-muted hover:border-gold/50'
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
                onClick={() => onFoilChange(f)}
                className={`px-3 py-1 rounded-sm border text-[10px] font-mono-stack ${
                  form.foil_treatment === f ? 'bg-gold text-black border-gold' : 'bg-ink-surface border-ink-border text-text-muted'
                }`}
              >
                {f.replace('_', ' ').toUpperCase()}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="space-y-4 bg-ink-surface/30 p-4 rounded border border-ink-border">
        <div>
          <h4 className="text-[10px] font-mono-stack uppercase text-gold mb-3">Identity Summary</h4>
          <div className="space-y-2 text-xs">
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Set:</span>
              <span className="font-bold">{form.set_code.toUpperCase()}</span>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Collector #:</span>
              <span>{form.collector_number}</span>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Treatment:</span>
              <span>{form.card_treatment.toUpperCase()}</span>
            </div>
            <div className="flex justify-between border-b border-ink-border/50 pb-1">
              <span className="text-text-muted">Foil:</span>
              <span className={form.foil_treatment !== 'non_foil' ? 'text-gold italic' : ''}>
                {form.foil_treatment.toUpperCase()}
              </span>
            </div>
          </div>
        </div>

        {isMTGSingles && (
          <div className="pt-4 border-t border-ink-border/50 space-y-4">
            <h4 className="text-[10px] font-mono-stack uppercase text-gold mb-2">MTG Metadata</h4>
            
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Language</label>
                <input type="text" className="text-xs py-1" value={form.language || 'en'} onChange={e => onUpdate({ language: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Rarity</label>
                <input type="text" className="text-xs py-1" value={form.rarity || 'common'} onChange={e => onUpdate({ rarity: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Colors</label>
                <input type="text" className="text-xs py-1" value={form.color_identity} onChange={e => onUpdate({ color_identity: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">CMC</label>
                <input type="number" step="1" className="text-xs py-1" value={form.cmc ?? ''} onChange={e => onUpdate({ cmc: e.target.value === '' ? '' : Number(e.target.value) })} />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-x-2 gap-y-1 py-2 border-y border-ink-border/30">
              {[
                { label: 'LEGENDARY', key: 'is_legendary' },
                { label: 'HISTORIC', key: 'is_historic' },
                { label: 'LAND', key: 'is_land' },
                { label: 'BASIC', key: 'is_basic_land' },
                { label: 'FULL ART', key: 'full_art' },
                { label: 'TEXTLESS', key: 'textless' }
              ].map(flag => (
                <label key={flag.key} className="flex items-center gap-1.5 text-[9px] font-mono-stack cursor-pointer hover:text-gold transition-colors">
                  <input type="checkbox" checked={(form as any)[flag.key]} onChange={e => onUpdate({ [flag.key]: e.target.checked })} className="w-3 h-3" />
                  {flag.label}
                </label>
              ))}
            </div>

            <div className="space-y-3">
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Type Line</label>
                <input type="text" className="text-xs py-1" value={form.type_line} onChange={e => onUpdate({ type_line: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Oracle Text</label>
                <textarea rows={2} className="text-xs py-1" value={form.oracle_text} onChange={e => onUpdate({ oracle_text: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Artist</label>
                <input type="text" className="text-xs py-1" value={form.artist} onChange={e => onUpdate({ artist: e.target.value })} />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
