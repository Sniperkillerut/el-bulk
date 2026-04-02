'use client';

import { FormState } from '../types';
import { ScryfallCard, FoilTreatment, CardTreatment } from '@/lib/types';
import MTGVariantSelector from '../../MTGVariantSelector';
import { useLanguage } from '@/context/LanguageContext';

interface VariantTabProps {
  form: FormState;
  prints: ScryfallCard[];
  isMTGSingles: boolean;
  onUpdate: (update: Partial<FormState>) => void;
  onTreatmentChange: (t: CardTreatment) => void;
  onArtChange: (a: string) => void;
  onPromoChange: (p: string) => void;
  onFoilChange: (f: FoilTreatment) => void;
  onSelectPrint: (p: ScryfallCard) => void;
}

export default function VariantTab({
  form, prints, isMTGSingles, onUpdate,
  onTreatmentChange, onArtChange, onPromoChange, onFoilChange, onSelectPrint
}: VariantTabProps) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-col gap-6">
      {isMTGSingles && prints.length > 1 && (
        <div className="bg-ink-surface/40 p-4 rounded-lg border border-ink-border/50 animate-in fade-in duration-500">
          <label className="text-[10px] font-mono-stack mb-3 block uppercase text-gold tracking-widest">{t('components.admin.product_modal.variant.prints_gallery', 'SELECT SPECIFIC PRINTING TO ADD')}</label>
          <div className="flex overflow-x-auto pb-2 gap-3 custom-scrollbar snap-x">
            {prints.map((p, idx) => {
              const isSelected = p.set?.toLowerCase() === form.set_code?.toLowerCase() && p.collector_number === form.collector_number;
              const img = p.image_uris?.small || p.image_uris?.normal || '';
              
              return (
                <button 
                  key={`${p.set}-${p.collector_number}-${idx}`}
                  onClick={(e) => { e.preventDefault(); onSelectPrint(p); }}
                  className={`
                    group relative flex-shrink-0 w-24 aspect-[63/88] rounded overflow-hidden border-2 transition-all snap-start
                    ${isSelected ? 'border-gold shadow-[0_0_15px_rgba(212,175,55,0.4)] scale-105 z-10' : 'border-ink-border/40 grayscale-[0.6] opacity-60 hover:opacity-100 hover:grayscale-0 hover:border-gold/50'}
                  `}
                >
                  {img ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={img} alt={p.set_name} className="w-full h-full object-cover" />
                  ) : (
                    <div className="w-full h-full bg-ink-border/10 flex items-center justify-center text-[8px] font-mono-stack text-center p-1">
                      {p.set_name}
                    </div>
                  )}
                  <div className="absolute inset-x-0 bottom-0 bg-black/80 px-1 py-0.5 text-[8px] font-mono-stack text-white opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap overflow-hidden text-ellipsis">
                    {p.set?.toUpperCase()} #{p.collector_number}
                  </div>
                  {isSelected && (
                    <div className="absolute top-1 right-1 bg-gold text-black rounded-full w-4 h-4 flex items-center justify-center text-[10px] shadow-lg">
                      ✓
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}

      {isMTGSingles && (
        <MTGVariantSelector
          tcg={form.tcg}
          setCode={form.set_code}
          setName={form.set_name}
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
      )}

      {!isMTGSingles && (
        <div className="p-12 text-center bg-ink-surface/20 border border-dashed border-ink-border rounded-lg text-text-muted animate-in fade-in duration-500">
          <p className="text-sm font-mono-stack">{t('components.admin.product_modal.variant.optimized_hint', 'VARIANT OPTIONS ARE CURRENTLY OPTIMIZED FOR MTG SINGLES.')}</p>
          <p className="text-[10px] mt-1 opacity-60">{t('components.admin.product_modal.variant.general_hint', 'General product details are managed in the main header and Pricing tab.')}</p>
        </div>
      )}

      {isMTGSingles && prints.length > 0 && (
        <div className="bg-ink-surface/30 p-4 rounded border border-ink-border animate-in fade-in slide-in-from-bottom-2 duration-300">
          <h4 className="text-[10px] font-mono-stack uppercase text-gold mb-2">{t('components.admin.product_modal.variant.metadata_title', 'MTG Metadata')}</h4>
          
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.language_label', 'Language')}</label>
              <input type="text" className="w-full text-xs py-1" value={form.language || 'en'} onChange={e => onUpdate({ language: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.rarity_label', 'Rarity')}</label>
              <input type="text" className="w-full text-xs py-1" value={form.rarity || 'common'} onChange={e => onUpdate({ rarity: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.colors_label', 'Colors')}</label>
              <input type="text" className="w-full text-xs py-1" value={form.color_identity} onChange={e => onUpdate({ color_identity: e.target.value })} />
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.cmc_label', 'CMC')}</label>
              <input type="number" step="1" className="w-full text-xs py-1" value={form.cmc ?? ''} onChange={e => onUpdate({ cmc: e.target.value === '' ? '' : Number(e.target.value) })} />
            </div>
          </div>

          <div className="grid grid-cols-3 sm:grid-cols-6 gap-x-2 gap-y-2 py-3 border-y border-ink-border/30 mb-4">
            {[
              { label: t('components.admin.product_modal.variant.legendary_label', 'LEGENDARY'), key: 'is_legendary' },
              { label: t('components.admin.product_modal.variant.historic_label', 'HISTORIC'), key: 'is_historic' },
              { label: t('components.admin.product_modal.variant.land_label', 'LAND'), key: 'is_land' },
              { label: t('components.admin.product_modal.variant.basic_label', 'BASIC'), key: 'is_basic_land' },
              { label: t('components.admin.product_modal.variant.full_art_label', 'FULL ART'), key: 'full_art' },
              { label: t('components.admin.product_modal.variant.textless_label', 'TEXTLESS'), key: 'textless' }
            ].map(flag => (
              <label key={flag.key} className="flex items-center gap-1.5 text-[9px] font-mono-stack cursor-pointer hover:text-gold transition-colors">
                <input type="checkbox" checked={!!form[flag.key as keyof typeof form]} onChange={e => onUpdate({ [flag.key]: e.target.checked })} className="w-3 h-3" />
                {flag.label}
              </label>
            ))}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-3">
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.type_line_label', 'Type Line')}</label>
                <input type="text" className="w-full text-xs py-1" value={form.type_line} onChange={e => onUpdate({ type_line: e.target.value })} />
              </div>
              <div>
                <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.artist_label', 'Artist')}</label>
                <input type="text" className="w-full text-xs py-1" value={form.artist} onChange={e => onUpdate({ artist: e.target.value })} />
              </div>
            </div>
            <div>
              <label className="text-[9px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.variant.oracle_text_label', 'Oracle Text')}</label>
              <textarea rows={4} className="w-full text-xs py-1 resize-none" value={form.oracle_text} onChange={e => onUpdate({ oracle_text: e.target.value })} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
