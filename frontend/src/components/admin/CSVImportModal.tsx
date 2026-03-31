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
import { lookupMTGCard, adminBulkCreateProducts, adminBatchLookupMTG } from '@/lib/api';
import CardImage from '@/components/CardImage';

interface Props {
  token: string;
  storageLocations: StoredIn[];
  categories: CustomCategory[];
  onClose: () => void;
  onImported: () => void;
}

type Step = 'upload' | 'mapping' | 'preview' | 'importing' | 'summary';

const FIELDS = [
  { key: 'name', label: 'Card Name', required: true },
  { key: 'set_code', label: 'Set Code', required: false },
  { key: 'collector_number', label: 'Collector #', required: false },
  { key: 'condition', label: 'Condition', required: false },
  { key: 'foil_treatment', label: 'Foil', required: false },
  { key: 'stock', label: 'Quantity', required: false },
] as const;

export default function CSVImportModal({ token, storageLocations, categories, onClose, onImported }: Props) {
  const [step, setStep] = useState<Step>('upload');
  const [csvData, setCsvData] = useState<any[]>([]);
  const [headers, setHeaders] = useState<string[]>([]);
  const [mapping, setMapping] = useState<Record<string, string>>({});
  const [previewData, setPreviewData] = useState<BulkProductInput[]>([]);
  const [loading, setLoading] = useState(false);
  const [processedCount, setProcessedCount] = useState(0);
  const [error, setError] = useState('');
  const [importResults, setImportResults] = useState<{ message: string; count: number } | null>(null);

  // Global settings for the import
  const [defaultStorage, setDefaultStorage] = useState(storageLocations[0]?.id || '');
  const [defaultCategories] = useState<string[]>([]);
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
        setCsvData(results.data);
        if (results.meta.fields) {
          setHeaders(results.meta.fields);
          // Initial auto-mapping
          const initialMapping: Record<string, string> = {};
          results.meta.fields.forEach(h => {
            const lower = h.toLowerCase().replace(/[^a-z0-9]/g, '');
            if (lower.includes('name') || lower === 'card') initialMapping['name'] = h;
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
    if (!mapping['name']) {
      setError('Card Name mapping is required.');
      return;
    }
    if (importDestination === 'deck' && !deckName.trim()) {
      setError('Deck Name is required for Deck imports.');
      return;
    }

    const data: BulkProductInput[] = csvData.map(row => {
      const p: BulkProductInput = {
        tcg: tcgType,
        category: importDestination === 'singles' ? 'singles' : 'store_exclusives',
        name: row[mapping['name']] || '',
        set_code: row[mapping['set_code']] || '',
        collector_number: row[mapping['collector_number']] || '',
        condition: (row[mapping['condition']] || 'NM').toUpperCase() as Condition,
        foil_treatment: (row[mapping['foil_treatment']]?.toLowerCase().includes('foil') ? 'foil' : 'non_foil') as FoilTreatment,
        stock: parseInt(row[mapping['stock']]) || 1,
        category_ids: [...defaultCategories],
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
          cn: item.collector_number
        }));

        try {
          const results = await adminBatchLookupMTG(token, identifiers);
          
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
                price_reference: (item.foil_treatment === 'non_foil' ? res.price_tcgplayer : res.price_cardmarket) || item.price_reference,
                price_source: item.foil_treatment === 'non_foil' ? 'tcgplayer' : 'cardmarket',
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
              };
              
              if (item.foil_treatment === 'non_foil' && res.price_tcgplayer) {
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
        if (!item.image_url && item.name) {
          try {
            const res = await lookupMTGCard(token, item.name, item.set_code, item.collector_number, item.foil_treatment);
            newData[i] = {
              ...item,
              name: res.name || item.name,
              set_code: res.set_code || item.set_code,
              set_name: res.set_name,
              collector_number: res.collector_number || item.collector_number,
              image_url: res.image_url,
              price_reference: (item.foil_treatment === 'non_foil' ? res.price_tcgplayer : res.price_cardmarket) || item.price_reference,
              price_source: item.foil_treatment === 'non_foil' ? 'tcgplayer' : 'cardmarket',
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
            language: item.language || 'en',
            color_identity: item.color_identity,
            cmc: item.cmc,
            is_legendary: !!item.is_legendary,
            is_historic: !!item.is_historic,
            is_land: !!item.is_land,
            is_basic_land: !!item.is_basic_land,
            art_variation: item.art_variation,
            type_line: item.type_line,
          })) as any,
          category_ids: [...defaultCategories],
          storage_items: defaultStorage ? [{ stored_in_id: defaultStorage, name: '', quantity: 1 }] : []
        };
        dataToImport = [deckProduct];
      }

      const res = await adminBulkCreateProducts(token, dataToImport);
      setImportResults(res);
      setStep('summary');
    } catch (e: any) {
      setError(e.message);
      setStep('preview');
    } finally {
      setLoading(false);
    }
  };

  const updatePreviewItem = (index: number, updates: Partial<BulkProductInput>) => {
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
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 overflow-y-auto">
      <div className="bg-[#1a1614] border-4 border-[#3c2a21] rounded-none shadow-[0_0_50px_rgba(0,0,0,0.5)] w-full max-w-[95vw] flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="p-6 border-b-4 border-[#3c2a21] bg-[#2a1e17] flex justify-between items-center">
          <div>
            <h2 className="text-3xl font-black text-[#d4c3b3] tracking-tighter uppercase italic drop-shadow-md">
              CSV IMPORT TOOL
            </h2>
            <p className="text-[#8b7355] text-sm font-medium">Bulk add products to the cardboard warehouse</p>
          </div>
          <button 
            onClick={onClose}
            className="w-10 h-10 border-2 border-[#3c2a21] bg-[#1a1614] text-[#8b7355] flex items-center justify-center hover:bg-[#b04b4b] hover:text-white transition-all transform hover:rotate-90 active:scale-90"
          >
            ✕
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
          {error && (
            <div className="mb-6 p-4 bg-[#b04b4b]/20 border-2 border-[#b04b4b] text-[#ffbaba] font-bold animate-pulse">
              ⚠️ {error}
            </div>
          )}

          {step === 'upload' && (
            <div className="flex flex-col items-center justify-center py-20 border-4 border-dashed border-[#3c2a21] bg-[#1a1614]/50 hover:bg-[#2a1e17]/50 transition-colors cursor-pointer relative group">
              <input 
                type="file" 
                accept=".csv" 
                onChange={handleFileUpload}
                className="absolute inset-0 opacity-0 cursor-pointer"
              />
              <div className="text-6xl mb-4 group-hover:scale-110 transition-transform">📄</div>
              <h3 className="text-2xl font-bold text-[#d4c3b3] mb-2 uppercase">Drop CSV File Here</h3>
              <p className="text-[#8b7355]">or click to browse your computer</p>
            </div>
          )}

          {step === 'mapping' && (
            <div>
              <div className="mb-8 p-6 bg-[#2a1e17] border-2 border-[#3c2a21] grid grid-cols-2 gap-8">
                <div>
                  <h4 className="text-[#d4c3b3] font-bold mb-4 uppercase text-sm tracking-widest">General Settings</h4>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-xs font-bold text-[#8b7355] uppercase mb-1">Destination</label>
                      <div className="flex gap-4 mb-4">
                        <button
                          onClick={() => setImportDestination('singles')}
                          className={`flex-1 py-3 px-4 border-2 font-black uppercase text-xs tracking-widest transition-all ${
                            importDestination === 'singles'
                              ? 'bg-[#d4c3b3] border-[#d4c3b3] text-[#1a1614]'
                              : 'bg-[#1a1614] border-[#3c2a21] text-[#8b7355] hover:border-[#d4c3b3]'
                          }`}
                        >
                          Singles
                        </button>
                        <button
                          onClick={() => setImportDestination('deck')}
                          className={`flex-1 py-3 px-4 border-2 font-black uppercase text-xs tracking-widest transition-all ${
                            importDestination === 'deck'
                              ? 'bg-[#d4c3b3] border-[#d4c3b3] text-[#1a1614]'
                              : 'bg-[#1a1614] border-[#3c2a21] text-[#8b7355] hover:border-[#d4c3b3]'
                          }`}
                        >
                          Deck
                        </button>
                      </div>
                    </div>

                    {importDestination === 'deck' && (
                      <div>
                        <label className="block text-xs font-bold text-[#8b7355] uppercase mb-1">Deck Name</label>
                        <input
                          type="text"
                          value={deckName}
                          onChange={(e) => setDeckName(e.target.value)}
                          placeholder="e.g. Mono Red Aggro"
                          className="w-full bg-[#1a1614] border-2 border-[#3c2a21] p-3 text-[#d4c3b3] focus:border-[#d4c3b3] outline-none transition-all placeholder:text-[#3c2a21]"
                        />
                      </div>
                    )}

                    <div>
                      <label className="block text-xs font-bold text-[#8b7355] uppercase mb-1">TCG Type</label>
                      <select 
                        value={tcgType}
                        onChange={(e) => setTcgType(e.target.value)}
                        className="w-full bg-[#1a1614] border-2 border-[#3c2a21] p-3 text-[#d4c3b3] focus:border-[#d4c3b3] outline-none transition-all"
                      >
                        {KNOWN_TCGS.map(id => (
                          <option key={id} value={id}>{TCG_LABELS[id] || id}</option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label className="block text-xs font-bold text-[#8b7355] uppercase mb-1">Default Storage</label>
                      <select 
                        value={defaultStorage}
                        onChange={(e) => setDefaultStorage(e.target.value)}
                        className="w-full bg-[#1a1614] border-2 border-[#3c2a21] p-3 text-[#d4c3b3] focus:border-[#d4c3b3] outline-none transition-all"
                      >
                        <option value="" disabled>Select Storage Location</option>
                        {storageLocations.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                      </select>
                    </div>
                  </div>
                </div>
                <div className="border-l-2 border-[#3c2a21] pl-8">
                  <h4 className="text-[#d4c3b3] font-bold mb-4 uppercase text-sm tracking-widest">Column Mapping</h4>
                  <div className="space-y-3">
                    {FIELDS.map(f => (
                      <div key={f.key} className="flex items-center gap-4">
                        <label className="w-32 text-xs font-bold text-[#8b7355] uppercase whitespace-nowrap">
                          {f.label} {f.required && <span className="text-[#b04b4b]">*</span>}
                        </label>
                        <select 
                          value={mapping[f.key] || ''}
                          onChange={(e) => setMapping(prev => ({ ...prev, [f.key]: e.target.value }))}
                          className="flex-1 bg-[#1a1614] border-2 border-[#3c2a21] p-2 text-[#d4c3b3] text-sm focus:border-[#d4c3b3] outline-none"
                        >
                          <option value="">-- [Skip Field] --</option>
                          {headers.map(h => <option key={h} value={h}>{h}</option>)}
                        </select>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              
              <div className="flex justify-end gap-4 mt-8">
                <button 
                  onClick={() => setStep('upload')}
                  className="px-8 py-4 border-2 border-[#3c2a21] text-[#8b7355] font-black uppercase hover:bg-[#3c2a21] hover:text-[#d4c3b3] transition-all active:scale-95"
                >
                  Back
                </button>
                <button 
                  onClick={handleStartPreview}
                  className="px-8 py-4 bg-[#3c2a21] text-[#d4c3b3] border-4 border-[#d4c3b3]/20 font-black uppercase tracking-widest hover:bg-[#d4c3b3] hover:text-[#3c2a21] transition-all hover:shadow-[0_0_20px_rgba(212,195,179,0.3)] active:scale-95"
                >
                  Generate Preview
                </button>
              </div>
            </div>
          )}

          {step === 'preview' && (
            <div className="flex flex-col h-full overflow-hidden">
              <div className="mb-6 flex flex-col gap-4 bg-[#2a1e17] p-4 border-2 border-[#3c2a21] relative overflow-hidden">
                <div className="flex justify-between items-center relative z-10">
                  <div className="flex gap-4">
                    <button 
                      onClick={handleEnrich}
                      disabled={loading}
                      className="px-6 py-2 bg-[#8b7355] text-white font-black uppercase text-xs tracking-tighter hover:bg-[#d4c3b3] hover:text-[#3c2a21] transition-all disabled:opacity-50 flex items-center gap-2 relative overflow-hidden group/btn"
                    >
                      <span className="relative z-10">{loading ? '⌛ ENRICHING...' : '✨ ENRICH WITH SCRYFALL'}</span>
                      {loading && (
                        <div 
                          className="absolute inset-0 bg-white/20 transition-all duration-300" 
                          style={{ width: `${(processedCount / previewData.length) * 100}%` }}
                        />
                      )}
                    </button>
                  </div>
                  <div className="text-[#8b7355] font-bold text-sm italic">
                    {previewData.length} cards detected in shipment
                  </div>
                </div>

                {loading && (
                  <div className="w-full bg-[#1a1614] h-1.5 border border-[#3c2a21] relative overflow-hidden">
                    <div 
                      className="absolute inset-y-0 left-0 bg-[#d4c3b3] transition-all duration-300 shadow-[0_0_10px_rgba(212,195,179,0.5)]"
                      style={{ width: `${(processedCount / previewData.length) * 100}%` }}
                    />
                    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                      <span className="text-[8px] font-black text-[#8b7355] uppercase tracking-widest">
                        Scanning Manifest: {Math.round((processedCount / previewData.length) * 100)}%
                      </span>
                    </div>
                  </div>
                )}
              </div>

              <div className="flex-1 overflow-auto border-2 border-[#3c2a21] bg-[#1a1614] custom-scrollbar">
                <table className="w-full text-left border-collapse min-w-[1200px]">
                  <thead className="sticky top-0 z-10 bg-[#3c2a21] text-[#d4c3b3] text-[10px] font-black uppercase tracking-widest">
                    <tr>
                      <th className="p-3 border-r border-[#1a1614] w-20">Art</th>
                      <th className="p-3 border-r border-[#1a1614]">Card Details</th>
                      <th className="p-3 border-r border-[#1a1614] w-32">Treatment</th>
                      <th className="p-3 border-r border-[#1a1614] w-32">Storage</th>
                      <th className="p-3 border-r border-[#1a1614] w-48">Collections</th>
                      <th className="p-3 border-r border-[#1a1614] w-24">Price</th>
                      <th className="p-3 w-20">Stock</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#3c2a21]">
                    {previewData.map((item, idx) => (
                      <tr key={idx} className="hover:bg-[#2a1e17] transition-colors group">
                        <td className="p-2 border-r border-[#3c2a21]">
                          <div className="w-16 aspect-[2.5/3.5] bg-[#3c2a21] relative overflow-hidden ring-2 ring-[#3c2a21] group-hover:ring-[#d4c3b3] transition-all">
                            {item.image_url ? (
                              <CardImage 
                                imageUrl={item.image_url} 
                                name={item.name!} 
                                height="100%"
                                enableHover={true}
                                enableModal={true}
                              />
                            ) : (
                              <div className="w-full h-full flex items-center justify-center text-[10px] text-[#8b7355] font-black uppercase text-center p-1">No Image</div>
                            )}
                          </div>
                        </td>
                        <td className="p-3 border-r border-[#3c2a21]">
                          <input 
                            value={item.name || ''} 
                            onChange={(e) => updatePreviewItem(idx, { name: e.target.value })}
                            placeholder="Card Name"
                            className="bg-transparent text-[#d4c3b3] font-bold w-full focus:outline-none focus:bg-[#3c2a21] px-1"
                          />
                          <div className="flex gap-2 mt-1">
                            <input 
                              value={item.set_code || ''} 
                              onChange={(e) => updatePreviewItem(idx, { set_code: e.target.value })}
                              placeholder="SET"
                              className="bg-[#1a1614] text-[#8b7355] text-[10px] font-black uppercase w-12 text-center border border-[#3c2a21] focus:outline-none focus:border-[#d4c3b3]"
                            />
                            <input 
                              value={item.collector_number || ''} 
                              onChange={(e) => updatePreviewItem(idx, { collector_number: e.target.value })}
                              placeholder="#"
                              className="bg-[#1a1614] text-[#8b7355] text-[10px] font-black uppercase w-16 text-center border border-[#3c2a21] focus:outline-none focus:border-[#d4c3b3]"
                            />
                            <select
                              value={item.condition || 'NM'}
                              onChange={(e) => updatePreviewItem(idx, { condition: e.target.value as Condition })}
                              className="bg-[#1a1614] text-[#8b7355] text-[10px] font-black uppercase border border-[#3c2a21] outline-none"
                            >
                              {['NM', 'LP', 'MP', 'HP', 'DMG'].map(c => <option key={c} value={c}>{c}</option>)}
                            </select>
                          </div>
                        </td>
                        <td className="p-3 border-r border-[#3c2a21]">
                          <select
                            value={item.foil_treatment}
                            onChange={(e) => updatePreviewItem(idx, { foil_treatment: e.target.value as FoilTreatment })}
                            className="w-full bg-[#1a1614] text-[#d4c3b3] text-[10px] font-bold uppercase border border-[#3c2a21] p-1 outline-none mb-1"
                          >
                            {Object.entries(FOIL_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                          </select>
                          <select
                            value={item.card_treatment}
                            onChange={(e) => updatePreviewItem(idx, { card_treatment: e.target.value as CardTreatment })}
                            className="w-full bg-[#1a1614] text-[#d4c3b3] text-[10px] font-bold uppercase border border-[#3c2a21] p-1 outline-none"
                          >
                            {Object.entries(TREATMENT_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                          </select>
                        </td>
                        <td className="p-3 border-r border-[#3c2a21]">
                          <select
                            value={item.storage_items?.[0]?.stored_in_id || ''}
                            onChange={(e) => updateStorage(idx, e.target.value, item.stock || 0)}
                            className="w-full bg-[#1a1614] text-[#d4c3b3] text-[10px] font-bold uppercase border border-[#3c2a21] p-1 outline-none"
                          >
                            {storageLocations.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                          </select>
                        </td>
                        <td className="p-3 border-r border-[#3c2a21]">
                          <div className="flex flex-wrap gap-1">
                            {categories.map(cat => (
                              <button
                                key={cat.id}
                                onClick={() => toggleCategory(idx, cat.id)}
                                className={`px-1.5 py-0.5 text-[9px] font-black uppercase border-2 transition-all ${
                                  (item.category_ids || []).includes(cat.id)
                                    ? 'bg-[#d4c3b3] border-[#d4c3b3] text-[#1a1614]'
                                    : 'border-[#3c2a21] text-[#8b7355] hover:border-[#d4c3b3]'
                                }`}
                              >
                                {cat.name}
                              </button>
                            ))}
                          </div>
                        </td>
                        <td className="p-3 border-r border-[#3c2a21]">
                          <div className="text-[10px] text-[#8b7355] font-black uppercase mb-1">
                            {item.price_source === 'tcgplayer' ? '$ USD' : '€ EUR'}
                          </div>
                          <input 
                            type="number"
                            value={item.price_reference || ''} 
                            onChange={(e) => updatePreviewItem(idx, { price_reference: parseFloat(e.target.value) || 0 })}
                            className="bg-[#1a1614] text-[#d4c3b3] font-mono font-bold w-full border border-[#3c2a21] p-1 text-sm focus:border-[#d4c3b3] outline-none"
                            step="0.01"
                          />
                        </td>
                        <td className="p-3">
                          <input 
                            type="number"
                            min="1"
                            value={item.stock || 0} 
                            onChange={(e) => {
                              const v = parseInt(e.target.value) || 0;
                              updateStorage(idx, item.storage_items?.[0]?.stored_in_id || '', v);
                            }}
                            className="bg-[#1a1614] text-[#d4c3b3] font-mono font-black w-full border border-[#3c2a21] p-1 text-sm text-center focus:border-[#d4c3b3] outline-none"
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="mt-6 flex justify-between items-center bg-[#2a1e17] p-6 border-4 border-[#3c2a21]">
                <button 
                  onClick={() => setStep('mapping')}
                  className="px-8 py-3 border-2 border-[#3c2a21] text-[#8b7355] font-black uppercase hover:bg-[#3c2a21] hover:text-[#d4c3b3] transition-all"
                >
                  Back to Mapping
                </button>
                <div className="flex gap-4">
                   <button 
                    onClick={handleImport}
                    disabled={loading || previewData.length === 0}
                    className="px-12 py-4 bg-[#8b7355] text-white border-4 border-white/20 font-black uppercase tracking-[0.2em] hover:bg-[#d4c3b3] hover:text-[#3c2a21] transition-all disabled:opacity-50 hover:shadow-[0_0_30px_rgba(212,195,179,0.4)] active:scale-95"
                  >
                    {loading ? 'IMPORTING...' : 'FINALIZE IMPORT'}
                  </button>
                </div>
              </div>
            </div>
          )}

          {step === 'importing' && (
            <div className="flex flex-col items-center justify-center py-32 text-center">
              <div className="relative mb-12">
                {/* Custom Box Animation */}
                <div className="w-32 h-32 border-4 border-[#3c2a21] bg-[#2a1e17] shadow-2xl relative animate-bounce">
                   <div className="absolute top-0 left-1/2 -translate-x-1/2 w-1 h-32 bg-[#3c2a21]/20"></div>
                   <div className="absolute top-1/2 left-0 -translate-y-1/2 w-32 h-1 bg-[#3c2a21]/20"></div>
                   <div className="absolute inset-2 border-2 border-[#3c2a21]/30 border-dashed"></div>
                   <div className="absolute -top-4 -right-4 text-4xl">📦</div>
                </div>
                {/* Warehouse Floor Shadow */}
                <div className="w-24 h-4 bg-black/40 blur-md rounded-full mt-4 mx-auto animate-pulse"></div>
              </div>
              
              <h3 className="text-4xl font-black text-[#d4c3b3] uppercase tracking-tighter mb-4 italic drop-shadow-lg">
                STORING SHIPMENT...
              </h3>
              <div className="max-w-md space-y-2">
                <p className="text-[#8b7355] font-bold text-lg italic animate-pulse">
                  &quot;Unpacking boxes and organizing the vault&quot;
                </p>
                <div className="flex justify-center gap-1">
                  {[1,2,3].map(i => (
                    <div key={i} className="w-2 h-2 bg-[#d4c3b3] animate-bounce" style={{ animationDelay: `${i*0.2}s` }}></div>
                  ))}
                </div>
              </div>
              
              <div className="mt-12 p-4 border-2 border-[#3c2a21] bg-[#2a1e17] text-[10px] font-black text-[#8b7355] uppercase tracking-[0.3em]">
                System processing batch of {previewData.length} records
              </div>
            </div>
          )}

          {step === 'summary' && importResults && (
            <div className="text-center py-20">
              <div className="text-8xl mb-6">📦</div>
              <h3 className="text-4xl font-black text-[#d4c3b3] uppercase mb-4 tracking-tighter italic">Import Complete!</h3>
              <p className="text-xl text-[#8b7355] mb-12 font-medium">
                Successfully processed <span className="text-[#d4c3b3] font-black">{importResults.count}</span> products into the system.
              </p>
              <button 
                onClick={() => { onImported(); onClose(); }}
                className="px-16 py-6 bg-[#d4c3b3] text-[#3c2a21] border-8 border-[#3c2a21]/20 font-black uppercase tracking-[0.3em] hover:bg-white transition-all shadow-2xl active:scale-95"
              >
                Return to Dashboard
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
