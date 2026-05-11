import { describe, it, expect } from 'vitest';
import { filterPlateLetterSegment, mapPlateLetterChar, normalizeCarPlate } from '../carPlate';

describe('normalizeCarPlate', () => {
  it('maps Cyrillic plate letters to Latin', () => {
    expect(normalizeCarPlate('А123ВС777')).toBe('A123BC777');
  });

  it('rejects letters not used on Russian plates', () => {
    expect(normalizeCarPlate('D123DD77')).toBe('');
    expect(normalizeCarPlate('А123ДС77')).toBe('');
  });
});

describe('mapPlateLetterChar', () => {
  it('accepts Latin plate letters', () => {
    expect(mapPlateLetterChar('a')).toBe('A');
    expect(mapPlateLetterChar('x')).toBe('X');
  });

  it('rejects non-plate Latin', () => {
    expect(mapPlateLetterChar('d')).toBe('');
    expect(mapPlateLetterChar('z')).toBe('');
  });
});

describe('filterPlateLetterSegment', () => {
  it('keeps only plate letters up to maxLen', () => {
    expect(filterPlateLetterSegment('ВСХ', 2)).toBe('BC');
  });
});
