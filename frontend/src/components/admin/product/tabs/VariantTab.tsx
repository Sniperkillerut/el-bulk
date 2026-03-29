import { FormState } from '../types';
import { ScryfallCard, FoilTreatment, CardTreatment } from '@/lib/types';
import MTGVariantSelector from '../../MTGVariantSelector';

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
  return (
    <div className="flex flex-col gap-6">
      <MTGVariantSelector
        tcg={form.tcg}
        setCode={form.set_code}
        cardTreatment={form.card_treatment}
        collectorNumber={form.collector_number}
        promoType={form.promo_type}
        foilTreatment={form.foil_treatment}
        prints={prints}
        onTreatmentChange={onTreatmentChange}
        onArtChange={onArtChange}
        onPromoChange={onPromoChange}
        onFoilChange={onFoilChange}
      />

      {isMTGSingles && prints.length > 0 && (
        <div className="bg-ink-surface/30 p-4 rounded border border-ink-border animate-in fade-in slide-in-from-bottom-2 duration-300">
          <h4 className="text-[10px] font-mono-stack uppercase text-gold mb-2">MTG Metadata</h4>
          
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Language</label>
              <input type="text" className="w-full text-xs py-1" value={form.language || 'en'} onChange={e => onUpdate({ language: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Rarity</label>
              <input type="text" className="w-full text-xs py-1" value={form.rarity || 'common'} onChange={e => onUpdate({ rarity: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Colors</label>
              <input type="text" className="w-full text-xs py-1" value={form.color_identity} onChange={e => onUpdate({ color_identity: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">CMC</label>
              <input type="number" step="1" className="w-full text-xs py-1" value={form.cmc ?? ''} onChange={e => onUpdate({ cmc: e.target.value === '' ? '' : Number(e.target.value) })} />
            </div>
          </div>

          <div className="grid grid-cols-3 sm:grid-cols-6 gap-x-2 gap-y-2 py-3 border-y border-ink-border/30 mb-4">
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

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-3">
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Type Line</label>
                <input type="text" className="w-full text-xs py-1" value={form.type_line} onChange={e => onUpdate({ type_line: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Artist</label>
                <input type="text" className="w-full text-xs py-1" value={form.artist} onChange={e => onUpdate({ artist: e.target.value })} />
              </div>
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">Oracle Text</label>
              <textarea rows={4} className="w-full text-xs py-1 resize-none" value={form.oracle_text} onChange={e => onUpdate({ oracle_text: e.target.value })} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
