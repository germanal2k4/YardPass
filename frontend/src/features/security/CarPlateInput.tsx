import { useRef, KeyboardEvent, ChangeEvent } from 'react';
import { Box, Typography, styled } from '@mui/material';

interface CarPlateInputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  disabled?: boolean;
}

const PlateContainer = styled(Box)(() => ({
  display: 'inline-flex',
  alignItems: 'center',
  background: 'linear-gradient(180deg, #FFFFFF 0%, #F5F5F5 100%)',
  border: '4px solid #000000',
  borderRadius: '12px',
  padding: '20px 24px',
  boxShadow: '0 6px 16px rgba(0, 0, 0, 0.4), inset 0 2px 6px rgba(255, 255, 255, 0.9)',
  position: 'relative',
  fontFamily: '"Roboto Mono", monospace',
  userSelect: 'none',
}));

const InputsContainer = styled(Box)(() => ({
  display: 'flex',
  alignItems: 'center',
  gap: '8px',
  marginTop: '8px',
}));

const PlateInput = styled('input')<{ width?: string }>(({ width }) => ({
  border: 'none',
  background: 'transparent',
  outline: 'none',
  fontSize: '3.5rem',
  fontWeight: 'bold',
  textTransform: 'uppercase',
  color: '#000000',
  textAlign: 'center',
  fontFamily: '"Roboto Mono", monospace',
  width: width || '50px',
  padding: '8px 4px',
  caretColor: '#FF6D00',
  letterSpacing: '2px',
  '&::placeholder': {
    color: '#DDDDDD',
  },
  '&:disabled': {
    color: '#999999',
  },
  '&:focus': {
    background: 'rgba(255, 109, 0, 0.05)',
    borderRadius: '4px',
  },
}));

const RegionBadge = styled(Box)(() => ({
  marginLeft: '20px',
  paddingLeft: '20px',
  borderLeft: '3px solid #000000',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  minWidth: '80px',
}));

const FlagImage = styled('img')(() => ({
  width: '50px',
  height: 'auto',
  marginBottom: '6px',
  borderRadius: '2px',
  border: '1px solid #DDDDDD',
}));

const RegionCode = styled(Typography)(() => ({
  fontSize: '1.2rem',
  fontWeight: 'bold',
  color: '#000000',
  fontFamily: '"Roboto Mono", monospace',
  lineHeight: 1.2,
  marginTop: '4px',
}));

const RegionInput = styled(PlateInput)(() => ({
  fontSize: '2.5rem',
  width: '100%',
  marginTop: '4px',
}));

export function CarPlateInput({ value, onChange, onSubmit, disabled }: CarPlateInputProps) {
  const letter1Ref = useRef<HTMLInputElement>(null);
  const digitsRef = useRef<HTMLInputElement>(null);
  const letters2Ref = useRef<HTMLInputElement>(null);
  const regionRef = useRef<HTMLInputElement>(null);

  // Parse value into parts: А123ВС777
  const letter1 = value.slice(0, 1) || '';
  const digits = value.slice(1, 4) || '';
  const letters2 = value.slice(4, 6) || '';
  const region = value.slice(6, 9) || '';

  const updateValue = (newLetter1: string, newDigits: string, newLetters2: string, newRegion: string) => {
    const combined = newLetter1 + newDigits + newLetters2 + newRegion;
    onChange(combined);
  };

  const handleLetter1Change = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value.toUpperCase().replace(/[^A-ZА-Я]/g, '').slice(0, 1);
    updateValue(newValue, digits, letters2, region);
    if (newValue.length === 1) {
      digitsRef.current?.focus();
    }
  };

  const handleDigitsChange = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value.replace(/\D/g, '').slice(0, 3);
    updateValue(letter1, newValue, letters2, region);
    if (newValue.length === 3) {
      letters2Ref.current?.focus();
    }
  };

  const handleLetters2Change = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value.toUpperCase().replace(/[^A-ZА-Я]/g, '').slice(0, 2);
    updateValue(letter1, digits, newValue, region);
    if (newValue.length === 2) {
      regionRef.current?.focus();
    }
  };

  const handleRegionChange = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value.replace(/\D/g, '').slice(0, 3);
    updateValue(letter1, digits, letters2, newValue);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>, currentRef: React.RefObject<HTMLInputElement>) => {
    if (e.key === 'Enter' && value.length >= 6) {
      e.preventDefault();
      onSubmit();
    } else if (e.key === 'Backspace' && currentRef.current?.value === '') {
      e.preventDefault();
      // Move to previous input on backspace
      if (currentRef === digitsRef) {
        letter1Ref.current?.focus();
      } else if (currentRef === letters2Ref) {
        digitsRef.current?.focus();
      } else if (currentRef === regionRef) {
        letters2Ref.current?.focus();
      }
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pastedText = e.clipboardData.getData('text').toUpperCase().replace(/\s/g, '');
    
    // Try to parse А123ВС777 format
    const letter1Match = pastedText.match(/^([A-ZА-Я])/);
    const digitsMatch = pastedText.match(/^[A-ZА-Я](\d{1,3})/);
    const letters2Match = pastedText.match(/^[A-ZА-Я]\d{1,3}([A-ZА-Я]{1,2})/);
    const regionMatch = pastedText.match(/^[A-ZА-Я]\d{1,3}[A-ZА-Я]{1,2}(\d{1,3})/);

    updateValue(
      letter1Match?.[1] || letter1,
      digitsMatch?.[1] || digits,
      letters2Match?.[1] || letters2,
      regionMatch?.[1] || region
    );

    // Focus last filled input
    if (regionMatch) {
      regionRef.current?.focus();
    } else if (letters2Match) {
      letters2Ref.current?.focus();
    } else if (digitsMatch) {
      digitsRef.current?.focus();
    }
  };

  return (
    <Box 
      sx={{ 
        display: 'flex', 
        justifyContent: 'center',
        alignItems: 'center',
        py: 3,
      }}
    >
      <PlateContainer>
        <InputsContainer>
          <PlateInput
            ref={letter1Ref}
            type="text"
            value={letter1}
            onChange={handleLetter1Change}
            onKeyDown={(e) => handleKeyDown(e, letter1Ref)}
            onPaste={handlePaste}
            placeholder="А"
            disabled={disabled}
            autoComplete="off"
            spellCheck={false}
            width="60px"
          />
          <PlateInput
            ref={digitsRef}
            type="text"
            value={digits}
            onChange={handleDigitsChange}
            onKeyDown={(e) => handleKeyDown(e, digitsRef)}
            placeholder="123"
            disabled={disabled}
            autoComplete="off"
            spellCheck={false}
            width="120px"
          />
          <PlateInput
            ref={letters2Ref}
            type="text"
            value={letters2}
            onChange={handleLetters2Change}
            onKeyDown={(e) => handleKeyDown(e, letters2Ref)}
            placeholder="ВС"
            disabled={disabled}
            autoComplete="off"
            spellCheck={false}
            width="90px"
          />
        </InputsContainer>
        <RegionBadge>
          <FlagImage src="/flag.png" alt="RUS" />
          <RegionCode>RUS</RegionCode>
          <RegionInput
            ref={regionRef}
            type="text"
            value={region}
            onChange={handleRegionChange}
            onKeyDown={(e) => handleKeyDown(e, regionRef)}
            placeholder="777"
            disabled={disabled}
            autoComplete="off"
            spellCheck={false}
          />
        </RegionBadge>
      </PlateContainer>
    </Box>
  );
}

