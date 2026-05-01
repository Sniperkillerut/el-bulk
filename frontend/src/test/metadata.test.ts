import { describe, it, expect } from 'vitest';
import { formatProductDescription } from '../lib/metadata';

describe('formatProductDescription', () => {
  it('formats a standard product correctly', () => {
    const mockProduct = {
      name: 'Black Lotus',
      set_name: 'Unlimited',
      condition: 'NM',
      foil_treatment: 'Non-Foil',
      price: 15000,
      stock: 1
    } as any;
    
    const result = formatProductDescription(mockProduct);
    expect(result).toBe('Buy Black Lotus ([Unlimited] NM Non-Foil) — COP 15,000 | 1 in stock. Secure shipping from El Bulk Bogotá.');
  });

  it('handles out of stock items', () => {
    const mockProduct = {
      name: 'Mox Emerald',
      set_name: 'Beta',
      condition: 'HP',
      foil_treatment: 'Non-Foil',
      price: 5000,
      stock: 0
    } as any;
    
    const result = formatProductDescription(mockProduct);
    expect(result).toContain('Out of stock');
  });
});
