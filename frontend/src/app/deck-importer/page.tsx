'use client';

import { useState } from 'react';
import { bulkSearchDeck, createClientRequestsBatch } from '@/lib/api';
import { useToast } from '@/context/ToastContext';
import { DeckMatch, ClientRequestBatchInput } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import { useLanguage } from '@/context/LanguageContext';
import { useUser } from '@/context/UserContext';
import CardImage from '@/components/CardImage';
import { ConditionBadge, FoilBadge } from '@/components/Badges';
import LoadingSpinner from '@/components/LoadingSpinner';
import SetIcon from '@/components/SetIcon';
import Modal from '@/components/ui/Modal';
import Button from '@/components/ui/Button';
import { useForm } from '@/hooks/useForm';
import ScryfallVariantPicker from '@/components/ScryfallVariantPicker';

interface WantedItem {
  card_name: string;
  set_name?: string;
  details?: string;
  quantity: number;
  tcg: string;
  original_line: string;
  scryfall_id?: string;
  image_url?: string;
  set_code?: string;
  collector_number?: string;
  foil_treatment?: string;
  card_treatment?: string;
  match_type?: 'any' | 'exact';
}

export default function DeckImporterPage() {
  const { toast } = useToast();
  const { t } = useLanguage();
  const { addItem } = useCart();
  const { user } = useUser();
  const [list, setList] = useState('');
  const [results, setResults] = useState<DeckMatch[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const foundCount = results ? results.filter(r => r.is_matched).length : 0;
  const totalCount = results ? results.length : 0;

  // Wanted list state
  const [wanted, setWanted] = useState<Record<string, WantedItem>>({});
  const [showRequestModal, setShowRequestModal] = useState(false);
  const [customizingItem, setCustomizingItem] = useState<{key: string, item: WantedItem} | null>(null);

  const handleAnalyze = async () => {
    if (!list.trim()) return;
    setLoading(true);
    setError('');
    setWanted({}); // Clear wanted on new analysis
    try {
      const resp = await bulkSearchDeck(list);
      setResults(resp.matches);
    } catch (err) {
      setError(t('pages.deck_importer.errors.analyze_failed', 'Failed to analyze list. Please check the format.'));
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddAll = () => {
    if (!results) return;
    let count = 0;
    results.forEach(match => {
      if (match.is_matched && match.matches && match.matches.length > 0) {
        const p = match.matches[0];
        const qtyToAdd = Math.min(match.quantity, p.stock);
        for (let i = 0; i < qtyToAdd; i++) {
          addItem(p);
        }
        count++;
      }
    });
    toast.success(t('pages.deck_importer.messages.added_count', 'Added products from {count} cards to your cart.', { count }));
  };

  const toggleWanted = (match: DeckMatch, quantity: number, isRemainder: boolean) => {
    const key = `${match.raw_line}_${isRemainder ? 'rem' : 'full'}`;
    if (wanted[key]) {
      const newWanted = { ...wanted };
      delete newWanted[key];
      setWanted(newWanted);
    } else {
      // Basic heuristic to get card name from raw line if possible, or just use it as is
      // Most common format: "4 Card Name (SET) 123"
      let cardName = match.raw_line.trim();
      const matchQty = cardName.match(/^(\d+)\s+(.*)/);
      if (matchQty) cardName = matchQty[2];

      const bestMatch = match.matches?.[0];

      setWanted(prev => ({
        ...prev,
        [key]: {
          card_name: bestMatch?.name || cardName,
          set_name: bestMatch?.set_name || match.requested_set,
          details: match.requested_cn ? `CN: ${match.requested_cn}` : (bestMatch?.collector_number ? `CN: ${bestMatch.collector_number}` : undefined),
          quantity: quantity,
          tcg: bestMatch?.tcg || 'mtg',
          original_line: match.raw_line,
          image_url: bestMatch?.image_url,
          scryfall_id: bestMatch?.scryfall_id,
          set_code: bestMatch?.set_code,
          collector_number: bestMatch?.collector_number,
          foil_treatment: bestMatch?.foil_treatment,
          card_treatment: bestMatch?.card_treatment,
          match_type: 'any' // Default to any unless they refine it
        }
      }));
    }
  };

  const wantedCount = Object.values(wanted).reduce((acc, curr) => acc + curr.quantity, 0);

  // Form for submission
  const {
    form,
    handleChange,
    submitting,
    handleSubmit
  } = useForm({
    customer_name: user ? `${user.first_name} ${user.last_name || ''}`.trim() : '',
    customer_contact: user ? (user.email || user.phone || '') : '',
  });

  const onSubmitBatch = async (data: Record<string, string>) => {
    const batch: ClientRequestBatchInput = {
      customer_name: data.customer_name,
      customer_contact: data.customer_contact,
      cards: Object.values(wanted).map(w => ({
        card_name: w.card_name,
        set_name: w.set_name,
        details: w.details,
        quantity: w.quantity,
        tcg: w.tcg,
        scryfall_id: w.scryfall_id,
        image_url: w.image_url,
        set_code: w.set_code,
        collector_number: w.collector_number,
        foil_treatment: w.foil_treatment,
        card_treatment: w.card_treatment,
        match_type: w.match_type || 'any'
      }))
    };

    try {
      await createClientRequestsBatch(batch);
      toast.success(t('pages.deck_importer.messages.request_success', 'We have received your acquisition mission! We will contact you once we find the cards.'));
      setWanted({});
      setShowRequestModal(false);
    } catch {
      toast.error(t('pages.deck_importer.errors.request_failed', 'Failed to submit requests. Please try again.'));
    }
  };

  return (
    <div className="centered-container px-4 py-12">
      <div className="mb-8 text-center">
        <h1 className="font-display text-5xl mb-4">{t('pages.deck_importer.title', 'DECK IMPORTER')}</h1>
        <p className="text-text-secondary max-w-2xl mx-auto">
          {t('pages.deck_importer.subtitle', 'Paste your card list below (works with Moxfield, Archideckt, MTGGoldfish, and more) and we\'ll match it against our current stock. If we don\'t have enough, you can request them in one go!')}
        </p>
        <div className="gold-line mt-6" />
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Input Side */}
        <div className="flex-1">
          <div className="card p-6">
            <label className="text-xs font-mono-stack mb-2 block text-text-muted uppercase tracking-widest">
              {t('pages.deck_importer.list_label', 'CARD LIST')}
            </label>
            <textarea
              value={list}
              onChange={(e) => setList(e.target.value)}
              placeholder={"1 Sheoldred, the Apocalypse (DMU) 107\n4 Birds of Paradise\n2 Lightning Bolt (M10)"}
              className="w-full h-80 font-mono text-sm p-4 bg-ink-surface border-ink-border focus:border-accent-primary transition-all rounded-md resize-none"
            />
            {error && <p className="text-red-500 text-xs mt-2">{error}</p>}
            <button
              onClick={handleAnalyze}
              disabled={loading || !list.trim()}
              className="btn-primary w-full mt-4 py-3 text-lg flex items-center justify-center gap-2"
            >
              {loading ? <LoadingSpinner /> : '🔍'}
              {loading ? t('pages.deck_importer.buttons.analyzing', 'ANALYZING...') : t('pages.deck_importer.buttons.analyze', 'ANALYZE LIST →')}
            </button>
          </div>
        </div>

        {/* Results Side */}
        <div className="flex-1 lg:max-w-[600px] flex flex-col gap-4">
          <div className="card p-6 min-h-[400px] flex flex-col flex-1">
            <div className="flex justify-between items-center mb-6">
              <h2 className="font-display text-2xl flex items-center gap-3">
                {t('pages.deck_importer.results_title', 'MATCHES')}
                {totalCount > 0 && (
                  <span className="text-xs font-sans font-normal text-text-muted bg-ink-border/30 px-2 py-0.5 rounded-full">
                    {t('pages.deck_importer.found_count', '{found} OF {total} FOUND', { found: foundCount, total: totalCount })}
                  </span>
                )}
              </h2>
              <div className="flex gap-2">
                {results && (
                  <button
                    onClick={handleAddAll}
                    className="btn-secondary text-xs py-1 px-3"
                  >
                    {t('pages.deck_importer.buttons.add_all', 'ADD ALL FOUND')}
                  </button>
                )}
              </div>
            </div>

            {!results && !loading && (
              <div className="flex-1 flex flex-col items-center justify-center opacity-30 italic text-center py-12">
                <p>{t('pages.deck_importer.no_results', 'Paste a list to see available cards.')}</p>
              </div>
            )}

            {results && (
              <div className="space-y-4 max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
                {results.map((match, idx) => {
                  const bestMatch = match.matches?.[0];
                  const inStock = bestMatch?.stock || 0;
                  const hasPartial = match.is_matched && inStock > 0 && inStock < match.quantity;
                   const isDifferentVersion = !!(match.requested_set && match.is_matched && bestMatch && (
                    (bestMatch.set_code?.toLowerCase().trim() !== match.requested_set.toLowerCase().trim() &&
                      bestMatch.set_name?.toLowerCase().trim() !== match.requested_set.toLowerCase().trim()) ||
                    (match.requested_cn && String(bestMatch.collector_number).toLowerCase().trim() !== String(match.requested_cn).toLowerCase().trim())
                  ));

                  return (
                    <div key={idx} className={`p-4 rounded-lg border transition-all ${match.is_matched ? 'bg-ink-surface/50 border-ink-border' : 'bg-hp-color/5 border-hp-color/20 opacity-90'}`}>
                      <div className="flex justify-between items-start mb-2">
                        <span className="text-xs font-mono-stack text-text-muted">{match.raw_line}</span>
                        <div className="flex gap-2">
                          {isDifferentVersion && (
                            <span className="text-[10px] bg-gold/20 text-gold-dark px-1.5 py-0.5 rounded border border-gold/30">
                              {t('pages.deck_importer.matches.different_version_details', 'DIFF. VERSION ({set} #{cn})', { 
                                set: bestMatch?.set_code?.toUpperCase() || '???', 
                                cn: bestMatch?.collector_number || '?'
                              })}
                            </span>
                          )}
                          {match.is_matched ? (
                            <span className="text-[10px] bg-green-500/20 text-green-400 px-1.5 py-0.5 rounded border border-green-500/30">
                              {hasPartial ? t('pages.deck_importer.matches.partial', 'PARTIAL STOCK ({n})', { n: inStock }) : t('pages.deck_importer.matches.matched', 'MATCHED')}
                            </span>
                          ) : (
                            <span className="text-[10px] bg-red-500/20 text-red-400 px-1.5 py-0.5 rounded border border-red-500/30">{t('pages.deck_importer.matches.no_stock', 'NO STOCK')}</span>
                          )}
                        </div>
                      </div>

                      {bestMatch && (
                        <div className="flex gap-3 mt-3">
                          <div className="w-12 h-16 flex-shrink-0">
                            <CardImage
                              imageUrl={bestMatch.image_url}
                              name={bestMatch.name}
                              tcg={bestMatch.tcg}
                              foilTreatment={bestMatch.foil_treatment}
                              enableHover={true}
                              enableModal={true}
                            />
                          </div>
                          <div className="flex-1">
                            <p className="text-sm font-bold truncate">{bestMatch.name}</p>
                            <div className="flex items-center gap-1.5 text-[10px] text-text-muted mb-1">
                              <SetIcon setCode={bestMatch.set_code} rarity={bestMatch.rarity} size="xs" />
                              <span className="truncate">{bestMatch.set_name}</span>
                            </div>
                            <div className="flex flex-wrap gap-1 mb-2">
                              <ConditionBadge condition={bestMatch.condition} />
                              <FoilBadge foil={bestMatch.foil_treatment} />
                            </div>
                            <div className="flex items-center justify-between">
                              <span className="price text-sm">${bestMatch.price.toLocaleString()} COP</span>
                              <div className="flex gap-2">
                                <button
                                  onClick={() => addItem(bestMatch)}
                                  className="btn-primary text-[10px] py-1 px-2"
                                >
                                  + {t('pages.common.buttons.add', 'ADD')}
                                </button>
                                {(hasPartial || isDifferentVersion) && (
                                  <button
                                    onClick={() => toggleWanted(match, isDifferentVersion ? match.quantity : (match.quantity - inStock), isDifferentVersion ? false : true)}
                                    className={`text-[10px] py-1 px-2 rounded border transition-colors ${wanted[`${match.raw_line}_${isDifferentVersion ? 'full' : 'rem'}`] ? 'bg-accent-primary text-text-on-accent border-accent-primary' : 'bg-gold/10 text-gold-dark border-gold/30 hover:bg-gold/20'}`}
                                  >
                                    {isDifferentVersion 
                                      ? (wanted[`${match.raw_line}_full`] ? '✓ WANTED' : t('pages.deck_importer.matches.request_original', 'REQUEST ORIGINAL'))
                                      : (wanted[`${match.raw_line}_rem`] ? '✓ WANTED' : t('pages.deck_importer.matches.add_wanted_rem', 'ADD REMAINDER'))
                                    }
                                  </button>
                                )}
                              </div>
                            </div>
                          </div>
                        </div>
                      )}

                      {!match.is_matched && (
                        <div className="mt-4 flex items-center justify-between">
                          <p className="text-[11px] text-text-muted italic">{t('pages.deck_importer.matches.no_stock_desc', 'Not in stock currently.')}</p>
                          <button
                            onClick={() => toggleWanted(match, match.quantity, false)}
                            className={`text-[10px] py-1 px-2 rounded border transition-colors ${wanted[`${match.raw_line}_full`] ? 'bg-accent-primary text-text-on-accent border-accent-primary' : 'bg-gold/10 text-gold-dark border-gold/30 hover:bg-gold/20'}`}
                          >
                            {wanted[`${match.raw_line}_full`] ? '✓' : `+ ${match.quantity} WANTED`}
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Wanted List Summary Section */}
          {wantedCount > 0 && (
            <div className="card p-6 border-gold/30 bg-gold/5 animate-in slide-in-from-bottom-4 duration-300">
              <div className="flex justify-between items-center mb-4">
                <div className="flex items-center gap-2">
                  <span className="text-xl">🕵️‍♂️</span>
                  <h3 className="font-display text-xl uppercase tracking-wider">{t('pages.deck_importer.wanted.title', 'MISSION LIST')}</h3>
                  <span className="bg-gold text-ink-page text-[10px] font-bold px-1.5 py-0.5 rounded-full">{wantedCount}</span>
                </div>
                <button onClick={() => setWanted({})} className="text-[10px] text-text-muted hover:text-red-400 uppercase tracking-widest">{t('pages.common.buttons.clear', 'CLEAR')}</button>
              </div>

              <div className="space-y-2 mb-6 max-h-[150px] overflow-y-auto custom-scrollbar">
                {Object.entries(wanted).map(([key, w]) => (
                  <div key={key} className="group relative flex justify-between items-center text-xs font-mono-stack p-2 bg-ink-page/50 rounded border border-gold/10 hover:border-gold/40 transition-all">
                    <div className="flex items-center gap-2 min-w-0">
                      {w.image_url && <img src={w.image_url} className="w-6 h-8 object-cover rounded border border-gold/20" />}
                      <div className="truncate">
                        <p className="text-gold-dark font-bold leading-none mb-1">{w.quantity}x {w.card_name}</p>
                        <p className="text-[9px] text-text-muted opacity-60 uppercase truncate">
                          {w.match_type === 'exact' ? `${w.set_name} #${w.collector_number}` : 'Any Version'}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <button 
                        onClick={() => setCustomizingItem({ key, item: w })}
                        className="text-[10px] text-gold-dark hover:text-gold opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        EDIT
                      </button>
                      <button onClick={() => {
                        const newWanted = { ...wanted };
                        delete newWanted[key];
                        setWanted(newWanted);
                      }} className="text-text-muted hover:text-red-400">✕</button>
                    </div>
                  </div>
                ))}
              </div>

              <button
                onClick={() => setShowRequestModal(true)}
                className="w-full py-4 bg-accent-primary hover:bg-accent-primary-hover text-text-on-accent font-bold rounded-xl transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2"
              >
                {t('pages.deck_importer.wanted.submit_btn', 'START ACQUISITION MISSION →')}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Batch Request Modal */}
      <Modal
        isOpen={showRequestModal}
        onClose={() => setShowRequestModal(false)}
        title={t('pages.deck_importer.modal.title', 'ACQUISITION MISSION')}
      >
        <div className="mb-6">
          <p className="text-xs text-text-muted uppercase tracking-widest leading-relaxed mb-4">
            {t('pages.deck_importer.modal.desc', "Tell us where to contact you once we've secured your items. This mission includes {n} cards.", { n: wantedCount })}
          </p>
        </div>

        <form onSubmit={(e) => { e.preventDefault(); handleSubmit(onSubmitBatch); }} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.name_label', 'Your Name *')}</label>
              <input
                name="customer_name"
                type="text"
                className="w-full text-sm"
                required
                value={form.customer_name}
                onChange={handleChange}
                placeholder={t('components.client_request_modal.form.name_placeholder', 'John Doe')}
              />
            </div>
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.contact_label', 'Contact Info *')}</label>
              <input
                name="customer_contact"
                type="text"
                className="w-full text-sm"
                required
                value={form.customer_contact}
                onChange={handleChange}
                placeholder={t('components.client_request_modal.form.contact_placeholder', 'Phone or Instagram')}
              />
            </div>
          </div>

          <div className="p-4 bg-ink-surface/50 rounded-lg border border-ink-border max-h-[200px] overflow-y-auto custom-scrollbar">
            <h4 className="text-[10px] font-mono-stack uppercase text-text-muted mb-2 tracking-widest">{t('pages.deck_importer.modal.items_list', 'ITEMS TO ACQUIRE')}</h4>
            {Object.values(wanted).map((w, i) => (
              <div key={i} className="text-xs py-1 border-b border-ink-border last:border-0 flex justify-between">
                <span>{w.quantity}x {w.card_name}</span>
                <span className="text-text-muted text-[10px]">{w.set_name || ''}</span>
              </div>
            ))}
          </div>

          <Button
            type="submit"
            loading={submitting}
            fullWidth
            size="lg"
            className="mt-6"
          >
            {t('pages.deck_importer.modal.confirm_btn', 'CONFIRM MISSION')}
          </Button>
        </form>
      </Modal>

      {/* Item Customization Modal */}
      {customizingItem && customizingItem.item && (
        <Modal
          isOpen={true}
          onClose={() => setCustomizingItem(null)}
          title={t('pages.deck_importer.customize.title', 'REFINE ACQUISITION')}
        >
          <div className="space-y-6">
            <div className="flex gap-4 items-start p-4 bg-ink-surface/50 rounded-lg border border-gold/10">
               <div className="w-16 h-20 bg-ink-page rounded flex-shrink-0 border border-gold/20 overflow-hidden shadow-sm">
                  {customizingItem.item.image_url ? (
                    <img src={customizingItem.item.image_url} alt={customizingItem.item.card_name} className="w-full h-full object-cover" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-2xl opacity-20">🎴</div>
                  )}
               </div>
               <div>
                  <h4 className="text-lg font-bold text-gold-dark">{customizingItem.item.card_name}</h4>
                  <p className="text-[10px] font-mono-stack text-text-muted uppercase tracking-widest">{t('pages.deck_importer.customize.current', 'CURRENT SPEC:')} {customizingItem.item.match_type === 'exact' ? `${customizingItem.item.set_name} (#${customizingItem.item.collector_number})` : 'Any Version'}</p>
               </div>
            </div>

            <ScryfallVariantPicker 
              cardName={customizingItem.item.card_name}
              selectedId={customizingItem.item.scryfall_id}
              onSelect={(print) => {
                if (!customizingItem) return;
                const updatedItem: WantedItem = {
                  ...customizingItem.item,
                  scryfall_id: print?.id,
                  match_type: print ? 'exact' : 'any',
                  image_url: print?.image_uris?.normal || print?.image_uris?.small || print?.card_faces?.[0]?.image_uris?.normal,
                  set_name: print?.set_name,
                  set_code: print?.set,
                  collector_number: print?.collector_number,
                  foil_treatment: print?.finishes?.includes('foil') ? 'foil' : 'non_foil',
                  card_treatment: print?.border_color === 'borderless' ? 'borderless' : (print?.frame_effects?.includes('showcase') ? 'showcase' : 'normal'),
                };
                setWanted(prev => ({ ...prev, [customizingItem.key]: updatedItem }));
                setCustomizingItem({ key: customizingItem.key, item: updatedItem });
              }}
            />

            <Button
              fullWidth
              onClick={() => setCustomizingItem(null)}
              className="mt-4"
            >
              {t('pages.common.buttons.save', 'SAVE CHANGES')}
            </Button>
          </div>
        </Modal>
      )}
    </div>
  );
}
