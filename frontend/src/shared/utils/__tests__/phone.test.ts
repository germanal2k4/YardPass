import { describe, it, expect } from 'vitest';
import {
  canonicalPhone,
  extractNsnDigits,
  formatNsnMask,
  formatPhoneInput,
  isValidOrEmptyPhone,
} from '../phone';

describe('formatNsnMask', () => {
  it('renders only the national part without the country code', () => {
    expect(formatNsnMask('')).toBe('');
    expect(formatNsnMask('9')).toBe('(9');
    expect(formatNsnMask('900')).toBe('(900)');
    expect(formatNsnMask('9001')).toBe('(900) 1');
    expect(formatNsnMask('900123456')).toBe('(900) 123-45-6');
    expect(formatNsnMask('9001234567')).toBe('(900) 123-45-67');
  });

  it('auto-normalizes pasted values with +7/8/7 prefixes', () => {
    expect(formatNsnMask('+7 (900) 123-45-67')).toBe('(900) 123-45-67');
    expect(formatNsnMask('89001234567')).toBe('(900) 123-45-67');
    expect(formatNsnMask('79001234567')).toBe('(900) 123-45-67');
  });
});

describe('extractNsnDigits', () => {
  it('strips +7 prefix and keeps the 10-digit national number', () => {
    expect(extractNsnDigits('+7 (900) 123-45-67')).toBe('9001234567');
  });

  it('treats leading 8 the same as 7 (Russian trunk prefix)', () => {
    expect(extractNsnDigits('89001234567')).toBe('9001234567');
  });

  it('keeps a bare 10-digit input', () => {
    expect(extractNsnDigits('9001234567')).toBe('9001234567');
  });

  it('drops everything beyond 11 digits so paste of long strings is bounded', () => {
    expect(extractNsnDigits('890012345670000')).toBe('9001234567');
  });

  it('returns empty for empty input', () => {
    expect(extractNsnDigits('')).toBe('');
  });
});

describe('formatPhoneInput', () => {
  it('renders the canonical mask progressively', () => {
    expect(formatPhoneInput('')).toBe('');
    expect(formatPhoneInput('9')).toBe('+7 (9');
    expect(formatPhoneInput('900')).toBe('+7 (900)');
    expect(formatPhoneInput('9001')).toBe('+7 (900) 1');
    expect(formatPhoneInput('900123')).toBe('+7 (900) 123');
    expect(formatPhoneInput('9001234')).toBe('+7 (900) 123-4');
    expect(formatPhoneInput('900123456')).toBe('+7 (900) 123-45-6');
    expect(formatPhoneInput('9001234567')).toBe('+7 (900) 123-45-67');
  });

  it('normalizes mixed input formats to the same mask', () => {
    const expected = '+7 (900) 123-45-67';
    expect(formatPhoneInput('+7 (900) 123-45-67')).toBe(expected);
    expect(formatPhoneInput('89001234567')).toBe(expected);
    expect(formatPhoneInput('79001234567')).toBe(expected);
    expect(formatPhoneInput('9001234567')).toBe(expected);
    expect(formatPhoneInput('абв 8 900 abc 123 -- 45 67')).toBe(expected);
  });
});

describe('canonicalPhone', () => {
  it('returns the +7XXXXXXXXXX form only when 10 NSN digits are present', () => {
    expect(canonicalPhone('+7 (900) 123-45-67')).toBe('+79001234567');
    expect(canonicalPhone('89001234567')).toBe('+79001234567');
    expect(canonicalPhone('9001234567')).toBe('+79001234567');
  });

  it('returns empty for partial or empty input', () => {
    expect(canonicalPhone('')).toBe('');
    expect(canonicalPhone('900')).toBe('');
    expect(canonicalPhone('+7 (900) 123')).toBe('');
  });
});

describe('isValidOrEmptyPhone', () => {
  it('accepts empty input (phone is optional)', () => {
    expect(isValidOrEmptyPhone('')).toBe(true);
  });

  it('accepts any input that yields a 10-digit NSN', () => {
    expect(isValidOrEmptyPhone('+7 (900) 123-45-67')).toBe(true);
    expect(isValidOrEmptyPhone('89001234567')).toBe(true);
    expect(isValidOrEmptyPhone('9001234567')).toBe(true);
  });

  it('rejects partial digits', () => {
    expect(isValidOrEmptyPhone('900')).toBe(false);
    expect(isValidOrEmptyPhone('+7 (900) 123')).toBe(false);
  });
});
