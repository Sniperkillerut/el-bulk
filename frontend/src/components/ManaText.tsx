import React from 'react';

const MANA_MAP: Record<string, string> = {
  't': 'tap',
  'q': 'untap',
  'chaos': 'chaos',
  'pw': 'planeswalker',
  'tk': 'ticket',
};

const formatSymbol = (symbol: string) => {
  // Scryfall symbols like {W/U} or {2/B} or {B/P} 
  // need to be cleaned of brackets and slashes for ManaFont
  const clean = symbol.replace(/[{}]/g, '').toLowerCase().replace(/\//g, '');
  return MANA_MAP[clean] || clean;
};

interface ManaTextProps {
  text?: string;
  className?: string;
}

/**
 * Parses MTG oracle text and replaces bracketed symbols (e.g., {W}, {T})
 * with their corresponding high-fidelity icons from ManaFont.
 */
const ManaText = ({ text, className = '' }: ManaTextProps) => {
  if (!text) return null;

  // Split by potential MTG symbols in braces, preserving the braces for detection
  const parts = text.split(/(\{.+?\})/g);

  return (
    <span className={className}>
      {parts.map((part, i) => {
        if (part.startsWith('{') && part.endsWith('}')) {
          const symbolClass = formatSymbol(part);
          return (
            <i 
              key={`${part}-${i}`} 
              className={`ms ms-${symbolClass} ms-cost ms-shadow align-middle inline-block mx-[0.1em] scale-[1.05]`} 
              title={part}
            />
          );
        }
        // Handle newlines if any are in the plain text
        return part;
      })}
    </span>
  );
};

export default ManaText;
