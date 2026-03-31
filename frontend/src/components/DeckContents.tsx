'use client';

import { DeckCard } from '@/lib/types';
import CardImage from './CardImage';
import { getDeckAnalytics } from '@/lib/mtg-logic';

interface DeckContentsProps {
  cards: DeckCard[];
  tcg?: string;
  className?: string;
}

/**
 * DeckContents renders a responsive grid of deck cards as thumbnails.
 * Each thumbnail shows the quantity and triggers a large preview on hover.
 */
export default function DeckContents({ cards, tcg, className = '' }: DeckContentsProps) {
  if (!cards || cards.length === 0) return null;

  const { total, summary } = getDeckAnalytics(cards);

  return (
    <div className={`mt-4 ${className}`}>
      <div className="mb-4">
        <h3 className="font-mono-stack text-xs uppercase text-text-muted opacity-70 flex justify-between items-center mb-1 font-bold tracking-widest">
          <span>Deck Contents</span>
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
      <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-8 gap-2 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
        {cards.map((card) => (
          <div key={card.id} className="relative group">
            <div className="border border-kraft-dark/30 rounded-sm overflow-hidden bg-ink-surface shadow-sm transition-transform hover:scale-105">
              <CardImage 
                imageUrl={card.image_url} 
                name={card.name} 
                tcg={tcg} 
                enableHover={true}
              />
            </div>
            <div className="absolute bottom-0 right-0 bg-ink-deep text-white text-[10px] font-mono-stack font-bold px-1.5 py-0.5 rounded-tl-sm pointer-events-none shadow-md z-10">
              x{card.quantity}
            </div>
            {/* Tooltip on bottom for name on hover if needed, but the big image shows the card name usually */}
          </div>
        ))}
      </div>
    </div>
  );
}
