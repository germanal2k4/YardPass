import { forwardRef, useLayoutEffect, useRef, useState } from 'react';
import type { ChangeEvent, FocusEvent } from 'react';
import { TextField, InputAdornment } from '@mui/material';
import type { TextFieldProps } from '@mui/material';
import { extractNsnDigits, formatNsnMask } from '@/shared/utils/phone';

type PhoneNsnFieldProps = Omit<
  TextFieldProps,
  'value' | 'onChange' | 'onBlur' | 'inputRef' | 'InputProps' | 'type'
> & {
  /** Raw NSN digits (no country code), e.g. "9001234567". */
  value: string;
  onChange: (nsn: string) => void;
  onBlur?: (e: FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
};

/**
 * Controlled phone input that displays the national part as "(XXX) XXX-XX-XX"
 * with a fixed "+7" adornment. The user types only digits; format characters
 * appear automatically. The caret is preserved across reformat: it always lands
 * after the same digit-index it was on, regardless of how many mask characters
 * shifted around it. Without that, editing mid-string sends the caret to the
 * end on every keystroke (MUI controlled input + length-changing reformat).
 */
export const PhoneNsnField = forwardRef<HTMLInputElement, PhoneNsnFieldProps>(
  function PhoneNsnField({ value, onChange, onBlur, ...rest }, forwardedRef) {
    const innerRef = useRef<HTMLInputElement | null>(null);
    const pendingCaret = useRef<number | null>(null);
    const [display, setDisplay] = useState(() => formatNsnMask(value));

    // Keep display in sync when the source `value` is updated from outside
    // (form reset, edit dialog opens, etc.). We intentionally don't depend on
    // `display` to avoid clobbering caret preservation during in-flight edits.
    const lastExternalValue = useRef(value);
    if (value !== lastExternalValue.current) {
      lastExternalValue.current = value;
      const formatted = formatNsnMask(value);
      if (formatted !== display) setDisplay(formatted);
    }

    useLayoutEffect(() => {
      if (pendingCaret.current === null) return;
      const node = innerRef.current;
      if (node) {
        const pos = Math.min(pendingCaret.current, node.value.length);
        node.setSelectionRange(pos, pos);
      }
      pendingCaret.current = null;
    }, [display]);

    const handleChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const input = e.target;
      const rawNew = input.value;
      const selectionEnd = input.selectionEnd ?? rawNew.length;
      const digitsBeforeCaret = countDigits(rawNew.slice(0, selectionEnd));

      const newNsn = extractNsnDigits(rawNew);
      const newDisplay = formatNsnMask(newNsn);

      // Map "Nth digit in new display" back to a character index so the caret
      // lands right after the digit the user just touched.
      pendingCaret.current = caretIndexAfterNthDigit(newDisplay, digitsBeforeCaret);

      setDisplay(newDisplay);
      onChange(newNsn);
    };

    return (
      <TextField
        {...rest}
        type="tel"
        value={display}
        onChange={handleChange}
        onBlur={onBlur}
        inputRef={(node: HTMLInputElement | null) => {
          innerRef.current = node;
          if (typeof forwardedRef === 'function') forwardedRef(node);
          else if (forwardedRef) forwardedRef.current = node;
        }}
        InputProps={{
          startAdornment: <InputAdornment position="start">+7</InputAdornment>,
        }}
      />
    );
  },
);

function countDigits(s: string): number {
  let n = 0;
  for (const ch of s) if (ch >= '0' && ch <= '9') n += 1;
  return n;
}

function caretIndexAfterNthDigit(s: string, n: number): number {
  if (n <= 0) return 0;
  let seen = 0;
  for (let i = 0; i < s.length; i += 1) {
    const ch = s[i];
    if (ch >= '0' && ch <= '9') {
      seen += 1;
      if (seen === n) return i + 1;
    }
  }
  return s.length;
}
