/** Латинские буквы, как на российском госномере (табличка). */
const PLATE_LATIN = new Set([...'ABCEHKMOPTXY']);

/** Кириллица «номерного» ряда → латиница с таблички (верхний и нижний регистр). */
const CYRILLIC_TO_LATIN: Record<string, string> = {
  А: 'A',
  а: 'A',
  В: 'B',
  в: 'B',
  С: 'C',
  с: 'C',
  Е: 'E',
  е: 'E',
  К: 'K',
  к: 'K',
  М: 'M',
  м: 'M',
  Н: 'H',
  н: 'H',
  О: 'O',
  о: 'O',
  Р: 'P',
  р: 'P',
  Т: 'T',
  т: 'T',
  У: 'Y',
  у: 'Y',
  Х: 'X',
  х: 'X',
};

/** Нормализация как на бэкенде: только допустимые буквы и цифры, иначе ''. */
export function normalizeCarPlate(plate: string): string {
  const trimmed = plate.trim();
  if (!trimmed) return '';

  let out = '';
  for (const ch of trimmed) {
    if (ch === ' ' || ch === '-' || ch === '.' || ch === '\u00a0' || ch === '_') {
      continue;
    }
    const mapped = CYRILLIC_TO_LATIN[ch];
    if (mapped) {
      out += mapped;
      continue;
    }
    const u = ch.toUpperCase();
    if (u >= '0' && u <= '9') {
      out += u;
      continue;
    }
    if (u.length === 1 && u >= 'A' && u <= 'Z') {
      if (!PLATE_LATIN.has(u)) {
        return '';
      }
      out += u;
      continue;
    }
    return '';
  }
  return out;
}

/** Одна допустимая «номерная» буква (любая раскладка) → латиница или ''. */
export function mapPlateLetterChar(ch: string): string {
  if (!ch) return '';
  const last = ch.slice(-1);
  const mapped = CYRILLIC_TO_LATIN[last];
  if (mapped) return mapped;
  const u = last.toUpperCase();
  if (u.length === 1 && u >= 'A' && u <= 'Z' && PLATE_LATIN.has(u)) return u;
  return '';
}

/** Отфильтровать сегмент до maxLen допустимых букв номера (уже в латинице). */
export function filterPlateLetterSegment(raw: string, maxLen: number): string {
  let out = '';
  for (const ch of raw) {
    const m = mapPlateLetterChar(ch);
    if (m) out += m;
    if (out.length >= maxLen) break;
  }
  return out;
}
