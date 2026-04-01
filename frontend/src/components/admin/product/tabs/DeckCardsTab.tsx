'use client';

import { useState } from 'react';
import { FormState } from '../types';
import { DeckCard, ScryfallCard } from '@/lib/types';
import ScryfallPopulate from '../ScryfallPopulate';
import { getScryfallImage, getDeckAnalytics, resolveCardTreatment, resolveFoilTreatment, resolveArtVariation } from '@/lib/mtg-logic';
import { FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import CardImage from '@/components/CardImage';

interface DeckCardsTabProps {
  form: FormState;
  onUpdate: (updates: Partial<FormState>) => void;
}

export default function DeckCardsTab({ form, onUpdate }: DeckCardsTabProps) {
  const [lookingUp, setLookingUp] = useState(false);
  const [scryfallPrints, setScryfallPrints] = useState<ScryfallCard[]>([]);
  const [searchName, setSearchName] = useState('');
  const [searchSet, setSearchSet] = useState('');
  const [searchCn, setSearchCn] = useState('');

  const handlePopulate = async (forceSearchName?: string) => {
    const name = forceSearchName || searchName.trim();
    const set = searchSet.trim().toLowerCase();
    const cn = searchCn.trim();
    if (!name && (!set || !cn)) return;

    setLookingUp(true);
    try {
      let searchQ = "";
      if (set && cn) searchQ = `set:${set} cn:"${cn}"`;
      else if (name) searchQ = `!"${name}"`;
      else if (set) searchQ = `set:${set}`;

      let nextUrl: string | null = `https://api.scryfall.com/cards/search?q=${encodeURIComponent(searchQ)}+game:paper&unique=prints&order=released`;
      let results: ScryfallCard[] = [];
      while (nextUrl) {
        const r: Response = await fetch(nextUrl);
        if (!r.ok) break;
        const b: { data: ScryfallCard[]; has_more: boolean; next_page?: string } = await r.json();
        if (b.data) {
          const paperOnly = (b.data as ScryfallCard[]).filter(c => !c.digital);
          results = results.concat(paperOnly);
        }
        nextUrl = b.has_more ? (b.next_page as string) : null;
        if (nextUrl) await new Promise(res => setTimeout(res, 100));
      }

      if (results.length === 0) throw new Error('No printings found');
      
      const oracleId = (results[0] as unknown as { oracle_id?: string }).oracle_id;
      if (oracleId) {
        let oNextUrl: string | null = `https://api.scryfall.com/cards/search?q=oracle_id:${oracleId}+game:paper&unique=prints&order=released`;
        let oResults: ScryfallCard[] = [];
        while (oNextUrl) {
          const r: Response = await fetch(oNextUrl);
          if (!r.ok) break;
          const b: { data: ScryfallCard[]; has_more: boolean; next_page?: string } = await r.json();
          if (b.data) {
            const paperOnly = (b.data as ScryfallCard[]).filter(c => !c.digital);
            oResults = oResults.concat(paperOnly);
          }
          oNextUrl = b.has_more ? (b.next_page as string) : null;
          if (oNextUrl) await new Promise(res => setTimeout(res, 100));
        }
        if (oResults.length > results.length) results = oResults;
      }

      setScryfallPrints(results);
    } catch (e: unknown) {
      console.error(e);
      setScryfallPrints([]);
    } finally {
      setLookingUp(false);
    }
  };

  const addCard = (card: ScryfallCard | null, isManual: boolean = false) => {
    const defaultDeckCard: DeckCard = {
      id: crypto.randomUUID(),
      name: isManual ? searchName : card?.name || 'Unknown',
      set_code: card?.set || searchSet || '',
      collector_number: card?.collector_number || searchCn || '',
      quantity: 1,
      type_line: card?.type_line || '',
      image_url: card ? (getScryfallImage(card) || '') : '',
      foil_treatment: card ? resolveFoilTreatment(card) : 'non_foil',
      card_treatment: card ? resolveCardTreatment(card) : 'normal',
      rarity: card?.rarity || 'common',
      art_variation: card ? resolveArtVariation(card) : ''
    };
    
    // check if it already exists, if so increment
    const existingIdx = form.deck_cards.findIndex(c => c.name === defaultDeckCard.name && c.set_code === defaultDeckCard.set_code && c.collector_number === defaultDeckCard.collector_number);
    if (existingIdx !== -1) {
      const newList = [...form.deck_cards];
      newList[existingIdx] = { ...newList[existingIdx], quantity: newList[existingIdx].quantity + 1 };
      onUpdate({ deck_cards: newList });
    } else {
      onUpdate({ deck_cards: [...form.deck_cards, defaultDeckCard] });
    }

    setSearchName('');
    setSearchSet('');
    setSearchCn('');
    setScryfallPrints([]);
  };

  const removeCard = (id: string) => {
    onUpdate({ deck_cards: form.deck_cards.filter(c => c.id !== id) });
  };

  const updateCardQty = (id: string, qty: number) => {
    onUpdate({ deck_cards: form.deck_cards.map(c => c.id === id ? { ...c, quantity: Math.max(1, qty) } : c) });
  };

  const editCard = (card: DeckCard) => {
    setSearchName(card.name);
    setSearchSet(card.set_code || '');
    setSearchCn(card.collector_number || '');
    
    // Trigger populate with the existing name
    // (Small delay to ensure setState is processed if needed, though handlePopulate uses params)
    handlePopulate(card.name);
    
    // Optional: Scroll back to top to see the results
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const { total, summary } = getDeckAnalytics(form.deck_cards);

  return (
    <div className="space-y-4">
      {form.tcg === 'mtg' ? (
        <div className="animate-in fade-in slide-in-from-top-2 duration-300">
          <div>
            <ScryfallPopulate 
              name={searchName}
              setCode={searchSet}
              collectorNumber={searchCn}
              setName={searchSet}
              scryfallPrints={scryfallPrints}
              lookingUp={lookingUp}
              onNameChange={val => { setSearchName(val); setScryfallPrints([]); }}
              onSetCodeChange={val => setSearchSet(val)}
              onCollectorNumberChange={val => setSearchCn(val)}
              onPopulate={() => handlePopulate()}
              onCardSelect={card => {
                setSearchName(card.name);
                handlePopulate(card.name);
              }}
              onSetSearchChange={val => setSearchSet(val)}
            />
            
            {scryfallPrints.length > 0 && (
              <div className="mt-3 border-t border-ink-border/10 pt-3">
                <label className="text-[10px] font-mono-stack block mb-1 uppercase tracking-widest opacity-50">Select Specific Printing to Add</label>
                <div className="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-8 gap-2 overflow-y-auto max-h-48 p-2 custom-scrollbar">
                  {scryfallPrints.map((print, i) => {
                    const img = getScryfallImage(print);
                    return (
                      <div 
                        key={i} 
                        className="cursor-pointer group relative hover:scale-105 transition-transform"
                        onClick={() => addCard(print)}
                      >
                        <div className="aspect-[63/88] rounded-md overflow-hidden bg-ink-border/10">
                          {img ? (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={img} alt={print.name} className="w-full h-full object-cover" />
                          ) : (
                            <div className="w-full h-full flex flex-col items-center justify-center p-1 text-[8px] font-mono text-center text-text-muted">
                              {print.set.toUpperCase()}<br/>#{print.collector_number}
                            </div>
                          )}
                        </div>
                        <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center pointer-events-none rounded-md">
                          <span className="text-white font-bold text-xs">+ ADD</span>
                        </div>
                        <div className="text-[9px] mt-1 text-center font-mono-stack truncate opacity-70">{print.set.toUpperCase()} · {print.collector_number}</div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="flex items-end gap-2 p-2 bg-white/40 border border-white/20 rounded-md">
          <div className="flex-1">
            <label className="text-[10px] font-mono-stack mb-1 block uppercase opacity-60">Card Name (Custom)</label>
            <input type="text" value={searchName} onChange={e => setSearchName(e.target.value)} placeholder="Type a custom card name..." />
          </div>
          <button className="btn-primary py-2 px-4 text-xs font-bold" onClick={() => addCard(null, true)} disabled={!searchName}>Add Custom Card</button>
        </div>
      )}

      <div>
        <div className="mb-2">
          <h3 className="font-mono-stack text-[10px] uppercase tracking-widest text-text-muted opacity-70 flex justify-between items-center mb-0.5">
            <span>Current Deck List</span>
            <span className="bg-ink-border/10 px-2 py-0.5 rounded-full text-[10px] font-bold text-ink-deep/60">
              {total} CARDS
            </span>
          </h3>
          {summary && (
            <p className="text-[10px] font-mono-stack text-text-muted opacity-60 uppercase tracking-tighter">
              {summary}
            </p>
          )}
        </div>
        {form.deck_cards.length === 0 ? (
          <div className="text-sm text-text-muted py-8 text-center bg-white/30 rounded-lg border border-white/20 font-mono-stack">
            Deck is currently empty.
          </div>
        ) : (
          <div className="space-y-3 max-h-[700px] overflow-y-auto custom-scrollbar p-1">
            {Object.entries(getDeckAnalytics(form.deck_cards).groups).map(([groupName, groupCards]) => {
              if (groupCards.length === 0) return null;
              const groupQty = groupCards.reduce((sum, c) => sum + c.quantity, 0);

              return (
                <div key={groupName} className="space-y-1">
                  <div className="flex items-center gap-2 py-0.5 px-2 bg-ink-border/5 rounded">
                    <span className="font-mono-stack text-[10px] font-bold uppercase tracking-wider text-text-muted">
                      {groupName} ({groupQty})
                    </span>
                    <div className="h-px flex-1 bg-ink-border/10"></div>
                  </div>
                  
                  <div className="flex flex-col gap-1.5">
                    {groupCards.map(card => (
                      <div key={card.id} className="flex items-center gap-2 p-1 bg-white/60 border border-white/40 rounded shadow-sm shrink-0 group/row hover:border-gold/30 transition-colors">
                        <div className="w-8 h-12 shrink-0 rounded-sm overflow-hidden border border-ink-border/20 shadow-sm shadow-black/5">
                          <CardImage 
                            imageUrl={card.image_url} 
                            name={card.name} 
                            tcg={form.tcg} 
                            enableHover={true} 
                          />
                        </div>
                        <div className="flex flex-col flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-bold text-[11px] truncate">{card.name}</span>
                            {card.rarity && (
                              <span className={`text-[8px] px-1.5 py-0.5 rounded-full font-bold uppercase ${
                                card.rarity === 'mythic' ? 'bg-hp-color/10 text-hp-color' :
                                card.rarity === 'rare' ? 'bg-gold/10 text-gold' :
                                card.rarity === 'uncommon' ? 'bg-ink-deep/10 text-ink-deep' :
                                'bg-text-muted/10 text-text-muted'
                              }`}>
                                {card.rarity}
                              </span>
                            )}
                          </div>
                          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
                            <span className="text-[10px] font-mono-stack text-text-muted">
                              {card.set_code?.toUpperCase()} #{card.collector_number}
                            </span>
                            {card.foil_treatment && card.foil_treatment !== 'non_foil' && (
                              <span className="text-[9px] font-bold text-ink-deep/60 flex items-center gap-0.5">
                                <span className="w-1.5 h-1.5 rounded-full bg-gradient-to-tr from-blue-400 via-purple-400 to-pink-400"></span>
                                {resolveLabel(card.foil_treatment, FOIL_LABELS)}
                              </span>
                            )}
                            {card.card_treatment && card.card_treatment !== 'normal' && (
                              <span className="text-[9px] font-mono-stack font-bold px-1 bg-ink-border/5 rounded text-text-muted border border-ink-border/5">
                                {resolveLabel(card.card_treatment, TREATMENT_LABELS)}
                              </span>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <button type="button" onClick={() => updateCardQty(card.id, card.quantity - 1)} className="w-6 h-6 rounded bg-ink-border/10 hover:bg-ink-border/20 transition-colors flex items-center justify-center font-mono font-bold" disabled={card.quantity <= 1}>-</button>
                          <input type="number" value={card.quantity} onChange={e => updateCardQty(card.id, parseInt(e.target.value) || 1)} className="w-12 text-center text-xs py-1 px-1 bg-white/50 m-0 border-none outline-none focus:ring-0" min={1} />
                          <button type="button" onClick={() => updateCardQty(card.id, card.quantity + 1)} className="w-6 h-6 rounded bg-ink-border/10 hover:bg-ink-border/20 transition-colors flex items-center justify-center font-mono font-bold">+</button>
                        </div>
                        <div className="flex items-center gap-1">
                          <button type="button" onClick={() => editCard(card)} className="w-8 h-8 rounded-full text-gold hover:bg-gold/10 flex items-center justify-center transition-colors" title="Edit/Repopulate">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                          </button>
                          <button type="button" onClick={() => removeCard(card.id)} className="w-8 h-8 rounded-full text-hp-color hover:bg-hp-color/10 flex items-center justify-center transition-colors" title="Remove">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
