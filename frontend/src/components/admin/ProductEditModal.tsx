'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  adminCreateProduct, adminUpdateProduct, adminUpdateProductStorage,
} from '@/lib/api';
import {
  Product, FOIL_LABELS, TREATMENT_LABELS, KNOWN_TCGS, TCG_SHORT,
  FoilTreatment, CardTreatment, PriceSource, Settings,
  StoredIn, StorageLocation, CustomCategory
} from '@/lib/types';

interface FormState {
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

const EMPTY_FORM: FormState = {
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
  tcgs: any[];
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
  const [scryfallPrints, setScryfallPrints] = useState<any[]>([]);
  const [lookingUp, setLookingUp] = useState(false);
  const [productStorage, setProductStorage] = useState<StorageLocation[]>([]);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');

  // Initialize form from editProduct or defaults
  useEffect(() => {
    if (editProduct) {
      setForm({
        name: editProduct.name, tcg: editProduct.tcg,
        category: editProduct.category as typeof EMPTY_FORM['category'],
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

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') { e.preventDefault(); handleSave(false); }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose, form, productStorage]);

  // Prevent body scroll
  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = 'unset'; };
  }, []);

  const isMTGSingles = form.tcg === 'mtg' && form.category === 'singles';

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


  const resolveLabel = (key: string, map: Record<string, string>) => {
    if (!key || key === 'none') return 'Standard / Regular';
    if (map[key]) return map[key];
    let label = key.replace(/([a-z])([A-Z])/g, '$1 $2');
    label = label.replace(/oilslick/gi, 'Oil Slick').replace(/stepandcompleat/gi, 'Step-and-Compleat');
    label = label.replace(/silverfoil/gi, 'Silver Foil').replace(/foil$/, ' Foil');
    label = label.replace(/_foil$/, ' Foil').replace(/_/g, ' ');
    return label.replace(/\b\w/g, l => l.toUpperCase()).trim();
  };

  const getTreatmentType = (print: any): CardTreatment => {
    const fx = print.frame_effects || [];
    const pt = print.promo_types || [];
    if (print.serialized) return 'serialized';
    if (pt.includes('stepandcompleat')) return 'step_and_compleat';
    if (pt.includes('galaxyfoil')) return 'galaxy_foil';
    if (fx.includes('showcase')) return 'showcase';
    if (fx.includes('extendedart')) return 'extended_art';
    if (print.full_art) return 'full_art';
    if (print.border_color === 'borderless') return 'borderless';
    if (fx.includes('retro') || ['1993', '1997'].includes(print.frame)) return 'retro_frame';
    if (print.frame === '2003') return 'legacy_border';
    if (print.textless) return 'textless';
    if (pt.includes('judgegift')) return 'judge_promo';
    if (print.promo) return 'promo';
    const ignoredPromos = ['boosterfun', 'promopack', 'setpromo', 'buyabox'];
    const validPt = pt.filter((p: string) => !ignoredPromos.includes(p) && !p.includes('foil'));
    if (validPt.length > 0) return validPt[0];
    if (fx.length > 0) return fx[0];
    return 'normal';
  };

  const extractPrices = (print: any, foil: FoilTreatment) => {
    const isEtched = foil === 'etched_foil';
    const isFoil = foil !== 'non_foil' && !isEtched;
    let usd = parseFloat(isEtched ? print?.prices?.usd_etched || print?.prices?.usd_foil || print?.prices?.usd : isFoil ? print?.prices?.usd_foil || print?.prices?.usd : print?.prices?.usd);
    let eur = parseFloat(isFoil || isEtched ? print?.prices?.eur_foil || print?.prices?.eur : print?.prices?.eur);
    return { usd: isNaN(usd) ? null : usd, eur: isNaN(eur) ? null : eur };
  };

  const applyPrintPrices = (print: any, foil: FoilTreatment, src: PriceSource) => {
    const { usd, eur } = extractPrices(print, foil);
    if (src === 'tcgplayer') return usd ?? 0;
    if (src === 'cardmarket') return eur ?? 0;
    return 0;
  };

  const extractMTGMetadata = (card: any) => {
    const typeLine = card?.type_line || '';
    const isLegendary = card?.is_legendary ?? (typeLine.toLowerCase().includes('legendary') || false);
    return {
      language: card?.language || card?.lang || 'en',
      color_identity: Array.isArray(card?.color_identity) ? card.color_identity.join(',') : (card?.color_identity || ''),
      rarity: card?.rarity || 'common',
      cmc: card?.cmc ?? 0,
      is_legendary: isLegendary,
      is_historic: card?.is_historic ?? (isLegendary || typeLine.toLowerCase().includes('artifact') || typeLine.toLowerCase().includes('saga') || false),
      is_land: card?.is_land ?? (typeLine.toLowerCase().includes('land') || false),
      is_basic_land: card?.is_basic_land ?? (typeLine.toLowerCase().includes('basic land') || false),
      art_variation: card?.art_variation || (card?.variation ? 'Variation' : ''),
      oracle_text: card?.oracle_text || '',
      artist: card?.artist || '',
      type_line: typeLine,
      border_color: card?.border_color || '',
      frame: card?.frame || '',
      full_art: !!card?.full_art,
      textless: !!card?.textless,
    };
  };

  const getTreatmentOptions = (setCode?: string) => {
    const targetSet = setCode || form.set_code;
    if (scryfallPrints.length === 0) return [{ value: form.card_treatment, label: resolveLabel(form.card_treatment, TREATMENT_LABELS) }];
    const options = new Map<string, string>();
    if (!setCode) options.set(form.card_treatment, resolveLabel(form.card_treatment, TREATMENT_LABELS));
    scryfallPrints.forEach(p => {
      if (p.set === targetSet || !targetSet) {
        const type = getTreatmentType(p);
        options.set(type, resolveLabel(type, TREATMENT_LABELS));
      }
    });
    return Array.from(options.entries()).map(([value, label]) => ({ value, label }));
  };

  const findMatchingPrint = (setCode: string, treatment: CardTreatment, collectorNumber: string, promoType: string, foil?: FoilTreatment) => {
    const targetSet = setCode || form.set_code;
    const match = scryfallPrints.find((p: any) => {
      if (!p) return false;
      const t = getTreatmentType(p);
      const finishes = p.finishes || [];
      const pt = p.promo_types || [];
      const ptKey = pt.join(',') || 'none';
      const hasFoil = !foil || (foil === 'non_foil' && finishes.includes('nonfoil')) ||
                      (foil === 'etched_foil' && finishes.includes('etched')) ||
                      (finishes.includes('foil') || pt.includes(foil || ''));
      return p.set === targetSet && t === treatment && p.collector_number === collectorNumber && ptKey === promoType && hasFoil;
    });
    if (match) return match;
    
    // Fallbacks
    return scryfallPrints.find((p: any) => p?.set === targetSet && p?.collector_number === collectorNumber)
       || scryfallPrints.find((p: any) => p?.set === targetSet)
       || scryfallPrints[0];
  };

  const getArtOptions = (treatment: CardTreatment, setCode?: string) => {
    const targetSet = setCode || form.set_code;
    const prints = scryfallPrints.filter(p => (p.set === targetSet || !targetSet) && getTreatmentType(p) === treatment);
    const seen = new Set();
    const results: { value: string, label: string }[] = [];
    prints.forEach(p => {
      if (!seen.has(p.collector_number)) {
        seen.add(p.collector_number);
        results.push({ value: p.collector_number, label: `Art by ${p.artist} (#${p.collector_number})` });
      }
    });
    if (results.length === 0 && form.collector_number) {
      results.push({ value: form.collector_number, label: form.artist ? `Art by ${form.artist} (#${form.collector_number})` : `Art #${form.collector_number}` });
    }
    return results;
  };

  const getPromoOptions = (treatment: CardTreatment, setCode: string, collectorNumber: string) => {
    const prints = scryfallPrints.filter(p => p.collector_number === collectorNumber && p.set === setCode && getTreatmentType(p) === treatment);
    const seen = new Set();
    const results: { value: string, label: string }[] = [];
    prints.forEach(p => {
      const pt = p.promo_types || [];
      const ptKey = pt.join(',') || 'none';
      if (!seen.has(ptKey)) {
        seen.add(ptKey);
        results.push({ value: ptKey, label: pt.length > 0 ? pt.map((t: string) => resolveLabel(t, FOIL_LABELS)).join(', ') : 'Standard / Regular' });
      }
    });
    if (results.length === 0) {
      results.push({ value: form.promo_type || 'none', label: (!form.promo_type || form.promo_type === 'none') ? 'Standard / Regular' : resolveLabel(form.promo_type, FOIL_LABELS) });
    }
    return results;
  };

  const getFoilOptions = (treatment: CardTreatment, setCode?: string, collectorNumber?: string, promoType?: string) => {
    const targetSet = setCode || form.set_code;
    const targetNum = collectorNumber || form.collector_number;
    const targetPromo = promoType || form.promo_type;
    const matchingPrints = scryfallPrints.filter(p =>
      getTreatmentType(p) === treatment && (p.set === targetSet || !targetSet) &&
      (p.collector_number === targetNum || !targetNum) &&
      ((p.promo_types?.join(',') || 'none') === targetPromo || !targetPromo)
    );
    const results: { value: FoilTreatment, label: string }[] = [];
    const seen = new Set<string>();
    const addOpt = (val: string, lbl?: string) => {
      if (!seen.has(val)) { seen.add(val); results.push({ value: val as FoilTreatment, label: lbl || resolveLabel(val, FOIL_LABELS) }); }
    };
    matchingPrints.forEach(print => {
      const finishes = print.finishes || [];
      const pt = (print.promo_types || []) as string[];
      finishes.forEach((f: string) => {
        const key = f === 'nonfoil' ? 'non_foil' : (f === 'foil' ? 'foil' : (f === 'etched' ? 'etched_foil' : f));
        addOpt(key);
      });
      const specialFinishesKw = ['foil', 'ink', 'slick', 'textured', 'gilded', 'glossy', 'rainbow', 'compleat', 'galaxy', 'surge', 'halo', 'raised'];
      pt.forEach((p: string) => { if (specialFinishesKw.some(kw => p.toLowerCase().includes(kw))) addOpt(p); });
    });
    if (results.length === 0) addOpt(form.foil_treatment || 'non_foil');
    return results;
  };

  const handleSetChange = (newSet: string) => {
    const treatments = getTreatmentOptions(newSet);
    const newTreatment = treatments.find(t => t.value === form.card_treatment)?.value || treatments[0]?.value || 'regular';
    const arts = getArtOptions(newTreatment as CardTreatment, newSet);
    const newArt = arts[0]?.value || '';
    const promos = getPromoOptions(newTreatment as CardTreatment, newSet, newArt);
    const newPromo = promos[0]?.value || 'none';
    const foils = getFoilOptions(newTreatment as CardTreatment, newSet, newArt, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    const bestPrint = findMatchingPrint(newSet, newTreatment as CardTreatment, newArt, newPromo, newFoil as FoilTreatment);
    setForm(f => ({
      ...f, set_code: newSet, set_name: bestPrint?.set_name || f.set_name,
      card_treatment: newTreatment as CardTreatment, collector_number: newArt,
      promo_type: newPromo, foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint?.image_uris?.normal || bestPrint?.image_uris?.large || (bestPrint?.card_faces ? bestPrint?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
      description: '', price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleTreatmentChange = (newTreatment: CardTreatment) => {
    const arts = getArtOptions(newTreatment, form.set_code);
    const newArt = arts[0]?.value || '';
    const promos = getPromoOptions(newTreatment, form.set_code, newArt);
    const newPromo = promos[0]?.value || 'none';
    const foils = getFoilOptions(newTreatment, form.set_code, newArt, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    const bestPrint = findMatchingPrint(form.set_code, newTreatment, newArt, newPromo, newFoil as FoilTreatment);
    setForm(f => ({
      ...f, card_treatment: newTreatment, collector_number: newArt,
      promo_type: newPromo, foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint?.image_uris?.normal || bestPrint?.image_uris?.large || (bestPrint?.card_faces ? bestPrint?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
      description: '', price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleArtChange = (newArt: string) => {
    const promos = getPromoOptions(form.card_treatment, form.set_code, newArt);
    const newPromo = promos[0]?.value || 'none';
    const foils = getFoilOptions(form.card_treatment, form.set_code, newArt, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    const bestPrint = findMatchingPrint(form.set_code, form.card_treatment, newArt, newPromo, newFoil as FoilTreatment);
    setForm(f => ({
      ...f, collector_number: newArt, promo_type: newPromo, foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint?.image_uris?.normal || bestPrint?.image_uris?.large || (bestPrint?.card_faces ? bestPrint?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
      description: '', price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handlePromoChange = (newPromo: string) => {
    const foils = getFoilOptions(form.card_treatment, form.set_code, form.collector_number, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    const bestPrint = findMatchingPrint(form.set_code, form.card_treatment, form.collector_number, newPromo, newFoil as FoilTreatment);
    setForm(f => ({
      ...f, promo_type: newPromo, foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint?.image_uris?.normal || bestPrint?.image_uris?.large || (bestPrint?.card_faces ? bestPrint?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
      price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleFoilChange = (newFoil: FoilTreatment) => {
    const print = findMatchingPrint(form.set_code, form.card_treatment, form.collector_number, form.promo_type, newFoil);
    setForm(f => ({
      ...f, foil_treatment: newFoil,
      image_url: print?.image_uris?.normal || print?.image_uris?.large || (print?.card_faces ? print?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
      price_reference: print ? applyPrintPrices(print, newFoil, f.price_source) : f.price_reference,
      ...(print ? extractMTGMetadata(print) : {})
    }));
  };

  const handleSourceChange = (src: PriceSource) => {
    setForm(f => {
      let ref = f.price_reference;
      const print = findMatchingPrint(f.set_code, f.card_treatment, f.collector_number, f.promo_type, f.foil_treatment);
      if (print) ref = applyPrintPrices(print, f.foil_treatment, src);
      return { ...f, price_source: src, price_reference: ref };
    });
  };

  const handlePopulate = async () => {
    const name = form.name.trim();
    const set = form.set_code.trim().toLowerCase();
    const cn = form.collector_number.trim();
    if (!name && (!set || !cn)) return;

    setLookingUp(true);
    setFormError('');
    try {
      let searchQ = "";
      if (set && cn) {
        searchQ = `set:${set} cn:"${cn}"`;
      } else if (name) {
        searchQ = `!"${name}"`;
      } else if (set) {
        searchQ = `set:${set}`;
      }

      const res = await fetch(`https://api.scryfall.com/cards/search?q=${encodeURIComponent(searchQ)}+is:paper&unique=prints&order=released`);
      const body = await res.json();
      if (!body.data || body.data.length === 0) throw new Error('No printings found for that search.');
      
      let prints = body.data;

      // If we have a specific card, fetch ALL its prints to populate variant cascade
      if (prints.length > 0 && (prints.length === 1 || (set && cn))) {
        const oracleId = prints[0].oracle_id;
        if (oracleId) {
          try {
            const allRes = await fetch(`https://api.scryfall.com/cards/search?q=oracle_id:${oracleId}+is:paper&unique=prints&order=released`);
            const allBody = await allRes.json();
            if (allBody.data && allBody.data.length > 0) prints = allBody.data;
          } catch (err) { console.error('Failed to fetch variants:', err); }
        }
      }

      setScryfallPrints(prints);
      
      let bestPrint = prints.find((p: any) => p?.set?.toLowerCase() === set && p?.collector_number === cn);
      if (!bestPrint && set) bestPrint = prints.find((p: any) => p?.set?.toLowerCase() === set);
      if (!bestPrint) bestPrint = prints[0];

      if (!bestPrint) throw new Error('Could not identify a matching print.');

      const initialTreatment = getTreatmentType(bestPrint);
      const initialFoil = (() => {
        if (bestPrint.finishes?.includes('nonfoil') && bestPrint.finishes?.length === 1) return 'non_foil';
        const pt = bestPrint.promo_types || [];
        const specializedFoilKeywords = ['slick', 'compleat', 'galaxy', 'surge', 'textured', 'raised', 'rainbow', 'confetti', 'ink', 'gilded', 'halo', 'glossy'];
        const bestPromo = pt.find((p: string) => specializedFoilKeywords.some((kw: string) => p.toLowerCase().includes(kw)));
        if (bestPromo) return bestPromo;
        if (bestPrint.finishes?.includes('etched')) return 'etched_foil';
        if (bestPrint.finishes?.includes('glossy')) return 'glossy';
        if (bestPrint.finishes?.includes('foil')) return 'foil';
        return 'non_foil';
      })();

      setForm(f => ({
        ...f, 
        name: bestPrint?.name || f.name,
        set_code: bestPrint?.set || f.set_code, 
        set_name: bestPrint?.set_name || f.set_name,
        card_treatment: initialTreatment, 
        collector_number: bestPrint?.collector_number || f.collector_number,
        promo_type: bestPrint?.promo_types?.join(',') || 'none',
        foil_treatment: initialFoil as FoilTreatment,
        image_url: bestPrint?.image_uris?.normal || bestPrint?.image_uris?.large || (bestPrint?.card_faces ? bestPrint?.card_faces[0]?.image_uris?.normal : '') || f.image_url,
        description: '', 
        price_reference: applyPrintPrices(bestPrint, initialFoil as FoilTreatment, f.price_source),
        ...extractMTGMetadata(bestPrint)
      }));
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Scryfall fetch failed');
    } finally { setLookingUp(false); }
  };

  const updateStoreQty = (id: string, delta: number) => {
    setProductStorage(prev => prev.map(loc => loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, loc.quantity + delta) } : loc));
  };
  const setStoreQty = (id: string, qty: number) => {
    setProductStorage(prev => prev.map(loc => loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, qty) } : loc));
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

        {/* â”€â”€ HEADER â”€â”€ */}
        <div className="flex items-center justify-between p-4 md:p-6 pb-0">
          <h2 className="font-display text-3xl m-0">{editProduct ? 'EDIT PRODUCT' : 'NEW PRODUCT'}</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        {/* â”€â”€ TCG + CATEGORY (top-level) â”€â”€ */}
        <div className="px-4 md:px-6 pt-4 flex gap-4 flex-wrap">
          <div style={{ minWidth: '160px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>TCG *</label>
            <select value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
              {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              {tcgs.length === 0 && KNOWN_TCGS.map(t => <option key={t} value={t}>{TCG_SHORT[t]}</option>)}
              <option value="accessories">Accessories</option>
            </select>
          </div>
          <div style={{ minWidth: '140px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>CATEGORY *</label>
            <select value={form.category} onChange={e => setForm(f => ({ ...f, category: e.target.value as typeof EMPTY_FORM['category'] }))}>
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

        {/* â”€â”€ HERO LOOKUP (MTG Singles only) â”€â”€ */}
        {isMTGSingles && (
          <div className="mx-4 md:mx-6 mt-4 p-4 rounded-sm" style={{ background: 'var(--kraft-light)', border: '2px dashed var(--kraft-dark)' }}>
            <div className="flex items-end gap-3 flex-wrap sm:flex-nowrap">
              <div style={{ width: '90px' }}>
                <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET</label>
                {scryfallPrints.length > 0 ? (
                  <select value={form.set_code} onChange={e => handleSetChange(e.target.value)} className="font-bold" style={{ fontSize: '0.85rem' }}>
                    {Array.from(new Map(scryfallPrints.filter(c => !!c).map(c => [c.set, c.set_name])).entries()).map(([code, name]) => (
                      <option key={code} value={code}>[{code.toUpperCase()}] {name}</option>
                    ))}
                  </select>
                ) : (
                  <input type="text" value={form.set_code} onChange={e => setForm(f => ({ ...f, set_code: e.target.value.toUpperCase() }))} placeholder="MH2" className="text-center font-bold uppercase" style={{ fontSize: '0.85rem' }} />
                )}
              </div>
              <div style={{ width: '70px' }}>
                <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}># CN</label>
                <input type="text" value={form.collector_number} onChange={e => handleArtChange(e.target.value)} placeholder="123" className="text-center font-bold" style={{ fontSize: '0.85rem' }} />
              </div>
              <div className="flex-1 min-w-[200px]">
                <div className="flex justify-between items-end mb-1">
                  <label className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>CARD NAME *</label>
                  {form.set_name && <span className="text-[10px] font-mono-stack truncate" style={{ color: 'var(--gold)', maxWidth: '200px' }}>{form.set_name}</span>}
                </div>
                <input type="text" value={form.name} onChange={e => { setForm(f => ({ ...f, name: e.target.value })); setScryfallPrints([]); }} style={{ fontSize: '1rem', fontWeight: 600 }} placeholder="e.g. Lightning Bolt" />
              </div>
              <button type="button" onClick={handlePopulate}
                disabled={lookingUp || (!form.name.trim() && (!form.set_code.trim() || !form.collector_number.trim()))}
                className="btn-primary px-5 transition-all"
                style={{ height: '42px', fontSize: '0.85rem', whiteSpace: 'nowrap', opacity: lookingUp ? 0.7 : 1 }}>
                {lookingUp ? '⏳ LOOKING UP...' : scryfallPrints.length > 0 ? '✓ RE-POPULATE' : '📥 POPULATE'}
              </button>
            </div>
          </div>
        )}

        {/* — For non-MTG singles: simple name input — */}
        {!isMTGSingles && (
          <div className="px-4 md:px-6 pt-4">
            <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRODUCT NAME *</label>
            <input type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} style={{ fontSize: '1rem', fontWeight: 600 }} placeholder="Product name" />
          </div>
        )}

        {/* — TAB BAR — */}
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

        {/* — MAIN CONTENT: Tab panels + Sidebar — */}
        <div className="flex gap-6 flex-col md:flex-row p-4 md:p-6 pt-4">
          {/* Tab Panels */}
          <div className="flex-1 min-w-0">

            {/* — TAB: VARIANT & IDENTITY — */}
            {activeTab === 'variant' && (
              <div className="space-y-4">
                {/* Variant Cascade */}
                {isMTGSingles && scryfallPrints.length > 0 && (
                  <div>
                    <h3 className="text-xs font-mono-stack mb-3 uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>VARIANT CASCADE</h3>
                    <div className="space-y-3">
                      {/* Treatment */}
                      <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono-stack w-20 text-right shrink-0" style={{ color: 'var(--text-muted)' }}>TREATMENT</span>
                        <span style={{ color: 'var(--kraft-dark)' }}>→</span>
                        <select className="flex-1" value={form.card_treatment} disabled={scryfallPrints.length === 0} onChange={e => handleTreatmentChange(e.target.value as CardTreatment)}>
                          {getTreatmentOptions().map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
                        </select>
                      </div>
                      {/* Art */}
                      <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono-stack w-20 text-right shrink-0" style={{ color: 'var(--text-muted)' }}>ART / CN</span>
                        <span style={{ color: 'var(--kraft-dark)' }}>→</span>
                        <select className="flex-1" value={form.collector_number} disabled={scryfallPrints.length === 0 || !form.card_treatment} onChange={e => handleArtChange(e.target.value)}>
                          {getArtOptions(form.card_treatment).map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
                        </select>
                      </div>
                      {/* Promo */}
                      <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono-stack w-20 text-right shrink-0" style={{ color: 'var(--text-muted)' }}>PROMO</span>
                        <span style={{ color: 'var(--kraft-dark)' }}>→</span>
                        <select className="flex-1" value={form.promo_type} disabled={scryfallPrints.length === 0 || !form.collector_number} onChange={e => handlePromoChange(e.target.value)}>
                          {getPromoOptions(form.card_treatment, form.set_code, form.collector_number).map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
                        </select>
                      </div>
                      {/* Foil */}
                      <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono-stack w-20 text-right shrink-0" style={{ color: 'var(--text-muted)' }}>FOIL</span>
                        <span style={{ color: 'var(--kraft-dark)' }}>→</span>
                        <select className="flex-1" value={form.foil_treatment} disabled={scryfallPrints.length === 0 || !form.promo_type} onChange={e => handleFoilChange(e.target.value as FoilTreatment)}>
                          {getFoilOptions(form.card_treatment, form.set_code, form.collector_number, form.promo_type).map(opt => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
                        </select>
                      </div>
                    </div>
                  </div>
                )}

                {/* Manual variant fields when no Scryfall data */}
                {isMTGSingles && scryfallPrints.length === 0 && (
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET NAME</label>
                      <input type="text" value={form.set_name} onChange={e => setForm(f => ({ ...f, set_name: e.target.value }))} placeholder="e.g. Modern Horizons 2" />
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CARD TREATMENT</label>
                      <select value={form.card_treatment} onChange={e => setForm(f => ({ ...f, card_treatment: e.target.value as CardTreatment }))}>
                        {Object.entries(TREATMENT_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                      </select>
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>FOIL TREATMENT</label>
                      <select value={form.foil_treatment} onChange={e => setForm(f => ({ ...f, foil_treatment: e.target.value as FoilTreatment }))}>
                        {Object.entries(FOIL_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                      </select>
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PROMO TYPE</label>
                      <input type="text" value={form.promo_type} onChange={e => setForm(f => ({ ...f, promo_type: e.target.value }))} placeholder="none" />
                    </div>
                  </div>
                )}

                {/* Non-MTG variant fields */}
                {!isMTGSingles && (
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET NAME</label>
                      <input type="text" value={form.set_name} onChange={e => setForm(f => ({ ...f, set_name: e.target.value }))} />
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET CODE</label>
                      <input type="text" value={form.set_code} onChange={e => setForm(f => ({ ...f, set_code: e.target.value.toUpperCase() }))} />
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* — TAB: PRICING & STOCK — */}
            {activeTab === 'pricing' && (
              <div className="space-y-6">
                {/* Pricing */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>PRICING</h3>
                    <div className="text-xs font-mono-stack px-2 py-1 rounded" style={{ background: 'var(--ink-surface)', color: 'var(--gold)' }}>
                      {form.price_source === 'tcgplayer' && `(x ${settings?.usd_to_cop_rate || 0} COP)`}
                      {form.price_source === 'cardmarket' && `(x ${settings?.eur_to_cop_rate || 0} COP)`}
                      {form.price_source !== 'manual' && typeof form.price_reference === 'number' && (
                        <span className="ml-2 font-bold text-sm" style={{ color: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--gold)' }}>
                          = ${(form.price_reference * (form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0) : (settings?.eur_to_cop_rate || 0))).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE SOURCE *</label>
                      <select value={form.price_source} onChange={e => handleSourceChange(e.target.value as PriceSource)}>
                        <option value="manual">Manual Override (COP)</option>
                        <option value="tcgplayer">External: TCGPlayer (USD)</option>
                        <option value="cardmarket">External: Cardmarket (EUR)</option>
                      </select>
                    </div>
                    {form.price_source === 'manual' ? (
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE (COP) *</label>
                        <input type="number" step="100" min="0" value={form.price_cop_override} onChange={e => setForm(f => ({ ...f, price_cop_override: e.target.value ? parseFloat(e.target.value) : '' }))} />
                      </div>
                    ) : (
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>
                          REFERENCE PRICE ({form.price_source === 'tcgplayer' ? 'USD' : 'EUR'}) *
                        </label>
                        <input type="number" step="0.01" min="0" value={form.price_reference}
                          onChange={e => setForm(f => ({ ...f, price_reference: e.target.value ? parseFloat(e.target.value) : '' }))}
                          style={{ color: form.price_reference === 0 ? 'var(--hp-color)' : 'inherit', borderColor: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--ink-border)' }} />
                      </div>
                    )}
                  </div>
                </div>

                {/* Storage */}
                <div className="pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>STORAGE LOCATIONS</h3>
                    <span className="text-xs font-bold bg-ink-surface px-2 py-1 text-gold rounded border border-ink-border">
                      TOTAL: {productStorage.reduce((acc, l) => acc + l.quantity, 0)}
                    </span>
                  </div>
                  <div className="space-y-2 max-h-60 overflow-y-auto mb-4">
                    {productStorage.length === 0 && <p className="text-xs text-text-muted italic text-center py-2">No storage assignments yet.</p>}
                    {productStorage.map(loc => (
                      <div key={loc.stored_in_id} className="flex items-center justify-between gap-2 text-sm border-b border-ink-border/50 pb-2">
                        <span className="truncate flex-1 font-semibold" title={loc.name}>{loc.name}</span>
                        <div className="flex items-center gap-1">
                          <button onClick={() => updateStoreQty(loc.stored_in_id, -1)} className="w-7 h-7 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-hp-color transition-colors rounded-sm" disabled={loc.quantity <= 0}>−</button>
                          <input type="number" value={loc.quantity === 0 ? '' : loc.quantity} min="0"
                            onChange={e => setStoreQty(loc.stored_in_id, parseInt(e.target.value) || 0)}
                            className="w-14 px-1 py-0 text-center text-sm font-mono-stack" style={{ height: '28px' }} placeholder="0" />
                          <button onClick={() => updateStoreQty(loc.stored_in_id, 1)} className="w-7 h-7 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-gold transition-colors rounded-sm">+</button>
                          <button onClick={() => setProductStorage(prev => prev.filter(p => p.stored_in_id !== loc.stored_in_id))} className="w-7 h-7 flex items-center justify-center hover:text-hp-color opacity-40 hover:opacity-100 transition-opacity" title="Remove">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div>
                    <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Quick Add Location</label>
                    <select className="w-full text-xs px-2 h-10" style={{ padding: '0 8px' }}
                      onChange={(e) => {
                        const id = e.target.value;
                        if (!id) return;
                        const loc = storageLocations.find(l => l.id === id);
                        if (!loc) return;
                        if (productStorage.find(p => p.stored_in_id === id)) return;
                        setProductStorage(prev => [...prev, { stored_in_id: loc.id, name: loc.name, quantity: 0 }]);
                        e.target.value = "";
                      }}
                    >
                      <option value="">-- Select Location --</option>
                      {storageLocations
                        .filter(sl => !productStorage.some(ps => ps.stored_in_id === sl.id))
                        .sort((a,b) => a.name.localeCompare(b.name))
                        .map(sl => <option key={sl.id} value={sl.id}>{sl.name}</option>)
                      }
                    </select>
                  </div>
                </div>
              </div>
            )}

            {/* — TAB: DETAILS & METADATA — */}
            {activeTab === 'details' && (
              <div className="space-y-6">
                {/* Image URL */}
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>IMAGE URL</label>
                  <input type="text" value={form.image_url} onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))} placeholder="https://..." />
                </div>

                {/* Description */}
                <div>
                  <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>DESCRIPTION</label>
                  <textarea className="w-full" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} rows={3} />
                </div>

                {/* Collections */}
                <div className="pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                  <label className="text-[10px] font-mono-stack mb-2 block uppercase" style={{ color: 'var(--text-muted)' }}>CUSTOM COLLECTIONS / CATEGORIES</label>
                  <div className="flex flex-wrap gap-2">
                    {categories.length === 0 && <span className="text-sm text-text-muted italic">No categories created yet.</span>}
                    {categories.map(c => {
                      const isSelected = form.category_ids.includes(c.id);
                      return (
                        <button key={c.id} type="button"
                          onClick={() => setForm(f => ({
                            ...f,
                            category_ids: isSelected ? f.category_ids.filter(id => id !== c.id) : [...f.category_ids, c.id]
                          }))}
                          className={`badge transition-colors cursor-pointer ${isSelected ? 'border-gold' : 'bg-ink-surface text-text-secondary border-ink-border hover:border-gold/50'}`}
                          style={{ background: isSelected ? 'var(--gold)' : '', color: isSelected ? 'var(--ink-deep)' : '' }}>
                          {c.name}
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* MTG Metadata (collapsible) */}
                {isMTGSingles && (
                  <div className="pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                    <h3 className="text-xs font-mono-stack mb-3 uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>MTG METADATA</h3>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-4">
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>LANGUAGE</label>
                        <select value={form.language} onChange={e => setForm(f => ({ ...f, language: e.target.value }))}>
                          <option value="en">English (EN)</option>
                          <option value="es">Spanish (ES)</option>
                          <option value="fr">French (FR)</option>
                          <option value="de">German (DE)</option>
                          <option value="it">Italian (IT)</option>
                          <option value="pt">Portuguese (PT)</option>
                          <option value="ja">Japanese (JA)</option>
                          <option value="ko">Korean (KO)</option>
                          <option value="ru">Russian (RU)</option>
                        </select>
                      </div>
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>COLOR IDENTITY</label>
                        <input type="text" value={form.color_identity} onChange={e => setForm(f => ({ ...f, color_identity: e.target.value }))} placeholder="e.g. W,U" />
                      </div>
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>RARITY</label>
                        <select value={form.rarity} onChange={e => setForm(f => ({ ...f, rarity: e.target.value }))}>
                          <option value="common">Common</option>
                          <option value="uncommon">Uncommon</option>
                          <option value="rare">Rare</option>
                          <option value="mythic">Mythic</option>
                          <option value="special">Special</option>
                          <option value="bonus">Bonus</option>
                        </select>
                      </div>
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CMC</label>
                        <input type="number" step="0.5" value={form.cmc} onChange={e => setForm(f => ({ ...f, cmc: e.target.value ? parseFloat(e.target.value) : '' }))} />
                      </div>
                      <div>
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>ARTIST</label>
                        <input type="text" value={form.artist} onChange={e => setForm(f => ({ ...f, artist: e.target.value }))} />
                      </div>
                      <div className="sm:col-span-2">
                        <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>TYPE LINE</label>
                        <input type="text" value={form.type_line} onChange={e => setForm(f => ({ ...f, type_line: e.target.value }))} />
                      </div>
                    </div>
                    <div className="mb-4">
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>ORACLE TEXT</label>
                      <textarea className="w-full text-[11px] font-mono-stack p-2 bg-transparent border border-ink-border rounded h-24"
                        value={form.oracle_text} onChange={e => setForm(f => ({ ...f, oracle_text: e.target.value }))} />
                    </div>
                    <div className="flex flex-wrap gap-x-6 gap-y-3 p-3 bg-kraft-light/10 border border-kraft-dark rounded">
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.is_legendary} onChange={e => setForm(f => ({ ...f, is_legendary: e.target.checked }))} /> LEGENDARY
                      </label>
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.is_historic} onChange={e => setForm(f => ({ ...f, is_historic: e.target.checked }))} /> HISTORIC
                      </label>
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.is_land} onChange={e => setForm(f => ({ ...f, is_land: e.target.checked }))} /> LAND
                      </label>
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.is_basic_land} onChange={e => setForm(f => ({ ...f, is_basic_land: e.target.checked }))} /> BASIC LAND
                      </label>
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.full_art} onChange={e => setForm(f => ({ ...f, full_art: e.target.checked }))} /> FULL ART
                      </label>
                      <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                        <input type="checkbox" checked={form.textless} onChange={e => setForm(f => ({ ...f, textless: e.target.checked }))} /> TEXTLESS
                      </label>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* — SIDEBAR: Image Preview — */}
          <div className="w-full md:w-72 flex-shrink-0">
            <div className="sticky top-4">
              <label className="text-[10px] font-mono-stack mb-1 block uppercase" style={{ color: 'var(--text-muted)' }}>IMAGE PREVIEW</label>
              <div className="cardbox overflow-hidden" style={{ aspectRatio: '3/4', padding: '0.5rem', background: 'var(--kraft-light)' }}>
                <div className="w-full h-full bg-ink-card border border-ink-border relative flex items-center justify-center">
                  {form.image_url ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={form.image_url} alt="Preview" style={{ width: '100%', height: '100%', objectFit: 'contain' }} />
                  ) : (
                    <span className="text-5xl opacity-20">🃏</span>
                  )}
                </div>
              </div>
              {/* Quick info summary */}
              {form.name && (
                <div className="mt-2 p-2 text-[10px] font-mono-stack rounded" style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)' }}>
                  <p className="font-bold truncate" style={{ color: 'var(--ink-deep)' }}>{form.name}</p>
                  {form.set_name && <p className="truncate" style={{ color: 'var(--text-muted)' }}>[{form.set_code}] {form.set_name}</p>}
                  <div className="flex gap-1 mt-1 flex-wrap">
                    <span className="badge text-[9px]" style={{ padding: '1px 4px' }}>{form.condition || 'NM'}</span>
                    {form.foil_treatment !== 'non_foil' && <span className="badge badge-foil text-[9px]" style={{ padding: '1px 4px' }}>✦ Foil</span>}
                    {form.card_treatment !== 'normal' && <span className="badge text-[9px]" style={{ padding: '1px 4px', background: 'var(--ink-surface)' }}>{resolveLabel(form.card_treatment, TREATMENT_LABELS)}</span>}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* — STICKY FOOTER — */}
        <div className="sticky bottom-0 bg-surface p-4 md:px-6 border-t border-ink-border" style={{ background: 'var(--kraft-light)', zIndex: 10 }}>
          {formError && (
            <p className="mb-3 text-sm font-mono-stack" style={{ color: 'var(--hp-color)' }}>{formError}</p>
          )}
          <div className="flex gap-3 flex-wrap">
            <button onClick={() => handleSave(false)} className="btn-primary flex-1 py-3 text-sm" disabled={saving}>
              {saving ? 'SAVING...' : editProduct ? 'SAVE CHANGES' : 'CREATE PRODUCT'}
            </button>
            {!editProduct && (
              <button onClick={() => handleSave(true)} className="btn-secondary flex-1 py-3 text-sm" disabled={saving}
                title="Save and immediately open a new form with TCG, category, and condition pre-filled">
                {saving ? 'SAVING...' : '💾 SAVE & ADD ANOTHER'}
              </button>
            )}
            <button onClick={onClose} className="btn-secondary py-3 text-sm px-6">CANCEL</button>
          </div>
          <p className="text-[9px] font-mono-stack mt-2 text-center" style={{ color: 'var(--text-muted)' }}>
            Press <kbd style={{ background: 'var(--ink-surface)', padding: '1px 4px', borderRadius: '2px', border: '1px solid var(--ink-border)' }}>Ctrl+Enter</kbd> to save
          </p>
        </div>

      </div>
    </div>
  );
}
