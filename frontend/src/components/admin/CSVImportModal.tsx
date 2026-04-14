'use client';

import { useState } from 'react';
import Papa from 'papaparse';
import {
  BulkProductInput,
  StoredIn,
  CustomCategory,
  FoilTreatment,
  CardTreatment,
  Condition,
  FOIL_LABELS,
  TREATMENT_LABELS,
  TCG_LABELS,
  KNOWN_TCGS
} from '@/lib/types';

const PROMO_LABELS: Record<string, string> = {
  '': 'Regular / No Promo',
  'judgepromo': 'Judge Promo',
  'prerelease': 'Prerelease Promo',
  'release': 'Release Promo',
  'bundle': 'Bundle Promo',
  'buyabox': 'Buy-A-Box',
  'giftbundle': 'Gift Bundle',
  'storechampionship': 'Store Champ',
  'gameday': 'Game Day',
  'starterdeck': 'Starter Deck',
  'planeswalkerdeck': 'PW Deck',
  'bringafriend': 'Bring-a-Friend',
};
import {
  lookupMTGCard,
  adminBulkCreateProducts,
  adminBatchLookupMTG
} from '@/lib/api';
import { identifyFoilFromString } from '@/lib/mtg-logic';
import CardImage from '@/components/CardImage';

interface Props {
  storageLocations: StoredIn[];
  categories: CustomCategory[];
  onClose: () => void;
  onImported: () => void;
}

interface PreviewItem extends BulkProductInput {
  is_loading?: boolean;
}

type Step = 'upload' | 'mapping' | 'preview' | 'importing' | 'summary';

const FIELDS = [
  { key: 'name', label: 'Card Name', required: false },
  { key: 'set_code', label: 'Set Code', required: false },
  { key: 'collector_number', label: 'Collector #', required: false },
  { key: 'scryfall_id', label: 'Scryfall ID', required: false },
  { key: 'condition', label: 'Condition', required: false },
  { key: 'foil_treatment', label: 'Foil', required: false },
  { key: 'stock', label: 'Quantity', required: false },
] as const;

export default function CSVImportModal({ storageLocations, categories, onClose, onImported }: Props) {
  const [step, setStep] = useState<Step>('upload');
  const [csvData, setCsvData] = useState<Record<string, string>[]>([]);
  const [headers, setHeaders] = useState<string[]>([]);
  const [mapping, setMapping] = useState<Record<string, string>>({});
  const [previewData, setPreviewData] = useState<PreviewItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [processedCount, setProcessedCount] = useState(0);
  const [error, setError] = useState('');
  const [importResults, setImportResults] = useState<{ message: string; count: number } | null>(null);

  // Global settings for the import
  const [defaultStorage, setDefaultStorage] = useState(storageLocations[0]?.id || '');
  const [bulkCategoryIds, setBulkCategoryIds] = useState<string[]>([]);
  const [importDestination, setImportDestination] = useState<'singles' | 'deck'>('singles');
  const [deckName, setDeckName] = useState('');
  const [tcgType, setTcgType] = useState('mtg');

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    Papa.parse(file, {
      header: true,
      skipEmptyLines: true,
      complete: (results) => {
        setCsvData(results.data as Record<string, string>[]);
        if (results.meta.fields) {
          setHeaders(results.meta.fields);
          // Initial auto-mapping
          const initialMapping: Record<string, string> = {};
          results.meta.fields.forEach(h => {
            const lower = h.toLowerCase().replace(/[^a-z0-9]/g, '');
            if (lower.includes('name') || lower === 'card') initialMapping['name'] = h;
            if (lower.includes('scryfall') || lower === 'sid' || lower === 'uuid') initialMapping['scryfall_id'] = h;
            if (lower.includes('set') || lower === 'code') initialMapping['set_code'] = h;
            if (lower.includes('number') || lower === 'cn' || lower === 'collector') initialMapping['collector_number'] = h;
            if (lower.includes('condition')) initialMapping['condition'] = h;
            if (lower.includes('foil')) initialMapping['foil_treatment'] = h;
            if (lower.includes('qty') || lower.includes('quantity') || lower.includes('stock')) initialMapping['stock'] = h;
          });
          setMapping(initialMapping);
          setStep('mapping');
        }
      },
      error: (err) => setError('Failed to parse CSV: ' + err.message)
    });
  };

  const handleStartPreview = () => {
    if (!mapping['name'] && !mapping['scryfall_id']) {
      setError('Either Card Name or Scryfall ID mapping is required.');
      return;
    }
    if (importDestination === 'deck' && !deckName.trim()) {
      setError('Deck Name is required for Deck imports.');
      return;
    }

    const data: PreviewItem[] = csvData.map(row => {
      const p: PreviewItem = {
        tcg: tcgType,
        category: importDestination === 'singles' ? 'singles' : 'store_exclusives',
        name: row[mapping['name']] || '',
        scryfall_id: row[mapping['scryfall_id']] || '',
        set_code: row[mapping['set_code']] || '',
        collector_number: row[mapping['collector_number']] || '',
        condition: (row[mapping['condition']] || 'NM').toUpperCase() as Condition,
        foil_treatment: identifyFoilFromString(row[mapping['foil_treatment']]),
        stock: parseInt(row[mapping['stock']]) || 1,
        category_ids: [...bulkCategoryIds], // Pre-populate with current global selection
        storage_items: defaultStorage ? [{ stored_in_id: defaultStorage, name: '', quantity: parseInt(row[mapping['stock']]) || 1 }] : []
      };
      return p;
    });

    setPreviewData(data);
    setStep('preview');
  };

  const handleEnrich = async () => {
    if (tcgType !== 'mtg') {
      setError('Automatic enrichment is only available for Magic: The Gathering.');
      return;
    }

    setLoading(true);
    setProcessedCount(0);
    setError('');
    const newData = [...previewData];
    const totalToEnrich = newData.length;
    const CHUNK_SIZE = 75;

    try {
      for (let i = 0; i < newData.length; i += CHUNK_SIZE) {
        const chunk = newData.slice(i, i + CHUNK_SIZE);
        const identifiers = chunk.map(item => ({
          name: item.name,
          set: item.set_code,
          cn: item.collector_number,
          scryfall_id: item.scryfall_id
        }));

        try {
          const results = await adminBatchLookupMTG(identifiers);

          results.forEach((res) => {
            // Find the best match in the chunk
            const matchingIndexInChunk = chunk.findIndex((item, idx) => {
              const dataIndex = i + idx;
              if (newData[dataIndex].image_url) return false;

              if (item.set_code && item.collector_number && res.set_code && res.collector_number) {
                return item.set_code.toLowerCase() === res.set_code.toLowerCase() &&
                  item.collector_number.toLowerCase() === res.collector_number.toLowerCase();
              }
              if (item.name && res.name) {
                const n1 = item.name.toLowerCase().replace(/[^a-z0-9]/g, '');
                const n2 = res.name.toLowerCase().replace(/[^a-z0-9]/g, '');
                return n1 === n2 || n2.includes(n1) || n1.includes(n2);
              }
              return false;
            });

            if (matchingIndexInChunk !== -1) {
              const dataIndex = i + matchingIndexInChunk;
              const item = newData[dataIndex];
              newData[dataIndex] = {
                ...item,
                name: res.name || item.name,
                set_code: res.set_code || item.set_code,
                set_name: res.set_name,
                collector_number: res.collector_number || item.collector_number,
                image_url: res.image_url,
                price_reference: item.tcg === 'mtg' && res.price_cardkingdom !== undefined ? res.price_cardkingdom : ((item.foil_treatment === 'non_foil' ? res.price_tcgplayer : res.price_cardmarket) || item.price_reference),
                price_source: item.tcg === 'mtg' ? 'cardkingdom' : (item.foil_treatment === 'non_foil' ? 'tcgplayer' : 'cardmarket'),
                rarity: res.rarity,
                cmc: res.cmc,
                color_identity: res.color_identity,
                language: res.language || 'en',
                is_legendary: res.is_legendary,
                is_historic: res.is_historic,
                is_land: res.is_land,
                is_basic_land: res.is_basic_land,
                art_variation: res.art_variation,
                oracle_text: res.oracle_text,
                artist: res.artist,
                type_line: res.type_line,
                border_color: res.border_color,
                frame: res.frame,
                full_art: res.full_art,
                textless: res.textless,
                foil_treatment: res.foil_treatment || item.foil_treatment,
                card_treatment: res.card_treatment || item.card_treatment,
                promo_type: res.promo_type || item.promo_type,
                scryfall_id: res.scryfall_id || item.scryfall_id,
              };

              if (item.tcg === 'mtg') {
                newData[dataIndex].price_source = 'cardkingdom';
              } else if (item.foil_treatment === 'non_foil' && res.price_tcgplayer) {
                newData[dataIndex].price_reference = res.price_tcgplayer;
                newData[dataIndex].price_source = 'tcgplayer';
              } else if (item.foil_treatment !== 'non_foil' && res.price_cardmarket) {
                newData[dataIndex].price_reference = res.price_cardmarket;
                newData[dataIndex].price_source = 'cardmarket';
              }
            }
          });

          setPreviewData([...newData]); // Update after batch
          setProcessedCount(prev => Math.min(prev + chunk.length, totalToEnrich));
        } catch (e) {
          console.warn(`Batch lookup failed for chunk ${i}-${i + CHUNK_SIZE}`, e);
        }
      }

      // Final pass: Single lookup for anything still missing data (fuzzy fallback)
      for (let i = 0; i < newData.length; i++) {
        const item = newData[i];
        if (!item.image_url && (item.name || item.scryfall_id)) {
          try {
            const res = await lookupMTGCard(item.name!, item.set_code, item.collector_number, item.foil_treatment, item.scryfall_id);
            newData[i] = {
              ...item,
              name: res.name || item.name,
              set_code: res.set_code || item.set_code,
              set_name: res.set_name,
              collector_number: res.collector_number || item.collector_number,
              image_url: res.image_url,
              price_reference: item.tcg === 'mtg' && res.price_cardkingdom !== undefined ? res.price_cardkingdom : ((item.foil_treatment === 'non_foil' ? res.price_tcgplayer : res.price_cardmarket) || item.price_reference),
              price_source: item.tcg === 'mtg' ? 'cardkingdom' : (item.foil_treatment === 'non_foil' ? 'tcgplayer' : 'cardmarket'),
              rarity: res.rarity,
              cmc: res.cmc,
              color_identity: res.color_identity,
              language: res.language || 'en',
              is_legendary: res.is_legendary,
              is_historic: res.is_historic,
              is_land: res.is_land,
              is_basic_land: res.is_basic_land,
              art_variation: res.art_variation,
              oracle_text: res.oracle_text,
              artist: res.artist,
              type_line: res.type_line,
              border_color: res.border_color,
              frame: res.frame,
              full_art: res.full_art,
              textless: res.textless,
              foil_treatment: res.foil_treatment || item.foil_treatment,
              card_treatment: res.card_treatment || item.card_treatment,
              promo_type: res.promo_type || item.promo_type,
              scryfall_id: res.scryfall_id || item.scryfall_id,
            };
          } catch (e) {
            console.warn(`Fuzzy fallback failed for ${item.name}`, e);
          }
        }
        if (i % 10 === 0) setProcessedCount(totalToEnrich * 0.9 + (i / newData.length) * 0.1 * totalToEnrich);
      }
      setProcessedCount(totalToEnrich);
      setPreviewData([...newData]);
    } finally {
      setLoading(false);
      setTimeout(() => setProcessedCount(0), 1000);
    }
  };

  const handleEnrichSingle = async (index: number) => {
    const item = previewData[index];
    if (!item.name && !item.scryfall_id) return;

    // Use a temporary flag for loading UI
    setPreviewData(prev => {
      const next = [...prev];
      next[index].is_loading = true;
      return next;
    });

    try {
      const res = await lookupMTGCard(item.name!, item.set_code, item.collector_number, item.foil_treatment, item.scryfall_id);
      setPreviewData(prev => {
        const next = [...prev];
        next[index] = {
          ...next[index],
          name: res.name || next[index].name,
          set_code: res.set_code || next[index].set_code,
          set_name: res.set_name,
          collector_number: res.collector_number || next[index].collector_number,
          image_url: res.image_url,
          price_reference: item.tcg === 'mtg' && res.price_cardkingdom !== undefined ? res.price_cardkingdom : ((item.foil_treatment === 'non_foil' ? res.price_tcgplayer : res.price_cardmarket) || item.price_reference),
          price_source: item.tcg === 'mtg' ? 'cardkingdom' : (item.foil_treatment === 'non_foil' ? 'tcgplayer' : 'cardmarket'),
          rarity: res.rarity,
          cmc: res.cmc,
          color_identity: res.color_identity,
          language: res.language || 'en',
          is_legendary: res.is_legendary,
          is_historic: res.is_historic,
          is_land: res.is_land,
          is_basic_land: res.is_basic_land,
          art_variation: res.art_variation,
          oracle_text: res.oracle_text,
          artist: res.artist,
          type_line: res.type_line,
          border_color: res.border_color,
          frame: res.frame,
          full_art: res.full_art,
          textless: res.textless,
          foil_treatment: res.foil_treatment || item.foil_treatment,
          card_treatment: res.card_treatment || item.card_treatment,
          promo_type: res.promo_type || item.promo_type,
          scryfall_id: res.scryfall_id || item.scryfall_id,
        };
        next[index].is_loading = false;
        return next;
      });
    } catch (e) {
      console.error("Single enrichment failed", e);
      setPreviewData(prev => {
        const next = [...prev];
        next[index].is_loading = false;
        return next;
      });
    }
  };

  const handleImport = async () => {
    setLoading(true);
    setStep('importing');
    try {
      let dataToImport = previewData;

      if (importDestination === 'deck') {
        const deckProduct: BulkProductInput = {
          name: deckName,
          tcg: tcgType,
          category: 'store_exclusives',
          stock: 1,
          price_source: 'manual',
          price_reference: 0,
          image_url: previewData[0]?.image_url, // Use first card as deck image
          deck_cards: previewData.map(item => ({
            name: item.name!,
            quantity: item.stock || 1,
            set_code: item.set_code,
            collector_number: item.collector_number,
            image_url: item.image_url,
            foil_treatment: item.foil_treatment,
            card_treatment: item.card_treatment,
            rarity: item.rarity,
            art_variation: item.art_variation,
            type_line: item.type_line,
          })),
          category_ids: [],
          storage_items: defaultStorage ? [{ stored_in_id: defaultStorage, name: '', quantity: 1 }] : []
        };
        dataToImport = [deckProduct];
      }

      const res = await adminBulkCreateProducts({
        products: dataToImport,
        category_ids: [] // Already pre-populated in individual objects to allow overrides
      });
      setImportResults(res);
      setStep('summary');
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
      setStep('preview');
    } finally {
      setLoading(false);
    }
  };

  const updatePreviewItem = (index: number, updates: Partial<PreviewItem>) => {
    setPreviewData(prev => {
      const next = [...prev];
      next[index] = { ...next[index], ...updates };
      return next;
    });
  };

  const toggleCategory = (index: number, catId: string) => {
    const item = previewData[index];
    const current = item.category_ids || [];
    const next = current.includes(catId) ? current.filter(id => id !== catId) : [...current, catId];
    updatePreviewItem(index, { category_ids: next });
  };

  const updateStorage = (index: number, storageId: string, qty: number) => {
    updatePreviewItem(index, {
      storage_items: [{ stored_in_id: storageId, name: '', quantity: qty }],
      stock: qty
    });
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[1000] flex items-center justify-center p-4">
      <div className="bg-bg-page border-4 border-border-main rounded-lg shadow-[0_0_50px_rgba(0,0,0,0.6)] w-full max-w-[95vw] flex flex-col max-h-[95vh] relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-accent-primary" />
        {/* Header */}
        <div className="p-6 border-b-4 border-border-main bg-bg-header flex justify-between items-center relative">
          <div>
            <h2 className="text-3xl font-black text-text-on-header tracking-tighter uppercase italic drop-shadow-md">
              CSV IMPORT TOOL
            </h2>
            <p className="text-text-on-header/60 text-sm font-medium">Bulk add products to the inventory manifest</p>
          </div>
          <button
            onClick={onClose}
            className="w-10 h-10 border-2 border-border-main/30 bg-bg-card/10 text-text-on-header flex items-center justify-center hover:bg-hp-color hover:text-white transition-all transform hover:rotate-90 active:scale-90"
          >
            ✕
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-8 custom-scrollbar bg-bg-surface/30">
          {error && (
            <div className="mb-6 p-4 bg-hp-color/10 border-2 border-hp-color text-hp-color font-bold animate-fade-up flex items-center gap-3">
              <span className="text-xl">⚠️</span> {error}
            </div>
          )}

          {step === 'upload' && (
            <div className="flex flex-col items-center justify-center py-24 border-4 border-dashed border-border-main bg-bg-card hover:bg-bg-surface transition-all cursor-pointer relative group rounded-xl shadow-inner">
              <input
                type="file"
                accept=".csv"
                onChange={handleFileUpload}
                className="absolute inset-0 opacity-0 cursor-pointer z-10"
              />
              <div className="text-7xl mb-6 group-hover:scale-110 group-hover:rotate-6 transition-transform drop-shadow-xl">📄</div>
              <h3 className="text-3xl font-black text-text-main mb-2 uppercase tracking-tight italic">Drop CSV File Here</h3>
              <p className="text-text-muted font-bold uppercase tracking-widest text-[10px]">or click to browse your local archives</p>
              <div className="mt-8 px-6 py-2 bg-accent-primary text-text-on-accent font-black uppercase text-xs tracking-widest shadow-lg">Select Master Manifest</div>
            </div>
          )}

          {step === 'mapping' && (
            <div className="animate-fade-up">
              <div className="mb-8 p-4 md:p-8 bg-bg-card border-2 border-border-main shadow-xl grid grid-cols-1 lg:grid-cols-2 gap-10 rounded-lg overflow-hidden relative">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent-primary/5 -rotate-45 translate-x-16 -translate-y-16" />
                <div>
                  <h4 className="text-accent-primary font-black mb-6 uppercase text-sm tracking-[0.2em] italic">System Configuration</h4>
                  <div className="space-y-6">
                    <div>
                      <label className="block text-[10px] font-black text-text-muted uppercase tracking-widest mb-3">Inventory Destination</label>
                      <div className="flex gap-3">
                        <button
                          onClick={() => setImportDestination('singles')}
                          className={`flex-1 py-4 px-4 border-2 font-black uppercase text-[10px] tracking-widest transition-all shadow-sm ${importDestination === 'singles'
                            ? 'bg-accent-primary border-accent-primary text-text-on-accent'
                            : 'bg-bg-page border-border-main text-text-secondary hover:border-accent-primary'
                            }`}
                        >
                          Individual Singles
                        </button>
                        <button
                          onClick={() => setImportDestination('deck')}
                          className={`flex-1 py-4 px-4 border-2 font-black uppercase text-[10px] tracking-widest transition-all shadow-sm ${importDestination === 'deck'
                            ? 'bg-accent-primary border-accent-primary text-text-on-accent'
                            : 'bg-bg-page border-border-main text-text-secondary hover:border-accent-primary'
                            }`}
                        >
                          Curated Deck
                        </button>
                      </div>
                    </div>

                    {importDestination === 'deck' && (
                      <div className="animate-fade-up">
                        <label className="block text-[10px] font-black text-text-muted uppercase tracking-widest mb-2">Manifest Identifier (Deck Name)</label>
                        <input
                          type="text"
                          value={deckName}
                          onChange={(e) => setDeckName(e.target.value)}
                          placeholder="e.g. MONO BLACK CONTROL"
                          className="w-full bg-bg-page border-2 border-border-main p-4 text-text-main font-bold focus:border-accent-primary outline-none transition-all placeholder:text-text-muted/30 italic"
                        />
                      </div>
                    )}

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-[10px] font-black text-text-muted uppercase tracking-widest mb-2">TCG Framework</label>
                        <select
                          value={tcgType}
                          onChange={(e) => setTcgType(e.target.value)}
                          className="w-full bg-bg-page border-2 border-border-main p-3 text-text-main font-bold text-xs focus:border-accent-primary outline-none transition-all"
                        >
                          {KNOWN_TCGS.map(id => (
                            <option key={id} value={id}>{TCG_LABELS[id] || id}</option>
                          ))}
                        </select>
                      </div>
                      <div>
                        <label className="block text-[10px] font-black text-text-muted uppercase tracking-widest mb-2">Vault Allocation</label>
                        <select
                          value={defaultStorage}
                          onChange={(e) => setDefaultStorage(e.target.value)}
                          className="w-full bg-bg-page border-2 border-border-main p-3 text-text-main font-bold text-xs focus:border-accent-primary outline-none transition-all"
                        >
                          <option value="" disabled>Select Storage Location</option>
                          {storageLocations.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                        </select>
                      </div>
                    </div>

                    <div>
                      <label className="block text-[10px] font-black text-text-muted uppercase tracking-widest mb-2">Metadata Tagging (Bulk)</label>
                      <div className="flex flex-wrap gap-2 p-4 bg-bg-page/50 border-2 border-border-main/50 rounded shadow-inner">
                        {categories.map(cat => (
                          <button
                            key={cat.id}
                            onClick={() => {
                              setBulkCategoryIds(prev =>
                                prev.includes(cat.id) ? prev.filter(id => id !== cat.id) : [...prev, cat.id]
                              );
                            }}
                            className={`px-3 py-1.5 text-[10px] font-black uppercase border-2 transition-all shadow-sm ${bulkCategoryIds.includes(cat.id)
                                ? 'bg-accent-primary border-accent-primary text-text-on-accent'
                                : 'bg-bg-card border-border-main text-text-secondary hover:border-accent-primary'
                              }`}
                          >
                            {cat.name}
                          </button>
                        ))}
                        {categories.length === 0 && <span className="text-text-muted italic text-xs">No collections defined</span>}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="border-l-4 border-border-main pl-10 relative">
                   <div className="absolute top-1/2 left-0 w-4 h-4 bg-border-main rotate-45 -translate-x-[10px] -translate-y-1/2" />
                  <h4 className="text-accent-primary font-black mb-6 uppercase text-sm tracking-[0.2em] italic">Column Topology Mapping</h4>
                  <div className="space-y-4">
                    {FIELDS.map(f => (
                      <div key={f.key} className="flex items-center gap-6 group/field">
                        <label className="w-36 text-[10px] font-black text-text-muted uppercase tracking-widest whitespace-nowrap group-hover/field:text-accent-primary transition-colors">
                          {f.label} {(f.key === 'name' || f.key === 'scryfall_id') && <span className="text-hp-color">*</span>}
                        </label>
                        <div className="flex-1 relative">
                          <select
                            value={mapping[f.key] || ''}
                            onChange={(e) => setMapping(prev => ({ ...prev, [f.key]: e.target.value }))}
                            className="w-full bg-bg-page border-2 border-border-main p-3 text-text-main font-bold text-xs focus:border-accent-primary outline-none transition-all shadow-sm appearance-none pr-10"
                          >
                            <option value="">-- [Skip Selection] --</option>
                            {headers.map(h => <option key={h} value={h}>{h}</option>)}
                          </select>
                          <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-text-muted/30">▼</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-6 mt-10">
                <button
                  onClick={() => setStep('upload')}
                  className="btn-secondary px-10 italic"
                >
                  Return
                </button>
                <button
                  onClick={handleStartPreview}
                  className="btn-primary px-12 italic"
                >
                  Generate Manifest
                </button>
              </div>
            </div>
          )}

          {step === 'preview' && (
            <div className="flex flex-col h-full overflow-hidden">
              <div className="mb-6 flex flex-col gap-4 bg-bg-card p-6 border-2 border-border-main relative overflow-hidden shadow-lg rounded-lg">
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent-primary/5 -rotate-45 translate-x-16 -translate-y-16" />
                <div className="flex justify-between items-center relative z-10">
                  <div className="flex gap-4">
                    <button
                      onClick={handleEnrich}
                      disabled={loading}
                      className="btn-secondary italic flex items-center gap-2 group/btn"
                    >
                      <span className="group-hover/btn:rotate-12 transition-transform">✨</span>
                      {loading ? 'SYNTHESIZING...' : 'ENRICH WITH SCRYFALL'}
                    </button>
                    {loading && (
                      <div className="flex items-center gap-4 px-6 border-l-2 border-border-main/30 ml-2">
                        <div className="w-48 h-2 bg-bg-page rounded-full overflow-hidden border border-border-main/50 relative">
                           <div 
                             className="absolute inset-y-0 left-0 bg-accent-primary transition-all duration-300 shadow-[0_0_10px_rgba(212,175,55,0.4)]"
                             style={{ width: `${(processedCount / previewData.length) * 100}%` }}
                           />
                        </div>
                        <span className="text-[10px] font-black text-text-muted uppercase tracking-widest">{Math.round((processedCount / previewData.length) * 100)}% Complete</span>
                      </div>
                    )}
                  </div>
                  <div className="flex gap-3">
                     <span className="text-[10px] font-black text-text-muted uppercase tracking-widest self-center border-r-2 border-border-main/30 pr-4 mr-1">Found {previewData.length} Signatures</span>
                  </div>
                </div>
              </div>

              <div className="flex-1 overflow-auto border-4 border-border-main bg-bg-card custom-scrollbar shadow-inner rounded-xl relative">
                <table className="w-full text-left border-collapse min-w-[1200px]">
                  <thead className="sticky top-0 z-20 bg-bg-header text-text-on-header text-[10px] font-black uppercase tracking-widest">
                    <tr>
                      <th className="p-4 border-r border-border-main/20 w-20">Art</th>
                      <th className="p-4 border-r border-border-main/20 w-[280px]">Product Info</th>
                      <th className="p-4 border-r border-border-main/20 w-16 text-center">Cond</th>
                      <th className="p-4 border-r border-border-main/20 w-36">Treatment</th>
                      <th className="p-4 border-r border-border-main/20 w-32">Vault</th>
                      <th className="p-4 border-r border-border-main/20 w-48">Collections</th>
                      <th className="p-4 border-r border-border-main/20 w-24">Reference</th>
                      <th className="p-4 w-24 text-center">Qty</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border-main/10">
                    {previewData.map((item, idx) => (
                      <tr key={idx} className="hover:bg-accent-primary/5 transition-colors group/row italic-rows">
                        <td className="p-4 border-r border-border-main/10 relative">
                          <div className="w-14 h-20 mx-auto overflow-hidden bg-bg-page border border-border-main shadow-sm transition-transform group-hover/row:scale-110">
                            {item.image_url ? (
                              <CardImage
                                imageUrl={item.image_url}
                                name={item.name!}
                                foilTreatment={item.foil_treatment}
                                height="100%"
                                enableHover={true}
                                enableModal={true}
                              />
                            ) : (
                              <div className="w-full h-full flex items-center justify-center text-[10px] text-text-muted font-black uppercase text-center p-2 italic">Scanning...</div>
                            )}
                          </div>
                        </td>
                        <td className="p-4 border-r border-border-main/10 w-[280px]">
                          <div className="flex items-center gap-2 mb-3">
                            <input
                              value={item.name || ''}
                              onChange={(e) => updatePreviewItem(idx, { name: e.target.value })}
                              placeholder="Signature Key"
                              className="bg-transparent text-text-main font-black flex-1 text-xs focus:outline-none focus:bg-bg-page/50 px-2 border-b-2 border-border-main/30"
                            />
                            <button
                              onClick={() => handleEnrichSingle(idx)}
                              disabled={item.is_loading}
                              className={`w-9 h-9 flex-shrink-0 flex items-center justify-center border-2 border-border-main bg-bg-surface text-text-main hover:bg-accent-primary hover:text-text-on-accent hover:border-accent-primary transition-all relative group/enrich active:scale-95 shadow-sm ${item.is_loading ? 'opacity-50 animate-pulse' : ''}`}
                              title="Single Enrich"
                            >
                              {item.is_loading ? (
                                <div className="w-5 h-5 border-2 border-current border-t-transparent animate-spin rounded-full" />
                                ) : (
                                  <span className="text-lg group-hover/enrich:scale-125 transition-transform">✨</span>
                                )}
                            </button>
                          </div>
                          <div className="flex items-center gap-2 px-1">
                            <input
                              value={item.scryfall_id || ''}
                              onChange={(e) => updatePreviewItem(idx, { scryfall_id: e.target.value })}
                              placeholder="System UUID"
                              className="flex-1 bg-bg-page text-text-muted text-[10px] font-mono-stack text-center border-2 border-border-main/50 focus:outline-none focus:border-accent-primary py-1.5 rounded"
                              title="Scryfall ID"
                            />
                            <div className="flex gap-1">
                              <input
                                value={item.set_code || ''}
                                onChange={(e) => updatePreviewItem(idx, { set_code: e.target.value })}
                                placeholder="SET"
                                className="bg-bg-page text-text-muted text-[10px] font-black uppercase w-10 text-center border-2 border-border-main/50 focus:outline-none focus:border-accent-primary py-1.5 rounded"
                              />
                              <input
                                value={item.collector_number || ''}
                                onChange={(e) => updatePreviewItem(idx, { collector_number: e.target.value })}
                                placeholder="#"
                                className="bg-bg-page text-text-muted text-[10px] font-black uppercase w-12 text-center border-2 border-border-main/50 focus:outline-none focus:border-accent-primary py-1.5 rounded"
                              />
                            </div>
                          </div>
                        </td>
                        <td className="p-4 border-r border-border-main/10 w-16 text-center">
                          <select
                            value={item.condition || 'NM'}
                            onChange={(e) => updatePreviewItem(idx, { condition: e.target.value as Condition })}
                            className="bg-bg-card text-accent-primary text-[11px] font-black uppercase border-2 border-border-main shadow-sm outline-none w-full text-center py-2 hover:border-accent-primary transition-colors cursor-pointer rounded"
                          >
                            {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
                          </select>
                        </td>
                        <td className="p-4 border-r border-border-main/10 w-36">
                          <select
                            value={item.foil_treatment || 'non_foil'}
                            onChange={(e) => updatePreviewItem(idx, { foil_treatment: e.target.value as FoilTreatment })}
                            className="bg-bg-card text-text-main text-[10px] font-bold w-full border border-border-main/50 outline-none mb-1 py-1 px-2 focus:border-accent-primary rounded"
                          >
                            {Object.entries(FOIL_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                          </select>
                          <select
                            value={item.card_treatment || 'normal'}
                            onChange={(e) => updatePreviewItem(idx, { card_treatment: e.target.value as CardTreatment })}
                            className="bg-bg-card text-text-main text-[10px] font-bold w-full border border-border-main/50 outline-none mb-1 py-1 px-2 focus:border-accent-primary rounded"
                          >
                            {Object.entries(TREATMENT_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                          </select>
                          <select
                            value={item.promo_type || ''}
                            onChange={(e) => updatePreviewItem(idx, { promo_type: e.target.value })}
                            className="bg-bg-card text-text-main text-[10px] font-bold w-full border border-border-main/50 outline-none py-1 px-2 focus:border-accent-primary rounded"
                          >
                            {Object.entries(PROMO_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                          </select>
                        </td>
                        <td className="p-4 border-r border-border-main/10">
                          <select
                            value={item.storage_items?.[0]?.stored_in_id || ''}
                            onChange={(e) => updateStorage(idx, e.target.value, item.stock || 0)}
                            className="w-full bg-bg-card text-text-main text-[10px] font-black uppercase border-2 border-border-main/50 p-2 outline-none focus:border-accent-primary rounded"
                          >
                            {storageLocations.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                          </select>
                        </td>
                        <td className="p-4 border-r border-border-main/10">
                          <div className="flex flex-wrap gap-1.5">
                            {categories.map(cat => (
                              <button
                                key={cat.id}
                                onClick={() => toggleCategory(idx, cat.id)}
                                className={`px-2 py-1 text-[9px] font-black uppercase border-2 transition-all shadow-sm ${(item.category_ids || []).includes(cat.id)
                                  ? 'bg-accent-primary border-accent-primary text-text-on-accent'
                                  : 'bg-bg-page border-border-main text-text-muted hover:border-accent-primary'
                                  }`}
                              >
                                {cat.name}
                              </button>
                            ))}
                          </div>
                        </td>
                        <td className="p-4 border-r border-border-main/10">
                          <div className="text-[10px] text-text-muted font-black uppercase mb-1 tracking-widest text-center">
                            {(item.price_source === 'tcgplayer' || item.price_source === 'cardkingdom') ? 'REF (USD)' : 'REF (EUR)'}
                          </div>
                          <input
                            type="number"
                            value={item.price_reference || ''}
                            onChange={(e) => updatePreviewItem(idx, { price_reference: parseFloat(e.target.value) || 0 })}
                            className="bg-bg-page text-text-main font-mono font-bold w-full border-2 border-border-main/50 p-2 text-xs focus:border-accent-primary outline-none text-center rounded placeholder:italic"
                            step="0.01"
                          />
                        </td>
                        <td className="p-4">
                          <input
                            type="number"
                            min="1"
                            value={item.stock || 0}
                            onChange={(e) => {
                              const v = parseInt(e.target.value) || 0;
                              updateStorage(idx, item.storage_items?.[0]?.stored_in_id || '', v);
                            }}
                            className="bg-bg-page text-accent-primary font-mono font-black w-full border-4 border-border-main/50 p-2 text-sm text-center focus:border-accent-primary outline-none shadow-inner rounded"
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="mt-8 flex justify-between items-center bg-bg-header p-8 border-4 border-border-main shadow-2xl relative overflow-hidden rounded-xl group">
                 <div className="absolute inset-0 bg-accent-primary/5 opacity-0 group-hover:opacity-100 transition-opacity" />
                <button
                  onClick={() => setStep('mapping')}
                  className="btn-secondary px-10 italic relative z-10"
                >
                  Return to Topology
                </button>
                <div className="flex gap-4 relative z-10">
                  <button
                    onClick={handleImport}
                    disabled={loading || previewData.length === 0}
                    className="btn-primary px-16 py-4 italic shadow-2xl scale-110 active:scale-105"
                  >
                    {loading ? 'COMMITING...' : '⚡ FINALIZE SHIPMENT'}
                  </button>
                </div>
              </div>
            </div>
          )}

          {step === 'importing' && (
            <div className="flex flex-col items-center justify-center py-40 text-center animate-fade-up">
              <div className="relative mb-16">
                {/* Enhanced Box Animation */}
                <div className="w-40 h-40 border-8 border-border-main bg-bg-card shadow-[0_0_80px_rgba(0,0,0,0.3)] relative animate-bounce overflow-hidden">
                  <div className="absolute top-0 left-1/2 -translate-x-1/2 w-2 h-40 bg-border-main/20"></div>
                  <div className="absolute top-1/2 left-0 -translate-y-1/2 w-40 h-2 bg-border-main/20"></div>
                  <div className="absolute inset-4 border-4 border-accent-primary/20 border-dashed"></div>
                  <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-7xl drop-shadow-2xl opacity-80 rotate-12">📦</div>
                </div>
                {/* Vault Floor Effect */}
                <div className="w-32 h-6 bg-black/40 blur-xl rounded-full mt-6 mx-auto animate-pulse"></div>
              </div>

              <h3 className="text-5xl font-black text-text-main uppercase tracking-tighter mb-6 italic drop-shadow-lg">
                STOCKING THE VAULT...
              </h3>
              <div className="max-w-md space-y-4">
                <p className="text-text-muted font-bold text-xl italic animate-pulse">
                  &quot;Unpacking shipment and organizing archives&quot;
                </p>
                <div className="flex justify-center gap-2">
                  {[1, 2, 3, 4, 5].map(i => (
                    <div key={i} className="w-3 h-3 bg-accent-primary animate-bounce shadow-[0_0_10px_var(--accent-primary)]" style={{ animationDelay: `${i * 0.15}s`, borderRadius: '2px' }}></div>
                  ))}
                </div>
              </div>

              <div className="mt-16 p-6 border-4 border-border-main bg-bg-header text-[12px] font-black text-text-on-header uppercase tracking-[0.3em] shadow-xl">
                System Processing Manifest of {previewData.length} Records
              </div>
            </div>
          )}

          {step === 'summary' && importResults && (
            <div className="text-center py-32 animate-fade-up">
              <div className="text-9xl mb-10 drop-shadow-[0_0_50px_rgba(212,175,55,0.4)] animate-bounce relative">
                 <span className="relative z-10">📦</span>
                 <div className="absolute inset-0 bg-accent-primary/20 blur-3xl rounded-full -z-10" />
              </div>
              <h3 className="text-6xl font-black text-text-main uppercase mb-6 tracking-tighter italic drop-shadow-md">Manifest Success!</h3>
              <div className="max-w-xl mx-auto bg-bg-card p-10 border-4 border-border-main shadow-2xl relative mb-16 rounded-xl overflow-hidden">
                 <div className="absolute top-0 left-0 w-full h-2 bg-accent-primary" />
                 <p className="text-2xl text-text-muted mb-4 font-bold italic">
                   Successfully archived <span className="text-accent-primary font-black text-4xl mx-2">{importResults.count}</span> SKU units.
                 </p>
                 <div className="h-px bg-border-main/30 w-1/2 mx-auto my-6" />
                 <p className="text-sm text-text-muted font-black uppercase tracking-widest opacity-70 italic">{importResults.message}</p>
              </div>
              <button
                onClick={() => { onImported(); onClose(); }}
                className="btn-primary px-20 py-8 text-xl italic shadow-[0_20px_50px_rgba(0,0,0,0.3)] hover:scale-110 active:scale-105 transition-all"
              >
                Return to Command Deck
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
