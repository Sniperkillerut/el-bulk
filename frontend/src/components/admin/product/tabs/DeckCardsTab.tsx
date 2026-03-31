'use client';

import { useState } from 'react';
import { FormState } from '../types';
import { DeckCard, ScryfallCard } from '@/lib/types';
import ScryfallPopulate from '../ScryfallPopulate';
import { getScryfallImage } from '@/lib/mtg-logic';

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
        const b: any = await r.json();
        if (b.data) {
          const paperOnly = (b.data as ScryfallCard[]).filter(c => !c.digital);
          results = results.concat(paperOnly);
        }
        nextUrl = b.has_more ? (b.next_page as string) : null;
        if (nextUrl) await new Promise(res => setTimeout(res, 100));
      }

      if (results.length === 0) throw new Error('No printings found');
      
      const oracleId = (results[0] as any).oracle_id;
      if (oracleId) {
        let oNextUrl: string | null = `https://api.scryfall.com/cards/search?q=oracle_id:${oracleId}+game:paper&unique=prints&order=released`;
        let oResults: ScryfallCard[] = [];
        while (oNextUrl) {
          const r: Response = await fetch(oNextUrl);
          if (!r.ok) break;
          const b: any = await r.json();
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
    } catch (e: any) {
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
      image_url: card ? (getScryfallImage(card) || '') : ''
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

  return (
    <div className="space-y-6">
      <div className="card p-5 bg-white/40 border-white/20 backdrop-blur-md">
        <h3 className="font-mono-stack text-xs uppercase tracking-widest text-text-muted mb-4 opacity-70">Deck Builder (Add Cards)</h3>
        
        {form.tcg === 'mtg' ? (
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
              <div className="mt-4 border-t border-ink-border/10 pt-4">
                <label className="text-[10px] font-mono-stack block mb-2 uppercase tracking-widest opacity-50">Select Specific Printing to Add</label>
                <div className="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-8 gap-2 overflow-y-auto max-h-64 p-2 custom-scrollbar">
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
        ) : (
          <div className="flex items-end gap-2">
            <div className="flex-1">
              <label className="text-xs mb-1 block">Card Name</label>
              <input type="text" value={searchName} onChange={e => setSearchName(e.target.value)} placeholder="Type a custom card name..." />
            </div>
            <button className="btn-primary py-2 px-4" onClick={() => addCard(null, true)} disabled={!searchName}>Add Custom Card</button>
          </div>
        )}
      </div>

      <div>
        <h3 className="font-mono-stack text-xs uppercase tracking-widest text-text-muted mb-4 opacity-70">
          Current Deck List ({form.deck_cards.reduce((sum, c) => sum + c.quantity, 0)} cards)
        </h3>
        {form.deck_cards.length === 0 ? (
          <div className="text-sm text-text-muted py-8 text-center bg-white/30 rounded-lg border border-white/20 font-mono-stack">
            Deck is currently empty.
          </div>
        ) : (
          <div className="flex flex-col gap-2 max-h-96 overflow-y-auto custom-scrollbar p-1">
            {form.deck_cards.map(card => (
              <div key={card.id} className="flex items-center gap-3 p-2 bg-white/60 border border-white/40 rounded-lg shrink-0">
                {card.image_url ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={card.image_url} alt={card.name} className="w-10 h-14 object-cover rounded-sm border border-ink-border/20 shadow-sm" />
                ) : (
                  <div className="w-10 h-14 bg-ink-border/10 rounded-sm flex items-center justify-center text-[8px] text-text-muted text-center leading-tight">
                    NO<br/>IMG
                  </div>
                )}
                <div className="flex flex-col flex-1 min-w-0">
                  <span className="font-bold text-sm truncate">{card.name}</span>
                  <span className="text-[10px] uppercase font-mono-stack text-text-muted truncate">
                    {card.set_code && `${card.set_code.toUpperCase()}`}
                    {card.set_code && card.collector_number && ' · '}
                    {card.collector_number && `#${card.collector_number}`}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <button type="button" onClick={() => updateCardQty(card.id, card.quantity - 1)} className="w-6 h-6 rounded bg-ink-border/10 hover:bg-ink-border/20 transition-colors flex items-center justify-center font-mono font-bold" disabled={card.quantity <= 1}>-</button>
                  <input type="number" value={card.quantity} onChange={e => updateCardQty(card.id, parseInt(e.target.value) || 1)} className="w-12 text-center text-xs py-1 px-1 bg-white/50 m-0" min={1} />
                  <button type="button" onClick={() => updateCardQty(card.id, card.quantity + 1)} className="w-6 h-6 rounded bg-ink-border/10 hover:bg-ink-border/20 transition-colors flex items-center justify-center font-mono font-bold">+</button>
                </div>
                <button type="button" onClick={() => removeCard(card.id)} className="w-8 h-8 rounded-full text-hp-color hover:bg-hp-color/10 flex items-center justify-center transition-colors">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
