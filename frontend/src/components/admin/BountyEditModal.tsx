'use client';

import { useEffect, useState } from 'react';
import { adminCreateBounty, adminUpdateBounty } from '@/lib/api';
import { Bounty, BountyInput, FoilTreatment, CardTreatment, Condition, TCG, ScryfallCard, Settings, PriceSource } from '@/lib/types';
import { extractMTGMetadata, getScryfallImage, resolveFoilTreatment, findMatchingPrint, applyPrintPrices, resolveCardTreatment, getSuggestedPrice } from '@/lib/mtg-logic';
import ScryfallPopulate from './product/ScryfallPopulate';
import MTGVariantSelector from './MTGVariantSelector';
import CardImage from '@/components/CardImage';
import { useLanguage } from '@/context/LanguageContext';

interface BountyEditModalProps {
  editBounty: Bounty | null;
  tcgs: TCG[];
  settings: Settings | undefined;
  onClose: () => void;
  onSaved: () => void;
  initialData?: Partial<BountyInput>;
}

export const EMPTY_BOUNTY: BountyInput = {
  name: '', tcg: 'mtg', set_name: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  collector_number: '', promo_type: '', language: 'en',
  target_price: 0, hide_price: false, quantity_needed: 1, image_url: '',
  price_source: 'tcgplayer', price_reference: 0
};

export default function BountyEditModal({
  editBounty, tcgs, settings, onClose, onSaved, initialData
}: BountyEditModalProps) {
  const { t } = useLanguage();
  const [form, setForm] = useState<BountyInput>(EMPTY_BOUNTY);
  const [scryfallPrints, setScryfallPrints] = useState<ScryfallCard[]>([]);
  const [lookingUp, setLookingUp] = useState(false);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');
  
  // Temporary fields for scryfall lookup
  const [setCode, setSetCode] = useState('');
  const [collectorNumber, setCollectorNumber] = useState('');
  
  useEffect(() => {
    if (editBounty) {
      setForm({
        name: editBounty.name,
        tcg: editBounty.tcg,
        set_name: editBounty.set_name || '',
        condition: editBounty.condition || 'NM',
        foil_treatment: editBounty.foil_treatment,
        card_treatment: editBounty.card_treatment || 'normal',
        collector_number: editBounty.collector_number || '',
        promo_type: editBounty.promo_type || '',
        language: editBounty.language || 'en',
        target_price: editBounty.target_price || 0,
        hide_price: editBounty.hide_price,
        quantity_needed: editBounty.quantity_needed,
        image_url: editBounty.image_url || '',
        price_source: editBounty.price_source || 'tcgplayer',
        price_reference: editBounty.price_reference || 0,
        ...extractMTGMetadata(editBounty as unknown as ScryfallCard)
      });
    } else {
      setForm({ ...EMPTY_BOUNTY, ...initialData });
    }
    setFormError('');
    setScryfallPrints([]);
  }, [editBounty, initialData]);

  const handleSave = async () => {
    if (!form.name || !form.tcg) {
      setFormError(t('components.admin.bounty_modal.error_required', 'Name and TCG are required.'));
      return;
    }
    setSaving(true);
    setFormError('');
    try {
      const payload: BountyInput = {
        ...form,
        target_price: form.target_price || undefined,
      };

      if (editBounty) {
        await adminUpdateBounty(editBounty.id, payload);
      } else {
        await adminCreateBounty(payload);
      }
      onSaved();
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : t('components.admin.bounty_modal.error_save', 'Failed to save bounty.'));
    } finally {
      setSaving(false);
    }
  };

  const handlePopulate = async (forceSearchName?: string) => {
    setFormError('');
    const name = forceSearchName || form.name.trim();
    const set = setCode.trim().toLowerCase();
    const cn = collectorNumber.trim();
    if (!name && (!set || !cn)) return;

    setLookingUp(true);
    try {
      let searchQ = "";
      if (set && cn) searchQ = `set:${set} cn:"${cn}"`;
      else if (name) searchQ = `!"${name}"`;
      else if (set) searchQ = `set:${set}`;

      const fetchAllPrints = async (q: string) => {
        let results: ScryfallCard[] = [];
        let nextUrl: string | null = `https://api.scryfall.com/cards/search?q=${encodeURIComponent(q)}+game:paper&unique=prints&order=released`;
        
        while (nextUrl) {
          const r = await fetch(nextUrl);
          if (!r.ok) break;
          const b = await r.json() as { data?: ScryfallCard[], has_more?: boolean, next_page?: string };
          if (b.data) {
            const paperOnly = (b.data as ScryfallCard[]).filter(c => !c.digital);
            results = results.concat(paperOnly);
          }
          nextUrl = b.has_more ? (b.next_page as string) : null;
          if (nextUrl) await new Promise(res => setTimeout(res, 100));
        }
        return results;
      };

      let prints = await fetchAllPrints(searchQ);
      if (prints.length === 0) throw new Error('No printings found for that search.');

      if (prints.length > 0) {
        const oracleId = (prints[0] as ScryfallCard).oracle_id;
        if (oracleId) {
          const oraclePrints = await fetchAllPrints(`oracle_id:${oracleId}`);
          if (oraclePrints.length > prints.length) prints = oraclePrints;
        }
      }

      setScryfallPrints(prints);
      
      let bestPrint = prints.find(p => p?.set?.toLowerCase() === set && p?.collector_number === cn);
      if (!bestPrint && set) bestPrint = prints.find(p => p?.set?.toLowerCase() === set);
      if (!bestPrint) bestPrint = prints[0];

      if (!bestPrint) throw new Error('Could not identify a matching print.');
      
      const treat = resolveCardTreatment(bestPrint);
      const foil = resolveFoilTreatment(bestPrint);
      const promo = bestPrint.promo_types?.join(',') || 'none';
      const ref = applyPrintPrices(bestPrint, foil, form.price_source);
      const suggested = getSuggestedPrice(bestPrint, foil, form.price_source, settings);

      setForm(f => ({
        ...f,
        name: bestPrint?.name || f.name,
        set_name: bestPrint?.set_name || f.set_name,
        image_url: getScryfallImage(bestPrint) || f.image_url,
        card_treatment: treat,
        foil_treatment: foil,
        promo_type: promo,
        price_reference: ref,
        target_price: suggested !== undefined ? suggested : f.target_price,
        ...extractMTGMetadata(bestPrint)
      }));
      setSetCode(bestPrint?.set || setCode);
      setCollectorNumber(bestPrint?.collector_number || collectorNumber);

    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Scryfall fetch failed');
    } finally {
      setLookingUp(false);
    }
  };


  const handlePriceSourceChange = (src: PriceSource) => {
    setForm(f => {
      const bestPrint = findMatchingPrint(scryfallPrints, setCode, f.card_treatment, f.collector_number || '', f.promo_type || '', f.foil_treatment);
      const ref = applyPrintPrices(bestPrint, f.foil_treatment, src);
      const suggested = getSuggestedPrice(bestPrint, f.foil_treatment, src, settings);
      return {
        ...f,
        price_source: src,
        price_reference: ref,
        target_price: suggested !== undefined ? suggested : f.target_price
      };
    });
  };

  const handleSetSearchChange = (newSet: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, newSet, form.card_treatment, form.collector_number || '', form.promo_type || '', form.foil_treatment);
    setSetCode(newSet);
    const ref = applyPrintPrices(bestPrint, form.foil_treatment, form.price_source);
    const suggested = getSuggestedPrice(bestPrint, form.foil_treatment, form.price_source, settings);
    setForm(f => ({
      ...f,
      set_name: bestPrint?.set_name || f.set_name,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price
    }));
  };

  const handleTreatmentChange = (t: CardTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, t, form.collector_number || '', form.promo_type || '', form.foil_treatment);
    const ref = applyPrintPrices(bestPrint, bestPrint?.finishes?.includes('foil') ? 'foil' : 'non_foil', form.price_source);
    const suggested = getSuggestedPrice(bestPrint, bestPrint?.finishes?.includes('foil') ? 'foil' : 'non_foil', form.price_source, settings);
    
    setForm(f => ({
      ...f,
      card_treatment: t,
      collector_number: bestPrint?.collector_number || f.collector_number,
      promo_type: (bestPrint?.promo_types || []).join(',') || 'none',
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      ...extractMTGMetadata(bestPrint)
    }));
    if (bestPrint?.collector_number) setCollectorNumber(bestPrint.collector_number);
  };

  const handleArtChange = (a: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, a, form.promo_type || '', form.foil_treatment);
    setCollectorNumber(a);
    const ref = applyPrintPrices(bestPrint, resolveFoilTreatment(bestPrint), form.price_source);
    const suggested = getSuggestedPrice(bestPrint, resolveFoilTreatment(bestPrint), form.price_source, settings);
    
    setForm(f => ({
      ...f,
      collector_number: a,
      promo_type: (bestPrint?.promo_types || []).join(',') || 'none',
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handlePromoChange = (p: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, form.collector_number || '', p, form.foil_treatment);
    const ref = applyPrintPrices(bestPrint, resolveFoilTreatment(bestPrint), form.price_source);
    const suggested = getSuggestedPrice(bestPrint, resolveFoilTreatment(bestPrint), form.price_source, settings);
    
    setForm(f => ({
      ...f,
      promo_type: p,
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleFoilChange = (f: FoilTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, form.collector_number || '', form.promo_type || '', f);
    const ref = applyPrintPrices(bestPrint, f, form.price_source);
    const suggested = getSuggestedPrice(bestPrint, f, form.price_source, settings);
    setForm(old => ({
      ...old,
      foil_treatment: f,
      image_url: getScryfallImage(bestPrint) || old.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : old.target_price
    }));
  };


  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
      style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(12px)', overflowY: 'auto' }}>
      <div className="card p-0 w-full max-w-4xl mb-8 border-white/20 shadow-2xl animate-in fade-in zoom-in duration-300"
        style={{ position: 'relative', background: 'rgba(255, 255, 255, 0.85)', backdropFilter: 'blur(20px)' }}>

        <div className="flex items-center justify-between p-4 md:p-6 pb-2 border-b border-ink-border/20">
          <div className="flex flex-col">
            <h2 className="font-display text-4xl m-0 tracking-tighter text-ink-deep">
              {editBounty ? t('components.admin.bounty_modal.title_edit', 'EDIT WANTED CARD') : t('components.admin.bounty_modal.title_add', 'ADD WANTED CARD')}
            </h2>
          </div>
          <button onClick={onClose} 
            className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-hp-color/10 text-text-muted hover:text-hp-color transition-all duration-300">
            ✕
          </button>
        </div>

        <div className="p-4 md:p-6 flex flex-col md:flex-row gap-6">
          <div className="flex-1 space-y-4">
            <div className="flex gap-4">
              <div className="flex-1">
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.tcg_label', 'TCG SYSTEM')}</label>
                <select className="bg-white/50 border-white/40 w-full" value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
                  {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                </select>
              </div>
              <div className="flex-1">
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.condition_label', 'CONDITION')}</label>
                <select className="bg-white/50 border-white/40 w-full" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value as Condition }))}>
                  <option value="">{t('pages.common.labels.any', 'Any')}</option>
                  {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
                </select>
              </div>
            </div>

            {form.tcg === 'mtg' ? (
              <>
                <ScryfallPopulate 
                  name={form.name} 
                  setCode={setCode} 
                  collectorNumber={collectorNumber}
                  setName={form.set_name || ''}
                  scryfallPrints={scryfallPrints}
                  lookingUp={lookingUp}
                  onNameChange={val => { setForm(f => ({ ...f, name: val })); setScryfallPrints([]); setFormError(''); }}
                  onSetCodeChange={val => { setSetCode(val); setFormError(''); }}
                  onCollectorNumberChange={val => { handleArtChange(val); setFormError(''); }}
                  onPopulate={() => handlePopulate()}
                  onCardSelect={(card: ScryfallCard) => {
                    setForm(f => ({ ...f, name: card.name }));
                    handlePopulate(card.name);
                  }}
                  onSetSearchChange={handleSetSearchChange}
                />
                
                {scryfallPrints.length > 0 && (
                  <div className="mt-6 border-t border-ink-border/20 pt-6">
                    <MTGVariantSelector
                      tcg={form.tcg}
                      setCode={setCode}
                      cardTreatment={form.card_treatment}
                      collectorNumber={collectorNumber}
                      promoType={form.promo_type}
                      foilTreatment={form.foil_treatment}
                      prints={scryfallPrints}
                      onTreatmentChange={handleTreatmentChange}
                      onArtChange={handleArtChange}
                      onPromoChange={handlePromoChange}
                      onFoilChange={handleFoilChange}
                    />
                  </div>
                )}
              </>
            ) : (
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.card_name', 'CARD NAME')}</label>
                <input type="text" className="w-full text-lg font-bold" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder={t('components.admin.bounty_modal.name_placeholder', 'Entity Name')} />
                
                <label className="text-[10px] font-mono-stack mt-4 mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.set_name', 'SET NAME / INFO')}</label>
                <input type="text" className="w-full" value={form.set_name} onChange={e => setForm(f => ({ ...f, set_name: e.target.value }))} placeholder={t('components.admin.bounty_modal.set_placeholder', 'Optional Set Name')} />
                
                <label className="text-[10px] font-mono-stack mt-4 mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.image_url', 'IMAGE URL')}</label>
                <input type="text" className="w-full" value={form.image_url} onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))} placeholder={t('components.admin.bounty_modal.image_placeholder', 'https://...')} />
              </div>
            )}

            <div className="mt-8 border-t border-ink-border/20 pt-6">
              <div className="flex justify-between items-center mb-3">
                <h3 className="text-xs font-mono-stack uppercase text-text-muted tracking-widest m-0">{t('components.admin.bounty_modal.pricing_title', 'PRICING')}</h3>
                <div className="text-xs font-mono-stack px-2 py-1 rounded bg-ink-surface text-gold shadow-sm">
                  {form.price_source === 'tcgplayer' && `(x ${settings?.usd_to_cop_rate || 0} COP)`}
                  {form.price_source === 'cardmarket' && `(x ${settings?.eur_to_cop_rate || 0} COP)`}
                  {form.price_source !== 'manual' && (
                    <span className="ml-2 font-bold text-sm">
                      = ${(form.target_price || 0).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
                    </span>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.price_source', 'PRICE SOURCE *')}</label>
                  <select className="bg-white/80 border-white/40 w-full" value={form.price_source} onChange={e => handlePriceSourceChange(e.target.value as PriceSource)}>
                    <option value="manual">{t('components.admin.bounty_modal.source_manual', 'Manual Override (COP)')}</option>
                    <option value="tcgplayer">{t('components.admin.bounty_modal.source_tcgplayer', 'External: TCGPlayer (USD)')}</option>
                    <option value="cardmarket">{t('components.admin.bounty_modal.source_cardmarket', 'External: Cardmarket (EUR)')}</option>
                  </select>
                </div>
                
                {form.price_source !== 'manual' ? (
                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">
                      {t('components.admin.bounty_modal.ref_price_label', 'REFERENCE PRICE ({source}) *', { source: form.price_source === 'tcgplayer' ? 'USD' : 'EUR' })}
                    </label>
                    <input 
                      type="number" step="0.01" className="w-full font-mono bg-white/90 border-white/40" 
                      value={form.price_reference || ''} 
                      onChange={e => {
                        const val = parseFloat(e.target.value) || 0;
                        const rate = form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0) : (settings?.eur_to_cop_rate || 0);
                        setForm(f => ({ ...f, price_reference: val, target_price: Math.round(val * rate) }));
                      }} 
                    />
                  </div>
                ) : (
                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.price_cop_label', 'PRICE (COP) *')}</label>
                    <input 
                      type="number" min="0" step="100" 
                      className="w-full font-mono bg-white/90 border-white/40" 
                      value={form.target_price || ''} 
                      onChange={e => setForm(f => ({ ...f, target_price: parseFloat(e.target.value) }))} 
                      placeholder="0" 
                    />
                  </div>
                )}
              </div>

              <div className="p-4 bg-ink-surface/30 rounded-sm border border-ink-border/50">
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.quantity_label', 'QUANTITY NEEDED')}</label>
                <input type="number" min="1" className="w-full max-w-[200px] font-mono bg-white/80 border-white/40 h-[42px]" value={form.quantity_needed} onChange={e => setForm(f => ({ ...f, quantity_needed: parseInt(e.target.value) || 1 }))} />
              </div>
            </div>

            <div className="mt-4 flex items-center gap-2">
              <input type="checkbox" id="hide_price" checked={form.hide_price} onChange={e => setForm(f => ({ ...f, hide_price: e.target.checked }))} className="w-4 h-4 accent-gold" />
              <label htmlFor="hide_price" className="text-sm font-bold text-text-secondary cursor-pointer">{t('components.admin.bounty_modal.hide_price_label', 'Hide target price from public view')}</label>
            </div>
            <p className="text-xs text-text-muted pl-6">{t('components.admin.bounty_modal.hide_price_desc', 'If checked, users will see "Contact for Price" instead of your target COP value.')}</p>
          </div>

          <div className="w-full md:w-64 shrink-0 flex flex-col items-center">
            <div className="relative aspect-[63/88] w-full max-w-[200px] bg-ink-border/5 rounded shadow-inner flex items-center justify-center overflow-hidden mb-6">
              <CardImage 
                imageUrl={form.image_url} 
                name={form.name} 
                tcg={form.tcg} 
                foilTreatment={form.foil_treatment} 
                enableHover={false} 
                height="100%"
              />
            </div>

            {formError && (
              <div className="p-3 mb-4 bg-hp-color/10 text-hp-color text-xs rounded w-full">
                <strong>{t('pages.common.labels.error', 'Error')}:</strong> {formError}
              </div>
            )}

            <button onClick={handleSave} disabled={saving} className="btn-primary w-full py-3 mb-2">
              {saving ? t('components.admin.bounty_modal.saving', 'SAVING...') : t('components.admin.bounty_modal.save_btn', '💾 SAVE BOUNTY')}
            </button>
            <button onClick={onClose} disabled={saving} className="btn-secondary w-full py-2 border-none">
              {t('pages.common.actions.cancel', 'CANCEL')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
