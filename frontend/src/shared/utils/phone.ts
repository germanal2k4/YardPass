// Phone utilities mirror the backend NormalizeResidentPhone (service/phone.go):
// canonical form is +7XXXXXXXXXX. The UI accepts only digits and renders them
// using the visual mask "+7 (XXX) XXX-XX-XX".

const MAX_NSN_DIGITS = 10;

/** Strip every non-digit character from input. */
function digitsOnly(value: string): string {
  let out = '';
  for (const ch of value) {
    if (ch >= '0' && ch <= '9') out += ch;
  }
  return out;
}

/**
 * Take any user-entered string and return up to 10 national-subscriber digits
 * (without the country code). Handles +7-, 7-, 8- prefixes and the bare 10-digit form.
 *
 * Russian mobile NSNs always start with 9, which we use to disambiguate a partial
 * digit string like "7900" — extracted from the live mask "+7 (900..." — from
 * an NSN that legitimately starts with 7. Without that rule the formatter is
 * not idempotent and re-rendering shifts the country-code digit into the NSN.
 */
export function extractNsnDigits(raw: string): string {
  let d = digitsOnly(raw);
  if (!d) return '';
  if (d.length > 11) d = d.slice(0, 11);
  if ((d[0] === '7' || d[0] === '8') && (d.length === 11 || (d.length > 1 && d[1] === '9'))) {
    d = d.slice(1);
  }
  if (d.length > MAX_NSN_DIGITS) d = d.slice(0, MAX_NSN_DIGITS);
  return d;
}

/** Render a partial NSN string as the mask "+7 (XXX) XXX-XX-XX". */
export function formatPhoneInput(raw: string): string {
  const d = extractNsnDigits(raw);
  if (d.length === 0) return '';
  const inner = formatNsnMask(d);
  return inner ? '+7 ' + inner : '+7';
}

/**
 * Render up to 10 NSN digits as the inner mask "(XXX) XXX-XX-XX" (without the
 * +7 country code). The +7 is rendered as a separate input adornment so the
 * user only ever interacts with the national part.
 */
export function formatNsnMask(raw: string): string {
  const d = extractNsnDigits(raw);
  if (d.length === 0) return '';
  let out = '(' + d.slice(0, 3);
  if (d.length >= 3) out += ')';
  if (d.length > 3) out += ' ' + d.slice(3, 6);
  if (d.length > 6) out += '-' + d.slice(6, 8);
  if (d.length > 8) out += '-' + d.slice(8, 10);
  return out;
}

/** Canonical form sent to the API: "+7XXXXXXXXXX" or '' if not yet complete. */
export function canonicalPhone(raw: string): string {
  const d = extractNsnDigits(raw);
  if (d.length !== MAX_NSN_DIGITS) return '';
  return '+7' + d;
}

/** True if the input is empty or yields a complete 10-digit NSN. */
export function isValidOrEmptyPhone(raw: string): boolean {
  const d = digitsOnly(raw);
  if (d.length === 0) return true;
  return extractNsnDigits(raw).length === MAX_NSN_DIGITS;
}
