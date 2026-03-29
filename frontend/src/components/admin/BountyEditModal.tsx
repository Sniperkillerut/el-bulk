'use client';

import { useEffect, useState } from 'react';
import { adminCreateBounty, adminUpdateBounty } from '@/lib/api';
import { Bounty, BountyInput, FoilTreatment, Condition, TCG, ScryfallCard } from '@/lib/types';
import { extractMTGMetadata, getScryfallImage, resolveFoilTreatment } from '@/lib/mtg-logic';
import ScryfallPopulate from './product/ScryfallPopulate';
import MTGVariantSelector from './MTGVariantSelector';

interface BountyEditModalProps {
  editBounty: Bounty | null;
  token: string;
  tcgs: TCG[];
  onClose: () => void;
  onSaved: () => void;
  initialData?: Partial<BountyInput>;
}

export const EMPTY_BOUNTY: BountyInput = {
  name: '', tcg: 'mtg', set_name: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  collector_number: '', promo_type: '', language: 'en',
  target_price: 0, hide_price: false, quantity_needed: 1, image_url: ''
};

export default function BountyEditModal({
  editBounty, token, tcgs, onClose, onSaved, initialData
}: BountyEditModalProps) {
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
      });
    } else {
      setForm({ ...EMPTY_BOUNTY, ...initialData });
    }
    setFormError('');
    setScryfallPrints([]);
  }, [editBounty, initialData]);

  const handleSave = async () => {
    if (!form.name || !form.tcg) {
      setFormError('Name and TCG are required.');
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
        await adminUpdateBounty(token, editBounty.id, payload);
      } else {
        await adminCreateBounty(token, payload);
      }
      onSaved();
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Failed to save bounty.');
    } finally {
      setSaving(false);
    }
  };

  const handlePopulate = async () => {
    const name = form.name.trim();
    const set = setCode.trim().toLowerCase();
    const cn = collectorNumber.trim();
    if (!name && (!set || !cn)) return;

    setLookingUp(true);
    setFormError('');
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
          const b: any = await r.json();
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
        const oracleId = (prints[0] as any).oracle_id;
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

      setForm(f => ({
        ...f,
        name: bestPrint?.name || f.name,
        set_name: bestPrint?.set_name || f.set_name,
        image_url: getScryfallImage(bestPrint) || f.image_url,
      }));
      setSetCode(bestPrint?.set || setCode);
      setCollectorNumber(bestPrint?.collector_number || collectorNumber);

    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Scryfall fetch failed');
    } finally {
      setLookingUp(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
      style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(12px)', overflowY: 'auto' }}>
      <div className="card p-0 w-full max-w-4xl mb-8 border-white/20 shadow-2xl animate-in fade-in zoom-in duration-300"
        style={{ position: 'relative', background: 'rgba(255, 255, 255, 0.85)', backdropFilter: 'blur(20px)' }}>

        <div className="flex items-center justify-between p-4 md:p-6 pb-2 border-b border-ink-border/20">
          <div className="flex flex-col">
            <h2 className="font-display text-4xl m-0 tracking-tighter text-ink-deep">{editBounty ? 'EDIT WANTED CARD' : 'ADD WANTED CARD'}</h2>
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
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">TCG SYSTEM</label>
                <select className="bg-white/50 border-white/40 w-full" value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
                  {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                </select>
              </div>
              <div className="flex-1">
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">CONDITION</label>
                <select className="bg-white/50 border-white/40 w-full" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value as Condition }))}>
                  <option value="">Any</option>
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
                  onNameChange={val => { setForm(f => ({ ...f, name: val })); setScryfallPrints([]); }}
                  onSetCodeChange={val => setSetCode(val)}
                  onCollectorNumberChange={val => setCollectorNumber(val)}
                  onPopulate={handlePopulate}
                  onSetSearchChange={val => {
                    setSetCode(val);
                    const p = scryfallPrints.find(pr => pr.set === val);
                    if (p) {
                      setForm(f => ({ ...f, set_name: p.set_name, image_url: getScryfallImage(p) }));
                    }
                  }}
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
                      onTreatmentChange={t => setForm(f => ({ ...f, card_treatment: t }))}
                      onArtChange={a => { 
                        setCollectorNumber(a);
                        setForm(f => ({ ...f, collector_number: a }));
                      }}
                      onPromoChange={p => setForm(old => ({ ...old, promo_type: p }))}
                      onFoilChange={f => setForm(old => ({ ...old, foil_treatment: f }))}
                    />
                  </div>
                )}
              </>
            ) : (
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">CARD NAME</label>
                <input type="text" className="w-full text-lg font-bold" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Entity Name" />
                
                <label className="text-[10px] font-mono-stack mt-4 mb-1 block uppercase text-text-muted">SET NAME / INFO</label>
                <input type="text" className="w-full" value={form.set_name} onChange={e => setForm(f => ({ ...f, set_name: e.target.value }))} placeholder="Optional Set Name" />
                
                <label className="text-[10px] font-mono-stack mt-4 mb-1 block uppercase text-text-muted">IMAGE URL</label>
                <input type="text" className="w-full" value={form.image_url} onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))} placeholder="https://..." />
              </div>
            )}

            <div className="grid grid-cols-2 gap-4 mt-6">
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">TARGET PRICE (COP)</label>
                <input type="number" min="0" step="50" className="w-full font-mono text-lg" value={form.target_price || ''} onChange={e => setForm(f => ({ ...f, target_price: parseFloat(e.target.value) }))} placeholder="0" />
              </div>
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">QUANTITY NEEDED</label>
                <input type="number" min="1" className="w-full font-mono text-lg" value={form.quantity_needed} onChange={e => setForm(f => ({ ...f, quantity_needed: parseInt(e.target.value) || 1 }))} />
              </div>
            </div>

            <div className="mt-4 flex items-center gap-2">
              <input type="checkbox" id="hide_price" checked={form.hide_price} onChange={e => setForm(f => ({ ...f, hide_price: e.target.checked }))} className="w-4 h-4 accent-gold" />
              <label htmlFor="hide_price" className="text-sm font-bold text-text-secondary cursor-pointer">Hide target price from public view</label>
            </div>
            <p className="text-xs text-text-muted pl-6">If checked, users will see "Contact for Price" instead of your target COP value.</p>
          </div>

          <div className="w-full md:w-64 shrink-0 flex flex-col items-center">
            <div className="relative aspect-[63/88] w-full max-w-[200px] bg-ink-border/5 rounded shadow-inner flex items-center justify-center overflow-hidden mb-6">
              {form.image_url ? (
                <img src={form.image_url} alt="Preview" className="w-full h-full object-contain" />
              ) : (
                <div className="text-[10px] font-mono-stack text-text-muted text-center p-4">NO IMAGE</div>
              )}
            </div>

            {formError && (
              <div className="p-3 mb-4 bg-hp-color/10 text-hp-color text-xs rounded w-full">
                <strong>Error:</strong> {formError}
              </div>
            )}

            <button onClick={handleSave} disabled={saving} className="btn-primary w-full py-3 mb-2">
              {saving ? 'SAVING...' : '💾 SAVE BOUNTY'}
            </button>
            <button onClick={onClose} disabled={saving} className="btn-secondary w-full py-2 border-none">
              CANCEL
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
