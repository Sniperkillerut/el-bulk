'use client';

import { useEffect, useState } from 'react';
import {
  adminCreateProduct, adminUpdateProduct, adminUpdateProductStorage,
} from '@/lib/api';
import {
  Product, FoilTreatment, CardTreatment, PriceSource,
  StorageLocation, ScryfallCard, Condition
} from '@/lib/types';
import { 
  getTreatmentType, applyPrintPrices, extractMTGMetadata, findMatchingPrint, getScryfallImage, resolveFoilTreatment,
} from '@/lib/mtg-logic';

import ScryfallPopulate from './product/ScryfallPopulate';
import VariantTab from './product/tabs/VariantTab';
import PricingTab from './product/tabs/PricingTab';
import { FormState, TabId, ProductEditModalProps } from './product/types';

export const EMPTY_FORM: FormState = {
  name: '', tcg: 'mtg', category: 'singles',
  set_name: '', set_code: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  price: 0, price_source: 'tcgplayer', price_reference: '', price_cop_override: '',
  stock: 0, description: '', category_ids: [], image_url: '',
  collector_number: '', promo_type: '',
  language: 'en', color_identity: '', rarity: 'common', cmc: '',
  is_legendary: false, is_historic: false, is_land: false, is_basic_land: false,
  art_variation: '', oracle_text: '', artist: '', type_line: '',
  border_color: '', frame: '', full_art: false, textless: false,
  storage_items: []
};

export default function ProductEditModal({
  editProduct, token, storageLocations, categories, tcgs, settings,
  storageFilter, onClose, onSaved, onSaveAndNew
}: ProductEditModalProps) {
  const [activeTab, setActiveTab] = useState<TabId>('variant');
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [scryfallPrints, setScryfallPrints] = useState<ScryfallCard[]>([]);
  const [lookingUp, setLookingUp] = useState(false);
  const [productStorage, setProductStorage] = useState<StorageLocation[]>([]);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');

  // Initialize form from editProduct or defaults
  useEffect(() => {
    if (editProduct) {
      setForm({
        id: editProduct.id,
        name: editProduct.name,
        tcg: editProduct.tcg,
        category: editProduct.category as 'singles' | 'sealed' | 'accessories',
        set_name: editProduct.set_name || '',
        set_code: editProduct.set_code || '',
        condition: (editProduct.condition as Condition) || 'NM',
        foil_treatment: editProduct.foil_treatment,
        card_treatment: editProduct.card_treatment,
        price: editProduct.price || 0,
        price_source: editProduct.price_source || 'manual',
        price_reference: editProduct.price_reference ?? '',
        price_cop_override: editProduct.price_cop_override ?? '',
        stock: editProduct.stock,
        description: editProduct.description || '',
        category_ids: editProduct.categories?.map(c => c.id) || [],
        image_url: editProduct.image_url || '',
        collector_number: editProduct.collector_number || '',
        promo_type: editProduct.promo_type || '',
        language: editProduct.language || 'en',
        color_identity: editProduct.color_identity || '',
        rarity: editProduct.rarity || '',
        cmc: editProduct.cmc ?? '',
        is_legendary: editProduct.is_legendary,
        is_historic: editProduct.is_historic,
        is_land: editProduct.is_land,
        is_basic_land: editProduct.is_basic_land,
        art_variation: editProduct.art_variation || '',
        oracle_text: editProduct.oracle_text || '',
        artist: editProduct.artist || '',
        type_line: editProduct.type_line || '',
        border_color: editProduct.border_color || '',
        frame: editProduct.frame || '',
        full_art: editProduct.full_art,
        textless: editProduct.textless,
        storage_items: editProduct.stored_in?.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })) || [],
      });
      setProductStorage((editProduct.stored_in || []).map(d => ({
        stored_in_id: d.stored_in_id, name: d.name, quantity: d.quantity
      })));
    } else {
      setForm({ ...EMPTY_FORM });
      const initialStorage: StorageLocation[] = [];
      if (storageFilter) {
        const loc = storageLocations.find(l => l.id === storageFilter);
        if (loc) initialStorage.push({ stored_in_id: loc.id, name: loc.name, quantity: 0 });
      }
      setProductStorage(initialStorage);
    }
    setFormError('');
    setScryfallPrints([]);
  }, [editProduct, storageFilter, storageLocations]);

  const handleSave = async (andNew: boolean) => {
    if (!form.name || !form.tcg || !form.category) { 
      setFormError('Name, TCG, and Category are required.'); 
      return; 
    }
    setSaving(true); 
    setFormError('');
    try {
      const payload: Partial<Product> & { category_ids?: string[] } = {
        name: form.name, 
        tcg: form.tcg, 
        category: form.category,
        set_name: form.set_name || undefined, 
        set_code: form.set_code || undefined,
        condition: (form.condition || undefined) as Condition,
        foil_treatment: form.foil_treatment, 
        card_treatment: form.card_treatment,
        price_source: form.price_source,
        price_reference: form.price_reference === '' ? undefined : Number(form.price_reference),
        price_cop_override: form.price_cop_override === '' ? undefined : Number(form.price_cop_override),
        stock: form.stock, 
        image_url: form.image_url || undefined,
        description: form.description || undefined, 
        category_ids: form.category_ids,
        collector_number: form.collector_number || undefined, 
        promo_type: form.promo_type || undefined,
        language: form.language, 
        color_identity: form.color_identity || undefined,
        rarity: form.rarity || undefined, 
        cmc: form.cmc === '' ? undefined : Number(form.cmc),
        is_legendary: form.is_legendary, 
        is_historic: form.is_historic,
        is_land: form.is_land, 
        is_basic_land: form.is_basic_land,
        art_variation: form.art_variation || undefined, 
        oracle_text: form.oracle_text || undefined,
        artist: form.artist || undefined, 
        type_line: form.type_line || undefined,
        border_color: form.border_color || undefined, 
        frame: form.frame || undefined,
        full_art: form.full_art, 
        textless: form.textless,
      };
      
      if (payload.price_source === 'manual') payload.price_reference = undefined;
      else payload.price_cop_override = undefined;

      if (editProduct) {
        const updated = await adminUpdateProduct(token, editProduct.id, payload);
        await adminUpdateProductStorage(token, updated.id, productStorage.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })));
      } else {
        const newP = await adminCreateProduct(token, payload);
        await adminUpdateProductStorage(token, newP.id, productStorage.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })));
      }

      if (andNew && onSaveAndNew) {
        // Keep sticky fields for faster entry
        const stickyTcg = form.tcg;
        const stickyCategory = form.category;
        const stickyCondition = form.condition;
        
        onSaveAndNew({ 
          tcg: stickyTcg, 
          category: stickyCategory, 
          condition: stickyCondition, 
          storageIds: productStorage.map(s => s.stored_in_id) 
        });

        // Reset form but keep sticky values
        setForm({
          ...EMPTY_FORM,
          tcg: stickyTcg,
          category: stickyCategory,
          condition: stickyCondition,
        });
        setProductStorage(prev => prev.map(s => ({ ...s, quantity: 0 })));
        setScryfallPrints([]);
        setActiveTab('variant');
      } else {
        onSaved();
      }
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Failed to save product.');
    } finally { 
      setSaving(false); 
    }
  };

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') { 
        e.preventDefault(); 
        handleSave(false); 
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [onClose, form, productStorage]);

  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = 'unset'; };
  }, []);

  const isMTGSingles = form.tcg === 'mtg' && form.category === 'singles';

  const updateStoreQty = (id: string, delta: number) => {
    setProductStorage(prev => prev.map(loc => loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, loc.quantity + delta) } : loc));
  };
  const setStoreQty = (id: string, qty: number) => {
    setProductStorage(prev => prev.map(loc => loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, qty) } : loc));
  };

  const handlePopulate = async (forceSearchName?: string) => {
    setFormError('');
    const name = forceSearchName || form.name.trim();
    const set = form.set_code.trim().toLowerCase();
    const cn = form.collector_number.trim();
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
          const r: Response = await fetch(nextUrl);
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

      let prints: ScryfallCard[] = await fetchAllPrints(searchQ);
      if (prints.length === 0) throw new Error('No printings found for that search.');

      if (prints.length > 0) {
        const oracleId = (prints[0] as any).oracle_id;
        if (oracleId) {
          const oraclePrints = await fetchAllPrints(`oracle_id:${oracleId}`);
          if (oraclePrints.length > prints.length) prints = oraclePrints;
        }
      }

      setScryfallPrints(prints);
      
      let bestPrint = prints.find((p: ScryfallCard) => p?.set?.toLowerCase() === set && p?.collector_number === cn);
      if (!bestPrint && set) bestPrint = prints.find((p: ScryfallCard) => p?.set?.toLowerCase() === set);
      if (!bestPrint) bestPrint = prints[0];

      if (!bestPrint) throw new Error('Could not identify a matching print.');

      const initialTreatment = getTreatmentType(bestPrint);
      const initialFoil = resolveFoilTreatment(bestPrint);

      setForm(f => ({
        ...f, 
        name: bestPrint?.name || f.name,
        set_code: bestPrint?.set || f.set_code, 
        set_name: bestPrint?.set_name || f.set_name,
        card_treatment: initialTreatment, 
        collector_number: bestPrint?.collector_number || f.collector_number,
        promo_type: bestPrint?.promo_types?.join(',') || 'none',
        foil_treatment: initialFoil as FoilTreatment,
        image_url: getScryfallImage(bestPrint) || f.image_url,
        description: '', 
        price_reference: applyPrintPrices(bestPrint, initialFoil as FoilTreatment, f.price_source),
        ...extractMTGMetadata(bestPrint)
      }));
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Scryfall fetch failed');
    } finally { 
      setLookingUp(false); 
    }
  };

  const handleSetSearchChange = (newSet: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, newSet, form.card_treatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, set_code: newSet, set_name: bestPrint?.set_name || f.set_name,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleTreatmentChange = (newTreatment: CardTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, newTreatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, card_treatment: newTreatment,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleArtChange = (newArt: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, newArt, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, collector_number: newArt,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handlePromoChange = (newPromo: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, newPromo, form.foil_treatment);
    setForm(f => ({
      ...f, promo_type: newPromo,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleFoilChange = (newFoil: FoilTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, form.promo_type, newFoil);
    setForm(f => ({
      ...f, foil_treatment: newFoil,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, newFoil, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleSourceChange = (src: PriceSource) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, price_source: src,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, src)
    }));
  };

  const TABS: { id: TabId; label: string; show: boolean }[] = [
    { id: 'variant', label: 'VARIANT & IDENTITY', show: true },
    { id: 'pricing', label: 'PRICING & STOCK', show: true },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
      style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(12px)', overflowY: 'auto' }}>
      <div className="card p-0 w-full max-w-6xl mb-8 border-white/20 shadow-2xl animate-in fade-in zoom-in duration-300" 
        style={{ position: 'relative', background: 'rgba(255, 255, 255, 0.85)', backdropFilter: 'blur(20px)' }}>

        <div className="flex items-center justify-between p-4 md:p-6 pb-2">
          <div className="flex flex-col">
            <h2 className="font-display text-4xl m-0 tracking-tighter text-ink-deep">{editProduct ? 'EDIT PRODUCT' : 'NEW PRODUCT'}</h2>
            <p className="font-mono-stack text-[10px] text-text-muted opacity-50">PRODUCT ID: {form.id || 'NEW'}</p>
          </div>
          <button onClick={onClose} 
            className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-hp-color/10 text-text-muted hover:text-hp-color transition-all duration-300">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div className="px-4 md:px-6 pt-2 flex gap-4 flex-wrap">
          <div style={{ minWidth: '160px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>TCG SYSTEM</label>
            <select 
              className="bg-white/50 border-white/40 focus:bg-white transition-all" 
              value={form.tcg} 
              onChange={e => {
                const newTcg = e.target.value;
                setForm(f => ({ ...EMPTY_FORM, tcg: newTcg, category: f.category, condition: f.condition }));
                setScryfallPrints([]);
              }}
            >
              {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              <option value="accessories">Accessories</option>
            </select>
          </div>
          <div style={{ minWidth: '140px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>CATEGORY</label>
            <select 
              className="bg-white/50 border-white/40 focus:bg-white transition-all" 
              value={form.category} 
              onChange={e => {
                const newCat = e.target.value as 'singles' | 'sealed' | 'accessories';
                setForm(f => ({ ...EMPTY_FORM, tcg: f.tcg, category: newCat, condition: f.condition }));
                setScryfallPrints([]);
              }}
            >
              <option value="singles">Singles</option>
              <option value="sealed">Sealed</option>
              <option value="accessories">Accessories</option>
            </select>
          </div>
          <div style={{ minWidth: '100px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>CONDITION</label>
            <select className="bg-white/50 border-white/40 focus:bg-white transition-all" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value as Condition }))}>
              {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>
        </div>

        {isMTGSingles ? (
          <ScryfallPopulate 
            name={form.name} 
            setCode={form.set_code} 
            collectorNumber={form.collector_number}
            setName={form.set_name}
            scryfallPrints={scryfallPrints}
            lookingUp={lookingUp}
            onNameChange={val => { setForm(f => ({ ...f, name: val })); setScryfallPrints([]); setFormError(''); }}
            onSetCodeChange={val => { setForm(f => ({ ...f, set_code: val })); setFormError(''); }}
            onCollectorNumberChange={val => { handleArtChange(val); setFormError(''); }}
            onPopulate={() => handlePopulate()}
            onCardSelect={(card: ScryfallCard) => {
              setForm(f => ({ ...f, name: card.name }));
              handlePopulate(card.name);
            }}
            onSetSearchChange={handleSetSearchChange}
          />
        ) : (
          <div className="px-4 md:px-6 pt-4">
            <label className="text-xs font-mono-stack mb-1 block uppercase opacity-50" style={{ color: 'var(--text-muted)' }}>PRODUCT NAME</label>
            <input type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} style={{ fontSize: '1.25rem', fontWeight: 800, background: 'rgba(255,255,255,0.5)', border: '1px solid rgba(255,255,255,0.3)' }} placeholder="Entity Name" />
          </div>
        )}

        <div className="px-4 md:px-6 mt-6 flex gap-3 border-b border-ink-border/20">
          {TABS.filter(t => t.show).map(tab => {
            const isActive = activeTab === tab.id;
            return (
              <button 
                key={tab.id} 
                onClick={() => setActiveTab(tab.id)}
                className={`
                  font-mono-stack text-sm px-10 py-6 transition-all tracking-widest uppercase relative -mb-px rounded-t-md border-x border-t-[8px] group
                  ${isActive 
                    ? 'text-gold bg-white border-gold border-x-ink-border/20 border-b-white z-20 shadow-[0_0_25px_rgba(186,155,74,0.5),0_0_12px_rgba(186,155,74,0.3)] font-black' 
                    : 'text-text-muted bg-kraft-dark/30 border-transparent hover:text-ink-deep hover:bg-kraft-dark/50 font-bold'
                  }
                `}
              >
                <div className="flex items-center gap-4">
                  <span className={`transition-all duration-300 w-5 flex justify-center ${isActive ? 'opacity-100 scale-125 text-gold' : 'opacity-20 group-hover:opacity-100'}`}>◈</span>
                  {tab.label}
                  {/* Balanced Spacer */}
                  <div className="w-5" />
                </div>
              </button>
            );
          })}
        </div>

        <div className="flex gap-8 flex-col md:flex-row p-4 md:p-6 pt-6">
          <div className="flex-1 min-w-0 animate-in fade-in slide-in-from-bottom-2 duration-500">
            {activeTab === 'variant' && (
              <VariantTab 
                form={form} 
                prints={scryfallPrints} 
                isMTGSingles={isMTGSingles}
                onUpdate={u => setForm(f => ({ ...f, ...u }))}
                onTreatmentChange={handleTreatmentChange}
                onArtChange={handleArtChange}
                onPromoChange={handlePromoChange}
                onFoilChange={handleFoilChange}
              />
            )}
            {activeTab === 'pricing' && (
              <PricingTab 
                form={form} 
                settings={settings} 
                categories={categories}
                productStorage={productStorage}
                storageLocations={storageLocations}
                onUpdate={u => setForm(f => ({ ...f, ...u }))}
                onSourceChange={handleSourceChange}
                onUpdateStoreQty={updateStoreQty}
                onSetStoreQty={setStoreQty}
                onRemoveStorage={id => setProductStorage(prev => prev.filter(p => p.stored_in_id !== id))}
                onAddStorage={id => {
                  const loc = storageLocations.find(l => l.id === id);
                  if (loc) setProductStorage(prev => [...prev, { stored_in_id: loc.id, name: loc.name, quantity: 0 }]);
                }}
              />
            )}
          </div>

          <div className="w-full md:w-80 shrink-0">
            <div className="flex justify-between items-center mb-3">
               <label className="text-[10px] font-mono-stack uppercase tracking-tighter opacity-50" style={{ color: 'var(--text-muted)' }}>IMAGE PREVIEW</label>
               <span className="text-[10px] font-mono-stack px-2 py-0.5 rounded-full font-bold shadow-sm" style={{ background: 'var(--nm-color)', color: 'white' }}>{form.condition}</span>
            </div>
            <div className="card p-2 bg-white/40 border-white/30 backdrop-blur-sm overflow-hidden group mb-8 shadow-xl">
              <div className="relative aspect-[63/88] w-full bg-ink-border/5 rounded shadow-inner flex items-center justify-center overflow-hidden">
                {form.image_url ? (
                  <img src={form.image_url} alt={form.name} className="w-full h-full object-contain drop-shadow-2xl transition-transform duration-700 group-hover:scale-110" />
                ) : (
                  <div className="text-[10px] font-mono-stack text-text-muted text-center p-4">NO IMAGE FOUND<br/>SEARCH TO LOAD</div>
                )}
                {form.foil_treatment !== 'non_foil' && (
                  <div className="absolute inset-0 pointer-events-none opacity-30 mix-blend-color-dodge transition-opacity group-hover:opacity-50"
                    style={{ background: 'linear-gradient(135deg, rgba(255,255,255,0) 0%, rgba(255,255,255,0.6) 50%, rgba(255,255,255,0) 100%)', backgroundSize: '200% 200%', animation: 'shimmer 3s infinite linear' }} />
                )}
              </div>
            </div>

            {formError && (
              <div className="p-4 mb-6 bg-hp-color/5 border border-hp-color/20 text-hp-color text-[11px] rounded font-mono-stack animate-in slide-in-from-right duration-300">
                <span className="font-bold mr-2">! ERROR:</span> {formError}
              </div>
            )}

            <div className="space-y-4">
              <button 
                onClick={() => handleSave(false)} 
                disabled={saving} 
                className="btn-primary w-full py-5 text-sm font-bold shadow-[0_10px_30px_rgba(212,175,55,0.3)] hover:shadow-[0_15px_40px_rgba(212,175,55,0.4)] transition-all active:scale-95"
              >
                {saving ? '📥 SAVING...' : '💾 SAVE PRODUCT'}
              </button>
              <button 
                onClick={() => handleSave(true)} 
                disabled={saving} 
                className="btn-secondary w-full py-4 text-[10px] font-bold font-mono-stack tracking-widest border-ink-border/20"
              >
                💾 SAVE & ADD NEW
              </button>
              <button onClick={onClose} disabled={saving} className="w-full py-2 text-[9px] font-mono-stack text-text-muted hover:text-hp-color transition-colors tracking-widest opacity-60">
                CANCEL
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
