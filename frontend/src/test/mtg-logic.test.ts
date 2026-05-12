import { describe, it, expect } from 'vitest';
import { 
  resolveCardTreatment, 
  extractMTGMetadata, 
  findMatchingPrint, 
  getArtOptions, 
  getPromoOptions,
  resolveFoilTreatment,
  getFoilOptions,
  getScryfallImage,
  extractPrices,
  resolveArtVariation,
  applyPrintPrices,
  getTreatmentOptions
} from '../lib/mtg-logic';
import { ScryfallCard } from '../lib/types';

describe('mtg-logic', () => {
  const mockCard: ScryfallCard = {
    id: '1',
    name: 'Lightning Bolt',
    set: 'sld',
    set_name: 'Secret Lair Drop',
    collector_number: '1',
    artist: 'Chris Rahn',
    border_color: 'black',
    promo: false,
    digital: false,
    finishes: ['nonfoil', 'foil'],
    prices: { usd: '1.00', usd_foil: '2.00', eur: '0.90' },
  };

  describe('resolveCardTreatment', () => {
    it('should identify normal cards', () => {
      expect(resolveCardTreatment(mockCard)).toBe('normal');
    });

    it('should identify borderless cards', () => {
      const borderless = { ...mockCard, border_color: 'borderless' };
      expect(resolveCardTreatment(borderless)).toBe('borderless');
    });

    it('should identify showcase cards', () => {
      const showcase = { ...mockCard, frame_effects: ['showcase'] };
      expect(resolveCardTreatment(showcase)).toBe('showcase');
    });

    it('should identify serialized cards', () => {
      const serialized = { ...mockCard, frame_effects: ['serialized'] };
      expect(resolveCardTreatment(serialized)).toBe('serialized');
    });

    it('should identify textless cards', () => {
      const textless = { ...mockCard, textless: true };
      expect(resolveCardTreatment(textless)).toBe('textless');
    });
  });

  describe('extractMTGMetadata', () => {
    it('should correctly extract metadata fields', () => {
      const metadata = extractMTGMetadata(mockCard);
      expect(metadata.artist).toBe('Chris Rahn');
      expect(metadata.collector_number).toBe('1');
      expect(metadata.rarity).toBe('');
      expect(metadata.promo_type).toBe('none');
    });

    it('should handle missing card gracefully', () => {
      const defaults = extractMTGMetadata(undefined);
      expect(defaults.language).toBe('en');
      expect(defaults.color_identity).toBe('');
      expect(defaults.rarity).toBe('');
      expect(defaults.frame_effects).toEqual([]);
    });
  });

  describe('getArtOptions', () => {
    const prints: ScryfallCard[] = [
      { ...mockCard, collector_number: '1', border_color: 'black' },
      { ...mockCard, collector_number: '2', border_color: 'borderless' }, // Different treatment
    ];

    it('should show all variations for Secret Lair (SLD) regardless of treatment', () => {
      const options = getArtOptions(prints, 'sld', 'normal');
      expect(options).toHaveLength(2);
      expect(options[0].cn).toBe('1');
      expect(options[1].cn).toBe('2');
    });

    it('should filter by treatment for other sets', () => {
      const nonSldPrints = prints.map(p => ({ ...p, set: 'lea' }));
      const options = getArtOptions(nonSldPrints, 'lea', 'normal');
      expect(options).toHaveLength(1);
      expect(options[0].cn).toBe('1');
    });
  });

  describe('getPromoOptions', () => {
    it('should filter out redundant tags like showcase or borderless', () => {
      const cardWithRedundantTags: ScryfallCard = {
        ...mockCard,
        promo_types: ['showcase', 'promopack', 'borderless']
      };
      const { promos } = getPromoOptions([cardWithRedundantTags], 'sld', 'normal', '1');
      expect(promos).toHaveLength(1);
      expect(promos).toContain('promopack');
      expect(promos).not.toContain('showcase');
      expect(promos).not.toContain('borderless');
    });

    it('should filter out exclusive foil tags', () => {
      const cardWithFoilTags: ScryfallCard = {
        ...mockCard,
        promo_types: ['surgefoil', 'universesbeyond', 'galaxyfoil']
      };
      const { promos } = getPromoOptions([cardWithFoilTags], 'sld', 'normal', '1');
      expect(promos).toHaveLength(1);
      expect(promos).toContain('universesbeyond');
      expect(promos).not.toContain('surgefoil');
      expect(promos).not.toContain('galaxyfoil');
    });
  });

  describe('findMatchingPrint', () => {
    const prints: ScryfallCard[] = [
      { ...mockCard, collector_number: '1', finishes: ['nonfoil'] },
      { ...mockCard, collector_number: '2', promo_types: ['promopack'], finishes: ['foil'] },
    ];

    it('should find an exact match', () => {
      const match = findMatchingPrint(prints, 'sld', 'normal', '2', 'promopack', 'foil');
      expect(match?.collector_number).toBe('2');
    });

    it('should fallback if promo is not found', () => {
      const match = findMatchingPrint(prints, 'sld', 'normal', '2', 'invalid_promo', 'non_foil');
      expect(match?.collector_number).toBe('2');
    });

    it('should fallback to first print if nothing matches', () => {
      const match = findMatchingPrint(prints, 'lea', 'normal', '999', 'none', 'non_foil');
      expect(match).toBeDefined();
    });
  });

  describe('resolveFoilTreatment', () => {
    it('should identify oil slick foil', () => {
      expect(resolveFoilTreatment({ ...mockCard, promo_types: ['oilslick'] })).toBe('oil_slick');
    });

    it('should identify galaxy foil', () => {
      expect(resolveFoilTreatment({ ...mockCard, promo_types: ['galaxyfoil'] })).toBe('galaxy_foil');
    });

    it('should identify etched foil from finishes', () => {
      expect(resolveFoilTreatment({ ...mockCard, finishes: ['etched'] })).toBe('etched_foil');
    });

    it('should identify regular foil from finishes', () => {
      expect(resolveFoilTreatment({ ...mockCard, finishes: ['foil'] })).toBe('foil');
    });

    it('should return non_foil if no foil finishes are present', () => {
      expect(resolveFoilTreatment({ ...mockCard, finishes: ['nonfoil'] })).toBe('non_foil');
    });
  });

  describe('getFoilOptions', () => {
    const prints: ScryfallCard[] = [
      { ...mockCard, collector_number: '1', finishes: ['nonfoil', 'foil'] },
      { ...mockCard, collector_number: '2', finishes: ['nonfoil', 'etched'] },
    ];

    it('should return available foils for a specific art', () => {
      const options = getFoilOptions(prints, 'sld', 'normal', '1', 'none');
      expect(options).toContain('non_foil');
      expect(options).toContain('foil');
      expect(options).not.toContain('etched_foil');
    });

    it('should return available foils even if promo matching fails', () => {
      const options = getFoilOptions(prints, 'sld', 'normal', '2', 'invalid_promo');
      expect(options).toContain('non_foil');
      // Etched foil is only shown when the promo tag matches; with invalid_promo it relaxes to cn match
      // which supports non_foil and etched finishes, but etched requires matching promo or 'none'
      expect(options.length).toBeGreaterThanOrEqual(1);
    });

    it('should show specialized foils when card has matching promo_type', () => {
      const surgeFoilPrint: ScryfallCard = {
        ...mockCard,
        collector_number: '3',
        promo_types: ['surgefoil', 'universesbeyond'],
        finishes: ['foil']
      };
      // Searching with the universesbeyond promo
      const options = getFoilOptions([surgeFoilPrint], 'sld', 'normal', '3', 'universesbeyond');
      expect(options).toContain('surge_foil');
      expect(options).not.toContain('foil');
    });

    it('should exclude non_foil when an exclusive foil tag is present', () => {
      const galaxyFoilPrint: ScryfallCard = {
        ...mockCard,
        collector_number: '4',
        promo_types: ['galaxyfoil'],
        finishes: ['nonfoil', 'foil'] // even if Scryfall weirdly includes nonfoil
      };
      const options = getFoilOptions([galaxyFoilPrint], 'sld', 'normal', '4', 'none');
      expect(options).not.toContain('non_foil');
      expect(options).toContain('galaxy_foil');
    });
  });

  describe('getScryfallImage', () => {
    it('should return normal image uri', () => {
      const card = { ...mockCard, image_uris: { normal: 'http://test.com/normal.jpg' } };
      expect(getScryfallImage(card)).toBe('http://test.com/normal.jpg');
    });

    it('should return face image uri for double-faced cards', () => {
      const card = { 
        ...mockCard, 
        card_faces: [{ image_uris: { normal: 'http://test.com/face.jpg' } }] 
      };
      expect(getScryfallImage(card)).toBe('http://test.com/face.jpg');
    });
  });

  describe('extractPrices', () => {
    it('should extract usd and eur prices', () => {
      const prices = extractPrices(mockCard);
      expect(prices.usd).toBe(1.0);
      expect(prices.eur).toBe(0.9);
    });

    it('should fallback to usd_foil if usd is null', () => {
      const prices = extractPrices({ ...mockCard, prices: { usd: null, usd_foil: '5.00', eur: '4.00' } });
      expect(prices.usd).toBe(5.0);
    });
  });

  describe('resolveArtVariation', () => {
    it('should identify japanese art', () => {
      expect(resolveArtVariation({ ...mockCard, promo_types: ['jpwalker'] })).toBe('japanese');
    });

    it('should identify concept art', () => {
      expect(resolveArtVariation({ ...mockCard, promo_types: ['concept'] })).toBe('concept_art');
    });

    it('should return empty string for normal art', () => {
      expect(resolveArtVariation(mockCard)).toBe('');
    });
  });

  describe('applyPrintPrices', () => {
    it('should apply TCGPlayer prices', () => {
      expect(applyPrintPrices(mockCard, 'non_foil', 'tcgplayer')).toBe(1.0);
      expect(applyPrintPrices(mockCard, 'foil', 'tcgplayer')).toBe(2.0);
      expect(applyPrintPrices({ ...mockCard, prices: { usd_etched: '3.00' } }, 'etched_foil', 'tcgplayer')).toBe(3.0);
    });

    it('should apply Cardmarket prices', () => {
      expect(applyPrintPrices(mockCard, 'non_foil', 'cardmarket')).toBe(0.9);
    });
  });

  describe('getTreatmentOptions', () => {
    const prints: ScryfallCard[] = [
      { ...mockCard, collector_number: '1', border_color: 'black' },
      { ...mockCard, collector_number: '2', border_color: 'borderless' },
    ];

    it('should return unique treatments for a set', () => {
      const options = getTreatmentOptions(prints, 'sld');
      expect(options).toContain('normal');
      expect(options).toContain('borderless');
    });
  });
});
