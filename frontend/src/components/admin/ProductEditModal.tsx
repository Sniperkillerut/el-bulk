'use client';

import { useEffect, useState } from 'react';
import {
  adminCreateProduct, adminUpdateProduct, adminUpdateProductStorage,
} from '@/lib/api';
import {
  Product, FoilTreatment, CardTreatment, PriceSource, Settings,
  StoredIn, StorageLocation, CustomCategory, TCG, ScryfallCard
} from '@/lib/types';
import { 
  getTreatmentType, applyPrintPrices, extractMTGMetadata, findMatchingPrint, getScryfallImage, resolveFoilTreatment,
} from '@/lib/mtg-logic';

import ScryfallPopulate from './product/ScryfallPopulate';
import VariantTab from './product/tabs/VariantTab';
import PricingTab from './product/tabs/PricingTab';
import DetailsTab from './product/tabs/DetailsTab';

export interface FormState {
  name: string; tcg: string; category: 'singles' | 'sealed' | 'accessories';
  set_name: string; set_code: string; condition: string;
  foil_treatment: FoilTreatment; card_treatment: CardTreatment;
  price_source: PriceSource; price_reference: number | '';
  price_cop_override: number | ''; stock: number; description: string;
  category_ids: string[]; image_url: string; collector_number: string;
  promo_type: string; language: string; color_identity: string;
  rarity: string; cmc: number | ''; is_legendary: boolean;
  is_historic: boolean; is_land: boolean; is_basic_land: boolean;
  art_variation: string; oracle_text: string; artist: string;
  type_line: string; border_color: string; frame: string;
  full_art: boolean; textless: boolean;
}

export const EMPTY_FORM: FormState = {
  name: '', tcg: 'mtg', category: 'singles',
  set_name: '', set_code: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  price_source: 'tcgplayer', price_reference: '', price_cop_override: '',
  stock: 0, description: '', category_ids: [], image_url: '',
  collector_number: '', promo_type: '',
  language: 'en', color_identity: '', rarity: 'common', cmc: '',
  is_legendary: false, is_historic: false, is_land: false, is_basic_land: false,
  art_variation: '', oracle_text: '', artist: '', type_line: '',
  border_color: '', frame: '', full_art: false, textless: false,
};

type TabId = 'variant' | 'pricing' | 'details';

interface ProductEditModalProps {
  editProduct: Product | null;
  token: string;
  storageLocations: StoredIn[];
  categories: CustomCategory[];
  tcgs: TCG[];
  settings: Settings | undefined;
  storageFilter: string;
  onClose: () => void;
  onSaved: () => void;
  onSaveAndNew: (lastForm: { tcg: string; category: string; condition: string; storageIds: string[] }) => void;
}

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
        name: editProduct.name, tcg: editProduct.tcg,
        category: editProduct.category as 'singles' | 'sealed' | 'accessories',
        set_name: editProduct.set_name || '', set_code: editProduct.set_code || '',
        condition: editProduct.condition || 'NM',
        foil_treatment: editProduct.foil_treatment, card_treatment: editProduct.card_treatment,
        price_source: editProduct.price_source || 'manual',
        price_reference: editProduct.price_reference ?? '',
        price_cop_override: editProduct.price_cop_override ?? '',
        stock: editProduct.stock, description: editProduct.description || '',
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
    if (!form.name || !form.tcg || !form.category) { setFormError('Name, TCG, and Category are required.'); return; }
    setSaving(true); setFormError('');
    try {
      const payload: Partial<Product> & { category_ids?: string[] } = {
        name: form.name, tcg: form.tcg, category: form.category,
        set_name: form.set_name || undefined, set_code: form.set_code || undefined,
        condition: (form.condition || undefined) as Product['condition'],
        foil_treatment: form.foil_treatment, card_treatment: form.card_treatment,
        price_source: form.price_source,
        price_reference: form.price_reference === '' ? undefined : Number(form.price_reference),
        price_cop_override: form.price_cop_override === '' ? undefined : Number(form.price_cop_override),
        stock: form.stock, image_url: form.image_url || undefined,
        description: form.description || undefined, category_ids: form.category_ids,
        collector_number: form.collector_number || undefined, promo_type: form.promo_type || undefined,
        language: form.language, color_identity: form.color_identity || undefined,
        rarity: form.rarity || undefined, cmc: form.cmc === '' ? undefined : Number(form.cmc),
        is_legendary: form.is_legendary, is_historic: form.is_historic,
        is_land: form.is_land, is_basic_land: form.is_basic_land,
        art_variation: form.art_variation || undefined, oracle_text: form.oracle_text || undefined,
        artist: form.artist || undefined, type_line: form.type_line || undefined,
        border_color: form.border_color || undefined, frame: form.frame || undefined,
        full_art: form.full_art, textless: form.textless,
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
      if (andNew) {
        onSaveAndNew({ tcg: form.tcg, category: form.category, condition: form.condition, storageIds: productStorage.map(s => s.stored_in_id) });
      } else {
        onSaved();
      }
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Failed to save product.');
    } finally { setSaving(false); }
  };

  // Keyboard shortcuts
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

  // Prevent body scroll
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

  const handlePopulate = async () => {
    const name = form.name.trim();
    const set = form.set_code.trim().toLowerCase();
    const cn = form.collector_number.trim();
    if (!name && (!set || !cn)) return;

    setLookingUp(true); setFormError('');
    try {
      let searchQ = "";
      if (set && cn) searchQ = `set:${set} cn:"${cn}"`;
      else if (name) searchQ = `!"${name}"`;
      else if (set) searchQ = `set:${set}`;

      const fetchAllPrints = async (q: string) => {
        let results: ScryfallCard[] = [];
        // Use game:paper to ensure only physical prints are included
        let nextUrl: string | null = `https://api.scryfall.com/cards/search?q=${encodeURIComponent(q)}+game:paper&unique=prints&order=released`;
        
        while (nextUrl) {
          const r: Response = await fetch(nextUrl);
          if (!r.ok) break;
          const b: any = await r.json();
          // Explicitly filter for non-digital cards as a secondary safety measure
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

      // If we found prints and have an oracle ID, ensure we have ALL prints for that oracle ID
      // (sometimes a broad name search or set/cn search misses variants)
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
        set_name: bestPrint?.set_name || f.set_name || f.set_name,
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
    } finally { setLookingUp(false); }
  };

  const handleSetSearchChange = (newSet: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, newSet, form.card_treatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, set_code: newSet, set_name: bestPrint?.set_name || f.set_name,
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleTreatmentChange = (newTreatment: CardTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, newTreatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, card_treatment: newTreatment,
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleArtChange = (newArt: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, newArt, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, collector_number: newArt,
      foil_treatment: resolveFoilTreatment(bestPrint),
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handlePromoChange = (newPromo: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, newPromo, form.foil_treatment);
    setForm(f => ({
      ...f, promo_type: newPromo,
      foil_treatment: resolveFoilTreatment(bestPrint),
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
    { id: 'details', label: isMTGSingles ? 'DETAILS & METADATA' : 'DETAILS', show: true },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
      style={{ background: 'rgba(0,0,0,0.7)', backdropFilter: 'blur(3px)', overflowY: 'auto' }}>
      <div className="card no-tilt p-0 w-full max-w-6xl mb-8" style={{ position: 'relative' }}>

        <div className="flex items-center justify-between p-4 md:p-6 pb-0">
          <h2 className="font-display text-3xl m-0">{editProduct ? 'EDIT PRODUCT' : 'NEW PRODUCT'}</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div className="px-4 md:px-6 pt-4 flex gap-4 flex-wrap">
          <div style={{ minWidth: '160px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>TCG *</label>
            <select value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
              {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              <option value="accessories">Accessories</option>
            </select>
          </div>
          <div style={{ minWidth: '140px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>CATEGORY *</label>
            <select value={form.category} onChange={e => setForm(f => ({ ...f, category: e.target.value as 'singles' | 'sealed' | 'accessories' }))}>
              <option value="singles">Singles</option>
              <option value="sealed">Sealed</option>
              <option value="accessories">Accessories</option>
            </select>
          </div>
          <div style={{ minWidth: '100px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>CONDITION</label>
            <select value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value }))}>
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
            onNameChange={val => { setForm(f => ({ ...f, name: val })); setScryfallPrints([]); }}
            onSetCodeChange={val => setForm(f => ({ ...f, set_code: val }))}
            onCollectorNumberChange={handleArtChange}
            onPopulate={handlePopulate}
            onSetSearchChange={handleSetSearchChange}
          />
        ) : (
          <div className="px-4 md:px-6 pt-4">
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRODUCT NAME *</label>
            <input type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} style={{ fontSize: '1rem', fontWeight: 600 }} placeholder="Product name" />
          </div>
        )}

        <div className="px-4 md:px-6 mt-4 flex gap-1 border-b border-ink-border">
          {TABS.filter(t => t.show).map(tab => (
            <button key={tab.id} onClick={() => setActiveTab(tab.id)}
              className="font-mono-stack text-xs px-4 py-2 transition-all"
              style={{
                background: activeTab === tab.id ? 'var(--ink-surface)' : 'transparent',
                color: activeTab === tab.id ? 'var(--gold)' : 'var(--text-muted)',
                fontWeight: activeTab === tab.id ? 700 : 400,
                cursor: 'pointer',
                borderTopWidth: 0, borderLeftWidth: 0, borderRightWidth: 0,
                borderBottomWidth: '2px', borderBottomStyle: 'solid',
                borderBottomColor: activeTab === tab.id ? 'var(--gold)' : 'transparent',
              }}>
              {tab.label}
            </button>
          ))}
        </div>

        <div className="flex gap-6 flex-col md:flex-row p-4 md:p-6 pt-4">
          <div className="flex-1 min-w-0">
            {activeTab === 'variant' && (
              <VariantTab 
                form={form} 
                prints={scryfallPrints} 
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
            {activeTab === 'details' && (
              <DetailsTab 
                form={form} 
                isMTGSingles={isMTGSingles} 
                categories={categories}
                onUpdate={u => setForm(f => ({ ...f, ...u }))}
              />
            )}
          </div>

          <div className="w-full md:w-80 shrink-0">
            <div className="flex justify-between items-center mb-2">
               <label className="text-[10px] font-mono-stack uppercase" style={{ color: 'var(--text-muted)' }}>PREVIEW</label>
               <span className="text-[10px] font-mono-stack px-1.5 py-0.5 rounded" style={{ background: 'var(--nm-color)', color: 'white' }}>{form.condition}</span>
            </div>
            <div className="card no-tilt p-2 bg-ink-surface border-ink-border overflow-hidden group mb-6">
              <div className="relative aspect-[63/88] w-full bg-ink-border/20 rounded shadow-inner flex items-center justify-center overflow-hidden">
                {form.image_url ? (
                  <img src={form.image_url} alt={form.name} className="w-full h-full object-contain" />
                ) : (
                  <div className="text-[10px] font-mono-stack text-text-muted text-center p-4">NO IMAGE<br/>POPULATE TO LOAD</div>
                )}
                {form.foil_treatment !== 'non_foil' && (
                  <div className="absolute inset-0 pointer-events-none opacity-40 mix-blend-color-dodge transition-opacity group-hover:opacity-60"
                    style={{ background: 'linear-gradient(135deg, rgba(255,255,255,0) 0%, rgba(255,255,255,0.4) 50%, rgba(255,255,255,0) 100%)', backgroundSize: '200% 200%', animation: 'shimmer 3s infinite linear' }} />
                )}
              </div>
            </div>

            {formError && (
              <div className="p-3 mb-4 bg-hp-color/10 border border-hp-color/30 text-hp-color text-xs rounded-sm font-mono-stack animate-pulse">
                ERR: {formError}
              </div>
            )}

            <div className="space-y-3">
              <button 
                onClick={() => handleSave(false)} 
                disabled={saving} 
                className="btn-primary w-full py-4 text-sm font-bold shadow-lg"
              >
                {saving ? '📥 SAVING...' : '💾 SAVE PRODUCT'}
              </button>
              <button 
                onClick={() => handleSave(true)} 
                disabled={saving} 
                className="btn-secondary w-full py-3 text-xs font-mono-stack"
              >
                💾 SAVE & NEW
              </button>
              <button onClick={onClose} disabled={saving} className="btn-secondary w-full py-2 text-xs opacity-50 hover:opacity-100">
                CANCEL
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
