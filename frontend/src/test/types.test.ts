import { describe, it, expect } from 'vitest';
import { resolveLabel } from '../lib/types';

describe('resolveLabel', () => {
  const mockMap = {
    'non_foil': 'Non-Foil',
    'foil': 'Foil',
    'etched_foil': 'Etched Foil',
    'surge_foil': 'Surge Foil',
  };

  it('returns empty string for null, empty or "none" keys', () => {
    expect(resolveLabel('', mockMap)).toBe('');
    // @ts-ignore
    expect(resolveLabel(null, mockMap)).toBe('');
    expect(resolveLabel('none', mockMap)).toBe('');
  });

  it('returns mapped value if key exists in map', () => {
    expect(resolveLabel('non_foil', mockMap)).toBe('Non-Foil');
    expect(resolveLabel('foil', mockMap)).toBe('Foil');
    expect(resolveLabel('surge_foil', mockMap)).toBe('Surge Foil');
  });

  it('formats camelCase keys', () => {
    expect(resolveLabel('surgeFoil', {})).toBe('Surge Foil');
    expect(resolveLabel('texturedFoil', {})).toBe('Textured Foil');
    expect(resolveLabel('galaxyFoil', {})).toBe('Galaxy Foil');
  });

  it('formats snake_case keys', () => {
    expect(resolveLabel('galaxy_foil', {})).toBe('Galaxy Foil');
    expect(resolveLabel('retro_frame', {})).toBe('Retro Frame');
  });

  it('handles special cases', () => {
    expect(resolveLabel('oilslick', {})).toBe('Oil Slick');
    expect(resolveLabel('stepandcompleat', {})).toBe('Step-and-Compleat');
    expect(resolveLabel('silverfoil', {})).toBe('Silver Foil');
    // Case insensitive special cases
    expect(resolveLabel('OILSLICK', {})).toBe('Oil Slick');
    // StepAndCompleat will be camelCase separated into "Step And Compleat"
    expect(resolveLabel('StepAndCompleat', {})).toBe('Step And Compleat');
  });

  it('adds space before foil suffix if missing', () => {
    expect(resolveLabel('texturedfoil', {})).toBe('Textured Foil');
    expect(resolveLabel('galaxyfoil', {})).toBe('Galaxy Foil');
  });

  it('handles complex future-proofing cases', () => {
    expect(resolveLabel('confettiFoil', {})).toBe('Confetti Foil');
    expect(resolveLabel('doubleRainbow', {})).toBe('Double Rainbow');
  });

  it('avoids double spaces or redundant "Foil" words', () => {
    // These are currently passing or failing depending on implementation
    expect(resolveLabel('surge_foil', {})).toBe('Surge Foil');
    expect(resolveLabel('surgeFoil', {})).toBe('Surge Foil');
  });
});
