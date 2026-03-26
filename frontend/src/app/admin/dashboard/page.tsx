'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import useSWR, { useSWRConfig } from 'swr';
import { useRouter } from 'next/navigation';
import {
  adminFetchProducts, adminCreateProduct, adminUpdateProduct, adminDeleteProduct,
  getAdminSettings, updateAdminSettings, lookupMTGCard, CardLookupResult,
  adminFetchStorage, adminCreateStorage, adminUpdateStorage, adminDeleteStorage,
  adminUpdateProductStorage,
  adminFetchCategories, adminCreateCategory, adminUpdateCategory, adminDeleteCategory,
  adminFetchTCGs, adminCreateTCG, adminUpdateTCG, adminDeleteTCG
} from '@/lib/api';
import { Product, FOIL_LABELS, TREATMENT_LABELS, KNOWN_TCGS, TCG_SHORT, FoilTreatment, CardTreatment, PriceSource, Settings, StoredIn, StorageLocation, CustomCategory } from '@/lib/types';
import OrdersPanel from '@/components/admin/OrdersPanel';
import TCGManager from '@/components/admin/TCGManager';
import CardImage from '@/components/CardImage';
import CSVImportModal from '@/components/admin/CSVImportModal';

interface FormState {
  name: string;
  tcg: string;
  category: 'singles' | 'sealed' | 'accessories';
  set_name: string;
  set_code: string;
  condition: string;
  foil_treatment: FoilTreatment;
  card_treatment: CardTreatment;
  price_source: PriceSource;
  price_reference: number | '';
  price_cop_override: number | '';
  stock: number;
  description: string;
  category_ids: string[];
  image_url: string;
  collector_number: string;
  promo_type: string;
  language: string;
  color_identity: string;
  rarity: string;
  cmc: number | '';
  is_legendary: boolean;
  is_historic: boolean;
  is_land: boolean;
  is_basic_land: boolean;
  art_variation: string;
  oracle_text: string;
  artist: string;
  type_line: string;
  border_color: string;
  frame: string;
  full_art: boolean;
  textless: boolean;
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

export default function AdminDashboard() {
  const router = useRouter();
  const { mutate: globalMutate } = useSWRConfig();
  const [token, setToken] = useState('');
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [storageFilter, setStorageFilter] = useState('');
  const [adminSortBy, setAdminSortBy] = useState('created_at');
  const [adminSortDir, setAdminSortDir] = useState<'asc' | 'desc'>('desc');
  const [deleteConfirm, setDeleteConfirm] = useState<{ id: string; name: string } | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Modal states
  const [showModal, setShowModal] = useState(false);
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');
  
  // Lookup states
  const [scryfallPrints, setScryfallPrints] = useState<any[]>([]);
  const [lookingUp, setLookingUp] = useState(false);

  // Storage Locations Global Modal
  const [showStorageModal, setShowStorageModal] = useState(false);
  // TCG Management Modal
  const [showTCGModal, setShowTCGModal] = useState(false);

  // Storage Locations Modal States
  const [newStorageName, setNewStorageName] = useState('');
  const [editingStorageId, setEditingStorageId] = useState<string | null>(null);
  const [editingStorageName, setEditingStorageName] = useState('');

  // Category Management Modal States
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newCategoryIsActive, setNewCategoryIsActive] = useState(true);
  const [newCategoryShowBadge, setNewCategoryShowBadge] = useState(true);
  const [newCategorySearchable, setNewCategorySearchable] = useState(true);
  const [editingCategoryId, setEditingCategoryId] = useState<string | null>(null);
  const [editingCategoryName, setEditingCategoryName] = useState('');
  const [editingCategoryIsActive, setEditingCategoryIsActive] = useState(true);
  const [editingCategoryShowBadge, setEditingCategoryShowBadge] = useState(true);
  const [editingCategorySearchable, setEditingCategorySearchable] = useState(true);
  
  // Product Storage State (inside Product Edit Modal)
  const [productStorage, setProductStorage] = useState<StorageLocation[]>([]);

  // Settings states
  const [showSettings, setShowSettings] = useState(false);
  const [showOrders, setShowOrders] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [editingSettings, setEditingSettings] = useState<Settings>({ 
    usd_to_cop_rate: 4200, 
    eur_to_cop_rate: 4600,
    contact_address: '',
    contact_phone: '',
    contact_email: '',
    contact_instagram: '',
    contact_hours: ''
  });
  const [savingSettings, setSavingSettings] = useState(false);

  // Auth check
  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) { router.push('/admin/login'); return; }
    setToken(t);
  }, [router]);

  const productKey = useMemo(() => 
    token ? ['/api/admin/products', debouncedSearch, storageFilter, page, adminSortBy, adminSortDir] : null,
    [token, debouncedSearch, storageFilter, page, adminSortBy, adminSortDir]
  );

  const { data: productRes, error: productError, isLoading: productsLoading, mutate: mutateProducts } = useSWR(
    productKey,
    ([, s, st, p, sb, sd]: any) => adminFetchProducts(token, { search: s, storage_id: st, page: p, page_size: 25, sort_by: sb, sort_dir: sd }),
    { keepPreviousData: true, revalidateOnFocus: false }
  );

  const products = productRes?.products || [];
  const total = productRes?.total || 0;
  const loading = productsLoading && !productRes;

  const { data: settings, mutate: mutateSettings } = useSWR(
    token ? '/api/admin/settings' : null,
    () => getAdminSettings(token)
  );

  const { data: storageLocations = [] } = useSWR(
    token ? '/api/admin/storage' : null,
    () => adminFetchStorage(token)
  );

  const { data: categories = [] } = useSWR(
    token ? '/api/admin/categories' : null,
    () => adminFetchCategories(token)
  );

  const { data: tcgs = [] } = useSWR(
    token ? '/api/admin/tcgs' : null,
    () => adminFetchTCGs(token)
  );

  // Sync Global locations into active Product form
  useEffect(() => {
    if (!storageLocations) return;
    setProductStorage(prev => {
      if (!showModal) return prev;
      const existingIds = new Set(prev.map(p => p.stored_in_id));
      const additions = storageLocations.filter(sl => !existingIds.has(sl.id)).map(sl => ({
        stored_in_id: sl.id,
        name: sl.name,
        quantity: 0
      }));
      if (additions.length === 0) return prev;
      return [...prev, ...additions].sort((a,b) => a.name.localeCompare(b.name));
    });
  }, [storageLocations, showModal]);

  const openCreate = () => {
    setEditProduct(null);
    setForm({ ...EMPTY_FORM });
    setProductStorage(storageLocations.map(s => ({ stored_in_id: s.id, name: s.name, quantity: 0 })));
    setFormError('');
    setScryfallPrints([]);
    setShowModal(true);
  };

  const openEdit = (p: Product) => {
    setEditProduct(p);
    setForm({
      name: p.name, tcg: p.tcg, category: p.category as typeof EMPTY_FORM['category'],
      set_name: p.set_name || '', set_code: p.set_code || '',
      condition: p.condition || 'NM',
      foil_treatment: p.foil_treatment, card_treatment: p.card_treatment,
      price_source: p.price_source || 'manual',
      price_reference: p.price_reference ?? '',
      price_cop_override: p.price_cop_override ?? '',
      stock: p.stock, description: p.description || '',
      category_ids: p.categories?.map(c => c.id) || [], 
      image_url: p.image_url || '',
      collector_number: p.collector_number || '',
      promo_type: p.promo_type || '',
      language: p.language || 'en',
      color_identity: p.color_identity || '',
      rarity: p.rarity || '',
      cmc: p.cmc ?? '',
      is_legendary: p.is_legendary,
      is_historic: p.is_historic,
      is_land: p.is_land,
      is_basic_land: p.is_basic_land,
      art_variation: p.art_variation || '',
      oracle_text: p.oracle_text || '',
      artist: p.artist || '',
      type_line: p.type_line || '',
      border_color: p.border_color || '',
      frame: p.frame || '',
      full_art: p.full_art,
      textless: p.textless,
    });
    const existingStorage = p.stored_in || [];
    setProductStorage(storageLocations.map(s => {
      const match = existingStorage.find(d => d.stored_in_id === s.id);
      return {
        stored_in_id: s.id,
        name: s.name,
        quantity: match ? match.quantity : 0
      };
    }));
    setFormError('');
    setScryfallPrints([]);
    setShowModal(true);
  };

  const openSettings = () => {
    if (settings) setEditingSettings(settings);
    setShowSettings(true);
  };

  const extractPrices = (print: any, foil: FoilTreatment) => {
    const isEtched = foil === 'etched_foil';
    const isFoil = foil !== 'non_foil' && !isEtched;
    // TCGplayer: We use Market Price exclusively (standard Scryfall 'usd' fields).
    let usd = parseFloat(isEtched ? print?.prices?.usd_etched || print?.prices?.usd_foil || print?.prices?.usd : isFoil ? print?.prices?.usd_foil || print?.prices?.usd : print?.prices?.usd);
    // Cardmarket: Scryfall's 'eur' already encapsulates Trend -> 1d -> 7d -> Avg fallback.
    let eur = parseFloat(isFoil || isEtched ? print?.prices?.eur_foil || print?.prices?.eur : print?.prices?.eur);
    return { usd: isNaN(usd) ? null : usd, eur: isNaN(eur) ? null : eur };
  };

  const extractMTGMetadata = (card: CardLookupResult | any) => {
    const isLegendary = card.is_legendary ?? (card.type_line?.toLowerCase().includes('legendary') || false);
    const typeLine = card.type_line || '';
    return {
      language: card.language || card.lang || 'en',
      color_identity: Array.isArray(card.color_identity) ? card.color_identity.join(',') : (card.color_identity || ''),
      rarity: card.rarity || 'common',
      cmc: card.cmc ?? 0,
      is_legendary: isLegendary,
      is_historic: card.is_historic ?? (
        isLegendary || 
        typeLine.toLowerCase().includes('artifact') || 
        typeLine.toLowerCase().includes('saga') || false
      ),
      is_land: card.is_land ?? (typeLine.toLowerCase().includes('land') || false),
      is_basic_land: card.is_basic_land ?? (typeLine.toLowerCase().includes('basic land') || false),
      art_variation: card.art_variation || (card.variation ? 'Variation' : ''),
      oracle_text: card.oracle_text || '',
      artist: card.artist || '',
      type_line: typeLine,
      border_color: card.border_color || '',
      frame: card.frame || '',
      full_art: !!card.full_art,
      textless: !!card.textless,
    };
  };

  const applyPrintPrices = (print: any, foil: FoilTreatment, src: PriceSource) => {
    const { usd, eur } = extractPrices(print, foil);
    if (src === 'tcgplayer') return usd ?? 0;
    if (src === 'cardmarket') return eur ?? 0;
    return 0;
  };

  const extractDescription = (print: any) => {
    return '';
  };

  const handlePopulate = async () => {
    if (!form.name.trim() && (!form.set_code.trim() || !form.collector_number.trim())) return;
    setLookingUp(true);
    setFormError('');
    try {
      const res = await fetch(`https://api.scryfall.com/cards/search?q=!"${encodeURIComponent(form.name.trim())}"+is:paper&unique=prints&order=released`);
      const body = await res.json();
      if (!body.data || body.data.length === 0) throw new Error('No printings found for that precise name.');
      const prints = body.data;
      setScryfallPrints(prints);
      
      // Smart matching: try set+cn, then just set, then first result
      let bestPrint = prints.find((p: any) => 
        p.set === form.set_code && p.collector_number === form.collector_number
      );
      if (!bestPrint && form.set_code) {
        bestPrint = prints.find((p: any) => p.set === form.set_code);
      }
      if (!bestPrint) {
        bestPrint = prints[0];
      }

      const initialTreatment = getTreatmentType(bestPrint);
      const initialFoil = (() => {
        if (bestPrint.finishes?.includes('nonfoil') && bestPrint.finishes?.length === 1) return 'non_foil';
        
        const pt = bestPrint.promo_types || [];
        const specializedFoilKeywords = ['slick', 'compleat', 'galaxy', 'surge', 'textured', 'raised', 'rainbow', 'confetti', 'ink', 'gilded', 'halo', 'glossy'];
        const bestPromo = pt.find((p: string) => specializedFoilKeywords.some(kw => p.toLowerCase().includes(kw)));
        
        if (bestPromo) return bestPromo;
        if (bestPrint.finishes?.includes('etched')) return 'etched_foil';
        if (bestPrint.finishes?.includes('glossy')) return 'glossy';
        if (bestPrint.finishes?.includes('foil')) return 'foil';
        return 'non_foil';
      })();

      setForm(f => ({
        ...f,
        set_code: bestPrint.set,
        set_name: bestPrint.set_name,
        card_treatment: initialTreatment,
        collector_number: bestPrint.collector_number,
        promo_type: bestPrint.promo_types?.join(',') || 'none',
        foil_treatment: initialFoil as FoilTreatment,
        image_url: bestPrint.image_uris?.normal || bestPrint.image_uris?.large || f.image_url,
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

  const resolveLabel = (key: string, map: Record<string, string>) => {
    if (!key || key === 'none') return 'Standard / Regular';
    if (map[key]) return map[key];
    // Dynamic formatting for future-proofing
    let label = key.replace(/([a-z])([A-Z])/g, '$1 $2'); // camelCase
    label = label.replace(/oilslick/gi, 'Oil Slick');
    label = label.replace(/stepandcompleat/gi, 'Step-and-Compleat');
    label = label.replace(/silverfoil/gi, 'Silver Foil');
    label = label.replace(/foil$/, ' Foil'); 
    label = label.replace(/_foil$/, ' Foil');
    label = label.replace(/_/g, ' '); 
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
    
    // Dynamic Fallback for unrecognized treatments
    const ignoredPromos = ['boosterfun', 'promopack', 'setpromo', 'buyabox'];
    const validPt = pt.filter((p: string) => !ignoredPromos.includes(p) && !p.includes('foil'));
    if (validPt.length > 0) return validPt[0];
    if (fx.length > 0) return fx[0];
    
    return 'normal';
  };

  const getTreatmentOptions = (setCode?: string) => {
    const targetSet = setCode || form.set_code;
    if (scryfallPrints.length === 0) return [{ value: form.card_treatment, label: resolveLabel(form.card_treatment, TREATMENT_LABELS) }];
    
    const options = new Map<string, string>();
    // Don't pre-set the current treatment as a ghost option if we are targeting a specific set
    if (!setCode) {
       options.set(form.card_treatment, resolveLabel(form.card_treatment, TREATMENT_LABELS));
    }

    scryfallPrints.forEach(p => {
      if (p.set === targetSet || !targetSet) {
        const type = getTreatmentType(p);
        options.set(type, resolveLabel(type, TREATMENT_LABELS));
      }
    });
    
    return Array.from(options.entries()).map(([value, label]) => ({ value, label }));
  };
  const findMatchingPrint = (setCode: string, treatment: CardTreatment, collectorNumber: string, promoType: string, foil?: FoilTreatment) => {
    // 1. Try to find a print that matches EVERYTHING
    let match = scryfallPrints.find(p => {
      const t = getTreatmentType(p);
      const finishes = p.finishes || [];
      const pt = p.promo_types || [];
      const ptKey = pt.join(',') || 'none';
      
      const hasFoil = !foil || (foil === 'non_foil' && finishes.includes('nonfoil')) ||
                      (foil === 'etched_foil' && finishes.includes('etched')) ||
                      (foil === 'galaxy_foil' && pt.includes('galaxyfoil')) ||
                      (finishes.includes('foil') || pt.includes(foil || ''));
      
      return p.set === setCode && t === treatment && p.collector_number === collectorNumber && ptKey === promoType && hasFoil;
    });

    if (match) return match;

    // 2. Fall back to Art + Promo match
    match = scryfallPrints.find(p => p.set === setCode && getTreatmentType(p) === treatment && p.collector_number === collectorNumber && (p.promo_types?.join(',') || 'none') === promoType);
    if (match) return match;

    // 3. Fall back to just Art match
    match = scryfallPrints.find(p => p.set === setCode && getTreatmentType(p) === treatment && p.collector_number === collectorNumber);
    if (match) return match;

    // 4. Fall back to Treatment match
    match = scryfallPrints.find(p => p.set === setCode && getTreatmentType(p) === treatment);
    
    return match || scryfallPrints.find(p => p.set === setCode) || scryfallPrints[0];
  };

  const getArtOptions = (treatment: CardTreatment, setCode?: string) => {
    const targetSet = setCode || form.set_code;
    const prints = scryfallPrints.filter(p => (p.set === targetSet || !targetSet) && getTreatmentType(p) === treatment);
    const seen = new Set();
    const results: { value: string, label: string }[] = [];
    
    prints.forEach(p => {
      if (!seen.has(p.collector_number)) {
        seen.add(p.collector_number);
        results.push({ 
          value: p.collector_number, 
          label: `Art by ${p.artist} (#${p.collector_number})` 
        });
      }
    });

    if (results.length === 0 && form.collector_number) {
       results.push({ 
         value: form.collector_number, 
         label: form.artist ? `Art by ${form.artist} (#${form.collector_number})` : `Art #${form.collector_number}` 
       });
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
        results.push({ 
          value: ptKey, 
          label: pt.length > 0 ? pt.map((t: string) => resolveLabel(t, FOIL_LABELS)).join(', ') : 'Standard / Regular' 
        });
      }
    });

    if (results.length === 0) {
       results.push({ 
         value: form.promo_type || 'none', 
         label: (!form.promo_type || form.promo_type === 'none') ? 'Standard / Regular' : resolveLabel(form.promo_type, FOIL_LABELS)
       });
    }
    
    return results;
  };

  const getFoilOptions = (treatment: CardTreatment, setCode?: string, collectorNumber?: string, promoType?: string) => {
    const targetSet = setCode || form.set_code;
    const targetNum = collectorNumber || form.collector_number;
    const targetPromo = promoType || form.promo_type;

    const matchingPrints = scryfallPrints.filter(p => 
      getTreatmentType(p) === treatment && 
      (p.set === targetSet || !targetSet) &&
      (p.collector_number === targetNum || !targetNum) &&
      ((p.promo_types?.join(',') || 'none') === targetPromo || !targetPromo)
    );
    
    const results: { value: FoilTreatment, label: string }[] = [];
    const seen = new Set<string>();

    const addOpt = (val: string, lbl?: string) => {
      if (!seen.has(val)) {
        seen.add(val);
        results.push({ value: val as FoilTreatment, label: lbl || resolveLabel(val, FOIL_LABELS) });
      }
    };

    matchingPrints.forEach(print => {
      const finishes = print.finishes || [];
      const pt = (print.promo_types || []) as string[];
      
      // 1. Physical Finishes
      finishes.forEach((f: string) => {
        const key = f === 'nonfoil' ? 'non_foil' : (f === 'foil' ? 'foil' : (f === 'etched' ? 'etched_foil' : f));
        addOpt(key);
      });
      
      // 2. Specialized Treatments (promo_types)
      const specialFinishesKw = ['foil', 'ink', 'slick', 'textured', 'gilded', 'glossy', 'rainbow', 'compleat', 'galaxy', 'surge', 'halo', 'raised'];
      pt.forEach((p: string) => {
        if (specialFinishesKw.some(kw => p.toLowerCase().includes(kw))) {
          addOpt(p); 
        }
      });
    });

    if (results.length === 0) {
      addOpt(form.foil_treatment || 'non_foil');
    }
    return results;
  };

  const handleSourceChange = (src: PriceSource) => {
    setForm(f => {
      let ref = f.price_reference;
      const print = findMatchingPrint(f.set_code, f.card_treatment, f.collector_number, f.promo_type, f.foil_treatment);
      if (print) {
        ref = applyPrintPrices(print, f.foil_treatment, src);
      }
      return { ...f, price_source: src, price_reference: ref };
    });
  };

  const handleSetChange = (newSet: string) => {
     // 1. Treatment
     const treatments = getTreatmentOptions(newSet);
     const newTreatment = treatments.find(t => t.value === form.card_treatment)?.value || treatments[0]?.value || 'regular';
     
     // 2. Art Variant
     const arts = getArtOptions(newTreatment as CardTreatment, newSet);
     const newArt = arts[0]?.value || '';

     // 3. Promo Version
     const promos = getPromoOptions(newTreatment as CardTreatment, newSet, newArt);
     const newPromo = promos[0]?.value || 'none';

     // 4. Foil
     const foils = getFoilOptions(newTreatment as CardTreatment, newSet, newArt, newPromo);
     const newFoil = foils[0]?.value || 'non_foil';
     
     // 5. Final Print
     const bestPrint = findMatchingPrint(newSet, newTreatment as CardTreatment, newArt, newPromo, newFoil as FoilTreatment);

     setForm(f => ({
       ...f,
       set_code: newSet,
       set_name: bestPrint.set_name,
       card_treatment: newTreatment as CardTreatment,
       collector_number: newArt,
       promo_type: newPromo,
       foil_treatment: newFoil as FoilTreatment,
       image_url: bestPrint.image_uris?.normal || bestPrint.image_uris?.large || f.image_url,
       description: extractDescription(bestPrint) || f.description,
       price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
       ...extractMTGMetadata(bestPrint)
     }));
  };

  const handleTreatmentChange = (newTreatment: CardTreatment) => {
    // 1. Art Variant
    const arts = getArtOptions(newTreatment, form.set_code);
    const newArt = arts[0]?.value || '';

    // 2. Promo Version
    const promos = getPromoOptions(newTreatment, form.set_code, newArt);
    const newPromo = promos[0]?.value || 'none';

    // 3. Foil
    const foils = getFoilOptions(newTreatment, form.set_code, newArt, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    
    // 4. Final Print
    const bestPrint = findMatchingPrint(form.set_code, newTreatment, newArt, newPromo, newFoil as FoilTreatment);

    setForm(f => ({
      ...f,
      card_treatment: newTreatment,
      collector_number: newArt,
      promo_type: newPromo,
      foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint.image_uris?.normal || bestPrint.image_uris?.large || f.image_url,
      description: '',
      price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleArtChange = (newArt: string) => {
    // 1. Promo Version
    const promos = getPromoOptions(form.card_treatment, form.set_code, newArt);
    const newPromo = promos[0]?.value || 'none';

    // 2. Foil
    const foils = getFoilOptions(form.card_treatment, form.set_code, newArt, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    
    // 3. Final Print
    const bestPrint = findMatchingPrint(form.set_code, form.card_treatment, newArt, newPromo, newFoil as FoilTreatment);

    setForm(f => ({
      ...f,
      collector_number: newArt,
      promo_type: newPromo,
      foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint.image_uris?.normal || bestPrint.image_uris?.large || f.image_url,
      description: '',
      price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handlePromoChange = (newPromo: string) => {
    // 1. Foil
    const foils = getFoilOptions(form.card_treatment, form.set_code, form.collector_number, newPromo);
    const newFoil = foils[0]?.value || 'non_foil';
    
    // 2. Final Print
    const bestPrint = findMatchingPrint(form.set_code, form.card_treatment, form.collector_number, newPromo, newFoil as FoilTreatment);

    setForm(f => ({
      ...f,
      promo_type: newPromo,
      foil_treatment: newFoil as FoilTreatment,
      image_url: bestPrint.image_uris?.normal || bestPrint.image_uris?.large || f.image_url,
      price_reference: applyPrintPrices(bestPrint, newFoil as FoilTreatment, f.price_source),
      ...extractMTGMetadata(bestPrint)
    }));
  };

  const handleFoilChange = (newFoil: FoilTreatment) => {
     const print = findMatchingPrint(form.set_code, form.card_treatment, form.collector_number, form.promo_type, newFoil);
     setForm(f => ({
       ...f,
       foil_treatment: newFoil,
       image_url: print?.image_uris?.normal || print?.image_uris?.large || f.image_url,
       price_reference: print ? applyPrintPrices(print, newFoil, f.price_source) : f.price_reference,
       ...(print ? extractMTGMetadata(print) : {})
      }));
  };


  const handleCreateStorage = async () => {
    if (!newStorageName.trim()) return;
    try {
      await adminCreateStorage(token, newStorageName);
      setNewStorageName('');
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleUpdateStorage = async (id: string) => {
    if (!editingStorageName.trim()) return;
    try {
      await adminUpdateStorage(token, id, editingStorageName);
      setEditingStorageId(null);
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleDeleteStorage = async (id: string, name: string, count: number = 0) => {
    let msg = `Delete location "${name}"?`;
    if (count > 0) {
      msg = `WARNING: "${name}" currently holds ${count} items!\n\nDeleting this location will instantly and permanently erase these items from your global stock.\n\nAre you sure you want to delete it?`;
    }
    if (!confirm(msg)) return;
    try {
      await adminDeleteStorage(token, id);
      globalMutate('/api/admin/storage');
    } catch (e: any) { alert(e.message); }
  };

  const handleCreateCategory = async () => {
    if (!newCategoryName.trim()) return;
    try {
      await adminCreateCategory(token, newCategoryName, undefined, newCategoryIsActive, newCategoryShowBadge, newCategorySearchable);
      setNewCategoryName('');
      setNewCategoryIsActive(true);
      setNewCategoryShowBadge(true);
      setNewCategorySearchable(true);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const handleUpdateCategory = async (id: string, slug: string) => {
    if (!editingCategoryName.trim()) return;
    try {
      await adminUpdateCategory(token, id, editingCategoryName, slug, editingCategoryIsActive, editingCategoryShowBadge, editingCategorySearchable);
      setEditingCategoryId(null);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const handleDeleteCategory = async (id: string, name: string) => {
    if (!confirm(`Delete custom category "${name}"?\nThis won't delete products, only remove the grouping.`)) return;
    try {
      await adminDeleteCategory(token, id);
      globalMutate('/api/admin/categories');
    } catch (e: any) { alert(e.message); }
  };

  const updateStoreQty = (id: string, delta: number) => {
    setProductStorage(prev => prev.map(loc => 
      loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, loc.quantity + delta) } : loc
    ));
  };

  const setStoreQty = (id: string, qty: number) => {
    setProductStorage(prev => prev.map(loc => 
      loc.stored_in_id === id ? { ...loc, quantity: Math.max(0, qty) } : loc
    ));
  };

  const handleSaveSettings = async () => {
    if (!token) return;
    setSavingSettings(true);
    try {
      const updated = await updateAdminSettings(token, editingSettings);
      mutateSettings(); // refresh settings
      setShowSettings(false);
      mutateProducts(); // refresh prices natively without hard reload
    } catch (e) {
      alert('Failed to save settings: ' + (e instanceof Error ? e.message : 'Unknown error'));
    } finally {
      setSavingSettings(false);
    }
  };

  const handleSave = async () => {
    if (!form.name || !form.tcg || !form.category) {
      setFormError('Name, TCG, and Category are required.');
      return;
    }
    setSaving(true);
    setFormError('');
    try {
      const payload: Partial<Product> & { category_ids?: string[] } = {
        name: form.name, tcg: form.tcg, category: form.category,
        set_name: form.set_name || undefined,
        set_code: form.set_code || undefined,
        condition: (form.condition || undefined) as Product['condition'],
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
      
      // Clean up irrelevant price fields depending on source
      if (payload.price_source === 'manual') payload.price_reference = undefined;
      else payload.price_cop_override = undefined;

      let pid = editProduct ? editProduct.id : '';
      if (editProduct) {
        const updated = await adminUpdateProduct(token, editProduct.id, payload);
        const storage = await adminUpdateProductStorage(token, updated.id, productStorage.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })));
        
        // Update list in-place
        mutateProducts();
      } else {
        const newP = await adminCreateProduct(token, payload);
        await adminUpdateProductStorage(token, newP.id, productStorage.map(s => ({ stored_in_id: s.stored_in_id, quantity: s.quantity })));
        mutateProducts(); // Re-fetch for new products to handle sorting/pagination
      }
      
      setShowModal(false);
    } catch (e: unknown) {
      setFormError(e instanceof Error ? e.message : 'Failed to save product.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    setDeleteConfirm({ id, name });
  };

  const confirmDelete = async () => {
    if (!deleteConfirm) return;
    try {
      await adminDeleteProduct(token, deleteConfirm.id);
      mutateProducts();
    } catch {
      alert('Failed to delete product.');
    } finally {
      setDeleteConfirm(null);
    }
  };

  const toggleAdminSort = (col: string) => {
    if (adminSortBy === col) {
      setAdminSortDir(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setAdminSortBy(col);
      setAdminSortDir(col === 'name' ? 'asc' : 'desc');
    }
    setPage(1);
  };

  const logout = () => {
    localStorage.removeItem('el_bulk_admin_token');
    router.push('/admin/login');
  };

  const totalPages = Math.ceil(total / 25);

  return (
    <div className="centered-container px-4 py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <p className="text-[10px] sm:text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>EL BULK / ADMIN</p>
          <h1 className="font-display text-4xl sm:text-5xl">PRODUCT MANAGEMENT</h1>
        </div>
        <div className="flex flex-wrap gap-2 sm:gap-3 w-full sm:w-auto">
          <button onClick={() => setShowOrders(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--nm-color)', color: 'var(--nm-color)' }}>📋 ÓRDENES</button>
          <button onClick={() => setShowCategoryModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--gold)', color: 'var(--gold)' }}>📋 COLLECTIONS</button>
          <button onClick={() => setShowTCGModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5" style={{ borderColor: 'var(--kraft-dark)', color: 'var(--kraft-dark)' }}>🃏 TCG REGISTRY</button>
          <button onClick={() => setShowStorageModal(true)} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">📦 STORAGE</button>
          <button id="admin-settings" onClick={openSettings} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">⚙ SETTINGS</button>
          <button id="admin-create-product" onClick={openCreate} className="btn-primary flex-1 sm:flex-none text-[10px] sm:text-[1.1rem] px-3 sm:px-6 py-2 sm:py-2.4">+ NEW PRODUCT</button>
          <button onClick={() => setShowImportModal(true)} className="btn-primary flex-1 sm:flex-none text-[10px] sm:text-[1rem] px-3 sm:px-6 py-2 sm:py-2.4" style={{ background: 'var(--nm-color)' }}>📂 IMPORT CSV</button>
          <button onClick={logout} className="btn-secondary flex-1 sm:flex-none text-[10px] sm:text-[0.85rem] px-3 sm:px-4 py-2 sm:py-2.5">LOG OUT</button>
        </div>
      </div>

      <div className="gold-line mb-6" />

      {/* Search & Filters */}
      <div className="flex flex-wrap gap-3 mb-4">
        <input
          id="admin-search"
          type="search"
          placeholder="Search products..."
          value={search}
          onChange={e => { setSearch(e.target.value); setPage(1); }}
          style={{ maxWidth: 300, flex: 1 }}
        />
        <select 
          value={storageFilter} 
          onChange={e => { setStorageFilter(e.target.value); setPage(1); }} 
          className="px-3 py-2 border border-kraft-dark bg-white" 
          style={{ fontSize: '0.9rem', flex: 1, maxWidth: 200, color: storageFilter ? 'var(--text-primary)' : 'var(--text-muted)' }}
        >
          <option value="">All Storage Locations</option>
          {storageLocations.map(l => (
            <option key={l.id} value={l.id}>{l.name}</option>
          ))}
        </select>
        <span className="flex items-center text-sm font-mono-stack ml-auto" style={{ color: 'var(--text-muted)' }}>
          {total} product{total !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Table */}
      <div className="card no-tilt overflow-x-auto">
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--ink-border)' }}>
              <th
                className="text-left px-2 py-3 text-xs font-mono-stack cursor-pointer select-none"
                style={{ color: 'var(--text-muted)', whiteSpace: 'nowrap', width: 40, transition: 'all 0.15s', borderTop: '2px solid transparent', borderBottom: '2px solid transparent' }}
                title="Reset sorting"
                onClick={() => { setAdminSortBy('created_at'); setAdminSortDir('desc'); setPage(1); }}
                onMouseEnter={e => {
                  if (adminSortBy !== 'created_at') {
                    e.currentTarget.style.background = 'var(--ink-surface)';
                    e.currentTarget.style.borderTopColor = 'var(--kraft-dark)';
                    e.currentTarget.style.borderBottomColor = 'var(--kraft-dark)';
                  }
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.background = 'transparent';
                  e.currentTarget.style.borderTopColor = 'transparent';
                  e.currentTarget.style.borderBottomColor = 'transparent';
                }}
              >{adminSortBy !== 'created_at' ? '↺' : ''}</th>
              {[
                { label: 'Name', key: 'name' },
                { label: 'TCG', key: 'tcg' },
                { label: 'Category', key: 'category' },
                { label: 'Set', key: 'set_name' },
                { label: 'Condition', key: 'condition' },
                { label: 'Stored In', key: '' },
                { label: 'Final Price', key: 'price' },
                { label: 'Stock', key: 'stock' },
                { label: 'Collections', key: '' },
              ].map((col, i) => {
                const isActive = col.key && adminSortBy === col.key;
                const isSortable = !!col.key;
                return (
                  <th
                    key={`${col.label}-${i}`}
                    className={`text-left px-4 py-3 text-xs font-mono-stack ${isSortable ? 'cursor-pointer select-none' : ''}`}
                    style={{
                      color: isActive ? 'var(--ink-deep)' : 'var(--text-muted)',
                      whiteSpace: 'nowrap',
                      transition: 'all 0.15s',
                      borderTop: '2px solid transparent',
                      borderBottom: '2px solid transparent',
                    }}
                    onClick={isSortable ? () => toggleAdminSort(col.key) : undefined}
                    onMouseEnter={isSortable ? e => {
                      e.currentTarget.style.background = 'var(--ink-surface)';
                      e.currentTarget.style.borderTopColor = 'var(--kraft-dark)';
                      e.currentTarget.style.borderBottomColor = 'var(--kraft-dark)';
                      e.currentTarget.style.color = 'var(--ink-deep)';
                    } : undefined}
                    onMouseLeave={isSortable ? e => {
                      e.currentTarget.style.background = 'transparent';
                      e.currentTarget.style.borderTopColor = 'transparent';
                      e.currentTarget.style.borderBottomColor = 'transparent';
                      e.currentTarget.style.color = isActive ? 'var(--ink-deep)' : 'var(--text-muted)';
                    } : undefined}
                  >
                    {col.label}
                    {isActive && (
                      <span className="ml-1 text-[10px]">{adminSortDir === 'asc' ? '▲' : '▼'}</span>
                    )}
                    {isSortable && !isActive && (
                      <span className="ml-1 text-[10px]" style={{ opacity: 0 }}>▲</span>
                    )}
                  </th>
                );
              })}
              <th className="text-left px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-muted)', whiteSpace: 'nowrap' }}></th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              Array.from({ length: 6 }).map((_, i) => (
                <tr key={i} style={{ borderBottom: '1px solid var(--ink-border)' }}>
                  {Array.from({ length: 11 }).map((_, j) => (
                    <td key={j} className="px-4 py-3">
                      <div className="skeleton" style={{ height: j === 0 ? 32 : 12, width: j === 0 ? 24 : j === 1 ? 140 : 60 }} />
                    </td>
                  ))}
                </tr>
              ))
            ) : products.length === 0 ? (
              <tr>
                <td colSpan={11} className="text-center py-12 text-sm" style={{ color: 'var(--text-muted)' }}>
                  No products found. Create one to get started.
                </td>
              </tr>
            ) : (
              products.map(p => (
                <tr key={p.id} style={{ borderBottom: '1px solid var(--ink-border)' }}
                  className="transition-colors"
                  onMouseEnter={e => (e.currentTarget.style.background = 'var(--ink-surface)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
                  <td key="thumb" className="px-2 py-2" style={{ width: 40, overflow: 'visible' }}>
                    <CardImage imageUrl={p.image_url} name={p.name} tcg={p.tcg} height={40} enableHover={true} enableModal={true} />
                  </td>
                  <td key="name" className="px-4 py-3 text-sm font-semibold" style={{ maxWidth: 200 }}>
                    <span className="line-clamp-1">{p.name}</span>
                  </td>
                  <td key="tcg" className="px-4 py-3">
                    <span className="badge" style={{ background: 'var(--ink-surface)', color: 'var(--kraft-mid)', border: '1px solid var(--ink-border)' }}>
                      {TCG_SHORT[p.tcg] || p.tcg}
                    </span>
                  </td>
                  <td key="type" className="px-4 py-3 text-xs font-mono-stack" style={{ color: 'var(--text-secondary)' }}>{p.category}</td>
                  <td key="set" className="px-4 py-3 text-xs" style={{ color: 'var(--text-muted)', maxWidth: 120 }}>
                    <span className="line-clamp-1">{p.set_name || '—'}</span>
                  </td>
                  <td key="cond" className="px-4 py-3">
                    {p.condition ? <span className={`badge badge-${p.condition.toLowerCase()}`}>{p.condition}</span> : <span style={{ color: 'var(--text-muted)' }}>—</span>}
                  </td>
                  <td key="storage" className="px-4 py-3 text-xs font-mono-stack" style={{ maxWidth: 150 }}>
                    {p.stored_in && p.stored_in.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {p.stored_in.map(s => (
                          <span key={s.stored_in_id} className="badge" style={{ background: 'var(--kraft-light)', color: 'var(--text-secondary)', border: '1px solid var(--kraft-dark)', padding: '0.1rem 0.3rem', fontSize: '0.65rem' }}>
                            {s.name}: {s.quantity}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <span className="text-text-muted italic">—</span>
                    )}
                  </td>
                  <td key="price" className="px-4 py-3 price text-sm" title={`Computed from: ${p.price_source}`}>
                    ${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
                  </td>
                  <td key="stock" className="px-4 py-3 text-sm font-mono-stack"
                    style={{ color: p.stock === 0 ? 'var(--hp-color)' : p.stock < 3 ? 'var(--mp-color)' : 'var(--text-primary)' }}>
                    {p.stock}
                  </td>
                  <td key="collections" className="px-4 py-3 text-center">
                    {p.categories && p.categories.length > 0 ? (
                      <div className="flex flex-wrap justify-center gap-1 max-w-[120px]">
                        {p.categories.map(c => (
                          <span key={c.id} className="text-[9px] px-1.5 py-0.5 rounded" style={{ background: 'var(--gold)', color: 'var(--ink-deep)' }} title={c.name}>{c.name}</span>
                        ))}
                      </div>
                    ) : (
                      <span style={{ color: 'var(--ink-border)' }}>—</span>
                    )}
                  </td>
                  <td key="actions" className="px-4 py-3">
                    <div className="flex gap-2 items-center">
                      <button
                        id={`edit-product-${p.id}`}
                        onClick={() => openEdit(p)}
                        className="btn-secondary"
                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.75rem' }}
                      >Edit</button>
                      <button
                        id={`delete-product-${p.id}`}
                        onClick={() => handleDelete(p.id, p.name)}
                        title="Delete product"
                        style={{ width: 28, height: 28, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(248,113,113,0.1)', border: '1px solid rgba(248,113,113,0.3)', color: 'var(--hp-color)', borderRadius: 4, cursor: 'pointer', fontSize: '0.85rem', transition: 'all 0.15s' }}
                        onMouseEnter={e => { e.currentTarget.style.background = 'rgba(248,113,113,0.25)'; }}
                        onMouseLeave={e => { e.currentTarget.style.background = 'rgba(248,113,113,0.1)'; }}
                      >🗑</button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-6">
          <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary" style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === 1 ? 0.4 : 1 }}>← Prev</button>
          <span className="flex items-center px-3 text-sm font-mono-stack" style={{ color: 'var(--text-secondary)' }}>{page} / {totalPages}</span>
          <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page === totalPages} className="btn-secondary" style={{ padding: '0.4rem 1rem', fontSize: '0.85rem', opacity: page === totalPages ? 0.4 : 1 }}>Next →</button>
        </div>
      )}

      {/* Settings Modal */}
      {showSettings && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-4xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)', position: 'relative' }}>
             {/* Decorative Corner */}
             <div className="absolute top-0 right-0 w-16 h-16 pointer-events-none opacity-20" style={{ borderTop: '8px solid var(--gold)', borderRight: '8px solid var(--gold)' }} />
            
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">GLOBAL SETTINGS</h2>
              <div className="px-3 py-1 bg-nm-color text-white text-xs font-mono-stack rounded shadow-sm">SYSTEM_CONFIG_V2</div>
            </div>

            <div className="grid md:grid-cols-2 gap-10">
              {/* Rates */}
              <div className="space-y-6">
                <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
                  <span className="text-2xl">📈</span>
                  <h4 className="text-lg font-display text-ink-deep m-0">EXCHANGE RATES</h4>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="cardbox p-4 bg-kraft-light/30">
                    <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>USD TO COP (TCG)</label>
                    <input type="number" className="font-bold text-lg" value={editingSettings.usd_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, usd_to_cop_rate: parseFloat(e.target.value) })} />
                  </div>
                  <div className="cardbox p-4 bg-kraft-light/30">
                    <label className="text-xs font-mono-stack mb-2 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>EUR TO COP (MCK)</label>
                    <input type="number" className="font-bold text-lg" value={editingSettings.eur_to_cop_rate} onChange={e => setEditingSettings({ ...editingSettings, eur_to_cop_rate: parseFloat(e.target.value) })} />
                  </div>
                </div>
                <p className="text-[10px] font-mono-stack text-text-muted mt-2">
                  * These rates are used to compute final COP prices from external sources.
                </p>
              </div>

              {/* Contact Info */}
              <div className="space-y-6">
                <div className="flex items-center gap-3 border-b border-kraft-dark pb-2 mb-4">
                  <span className="text-2xl">📦</span>
                  <h4 className="text-lg font-display text-ink-deep m-0">STORE IDENTITY</h4>
                </div>
                
                <div className="space-y-4">
                  <div>
                    <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>PHYSICAL ADDRESS</label>
                    <input type="text" className="bg-white" value={editingSettings.contact_address} onChange={e => setEditingSettings({ ...editingSettings, contact_address: e.target.value })} />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>WHATSAPP</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_phone} onChange={e => setEditingSettings({ ...editingSettings, contact_phone: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>INSTAGRAM</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_instagram} onChange={e => setEditingSettings({ ...editingSettings, contact_instagram: e.target.value })} />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>STORE EMAIL</label>
                      <input type="email" className="bg-white" value={editingSettings.contact_email} onChange={e => setEditingSettings({ ...editingSettings, contact_email: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block uppercase tracking-tighter" style={{ color: 'var(--text-muted)' }}>BUSINESS HOURS</label>
                      <input type="text" className="bg-white" value={editingSettings.contact_hours} onChange={e => setEditingSettings({ ...editingSettings, contact_hours: e.target.value })} />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-4 mt-12 bg-kraft-light/20 p-4 -m-8 mt-8 border-t border-kraft-dark">
              <button onClick={handleSaveSettings} className="btn-primary flex-1 shadow-md" disabled={savingSettings}>
                {savingSettings ? 'SYNCING...' : 'SAVE ENTIRE DB CONFIG →'}
              </button>
              <button onClick={() => setShowSettings(false)} className="btn-secondary px-10">DISCARD</button>
            </div>
          </div>
        </div>
      )}

      {/* Storage Locations Modal */}
      {showStorageModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">STORAGE LOCATIONS</h2>
              <button onClick={() => setShowStorageModal(false)} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <div className="flex gap-2 mb-6">
              <input type="text" placeholder="New Location Name (e.g. Binder A)" value={newStorageName} onChange={e => setNewStorageName(e.target.value)} className="flex-1 bg-white" />
              <button onClick={handleCreateStorage} className="btn-primary px-6">ADD</button>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto pr-2">
              {storageLocations.map(loc => (
                <div key={loc.id} className="flex items-center justify-between p-3 border border-kraft-dark bg-kraft-light/10">
                  {editingStorageId === loc.id ? (
                    <div className="flex gap-2 flex-1 mr-4">
                      <input type="text" value={editingStorageName} onChange={e => setEditingStorageName(e.target.value)} className="flex-1 py-1 bg-white" />
                      <button onClick={() => handleUpdateStorage(loc.id)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                      <button onClick={() => setEditingStorageId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center gap-3">
                        <span className="font-semibold text-lg">{loc.name}</span>
                        <span className="text-xs font-mono-stack text-text-muted bg-kraft-light px-2 py-0.5 rounded border border-kraft-dark">
                          {loc.item_count || 0} items
                        </span>
                      </div>
                      <div className="flex gap-2">
                        <button onClick={() => { setEditingStorageId(loc.id); setEditingStorageName(loc.name); }} className="btn-secondary px-3 py-1 text-xs">EDIT</button>
                        <button onClick={() => handleDeleteStorage(loc.id, loc.name, loc.item_count || 0)} className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" style={{ borderRadius: 4 }}>DELETE</button>
                      </div>
                    </>
                  )}
                </div>
              ))}
              {storageLocations.length === 0 && <p className="text-center text-text-muted py-8">No storage locations configured.</p>}
            </div>
          </div>
        </div>
      )}

      {/* Category Management Modal */}
      {showCategoryModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--gold)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">CUSTOM COLLECTIONS</h2>
              <button onClick={() => setShowCategoryModal(false)} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <div className="flex flex-col gap-3 mb-6 p-4 bg-kraft-light/10 border border-kraft-dark rounded">
              <div className="flex gap-2">
                <input type="text" placeholder="New Collection Name (e.g. Staples)" value={newCategoryName} onChange={e => setNewCategoryName(e.target.value)} className="flex-1 bg-white" />
                <button onClick={handleCreateCategory} className="btn-primary px-6 border-gold text-gold">ADD</button>
              </div>
              <div className="flex flex-wrap gap-x-4 gap-y-2">
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="HOME SECTION: Shows the collection as a dedicated row on the landing page.">
                  <input type="checkbox" checked={newCategoryIsActive} onChange={e => setNewCategoryIsActive(e.target.checked)} />
                  HOME SECTION
                </label>
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="NAVBAR/SEARCH: Includes the collection in navigation links and global search filters.">
                  <input type="checkbox" checked={newCategorySearchable} onChange={e => setNewCategorySearchable(e.target.checked)} />
                  NAVBAR/SEARCH
                </label>
                <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="SHOW BADGE: Displays the collection name as a tag on product cards.">
                  <input type="checkbox" checked={newCategoryShowBadge} onChange={e => setNewCategoryShowBadge(e.target.checked)} />
                  SHOW BADGE
                </label>
              </div>
            </div>

            <div className="space-y-2 max-h-96 overflow-y-auto pr-2">
              {categories.map(cat => (
                <div key={cat.id} className={`flex flex-col sm:flex-row items-start sm:items-center justify-between p-3 border ${cat.is_active ? 'border-kraft-dark bg-kraft-light/10' : 'border-ink-border bg-ink-surface'} gap-2`}>
                  {editingCategoryId === cat.id ? (
                    <div className="flex flex-col gap-3 flex-1 w-full">
                      <div className="flex gap-2">
                        <input type="text" value={editingCategoryName} onChange={e => setEditingCategoryName(e.target.value)} className="flex-1 py-1 bg-white" />
                        <button onClick={() => handleUpdateCategory(cat.id, cat.slug)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                        <button onClick={() => setEditingCategoryId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                      </div>
                      <div className="flex flex-wrap gap-x-4 gap-y-2">
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="HOME SECTION: Shows the collection as a dedicated row on the landing page.">
                          <input type="checkbox" checked={editingCategoryIsActive} onChange={e => setEditingCategoryIsActive(e.target.checked)} />
                          IS ACTIVE (HOME)
                        </label>
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="NAVBAR/SEARCH: Includes the collection in navigation links and global search filters.">
                          <input type="checkbox" checked={editingCategorySearchable} onChange={e => setEditingCategorySearchable(e.target.checked)} />
                          SEARCHABLE
                        </label>
                        <label className="flex items-center gap-2 text-[10px] font-mono-stack cursor-pointer" title="SHOW BADGE: Displays the collection name as a tag on product cards.">
                          <input type="checkbox" checked={editingCategoryShowBadge} onChange={e => setEditingCategoryShowBadge(e.target.checked)} />
                          SHOW BADGE
                        </label>
                      </div>
                    </div>
                  ) : (
                    <>
                      <div className="flex flex-col gap-1">
                        <div className="flex items-center gap-3">
                          <span className="font-semibold text-lg">{cat.name}</span>
                          <div className="flex gap-1.5 items-center">
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.is_active ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.is_active ? 'var(--gold)' : 'transparent',
                                color: cat.is_active ? 'var(--ink-deep)' : 'var(--text-muted)',
                                borderColor: cat.is_active ? 'var(--gold)' : 'var(--ink-border)',
                                opacity: cat.is_active ? 1 : 0.4
                              }}
                            >HOME</span>
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.searchable ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.searchable ? 'var(--hp-color)' : 'transparent',
                                color: cat.searchable ? 'var(--ink-surface)' : 'var(--text-muted)',
                                borderColor: cat.searchable ? 'var(--hp-color)' : 'var(--ink-border)',
                                opacity: cat.searchable ? 1 : 0.4
                              }}
                            >SEARCH</span>
                            <span 
                              className={`text-[10px] px-1.5 py-0.5 rounded border transition-colors ${cat.show_badge ? 'font-bold' : ''}`}
                              style={{ 
                                backgroundColor: cat.show_badge ? 'var(--kraft-dark)' : 'transparent',
                                color: cat.show_badge ? 'var(--ink-surface)' : 'var(--text-muted)',
                                borderColor: cat.show_badge ? 'var(--kraft-dark)' : 'var(--ink-border)',
                                opacity: cat.show_badge ? 1 : 0.4
                              }}
                            >BADGE</span>
                          </div>
                          <span className="text-xs font-mono-stack text-text-muted border border-ink-border px-2 py-0.5 rounded">
                            {cat.item_count || 0} items
                          </span>
                        </div>
                        <span className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>/collection/{cat.slug}</span>
                      </div>
                      <div className="flex gap-2">
                        <button onClick={() => { 
                          setEditingCategoryId(cat.id); 
                          setEditingCategoryName(cat.name); 
                          setEditingCategoryIsActive(cat.is_active);
                          setEditingCategoryShowBadge(cat.show_badge);
                          setEditingCategorySearchable(cat.searchable);
                        }} className="btn-secondary px-3 py-1 text-xs">EDIT</button>
                        <button onClick={() => handleDeleteCategory(cat.id, cat.name)} className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" style={{ borderRadius: 4 }}>DELETE</button>
                      </div>
                    </>
                  )}
                </div>
              ))}
              {categories.length === 0 && <p className="text-center text-text-muted py-8">No collections created.</p>}
            </div>

            <div className="mt-8 pt-6 border-t border-kraft-dark/30">
              <p className="font-mono-stack text-[10px] uppercase text-hp-color font-bold mb-3 tracking-widest flex items-center gap-2">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                VISIBILITY GUIDE
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-[9px] font-mono-stack leading-relaxed opacity-70">
                <div>
                  <span className="text-gold font-bold">HOME SECTION:</span> Shows the collection as a dedicated row on the landing page.
                </div>
                <div>
                  <span className="text-hp-color font-bold">NAVBAR/SEARCH:</span> Includes the collection in navigation links and global search filters.
                </div>
                <div>
                  <span className="text-kraft-dark font-bold">SHOW BADGE:</span> Displays the collection name as a tag on product cards.
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Product Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-4 md:pt-8 px-2 md:px-4"
          style={{ background: 'rgba(0,0,0,0.7)', backdropFilter: 'blur(3px)', overflowY: 'auto' }}>
          <div className="card no-tilt p-4 md:p-6 w-full max-w-5xl mb-8" style={{ position: 'relative' }}>
            <div className="flex items-center justify-between mb-6">
              <h2 className="font-display text-3xl">{editProduct ? 'EDIT PRODUCT' : 'NEW PRODUCT'}</h2>
              <button onClick={() => setShowModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>

            <div className="flex gap-6 flex-col md:flex-row">
              <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="sm:col-span-2 flex items-end gap-3 flex-wrap sm:flex-nowrap">
                  <div style={{ width: '100px' }}>
                    <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>SET</label>
                    {scryfallPrints.length > 0 ? (
                      <select id="form-set-code-top" value={form.set_code} onChange={e => handleSetChange(e.target.value)} className="font-bold">
                        {Array.from(new Map(scryfallPrints.map(c => [c.set, c.set_name])).entries()).map(([code, name]) => (
                          <option key={code} value={code}>[{code.toUpperCase()}] {name}</option>
                        ))}
                      </select>
                    ) : (
                      <input id="form-set-code-top" type="text" value={form.set_code} onChange={e => setForm(f => ({ ...f, set_code: e.target.value.toUpperCase() }))} placeholder="MH2" className="text-center font-bold uppercase" />
                    )}
                  </div>
                  <div style={{ width: '80px' }}>
                    <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}># CN</label>
                    <input id="form-cn-top" type="text" value={form.collector_number} onChange={e => handleArtChange(e.target.value)} placeholder="e.g. 123" className="text-center font-bold" />
                  </div>
                  <div className="flex-1 min-w-[200px]">
                    <div className="flex justify-between items-end mb-1">
                      <label className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>CARD / PRODUCT NAME *</label>
                      {form.set_name && <span className="text-[10px] font-mono-stack truncate" style={{ color: 'var(--gold)', maxWidth: '200px' }}>{form.set_name}</span>}
                    </div>
                    <input id="form-name" type="text" value={form.name} onChange={e => {
                      const val = e.target.value;
                      setForm(f => ({ ...f, name: val }));
                      setScryfallPrints([]);
                    }} />
                  </div>
                  {form.tcg === 'mtg' && form.category === 'singles' && (
                    <button type="button" onClick={handlePopulate} disabled={lookingUp || (!form.name.trim() && (!form.set_code.trim() || !form.collector_number.trim()))} className="btn-secondary px-4 transition-colors hover:text-gold" style={{ height: '42px', padding: '0 1rem', fontSize: '0.8rem' }} title="Lookup Scryfall Data">
                      {lookingUp ? '⏳...' : '📥 POPULATE'}
                    </button>
                  )}
                </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>TCG *</label>
                <select id="form-tcg" value={form.tcg} onChange={e => setForm(f => ({ ...f, tcg: e.target.value }))}>
                  {tcgs.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                  {tcgs.length === 0 && KNOWN_TCGS.map(t => <option key={t} value={t}>{TCG_SHORT[t]}</option>)}
                  <option value="accessories">Accessories</option>
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CATEGORY *</label>
                <select id="form-category" value={form.category} onChange={e => setForm(f => ({ ...f, category: e.target.value as typeof EMPTY_FORM['category'] }))}>
                  <option value="singles">Singles</option>
                  <option value="sealed">Sealed</option>
                  <option value="accessories">Accessories</option>
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CONDITION</label>
                <select id="form-condition" value={form.condition} onChange={e => setForm(f => ({ ...f, condition: e.target.value }))}>
                  {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CARD TREATMENT / VERSION</label>
                <select id="form-treatment" 
                  value={form.card_treatment} 
                  disabled={scryfallPrints.length === 0}
                  onChange={e => handleTreatmentChange(e.target.value as CardTreatment)}>
                  {getTreatmentOptions().map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>ARTIST / COLLECTOR #</label>
                <select id="form-art" 
                  value={form.collector_number} 
                  disabled={scryfallPrints.length === 0 || !form.card_treatment}
                  onChange={e => handleArtChange(e.target.value)}>
                  {getArtOptions(form.card_treatment).map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PROMO VERSION</label>
                <select id="form-promo" 
                  value={form.promo_type} 
                  disabled={scryfallPrints.length === 0 || !form.collector_number}
                  onChange={e => handlePromoChange(e.target.value)}>
                  {getPromoOptions(form.card_treatment, form.set_code, form.collector_number).map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>FOIL TREATMENT</label>
                <select id="form-foil" 
                  value={form.foil_treatment} 
                  disabled={scryfallPrints.length === 0 || !form.promo_type}
                  onChange={e => handleFoilChange(e.target.value as FoilTreatment)}>
                  {getFoilOptions(form.card_treatment, form.set_code, form.collector_number, form.promo_type).map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              
              {/* Pricing section */}
              <div className="sm:col-span-2 mt-2 pt-2" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                <div className="flex justify-between items-center mb-3">
                  <h3 className="text-sm font-mono-stack" style={{ color: 'var(--text-primary)' }}>PRICING</h3>
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
                    <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE SOURCE *</label>
                    <select id="form-price-source" value={form.price_source} onChange={e => handleSourceChange(e.target.value as PriceSource)}>
                      <option value="manual">Manual Override (COP)</option>
                      <option value="tcgplayer">External: TCGPlayer (USD)</option>
                      <option value="cardmarket">External: Cardmarket (EUR)</option>
                    </select>
                  </div>
                  {form.price_source === 'manual' ? (
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE (COP) *</label>
                      <input id="form-price-override" type="number" step="100" min="0" value={form.price_cop_override} onChange={e => setForm(f => ({ ...f, price_cop_override: e.target.value ? parseFloat(e.target.value) : '' }))} />
                    </div>
                  ) : (
                    <div>
                      <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>
                        REFERENCE PRICE ({form.price_source === 'tcgplayer' ? 'USD' : 'EUR'}) *
                      </label>
                       <input id="form-price-reference" type="number" step="0.01" min="0" value={form.price_reference} 
                        onChange={e => setForm(f => ({ ...f, price_reference: e.target.value ? parseFloat(e.target.value) : '' }))} 
                        style={{ color: form.price_reference === 0 ? 'var(--hp-color)' : 'inherit', borderColor: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--ink-border)' }} />
                    </div>
                  )}
                </div>
              </div>

              {/* MTG Metadata section */}
              {form.tcg === 'mtg' && form.category === 'singles' && (
                <div className="sm:col-span-2 mt-4 pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                  <h3 className="text-sm font-mono-stack mb-3" style={{ color: 'var(--text-primary)' }}>MTG METADATA</h3>
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

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
                    <div>
                      <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>ORACLE TEXT</label>
                      <textarea 
                        className="w-full text-[11px] font-mono-stack p-2 bg-transparent border border-ink-border rounded h-24"
                        value={form.oracle_text} 
                        onChange={e => setForm(f => ({ ...f, oracle_text: e.target.value }))}
                      />
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-x-6 gap-y-3 p-3 bg-kraft-light/10 border border-kraft-dark rounded mb-4">
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.is_legendary} onChange={e => setForm(f => ({ ...f, is_legendary: e.target.checked }))} />
                      LEGENDARY
                    </label>
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.is_historic} onChange={e => setForm(f => ({ ...f, is_historic: e.target.checked }))} />
                      HISTORIC
                    </label>
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.is_land} onChange={e => setForm(f => ({ ...f, is_land: e.target.checked }))} />
                      LAND
                    </label>
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.is_basic_land} onChange={e => setForm(f => ({ ...f, is_basic_land: e.target.checked }))} />
                      BASIC LAND
                    </label>
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.full_art} onChange={e => setForm(f => ({ ...f, full_art: e.target.checked }))} />
                      FULL ART
                    </label>
                    <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                      <input type="checkbox" checked={form.textless} onChange={e => setForm(f => ({ ...f, textless: e.target.checked }))} />
                      TEXTLESS
                    </label>
                  </div>
                </div>
              )}

              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>IMAGE URL</label>
                <input id="form-image" type="text" value={form.image_url} onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))} />
              </div>
              <div className="sm:col-span-2">
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>DESCRIPTION</label>
                <textarea id="form-description" className="w-full" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} rows={3} />
              </div>
              <div className="sm:col-span-2 mt-4 pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
                <label className="text-xs font-mono-stack mb-2 block" style={{ color: 'var(--text-muted)' }}>CUSTOM COLLECTIONS / CATEGORIES</label>
                <div className="flex flex-wrap gap-2">
                  {categories.length === 0 && <span className="text-sm text-text-muted italic">No categories created yet. Create them from the dashboard header.</span>}
                  {categories.map(c => {
                    const isSelected = form.category_ids.includes(c.id);
                    return (
                      <button
                        key={c.id}
                        type="button"
                        onClick={() => setForm(f => ({
                          ...f,
                          category_ids: isSelected ? f.category_ids.filter(id => id !== c.id) : [...f.category_ids, c.id]
                        }))}
                        className={`badge transition-colors cursor-pointer ${isSelected ? 'border-gold' : 'bg-ink-surface text-text-secondary border-ink-border hover:border-gold/50'}`}
                        style={{
                           background: isSelected ? 'var(--gold)' : '',
                           color: isSelected ? 'var(--ink-deep)' : ''
                        }}
                      >
                        {c.name}
                      </button>
                    );
                  })}
                </div>
              </div>
            </div>

            <div className="w-full md:w-64 flex-shrink-0">
              <div className="sticky top-4">
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>IMAGE PREVIEW</label>
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

                {/* STORAGE TABLE */}
                <div className="cardbox p-3 bg-kraft-light/30">
                  <div className="flex justify-between items-center mb-3">
                    <label className="text-xs font-mono-stack uppercase text-text-muted m-0">STORAGE</label>
                    <span className="text-xs font-bold bg-ink-surface px-2 py-1 text-gold rounded border border-ink-border">
                      TOTAL: {productStorage.reduce((acc, l) => acc + l.quantity, 0)}
                    </span>
                  </div>
                  
                  <div className="space-y-2 max-h-60 overflow-y-auto">
                    {productStorage.length === 0 && <p className="text-xs text-text-muted italic text-center py-2">No storage locations available.</p>}
                    {productStorage.map(loc => (
                      <div key={loc.stored_in_id} className="flex items-center justify-between gap-2 text-sm border-b border-ink-border/50 pb-2">
                        <span className="truncate flex-1 font-semibold" title={loc.name}>{loc.name}</span>
                        <div className="flex items-center gap-1">
                          <button onClick={() => updateStoreQty(loc.stored_in_id, -1)} className="w-6 h-6 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-hp-color transition-colors" disabled={loc.quantity <= 0}>-</button>
                          <input type="number" value={loc.quantity === 0 ? '' : loc.quantity} min="0" 
                            onChange={e => setStoreQty(loc.stored_in_id, parseInt(e.target.value) || 0)}
                            className="w-12 px-1 py-0 text-center text-sm font-mono-stack" style={{ height: '24px' }} placeholder="0" />
                          <button onClick={() => updateStoreQty(loc.stored_in_id, 1)} className="w-6 h-6 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-gold transition-colors">+</button>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>

            {formError && (
              <p className="mt-4 text-sm font-mono-stack" style={{ color: 'var(--hp-color)' }}>{formError}</p>
            )}

            <div className="flex gap-3 mt-6">
              <button id="admin-save-product" onClick={handleSave} className="btn-primary flex-1 py-3 text-sm" disabled={saving}>
                {saving ? 'SAVING...' : editProduct ? 'SAVE CHANGES' : 'CREATE PRODUCT'}
              </button>
              <button onClick={() => setShowModal(false)} className="btn-secondary py-3 text-sm">CANCEL</button>
            </div>
          </div>
        </div>
      )}

      {/* TCG Registry Modal */}
      {showTCGModal && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
          <div className="card no-tilt max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface', border: '4px solid var(--kraft-dark)' }}>
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-display text-4xl m-0">TCG MANAGEMENT</h2>
              <button onClick={() => { setShowTCGModal(false); globalMutate('/api/admin/tcgs'); }} className="text-text-muted hover:text-text-primary text-xl">✕</button>
            </div>
            
            <TCGManager token={token} />
          </div>
        </div>
      )}

      {/* Orders Panel */}
      {showOrders && (
        <OrdersPanel token={token} onClose={() => { setShowOrders(false); mutateProducts(); }} />
      )}

      {showImportModal && (
        <CSVImportModal 
          token={token} 
          storageLocations={storageLocations} 
          categories={categories}
          onClose={() => setShowImportModal(false)} 
          onImported={() => mutateProducts()} 
        />
      )}
      {/* Delete Confirmation Dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }} onClick={() => setDeleteConfirm(null)}>
          <div className="card no-tilt p-6 max-w-sm w-full mx-4" onClick={e => e.stopPropagation()} style={{ background: 'var(--kraft-light)' }}>
            <div className="flex items-center gap-3 mb-4">
              <span style={{ fontSize: '1.5rem' }}>🗑️</span>
              <h3 className="font-display text-xl">DELETE PRODUCT</h3>
            </div>
            <p className="text-sm mb-1" style={{ color: 'var(--text-secondary)' }}>
              Are you sure you want to delete:
            </p>
            <p className="text-sm font-semibold mb-4" style={{ color: 'var(--ink-deep)' }}>
              &quot;{deleteConfirm.name}&quot;
            </p>
            <p className="text-xs mb-6" style={{ color: 'var(--hp-color)' }}>
              This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeleteConfirm(null)}
                className="btn-secondary"
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem' }}
              >Cancel</button>
              <button
                onClick={confirmDelete}
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem', background: 'var(--hp-color)', color: 'white', border: 'none', borderRadius: 4, cursor: 'pointer', fontWeight: 600 }}
              >Delete</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
