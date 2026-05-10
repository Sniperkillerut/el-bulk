'use client';

import { useEffect, useState } from 'react';
import {
  adminCreateProduct, adminUpdateProduct, adminFetchExternalPrice
} from '@/lib/api';
import {
  Product, FoilTreatment, CardTreatment, PriceSource,
  StorageLocation, ScryfallCard, Condition
} from '@/lib/types';
import { 
  getTreatmentType, applyPrintPrices, extractMTGMetadata, findMatchingPrint, getScryfallImage, resolveFoilTreatment,
} from '@/lib/mtg-logic';
import { useLanguage } from '@/context/LanguageContext';

import ScryfallPopulate from './product/ScryfallPopulate';
import VariantTab from './product/tabs/VariantTab';
import PricingTab from './product/tabs/PricingTab';
import DeckCardsTab from './product/tabs/DeckCardsTab';
import CardImage from '../CardImage';
import { FormState, TabId, ProductEditModalProps } from './product/types';

export const EMPTY_FORM: FormState = {
  name: '', tcg: 'mtg', category: 'singles',
  set_name: '', set_code: '', condition: 'NM',
  foil_treatment: 'non_foil', card_treatment: 'normal',
  price: 0, price_source: 'cardkingdom', price_reference: '', price_cop_override: '', cost_basis_cop: '',
  stock: 0, description: '', category_ids: [], image_url: '',
  collector_number: '', promo_type: '', scryfall_id: '',
  language: 'en', color_identity: '', rarity: 'common', cmc: '',
  is_legendary: false, is_historic: false, is_land: false, is_basic_land: false,
  art_variation: '', oracle_text: '', artist: '', type_line: '',
  border_color: '', frame: '', full_art: false, textless: false,
  storage_items: [], deck_cards: []
};

export default function ProductEditModal({
  editProduct, storageLocations, categories, tcgs, settings,
  storageFilter, onClose, onSaved, onSaveAndNew
}: ProductEditModalProps) {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<TabId>('variant');
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [scryfallPrints, setScryfallPrints] = useState<ScryfallCard[]>([]);
  const [lookingUp, setLookingUp] = useState(false);
  const [productStorage, setProductStorage] = useState<StorageLocation[]>([]);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');

  // Mapping: Product -> ScryfallCard (for instant local pre-population)
  const productToScryfallCard = (p: Product): ScryfallCard => {
    const promoTypes = p.promo_type ? p.promo_type.split(',').map(s => s.trim().toLowerCase()) : [];
    
    // Reverse map foil_treatment to scryfall finishes
    const finishes = [
      p.foil_treatment === 'non_foil' ? 'nonfoil' : 
      p.foil_treatment === 'foil' ? 'foil' : 
      p.foil_treatment === 'etched_foil' ? 'etched' : 
      'foil' // default fallback for other foils
    ];
    
    // Ensure the specific foil treatment is in promo_types if it's indexed there by resolve...
    const specialFoils = ['ripple_foil', 'galaxy_foil', 'surge_foil', 'textured_foil', 'glossy', 'confetti_foil', 'oil_slick', 'step_and_compleat', 'double_rainbow', 'platinum_foil'];
    if (specialFoils.includes(p.foil_treatment)) {
      const ptKey = p.foil_treatment.replace('_', '');
      if (!promoTypes.includes(ptKey)) promoTypes.push(ptKey);
    }

    // Reverse map card_treatment to scryfall frame/border/effects
    const frameEffects: string[] = [];
    let borderColor = p.border_color || 'black';
    let frame = p.frame || '';

    if (p.card_treatment === 'showcase') frameEffects.push('showcase');
    if (p.card_treatment === 'extended_art') frameEffects.push('extendedart');
    if (p.card_treatment === 'borderless') borderColor = 'borderless';
    if (p.card_treatment === 'retro_frame') frameEffects.push('retro');
    if (p.card_treatment === 'legacy_border') frame = '1997';
    if (p.card_treatment === 'etched') frameEffects.push('etched');
    if (p.card_treatment === 'serialized') frameEffects.push('serialized');
    
    // Only set card.promo to true if the treatment itself is 'promo' 
    // or it's a known promo type that resolveCardTreatment handles
    if (p.card_treatment === 'promo' && !promoTypes.includes('promo')) {
      promoTypes.push('promo');
    }
    if (p.card_treatment === 'judge_promo' && !promoTypes.includes('judgegift')) {
      promoTypes.push('judgegift');
    }

    return {
      id: p.scryfall_id || p.id,
      name: p.name,
      set: p.set_code || '',
      set_name: p.set_name || '',
      collector_number: p.collector_number || '',
      image_uris: p.image_url ? { normal: p.image_url, small: p.image_url } : undefined,
      type_line: p.type_line || '',
      oracle_text: p.oracle_text || '',
      cmc: p.cmc,
      rarity: p.rarity,
      color_identity: p.color_identity ? p.color_identity.split(',') : [],
      full_art: p.full_art || p.card_treatment === 'full_art',
      textless: p.textless || p.card_treatment === 'textless',
      artist: p.artist,
      border_color: borderColor,
      frame: frame,
      promo_types: promoTypes,
      finishes: finishes,
      promo: p.card_treatment === 'promo'
    };
  };

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
        cost_basis_cop: editProduct.cost_basis_cop ?? '',
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
        scryfall_id: editProduct.scryfall_id || '',
        storage_items: editProduct.stored_in?.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })) || [],
        deck_cards: editProduct.deck_cards || [],
      });
      setProductStorage((editProduct.stored_in || []).map(d => ({
        stored_in_id: d.stored_in_id, name: d.name, quantity: d.quantity
      })));
      
      // INSTANT PRE-POPULATION: Use local data to fill Scryfall data sections before fetching
      if (editProduct.tcg === 'mtg' && editProduct.category === 'singles') {
        setScryfallPrints([productToScryfallCard(editProduct)]);
      } else {
        setScryfallPrints([]);
      }
    } else {
      setForm({ ...EMPTY_FORM });
      const initialStorage: StorageLocation[] = [];
      if (storageFilter) {
        const loc = storageLocations.find(l => l.id === storageFilter);
        if (loc) initialStorage.push({ stored_in_id: loc.id, name: loc.name, quantity: 0 });
      }
      setProductStorage(initialStorage);
      setScryfallPrints([]);
    }
    setFormError('');
  }, [editProduct, storageFilter, storageLocations]);

  // Auto-fetch CardKingdom prices when variations or source change seamlessly
  useEffect(() => {
    if (form.tcg !== 'mtg' || !form.name || form.price_source !== 'cardkingdom') return;

    let isMounted = true;
    const fetchCK = async () => {
      try {
        const res = await adminFetchExternalPrice(
          form.name,
          form.set_code || '',
          form.set_name || '',
          form.collector_number || '',
          form.foil_treatment || '',
          form.card_treatment || '',
          form.price_source,
          form.scryfall_id
        );
        if (isMounted && res.price !== undefined) {
          setForm(f => ({ ...f, price_reference: res.price }));
        }
      } catch {
        if (isMounted) {
          setForm(f => ({ ...f, price_reference: 0 }));
        }
      }
    };

    const timer = setTimeout(fetchCK, 300); // Debounce typing/selections
    return () => {
      isMounted = false;
      clearTimeout(timer);
    };
  }, [
    form.tcg,
    form.name,
    form.set_code,
    form.set_name,
    form.collector_number,
    form.foil_treatment,
    form.card_treatment,
    form.price_source,
    form.scryfall_id
  ]);

  // AUTO-SYNC: Trigger Scryfall fetch in background if essential metadata is missing
  useEffect(() => {
    if (editProduct && editProduct.tcg === 'mtg' && editProduct.category === 'singles' && scryfallPrints.length <= 1) {
      const isMissingEssentialMetadata = !editProduct.image_url || !editProduct.oracle_text || !editProduct.type_line || !editProduct.cmc;
      
      // If we only have our local "faked" print or are missing major fields, attempt population
      if (isMissingEssentialMetadata || (scryfallPrints.length === 1 && scryfallPrints[0].id === editProduct.id)) {
        handlePopulate();
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editProduct]);

  // Handle automatic tab selection on product switch/load
  useEffect(() => {
    if (editProduct) {
      if (editProduct.category === 'store_exclusives') setActiveTab('deck');
      else if (editProduct.tcg === 'mtg' && editProduct.category === 'singles') setActiveTab('variant');
      else setActiveTab('pricing');
    } else {
      if (form.category === 'store_exclusives') setActiveTab('deck');
      else if (form.tcg === 'mtg' && form.category === 'singles') setActiveTab('variant');
      else setActiveTab('pricing');
    }
  }, [editProduct, form.category, form.tcg]);

  const handleSave = async (andNew: boolean) => {
    if (!form.name || !form.tcg || !form.category) { 
      setFormError(t('components.admin.product_modal.error_required', 'Name, TCG, and Category are required.')); 
      return; 
    }
    setSaving(true); 
    setFormError('');
    try {
      const payload: Partial<Product> & { category_ids?: string[], storage_items?: { stored_in_id: string; quantity: number }[] } = {
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
        cost_basis_cop: form.cost_basis_cop === '' ? undefined : Number(form.cost_basis_cop),
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
        scryfall_id: form.scryfall_id || undefined,
        deck_cards: form.deck_cards,
        storage_items: productStorage.map(s => ({ 
          stored_in_id: s.stored_in_id, 
          quantity: s.quantity 
        })),
      };
      
      if (payload.price_source === 'manual') payload.price_reference = undefined;
      else payload.price_cop_override = undefined;

      if (editProduct) {
        await adminUpdateProduct(editProduct.id, payload);
      } else {
        await adminCreateProduct(payload);
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
        if (stickyCategory === 'store_exclusives') setActiveTab('deck');
        else if (stickyTcg === 'mtg' && stickyCategory === 'singles') setActiveTab('variant');
        else setActiveTab('pricing');
      } else {
        onSaved();
      }
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : t('components.admin.product_modal.error_save', 'Failed to save product.'));
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
    const sid = form.scryfall_id?.trim();
    if (!name && !sid && (!set || !cn)) return;

    setLookingUp(true); 

    try {
      let bestPrint: ScryfallCard | null = null;

      // STAGE 1: Explicit Scryfall ID (UUID)
      if (sid) {
        try {
          const sidRes = await fetch(`https://api.scryfall.com/cards/${sid}`);
          if (sidRes.ok) {
            bestPrint = await sidRes.json();
          }
        } catch (e) {
          console.warn(`Scryfall ID lookup failed for ${sid}, falling back...`, e);
        }
      }

      // STAGE 2: Set Code + Collector Number
      if (!bestPrint && set && cn) {
        try {
          const scnRes = await fetch(`https://api.scryfall.com/cards/search?q=${encodeURIComponent(`set:${set} cn:"${cn}"`)}&unique=prints`);
          if (scnRes.ok) {
            const body = await scnRes.json();
            if (body.data && body.data.length > 0) {
              bestPrint = body.data[0];
            }
          }
        } catch (e) {
          console.warn(`Set/CN lookup failed for ${set}/${cn}, falling back...`, e);
        }
      }

      // STAGE 3: Fuzzy Name Search (Failsafe)
      if (!bestPrint && name) {
        try {
          const nameRes = await fetch(`https://api.scryfall.com/cards/named?fuzzy=${encodeURIComponent(name)}`);
          if (nameRes.ok) {
            bestPrint = await nameRes.json();
          }
        } catch (e) {
          console.warn(`Name lookup failed for ${name}`, e);
        }
      }

      if (!bestPrint) {
        throw new Error(t('components.admin.product_modal.error_no_match', 'Could not identify a matching print on Scryfall.'));
      }

      // ORACLE RESOLUTION: Once we have a specific print, fetch its full list of prints (Oracle version)
      const fetchAllPrints = async (q: string) => {
        let results: ScryfallCard[] = [];
        let nextUrl: string | null = `https://api.scryfall.com/cards/search?q=${encodeURIComponent(q)}+game:paper&unique=prints&order=released`;
        while (nextUrl) {
          const r: Response = await fetch(nextUrl);
          if (!r.ok) break;
          const b: { data: ScryfallCard[]; has_more: boolean; next_page?: string } = await r.json();
          if (b.data) {
            const paperOnly = (b.data as ScryfallCard[]).filter(c => !c.digital);
            results = results.concat(paperOnly);
          }
          nextUrl = b.has_more ? (b.next_page as string) : null;
          if (nextUrl) await new Promise(res => setTimeout(res, 100)); // Rate limit
        }
        return results;
      };

      const oracleId = (bestPrint as unknown as { oracle_id?: string }).oracle_id;
      let oraclePrints: ScryfallCard[] = [];
      if (oracleId) {
        oraclePrints = await fetchAllPrints(`oracle_id:${oracleId}`);
      } else {
        // Fallback: search for its exact name to get prints
        oraclePrints = await fetchAllPrints(`!"${bestPrint.name}"`);
      }

      setScryfallPrints(oraclePrints);
      
      // Update form with metadata from the best matching print
      const initialTreatment = getTreatmentType(bestPrint);

      setForm(f => {
        const availableFoils = bestPrint?.finishes || [];
        const currentReq = f.foil_treatment === 'non_foil' ? 'nonfoil' : (f.foil_treatment === 'foil' ? 'foil' : (f.foil_treatment === 'etched_foil' ? 'etched' : f.foil_treatment));

        let foilToSet = resolveFoilTreatment(bestPrint) as FoilTreatment;
        if (availableFoils.includes(currentReq)) {
          foilToSet = f.foil_treatment;
        }

        return {
          ...f, 
          ...extractMTGMetadata(bestPrint),
          name: bestPrint?.name || f.name,
          set_code: bestPrint?.set || f.set_code, 
          set_name: bestPrint?.set_name || f.set_name,
          scryfall_id: bestPrint?.id || f.scryfall_id,
          // Override with Scryfall-determined best-guess (User wants "real" Scryfall data)
          card_treatment: initialTreatment, 
          collector_number: bestPrint?.collector_number || f.collector_number,
          promo_type: bestPrint?.promo_types?.join(',') || 'none',
          foil_treatment: foilToSet,
          image_url: getScryfallImage(bestPrint) || f.image_url,
          price_reference: applyPrintPrices(bestPrint, foilToSet, f.price_source, f.price_reference),
        };
      });

    } catch (e) {
      setFormError(e instanceof Error ? e.message : String(e));
    } finally {
      setLookingUp(false);
    }
  };

  const handleSetSearchChange = (newSet: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, newSet, form.card_treatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      set_code: newSet, set_name: bestPrint?.set_name || f.set_name,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      scryfall_id: bestPrint?.id || f.scryfall_id,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source, f.price_reference),
    }));
  };

  const handleTreatmentChange = (newTreatment: CardTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, newTreatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      card_treatment: newTreatment,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      scryfall_id: bestPrint?.id || f.scryfall_id,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source, f.price_reference),
    }));
  };

  const handleArtChange = (newArt: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, newArt, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      collector_number: newArt,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      scryfall_id: bestPrint?.id || f.scryfall_id,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source, f.price_reference),
    }));
  };

  const handlePromoChange = (newPromo: string) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, newPromo, form.foil_treatment);
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      promo_type: newPromo,
      foil_treatment: resolveFoilTreatment(bestPrint) as FoilTreatment,
      scryfall_id: bestPrint?.id || f.scryfall_id,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, f.price_source, f.price_reference),
    }));
  };

  const handleFoilChange = (newFoil: FoilTreatment) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, form.promo_type, newFoil);
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(bestPrint),
      foil_treatment: newFoil,
      scryfall_id: bestPrint?.id || f.scryfall_id,
      image_url: getScryfallImage(bestPrint) || f.image_url,
      price_reference: applyPrintPrices(bestPrint, newFoil, f.price_source, f.price_reference),
    }));
  };

  const handleSelectPrint = (p: ScryfallCard) => {
    const treatment = getTreatmentType(p);
    const foil = resolveFoilTreatment(p);
    
    setForm(f => ({
      ...f,
      ...extractMTGMetadata(p),
      name: p.name || f.name,
      set_code: p.set || f.set_code,
      set_name: p.set_name || f.set_name,
      card_treatment: treatment,
      collector_number: p.collector_number || f.collector_number,
      promo_type: p.promo_types?.join(',') || 'none',
      foil_treatment: foil as FoilTreatment,
      scryfall_id: p.id,
      image_url: getScryfallImage(p) || f.image_url,
      price_reference: applyPrintPrices(p, foil as FoilTreatment, f.price_source, f.price_reference),
    }));
  };

  const handleSourceChange = (src: PriceSource) => {
    const bestPrint = findMatchingPrint(scryfallPrints, form.set_code, form.card_treatment, form.collector_number, form.promo_type, form.foil_treatment);
    setForm(f => ({
      ...f, price_source: src,
      price_reference: applyPrintPrices(bestPrint, f.foil_treatment, src, f.price_reference)
    }));
  };

  const isStoreExclusive = form.category === 'store_exclusives';

  const TABS: { id: TabId; label: string; show: boolean }[] = [
    { id: 'deck', label: t('components.admin.product_modal.tab_deck', 'DECK BUILDER'), show: isStoreExclusive },
    { id: 'variant', label: t('components.admin.product_modal.tab_variant', 'VARIANT & IDENTITY'), show: isMTGSingles },
    { id: 'pricing', label: t('components.admin.product_modal.tab_pricing', 'PRICING & STOCK'), show: true },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-2 md:p-4"
      style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(12px)' }}>
      <div className="card p-0 w-full max-w-6xl max-h-[94vh] flex flex-col border-white/20 shadow-2xl animate-in fade-in zoom-in duration-300" 
        style={{ position: 'relative', background: 'rgba(255, 255, 255, 0.85)', backdropFilter: 'blur(20px)', overflow: 'visible' }}>

        <div className="flex items-center justify-between p-3 md:p-6 pb-2 border-b border-ink-border/5">
          <div className="flex flex-col">
            <h2 className="font-display text-2xl m-0 tracking-tighter text-ink-deep">{editProduct ? t('components.admin.product_modal.title_edit', 'EDIT PRODUCT') : t('components.admin.product_modal.title_new', 'NEW PRODUCT')}</h2>
            <p className="font-mono-stack text-[10px] text-text-muted opacity-50">{t('components.admin.product_modal.product_id', 'PRODUCT ID: {id}', { id: form.id || 'NEW' })}</p>
          </div>
          <button onClick={onClose} 
            className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-hp-color/10 text-text-muted hover:text-hp-color transition-all duration-300">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <div className="flex-1 overflow-y-auto overflow-x-hidden custom-scrollbar">
          <div className="px-4 md:px-6 pt-2 grid grid-cols-2 md:grid-cols-3 lg:flex gap-3 md:gap-4">
          <div style={{ minWidth: '160px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.tcg_system_label', 'TCG SYSTEM')}</label>
            <select 
              className="bg-white/50 border-white/40 focus:bg-white transition-all" 
              value={form.tcg} 
              onChange={e => {
                const newTcg = e.target.value;
                const newSource: PriceSource = (newTcg === 'mtg' && (form.category === 'singles' || form.category === 'store_exclusives')) ? 'cardkingdom' : 'manual';
                setForm(f => ({ ...EMPTY_FORM, tcg: newTcg, category: f.category, condition: f.condition, price_source: newSource }));
                setScryfallPrints([]);
                if (!editProduct) {
                  if (form.category === 'store_exclusives') setActiveTab('deck');
                  else if (newTcg === 'mtg' && form.category === 'singles') setActiveTab('variant');
                  else setActiveTab('pricing');
                }
              }}
            >
              {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              <option value="accessories">{t('pages.common.accessories', 'Accessories')}</option>
            </select>
          </div>
          <div style={{ minWidth: '140px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.category_label', 'CATEGORY')}</label>
            <select 
              className="bg-white/50 border-white/40 focus:bg-white transition-all" 
              value={form.category} 
              onChange={e => {
                const newCat = e.target.value as 'singles' | 'sealed' | 'accessories' | 'store_exclusives';
                const newSource: PriceSource = (form.tcg === 'mtg' && (newCat === 'singles' || newCat === 'store_exclusives')) ? 'cardkingdom' : 'manual';
                setForm(f => ({ ...EMPTY_FORM, tcg: f.tcg, category: newCat, condition: f.condition, price_source: newSource }));
                setScryfallPrints([]);
                if (!editProduct) {
                  if (newCat === 'store_exclusives') setActiveTab('deck');
                  else if (newCat === 'singles' && form.tcg === 'mtg') setActiveTab('variant');
                  else setActiveTab('pricing');
                }
              }}
            >
              <option value="singles">{t('pages.common.singles', 'Singles')}</option>
              <option value="sealed">{t('pages.common.sealed', 'Sealed')}</option>
              <option value="accessories">{t('pages.common.accessories', 'Accessories')}</option>
              <option value="store_exclusives">{t('pages.common.store_exclusives', 'Store Exclusives')}</option>
            </select>
          </div>
          <div style={{ minWidth: '100px' }}>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase tracking-widest" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.condition_label', 'CONDITION')}</label>
            <select className="bg-white/50 border-white/40 focus:bg-white transition-all text-xs" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value as Condition }))}>
              {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>

          {!isMTGSingles && (
            <div className="flex-1 min-w-[120px]">
              <label className="text-[10px] font-mono-stack mb-1 block uppercase opacity-50 tracking-widest" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.product_name_label', 'PRODUCT NAME')}</label>
              <input 
                type="text" 
                value={form.name} 
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))} 
                className="bg-white/50 border-white/40 focus:bg-white transition-all"
                style={{ fontSize: '0.9rem', fontWeight: 600, height: '36px' }} 
                placeholder={t('components.admin.product_modal.product_name_label', 'Product Name')} 
              />
            </div>
          )}
        </div>

        {isMTGSingles && (
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
        )}

        <div className="sticky top-0 z-30 bg-white/95 backdrop-blur-sm px-4 md:px-6 mt-2 md:mt-6 flex gap-3 border-b border-ink-border/20 overflow-x-auto overflow-y-hidden no-scrollbar whitespace-nowrap">
          {TABS.filter(t => t.show).map(tab => {
            const isActive = activeTab === tab.id;
            return (
              <button 
                key={tab.id} 
                onClick={() => setActiveTab(tab.id)}
                className={`
                  font-mono-stack text-xs px-6 py-2 transition-all tracking-widest uppercase relative -mb-px rounded-t-md border-x border-t-[4px] group
                  ${isActive 
                    ? 'text-gold bg-white border-gold border-x-ink-border/20 border-b-white z-20 shadow-[0_-5px_15px_rgba(186,155,74,0.1)] font-black' 
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

          <div className="flex gap-4 flex-col md:flex-row p-4 md:p-6 pb-4">
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
                onSelectPrint={handleSelectPrint}
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
            {activeTab === 'deck' && isStoreExclusive && (
              <>
                <DeckCardsTab 
                  form={form} 
                  onUpdate={u => setForm(f => ({ ...f, ...u }))} 
                />
                
                {/* Horizontal actions for Deck Builder since sidebar is hidden */}
                <div className="flex flex-row items-center justify-end gap-3 mt-8 pt-6 border-t border-ink-border/20">
                  <button 
                    onClick={() => handleSave(false)} 
                    disabled={saving} 
                    className="btn-primary px-8 py-3 text-sm font-bold shadow-[0_10px_30px_rgba(212,175,55,0.3)] hover:shadow-[0_15px_40px_rgba(212,175,55,0.4)] transition-all active:scale-95"
                  >
                    {saving ? '📥 SAVING...' : '💾 SAVE PRODUCT'}
                  </button>
                  <button 
                    onClick={() => handleSave(true)} 
                    disabled={saving} 
                    className="btn-secondary px-6 py-3 text-[10px] font-bold font-mono-stack tracking-widest border-ink-border/20"
                  >
                    💾 SAVE & ADD NEW
                  </button>
                  <button 
                    onClick={onClose} 
                    disabled={saving} 
                    className="px-4 text-[9px] font-mono-stack text-text-muted hover:text-hp-color transition-colors tracking-widest opacity-60"
                  >
                    CANCEL
                  </button>
                </div>
              </>
            )}
          </div>

          {activeTab !== 'deck' && (
            <div className="w-full md:w-80 shrink-0">
              <div className="flex justify-between items-center mb-3">
                 <label className="text-[10px] font-mono-stack uppercase tracking-tighter opacity-50" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.image_preview_label', 'IMAGE PREVIEW')}</label>
                 <span className="text-[10px] font-mono-stack px-2 py-0.5 rounded-full font-bold shadow-sm" style={{ background: 'var(--nm-color)', color: 'white' }}>{form.condition}</span>
              </div>
              <div className="card p-2 bg-white/40 border-white/30 backdrop-blur-sm overflow-visible group mb-8 shadow-xl">
                <div className="relative aspect-[63/88] w-full bg-ink-border/5 rounded shadow-inner flex items-center justify-center overflow-hidden">
                  <CardImage imageUrl={form.image_url} name={form.name} tcg={form.tcg} foilTreatment={form.foil_treatment} />
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
                  {saving ? t('components.admin.product_modal.saving', '📥 SAVING...') : t('components.admin.product_modal.save_btn', '💾 SAVE PRODUCT')}
                </button>
                <button 
                  onClick={() => handleSave(true)} 
                  disabled={saving} 
                  className="btn-secondary w-full py-4 text-[10px] font-bold font-mono-stack tracking-widest border-ink-border/20"
                >
                  {t('components.admin.product_modal.save_and_new_btn', '💾 SAVE & ADD NEW')}
                </button>
                <button onClick={onClose} disabled={saving} className="w-full py-2 text-[9px] font-mono-stack text-text-muted hover:text-hp-color transition-colors tracking-widest opacity-60">
                  {t('components.admin.product_modal.cancel_btn', 'CANCEL')}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  </div>
);
}
