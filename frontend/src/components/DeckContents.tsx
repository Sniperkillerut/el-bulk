import { useState } from 'react';
import { DeckCard } from '@/lib/types';
import CardImage from './CardImage';
import { getDeckAnalytics } from '@/lib/mtg-logic';
import { useLanguage } from '@/context/LanguageContext';

interface DeckContentsProps {
  cards: DeckCard[];
  tcg?: string;
  className?: string;
}

export default function DeckContents({ cards, tcg, className = '' }: DeckContentsProps) {
  const { total, summary, groups } = getDeckAnalytics(cards);
  const { t } = useLanguage();
  
  // Track expanded groups. Default is all expanded.
  const [expandedGroups, setExpandedGroups] = useState<string[]>(
    Object.keys(groups).filter(key => groups[key].length > 0)
  );

  if (!cards || cards.length === 0) return null;

  const toggleGroup = (groupName: string) => {
    setExpandedGroups(prev => 
      prev.includes(groupName) 
        ? prev.filter(g => g !== groupName)
        : [...prev, groupName]
    );
  };

  const expandAll = () => setExpandedGroups(Object.keys(groups).filter(key => groups[key].length > 0));
  const collapseAll = () => setExpandedGroups([]);

  return (
    <div className={`mt-4 ${className}`}>
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 mb-4">
        <div>
          <h3 className="font-mono-stack text-xs uppercase text-text-muted opacity-70 flex items-center gap-2 mb-1 font-bold tracking-widest">
            <span>{t('components.deck_contents.title', 'Deck Contents')}</span>
            <span className="bg-ink-border/10 px-2 py-0.5 rounded-full text-[10px] font-bold text-ink-deep/60">
              {total} {t('pages.checkout.summary.items', 'CARDS')}
            </span>
          </h3>
          {summary && (
            <p className="text-[10px] font-mono-stack text-text-muted opacity-60 uppercase tracking-tighter">
              {summary}
            </p>
          )}
        </div>
        <div className="flex gap-4">
          <button type="button" onClick={expandAll} className="text-[10px] uppercase font-mono-stack font-bold text-gold hover:underline">{t('components.deck_contents.expand_all', 'Expand All')}</button>
          <button type="button" onClick={collapseAll} className="text-[10px] uppercase font-mono-stack font-bold text-text-muted hover:underline">{t('components.deck_contents.collapse_all', 'Collapse All')}</button>
        </div>
      </div>

      <div className="space-y-4 max-h-[500px] overflow-y-auto pr-2 custom-scrollbar">
        {Object.entries(groups).map(([groupName, groupCards]) => {
          if (groupCards.length === 0) return null;
          const isExpanded = expandedGroups.includes(groupName);
          const groupQty = groupCards.reduce((sum, c) => sum + c.quantity, 0);

          return (
            <div key={groupName} className="border-b border-ink-border/5 last:border-0 pb-4">
              <button 
                type="button"
                onClick={() => toggleGroup(groupName)}
                className="w-full flex items-center justify-between py-2 group/header"
              >
                <div className="flex items-center gap-2">
                  <span className={`text-xs transition-transform duration-200 ${isExpanded ? 'rotate-90' : ''}`}>
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="9 18 15 12 9 6"></polyline></svg>
                  </span>
                  <span className="font-mono-stack text-xs font-bold uppercase tracking-widest text-ink-deep group-hover/header:text-gold transition-colors">
                    {groupName} ({groupQty})
                  </span>
                </div>
                <div className="h-px flex-1 bg-ink-border/10 mx-4 opacity-50"></div>
              </button>

              {isExpanded && (
                <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-8 gap-2 mt-2 px-1">
                  {groupCards.map((card) => (
                    <div key={card.id} className="relative group/card">
                      <div className="border border-kraft-dark/30 rounded-sm overflow-hidden bg-ink-surface shadow-sm transition-transform hover:scale-105">
                        <CardImage 
                          imageUrl={card.image_url} 
                          name={card.name} 
                          tcg={tcg} 
                          foilTreatment={card.foil_treatment}
                          enableHover={true}
                        />
                      </div>
                      <div className="absolute bottom-0 right-0 bg-ink-deep text-white text-[10px] font-mono-stack font-bold px-1.5 py-0.5 rounded-tl-sm pointer-events-none shadow-md z-10">
                        x{card.quantity}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
