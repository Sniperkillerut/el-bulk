'use client';

import { useCallback, useEffect, useState } from 'react';
import { adminCreateBounty, adminUpdateBounty, adminFetchBountyRequests, adminFetchBountyOffers, adminFetchExternalPrice } from '@/lib/api';
import { Bounty, BountyInput, FoilTreatment, CardTreatment, Condition, TCG, ScryfallCard, Settings, PriceSource, ClientRequest, BountyOffer } from '@/lib/types';
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
  price_source: 'tcgplayer', price_reference: 0, is_generic: false, scryfall_id: '',
  frame_effects: []
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

  const resolvePrice = useCallback(async (card: ScryfallCard | undefined, foil: FoilTreatment, source: PriceSource, treat: CardTreatment) => {
    if (!card) return { ref: 0, suggested: 0 };

    let ref = applyPrintPrices(card, foil, source);

    if (source === 'cardkingdom') {
      try {
        const ck = await adminFetchExternalPrice(
          card.name,
          card.set || '',
          card.set_name || '',
          card.collector_number || '',
          foil,
          treat,
          'cardkingdom',
          card.id
        );
        if (ck && ck.price) {
          ref = ck.price;
        }
      } catch (ckErr) {
        console.warn('CK price fetch failed:', ckErr);
      }
    }

    const suggested = getSuggestedPrice(card, foil, source, settings);
    const finalSuggested = source === 'cardkingdom'
      ? (Math.round((ref * (settings?.ck_to_cop_rate || settings?.usd_to_cop_rate || 0)) / 100) * 100)
      : suggested;

    return { ref, suggested: finalSuggested };
  }, [settings]);

  const [relatedRequests, setRelatedRequests] = useState<ClientRequest[]>([]);
  const [loadingRequests, setLoadingRequests] = useState(false);

  const [relatedOffers, setRelatedOffers] = useState<BountyOffer[]>([]);
  const [loadingOffers, setLoadingOffers] = useState(false);

  useEffect(() => {
    if (editBounty) {
      setLoadingRequests(true);
      adminFetchBountyRequests(editBounty.id)
        .then(data => {
          setRelatedRequests(data);
        })
        .finally(() => setLoadingRequests(false));

      setLoadingOffers(true);
      adminFetchBountyOffers(editBounty.id)
        .then(data => {
          setRelatedOffers(data);
        })
        .finally(() => setLoadingOffers(false));
    } else {
      setRelatedRequests([]);
      setRelatedOffers([]);
    }
  }, [editBounty]);

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
        frame_effects: form.frame_effects || undefined,
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

  const handlePopulate = useCallback(async (forceSearchName?: string) => {
    setFormError('');
    const name = forceSearchName || form.name.trim();
    const set = setCode.trim().toLowerCase();
    const cn = collectorNumber.trim();
    const scryfallId = form.scryfall_id;

    if (!scryfallId && !name && (!set || !cn)) return;

    setLookingUp(true);
    try {
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

      let prints: ScryfallCard[] = [];

      // 1. Direct ID lookup if possible
      if (scryfallId) {
        const r = await fetch(`https://api.scryfall.com/cards/${scryfallId}`);
        if (r.ok) {
          const card = await r.json();
          prints = [card];
        }
      }

      // 2. If no ID or ID fetch failed, search by name/set/cn
      if (prints.length === 0) {
        let searchQ = "";
        if (set && cn) searchQ = `set:${set} cn:"${cn}"`;
        else if (set && name) searchQ = `!"${name}" set:${set}`; // PRIORITIZE SET IF PROVIDED
        else if (name) searchQ = `!"${name}"`;
        else if (set) searchQ = `set:${set}`;

        prints = await fetchAllPrints(searchQ);
      }

      if (prints.length === 0) throw new Error('No printings found for that search.');

      // 3. Populate all oracle prints so the switcher works
      if (prints.length > 0) {
        const oracleId = (prints[0] as ScryfallCard).oracle_id;
        if (oracleId) {
          const oraclePrints = await fetchAllPrints(`oracle_id:${oracleId}`);
          if (oraclePrints.length > prints.length) prints = oraclePrints;
        }
      }

      setScryfallPrints(prints);

      let bestPrint = prints.find(p => p.id === scryfallId);
      if (!bestPrint) bestPrint = prints.find(p => p?.set?.toLowerCase() === set && p?.collector_number === cn);
      if (!bestPrint && set) bestPrint = prints.find(p => p?.set?.toLowerCase() === set);
      if (!bestPrint) bestPrint = prints[0];

      if (!bestPrint) throw new Error('Could not identify a matching print.');

      const treat = resolveCardTreatment(bestPrint);
      const foil = resolveFoilTreatment(bestPrint);
      const promo = bestPrint.promo_types?.join(',') || 'none';
      const { ref, suggested } = await resolvePrice(bestPrint, foil, form.price_source, treat);

      setForm(f => {
        // Only update treatments if they were 'normal'/'non_foil' and the print offers something more specific,
        // or if we are populating for the first time.
        const keepTreat = f.card_treatment && f.card_treatment !== 'normal';
        const keepFoil = f.foil_treatment && f.foil_treatment !== 'non_foil';

        return {
          ...f,
          ...extractMTGMetadata(bestPrint),
          name: bestPrint?.name || f.name,
          set_name: bestPrint?.set_name || f.set_name,
          image_url: getScryfallImage(bestPrint) || f.image_url,
          card_treatment: keepTreat ? f.card_treatment : treat,
          foil_treatment: keepFoil ? f.foil_treatment : foil,
          promo_type: promo,
          price_reference: ref,
          target_price: suggested !== undefined ? suggested : f.target_price,
          scryfall_id: bestPrint?.id || '',
          frame_effects: bestPrint?.frame_effects || [],
        };
      });
      setSetCode(bestPrint?.set || setCode);
      setCollectorNumber(bestPrint?.collector_number || collectorNumber);

    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Scryfall fetch failed');
    } finally {
      setLookingUp(false);
    }
  }, [form.name, form.scryfall_id, form.price_source, setCode, collectorNumber, resolvePrice]);

  useEffect(() => {
    if (editBounty) {
      const data = {
        ...extractMTGMetadata(editBounty as unknown as ScryfallCard),
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
        is_generic: editBounty.is_generic,
        scryfall_id: editBounty.scryfall_id || '',
        frame_effects: editBounty.frame_effects || [],
        set_code: editBounty.set_code || '',
      };
      setForm(data);
      if (data.collector_number) setCollectorNumber(data.collector_number as string);
      if (data.set_code) setSetCode(data.set_code as string);
      // Trigger auto-populate if we have an ID to fetch full details/prints
      if (data.scryfall_id && data.tcg === 'mtg') {
        setTimeout(() => handlePopulate(), 100);
      }
    } else {
      setForm({ ...EMPTY_BOUNTY, ...initialData });
      if (initialData?.collector_number) setCollectorNumber(initialData.collector_number);
      if (initialData?.scryfall_id && initialData?.tcg === 'mtg') {
        setTimeout(() => handlePopulate(), 100);
      }
    }
    setFormError('');
  }, [editBounty, initialData, handlePopulate]);


  const handlePriceSourceChange = async (src: PriceSource) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, form.collector_number || '', form.promo_type || '', form.foil_treatment);
    const { ref, suggested } = await resolvePrice(bestPrint, form.foil_treatment, src, form.card_treatment);
    setForm(f => {
      return {
        ...f,
        price_source: src,
        price_reference: ref,
        target_price: suggested !== undefined ? suggested : f.target_price
      };
    });
  };

  const handleSetSearchChange = async (newSet: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, newSet, form.card_treatment, form.collector_number || '', form.promo_type || '', form.foil_treatment);
    setSetCode(newSet);
    const { ref, suggested } = await resolvePrice(bestPrint, form.foil_treatment, form.price_source, form.card_treatment);
    setForm(f => ({
      ...f,
      set_name: bestPrint?.set_name || f.set_name,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price
    }));
  };

  const handleTreatmentChange = async (t: CardTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, t, form.collector_number || '', form.promo_type || '', form.foil_treatment);
    const foil = bestPrint?.finishes?.includes('foil') ? 'foil' as FoilTreatment : 'non_foil' as FoilTreatment;
    const { ref, suggested } = await resolvePrice(bestPrint, foil, form.price_source, t);

    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      card_treatment: t,
      collector_number: bestPrint?.collector_number || f.collector_number,
      promo_type: (bestPrint?.promo_types || []).join(',') || 'none',
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      scryfall_id: bestPrint?.id || '',
    }));
    if (bestPrint?.collector_number) setCollectorNumber(bestPrint.collector_number);
  };

  const handleArtChange = async (a: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, a, form.promo_type || '', form.foil_treatment);
    setCollectorNumber(a);
    const foil = resolveFoilTreatment(bestPrint);
    const { ref, suggested } = await resolvePrice(bestPrint, foil, form.price_source, form.card_treatment);

    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      collector_number: a,
      promo_type: (bestPrint?.promo_types || []).join(',') || 'none',
      foil_treatment: foil,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      scryfall_id: bestPrint?.id || '',
    }));
  };

  const handlePromoChange = async (p: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, form.collector_number || '', p, form.foil_treatment);
    const foil = resolveFoilTreatment(bestPrint);
    const { ref, suggested } = await resolvePrice(bestPrint, foil, form.price_source, form.card_treatment);

    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      promo_type: p,
      foil_treatment: foil,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : f.target_price,
      scryfall_id: bestPrint?.id || '',
    }));
  };

  const handleFoilChange = async (f: FoilTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, setCode, form.card_treatment, form.collector_number || '', form.promo_type || '', f);
    const { ref, suggested } = await resolvePrice(bestPrint, f, form.price_source, form.card_treatment);
    setForm(old => ({
      ...old,
      foil_treatment: f,
      image_url: getScryfallImage(bestPrint) || old.image_url,
      price_reference: ref,
      target_price: suggested !== undefined ? suggested : old.target_price,
      scryfall_id: bestPrint?.id || ''
    }));
  };


  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
      style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(12px)', overflowY: 'auto' }}>
      <div className="card p-0 w-full max-w-4xl mb-8 border-white/20 shadow-2xl animate-in fade-in zoom-in duration-300"
        style={{ position: 'relative', background: 'rgba(255, 255, 255, 0.85)', backdropFilter: 'blur(20px)' }}>

        <div className="flex items-center justify-between p-4 md:p-6 pb-2 border-b border-ink-border/20">
          <div className="flex flex-col">
            <h2 className="font-display text-2xl md:text-4xl m-0 tracking-tighter text-ink-deep">
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
                <select className="bg-white/50 border-white/40 w-full" value={form.tcg} onChange={e => {
                  const newTcg = e.target.value;
                  const newSource: PriceSource = newTcg === 'mtg' ? 'cardkingdom' : 'manual';
                  setForm(f => ({ ...f, tcg: newTcg, price_source: newSource }));
                }}>
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
                  {form.price_source === 'cardkingdom' && `(x ${settings?.ck_to_cop_rate || settings?.usd_to_cop_rate || 0} COP)`}
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
                    <option value="cardkingdom">{t('components.admin.bounty_modal.source_cardkingdom', 'External: CardKingdom (USD)')}</option>
                    <option value="tcgplayer">{t('components.admin.bounty_modal.source_tcgplayer', 'External: TCGPlayer (USD)')}</option>
                    <option value="cardmarket">{t('components.admin.bounty_modal.source_cardmarket', 'External: Cardmarket (EUR)')}</option>
                  </select>
                </div>

                {form.price_source !== 'manual' ? (
                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">
                      {t('components.admin.bounty_modal.ref_price_label', 'REFERENCE PRICE ({source}) *', { source: (form.price_source === 'tcgplayer' || form.price_source === 'cardkingdom') ? 'USD' : 'EUR' })}
                    </label>
                    <input
                      type="number" step="0.01" className="w-full font-mono bg-white/90 border-white/40"
                      value={form.price_reference || ''}
                      onChange={e => {
                        const val = parseFloat(e.target.value) || 0;
                        const rate = form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0)
                          : form.price_source === 'cardkingdom' ? (settings?.ck_to_cop_rate || settings?.usd_to_cop_rate || 0)
                            : (settings?.eur_to_cop_rate || 0);
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

              <div className="p-4 bg-ink-surface/30 rounded-sm border border-ink-border/50">
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.bounty_modal.is_generic', 'MATCH MODE')}</label>
                <div
                  onClick={() => setForm(f => ({ ...f, is_generic: !f.is_generic }))}
                  className={`flex items-center gap-3 p-2 rounded-lg border-2 cursor-pointer transition-all ${form.is_generic ? 'border-gold bg-gold/5' : 'border-ink-border/20 bg-white/50'}`}
                >
                  <div className={`w-10 h-6 rounded-full transition-colors relative flex-shrink-0 ${form.is_generic ? 'bg-gold' : 'bg-kraft-dark/40'}`}>
                    <div className={`w-4 h-4 bg-white rounded-full absolute top-1 transition-all shadow ${form.is_generic ? 'left-5' : 'left-1'}`} />
                  </div>
                  <span className="text-[10px] font-bold uppercase font-mono-stack tracking-widest">{form.is_generic ? t('pages.admin.bounties.generic_badge', 'ANY VERSION') : t('pages.common.labels.specific', 'SPECIFIC')}</span>
                </div>
              </div>
            </div>

            <div className="mt-4 flex items-center gap-2">
              <input type="checkbox" id="hide_price" checked={form.hide_price} onChange={e => setForm(f => ({ ...f, hide_price: e.target.checked }))} className="w-4 h-4 accent-gold" />
              <label htmlFor="hide_price" className="text-sm font-bold text-text-secondary cursor-pointer">{t('components.admin.bounty_modal.hide_price_label', 'Hide target price from public view')}</label>
            </div>
            <p className="text-xs text-text-muted pl-6">{t('components.admin.bounty_modal.hide_price_desc', 'If checked, users will see "Contact for Price" instead of your target COP value.')}</p>

            {editBounty && (
              <div className="mt-8 border-t border-ink-border/20 pt-6">
                <h3 className="text-xs font-mono-stack uppercase text-text-muted tracking-widest mb-4">
                  {t('components.admin.bounty_modal.linked_requests', 'LINKED CLIENT REQUESTS')} ({relatedRequests.length})
                </h3>
                {loadingRequests ? (
                  <div className="text-xs text-text-muted animate-pulse font-mono uppercase tracking-widest p-4 text-center">
                    {t('pages.common.labels.loading', 'Loading requests...')}
                  </div>
                ) : relatedRequests.length > 0 ? (
                  <div className="bg-white/50 rounded overflow-hidden border border-ink-border/10">
                    <table className="w-full text-[10px] font-mono-stack">
                      <thead className="bg-ink-surface/30">
                        <tr>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.common.labels.customer', 'Customer')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.profile.table.qty', 'Qty')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.profile.table.status', 'Status')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.common.labels.contact', 'Contact')}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-ink-border/10">
                        {relatedRequests.map((req: ClientRequest) => (
                          <tr key={req.id} className="hover:bg-gold/5">
                            <td className="px-3 py-2 font-bold">{req.customer_name}</td>
                            <td className="px-3 py-2">{req.quantity}</td>
                            <td className="px-3 py-2">
                              <span className={`px-1 rounded ${req.status === 'accepted' ? 'bg-green-100 text-green-700' :
                                  req.status === 'solved' ? 'bg-blue-100 text-blue-700' :
                                    'bg-gray-100 text-gray-700'
                                }`}>
                                {req.status.toUpperCase()}
                              </span>
                            </td>
                            <td className="px-3 py-2 text-text-muted">{req.customer_contact}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <div className="p-4 border border-dashed border-ink-border/20 rounded text-center text-[10px] text-text-muted font-mono uppercase tracking-widest">
                    {t('components.admin.bounty_modal.no_requests', 'No client requests linked to this bounty.')}
                  </div>
                )}
              </div>
            )}

            {editBounty && (
              <div className="mt-8 border-t border-ink-border/20 pt-6">
                <h3 className="text-xs font-mono-stack uppercase text-text-muted tracking-widest mb-4">
                  {t('components.admin.bounty_modal.linked_offers', 'LINKED CLIENT OFFERS')} ({relatedOffers.length})
                </h3>
                {loadingOffers ? (
                  <div className="text-xs text-text-muted animate-pulse font-mono uppercase tracking-widest p-4 text-center">
                    {t('pages.common.labels.loading', 'Loading offers...')}
                  </div>
                ) : relatedOffers.length > 0 ? (
                  <div className="bg-white/50 rounded overflow-hidden border border-ink-border/10">
                    <table className="w-full text-[10px] font-mono-stack">
                      <thead className="bg-ink-surface/30">
                        <tr>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.common.labels.customer', 'Customer')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.profile.table.qty', 'Qty')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.profile.table.condition', 'Cond')}</th>
                          <th className="px-3 py-2 text-left uppercase text-text-muted">{t('pages.profile.table.status', 'Status')}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-ink-border/10">
                        {relatedOffers.map((offer: BountyOffer) => (
                          <tr key={offer.id} className="hover:bg-gold/5">
                            <td className="px-3 py-2 font-bold">{offer.customer_name}</td>
                            <td className="px-3 py-2">{offer.quantity}</td>
                            <td className="px-3 py-2">{offer.condition}</td>
                            <td className="px-3 py-2">
                              <span className={`px-1 rounded ${offer.status === 'accepted' ? 'bg-green-100 text-green-700' :
                                  offer.status === 'fulfilled' ? 'bg-blue-100 text-blue-700' :
                                    'bg-gray-100 text-gray-700'
                                }`}>
                                {offer.status.toUpperCase()}
                              </span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <div className="p-4 border border-dashed border-ink-border/20 rounded text-center text-[10px] text-text-muted font-mono uppercase tracking-widest">
                    {t('components.admin.bounty_modal.no_offers', 'No client offers linked to this bounty.')}
                  </div>
                )}
              </div>
            )}
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
