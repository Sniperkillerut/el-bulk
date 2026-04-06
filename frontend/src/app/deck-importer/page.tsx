'use client';

import { useState } from 'react';
import { bulkSearchDeck } from '@/lib/api';
import { DeckMatch } from '@/lib/types';
import { useCart } from '@/lib/CartContext';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from '@/components/CardImage';
import { ConditionBadge, FoilBadge } from '@/components/Badges';
import LoadingSpinner from '@/components/LoadingSpinner';
import SetIcon from '@/components/SetIcon';

export default function DeckImporterPage() {
  const { t } = useLanguage();
  const { addItem } = useCart();
  const [list, setList] = useState('');
  const [results, setResults] = useState<DeckMatch[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleAnalyze = async () => {
    if (!list.trim()) return;
    setLoading(true);
    setError('');
    try {
      const resp = await bulkSearchDeck(list);
      setResults(resp.matches);
    } catch (err) {
      setError('Failed to analyze list. Please check the format.');
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
        // Add the first (best) match
        const p = match.matches[0];
        const qty = Math.min(match.quantity, p.stock);
        for (let i = 0; i < qty; i++) {
          addItem(p);
        }
        count++;
      }
    });
    alert(`Added products from ${count} cards to your cart.`);
  };

  return (
    <div className="centered-container px-4 py-12">
      <div className="mb-8 text-center">
        <h1 className="font-display text-5xl mb-4">{t('pages.deck_importer.title', 'DECK IMPORTER')}</h1>
        <p className="text-text-secondary max-w-2xl mx-auto">
          {t('pages.deck_importer.subtitle', 'Paste your card list below (e.g. "4 Birds of Paradise") and we\'ll match it against our current stock.')}
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
              placeholder="4 Birds of Paradise&#10;2 Lightning Bolt (M10)&#10;1 Sheoldred, the Apocalypse"
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
        <div className="w-full lg:w-[500px]">
          <div className="card p-6 min-h-[400px] flex flex-col">
            <div className="flex justify-between items-center mb-6">
              <h2 className="font-display text-2xl">{t('pages.deck_importer.results_title', 'MATCHES')}</h2>
              {results && (
                <button
                  onClick={handleAddAll}
                  className="btn-secondary text-xs py-1 px-3"
                >
                  {t('pages.deck_importer.buttons.add_all', 'ADD ALL FOUND')}
                </button>
              )}
            </div>

            {!results && !loading && (
              <div className="flex-1 flex flex-col items-center justify-center opacity-30 italic text-center py-12">
                <p>{t('pages.deck_importer.no_results', 'Paste a list to see available cards.')}</p>
              </div>
            )}

            {results && (
              <div className="space-y-4 max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
                {results.map((match, idx) => (
                  <div key={idx} className={`p-4 rounded-lg border transition-all ${match.is_matched ? 'bg-ink-surface/50 border-ink-border' : 'bg-hp-color/5 border-hp-color/20 opacity-60'}`}>
                     <div className="flex justify-between items-start mb-2">
                        <span className="text-xs font-mono-stack text-text-muted">{match.raw_line}</span>
                        <div className="flex gap-2">
                           {match.requested_set && match.is_matched && match.matches && match.matches.length > 0 && (
                             (match.matches[0].set_code?.toLowerCase() !== match.requested_set.toLowerCase() && 
                              match.matches[0].set_name?.toLowerCase() !== match.requested_set.toLowerCase()) ||
                             (match.requested_cn && match.matches[0].collector_number !== match.requested_cn)
                           ) && (
                             <span className="text-[10px] bg-gold/20 text-gold-dark px-1.5 py-0.5 rounded border border-gold/30">DIFFERENT VERSION</span>
                           )}
                           {match.is_matched ? (
                             <span className="text-[10px] bg-green-500/20 text-green-400 px-1.5 py-0.5 rounded border border-green-500/30">MATCHED</span>
                           ) : (
                             <span className="text-[10px] bg-red-500/20 text-red-400 px-1.5 py-0.5 rounded border border-red-500/30">NO STOCK</span>
                           )}
                        </div>
                     </div>

                    {match.matches && match.matches.length > 0 && (
                      <div className="flex gap-3 mt-3">
                        <div className="w-12 h-16 flex-shrink-0">
                          <CardImage 
                            imageUrl={match.matches[0].image_url} 
                            name={match.matches[0].name} 
                            tcg={match.matches[0].tcg} 
                            foilTreatment={match.matches[0].foil_treatment}
                            enableHover={true}
                            enableModal={true}
                          />
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-bold truncate">{match.matches[0].name}</p>
                          <div className="flex items-center gap-1.5 text-[10px] text-text-muted mb-1">
                            <SetIcon setCode={match.matches[0].set_code} rarity={match.matches[0].rarity} size="xs" />
                            <span className="truncate">{match.matches[0].set_name}</span>
                          </div>
                          <div className="flex flex-wrap gap-1 mb-2">
                            <ConditionBadge condition={match.matches[0].condition} />
                            <FoilBadge foil={match.matches[0].foil_treatment} />
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="price text-sm">${match.matches[0].price.toLocaleString()} COP</span>
                            <button 
                              onClick={() => addItem(match.matches[0])}
                              className="btn-primary text-[10px] py-1 px-2"
                            >
                              + {t('pages.common.buttons.add', 'ADD')}
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
