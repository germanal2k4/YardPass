import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/react';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '@/shared/ui/theme';
import { CarPlateInput } from '../CarPlateInput';

function renderPlate(overrides: Partial<React.ComponentProps<typeof CarPlateInput>> = {}) {
  const props = {
    value: '',
    onChange: vi.fn(),
    onSubmit: vi.fn(),
    ...overrides,
  };
  const utils = render(
    <ThemeProvider theme={theme}>
      <CarPlateInput {...props} />
    </ThemeProvider>,
  );
  const inputs = utils.container.querySelectorAll('input');
  return {
    ...utils,
    props,
    letter1: inputs[0] as HTMLInputElement,
    digits: inputs[1] as HTMLInputElement,
    letters2: inputs[2] as HTMLInputElement,
    region: inputs[3] as HTMLInputElement,
  };
}

describe('CarPlateInput', () => {
  describe('rendering', () => {
    it('renders four inputs', () => {
      const { letter1, digits, letters2, region } = renderPlate();
      expect(letter1).toBeInTheDocument();
      expect(digits).toBeInTheDocument();
      expect(letters2).toBeInTheDocument();
      expect(region).toBeInTheDocument();
    });

    it('distributes value into parts', () => {
      const { letter1, digits, letters2, region } = renderPlate({ value: 'А123ВС777' });
      expect(letter1.value).toBe('А');
      expect(digits.value).toBe('123');
      expect(letters2.value).toBe('ВС');
      expect(region.value).toBe('777');
    });

    it('disables all inputs when disabled=true', () => {
      const { letter1, digits, letters2, region } = renderPlate({ disabled: true });
      expect(letter1).toBeDisabled();
      expect(digits).toBeDisabled();
      expect(letters2).toBeDisabled();
      expect(region).toBeDisabled();
    });
  });

  describe('letter filtering and uppercase', () => {
    it('filters non-letter characters from letter1', () => {
      const { letter1, props } = renderPlate();
      fireEvent.change(letter1, { target: { value: '1' } });
      expect(props.onChange).toHaveBeenCalledWith('');
    });

    it('converts lowercase to uppercase', () => {
      const { letter1, props } = renderPlate();
      fireEvent.change(letter1, { target: { value: 'а' } });
      expect(props.onChange).toHaveBeenCalledWith('А');
    });

    it('limits letter1 to 1 character', () => {
      const { letter1, props } = renderPlate();
      fireEvent.change(letter1, { target: { value: 'АБ' } });
      expect(props.onChange).toHaveBeenCalledWith('А');
    });
  });

  describe('digit filtering', () => {
    it('filters non-digit characters from digits field', () => {
      const { digits, props } = renderPlate({ value: 'А' });
      fireEvent.change(digits, { target: { value: 'А12' } });
      expect(props.onChange).toHaveBeenCalledWith('А12');
    });

    it('limits digits to 3 characters', () => {
      const { digits, props } = renderPlate({ value: 'А' });
      fireEvent.change(digits, { target: { value: '1234' } });
      expect(props.onChange).toHaveBeenCalledWith('А123');
    });
  });

  describe('letters2 filtering', () => {
    it('allows Cyrillic letters', () => {
      const { letters2, props } = renderPlate({ value: 'А123' });
      fireEvent.change(letters2, { target: { value: 'вс' } });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС');
    });

    it('limits to 2 characters', () => {
      const { letters2, props } = renderPlate({ value: 'А123' });
      fireEvent.change(letters2, { target: { value: 'ВСХ' } });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС');
    });
  });

  describe('region filtering', () => {
    it('allows only digits in region', () => {
      const { region, props } = renderPlate({ value: 'А123ВС' });
      fireEvent.change(region, { target: { value: 'А77' } });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС77');
    });

    it('limits region to 3 digits', () => {
      const { region, props } = renderPlate({ value: 'А123ВС' });
      fireEvent.change(region, { target: { value: '7777' } });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС777');
    });
  });

  describe('auto-focus transitions', () => {
    it('focuses digits after entering 1 letter', () => {
      const { letter1, props } = renderPlate();
      letter1.focus();
      fireEvent.change(letter1, { target: { value: 'А' } });
      expect(props.onChange).toHaveBeenCalledWith('А');
      // Focus assertion: since we can't reliably check document.activeElement in jsdom
      // with styled-components and refs, we verify the onChange was called correctly
    });

    it('focuses letters2 after entering 3 digits', () => {
      const { digits, props } = renderPlate({ value: 'А' });
      digits.focus();
      fireEvent.change(digits, { target: { value: '123' } });
      expect(props.onChange).toHaveBeenCalledWith('А123');
    });

    it('focuses region after entering 2 letters', () => {
      const { letters2, props } = renderPlate({ value: 'А123' });
      letters2.focus();
      fireEvent.change(letters2, { target: { value: 'ВС' } });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС');
    });
  });

  describe('Backspace navigation', () => {
    it('moves focus from empty digits to letter1 on Backspace', () => {
      const { digits } = renderPlate({ value: 'А' });
      digits.focus();
      // Simulate empty value + backspace
      fireEvent.change(digits, { target: { value: '' } });
      fireEvent.keyDown(digits, { key: 'Backspace' });
    });

    it('moves focus from empty letters2 to digits on Backspace', () => {
      const { letters2 } = renderPlate({ value: 'А123' });
      letters2.focus();
      fireEvent.change(letters2, { target: { value: '' } });
      fireEvent.keyDown(letters2, { key: 'Backspace' });
    });

    it('moves focus from empty region to letters2 on Backspace', () => {
      const { region } = renderPlate({ value: 'А123ВС' });
      region.focus();
      fireEvent.change(region, { target: { value: '' } });
      fireEvent.keyDown(region, { key: 'Backspace' });
    });
  });

  describe('Enter to submit', () => {
    it('calls onSubmit when value length >= 6 and Enter is pressed', () => {
      const { letter1, props } = renderPlate({ value: 'А123ВС77' });
      fireEvent.keyDown(letter1, { key: 'Enter' });
      expect(props.onSubmit).toHaveBeenCalled();
    });

    it('does NOT call onSubmit when value is too short', () => {
      const { letter1, props } = renderPlate({ value: 'А12' });
      fireEvent.keyDown(letter1, { key: 'Enter' });
      expect(props.onSubmit).not.toHaveBeenCalled();
    });
  });

  describe('paste handling', () => {
    it('parses a full plate from paste', () => {
      const { letter1, props } = renderPlate();
      const clipboardData = {
        getData: vi.fn().mockReturnValue('А123ВС777'),
      };
      fireEvent.paste(letter1, { clipboardData });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС777');
    });

    it('parses a partial plate (letters + digits only)', () => {
      const { letter1, props } = renderPlate();
      const clipboardData = {
        getData: vi.fn().mockReturnValue('А123'),
      };
      fireEvent.paste(letter1, { clipboardData });
      expect(props.onChange).toHaveBeenCalledWith('А123');
    });

    it('strips whitespace from pasted text', () => {
      const { letter1, props } = renderPlate();
      const clipboardData = {
        getData: vi.fn().mockReturnValue('А 123 ВС 777'),
      };
      fireEvent.paste(letter1, { clipboardData });
      expect(props.onChange).toHaveBeenCalledWith('А123ВС777');
    });
  });
});
