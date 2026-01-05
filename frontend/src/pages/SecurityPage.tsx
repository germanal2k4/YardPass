import { useState, useRef, useEffect, KeyboardEvent } from 'react';
import {
  Container,
  Paper,
  TextField,
  Typography,
  Box,
  Alert,
  CircularProgress,
} from '@mui/material';
import { Layout } from '@/shared/ui/Layout';
import { useMutation } from '@tanstack/react-query';
import { passesApi } from '@/shared/api/passes';
import { PassDetailsCard } from '@/features/security/PassDetailsCard';
import { EventsLog } from '@/features/security/EventsLog';
import type { ValidatePassResponse } from '@/shared/types/api';
import { ERROR_MESSAGES } from '@/shared/config/constants';
import { AxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';

export function SecurityPage() {
  const [qrInput, setQrInput] = useState('');
  const [validationResult, setValidationResult] = useState<ValidatePassResponse | null>(null);
  const [errorMsg, setErrorMsg] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const validateMutation = useMutation({
    mutationFn: passesApi.validate,
    onSuccess: (data) => {
      setValidationResult(data);
      setErrorMsg('');
      setQrInput('');
      // Play success or error sound based on result
      playFeedbackSound(data.valid);
      // Return focus to input after a delay
      setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      const errorCode = error.response?.data?.error?.code || 'UNKNOWN_ERROR';
      setErrorMsg(ERROR_MESSAGES[errorCode] || ERROR_MESSAGES.UNKNOWN_ERROR);
      setValidationResult(null);
      setQrInput('');
      playFeedbackSound(false);
      setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
    },
  });

  const playFeedbackSound = (success: boolean) => {
    // Create simple beep sound using Web Audio API
    try {
      const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();

      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);

      oscillator.frequency.value = success ? 800 : 400;
      oscillator.type = 'sine';

      gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.2);

      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.2);
    } catch (e) {
      // Audio not supported, ignore
    }
  };

  const handleKeyPress = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Enter' && qrInput.trim()) {
      e.preventDefault();
      validateMutation.mutate(qrInput.trim());
    }
  };

  useEffect(() => {
    // Auto-focus input on mount and keep focused
    inputRef.current?.focus();
  }, []);

  const handleInputFocus = () => {
    // Clear previous results when focusing input
    setValidationResult(null);
    setErrorMsg('');
  };

  return (
    <Layout title="Сканирование пропусков">
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Paper 
          elevation={4} 
          sx={{ 
            p: 5, 
            mb: 4,
            borderRadius: 3,
            background: 'linear-gradient(135deg, rgba(255, 109, 0, 0.05) 0%, rgba(255, 255, 255, 1) 100%)',
            border: '2px solid',
            borderColor: 'rgba(255, 109, 0, 0.2)',
          }}
        >
          <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
            Проверка QR-кода
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
            Отсканируйте QR-код пропуска или введите код вручную
          </Typography>

          <TextField
            inputRef={inputRef}
            value={qrInput}
            onChange={(e) => setQrInput(e.target.value)}
            onKeyPress={handleKeyPress}
            onFocus={handleInputFocus}
            placeholder="Сканируйте QR-код или введите UUID..."
            fullWidth
            size="large"
            autoComplete="off"
            disabled={validateMutation.isPending}
            InputProps={{
              endAdornment: validateMutation.isPending && <CircularProgress size={28} sx={{ color: '#FF6D00' }} />,
              style: { fontSize: '1.3rem', padding: '20px' },
            }}
            sx={{
              '& .MuiOutlinedInput-root': {
                backgroundColor: '#FFFFFF',
                borderRadius: 2,
                '&:hover fieldset': {
                  borderColor: '#FF6D00',
                },
                '&.Mui-focused fieldset': {
                  borderColor: '#FF6D00',
                  borderWidth: 3,
                },
              },
            }}
          />

          {errorMsg && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {errorMsg}
            </Alert>
          )}
        </Paper>

        {validationResult && (
          <Box sx={{ mb: 4 }}>
            <PassDetailsCard result={validationResult} />
          </Box>
        )}

        <EventsLog />
      </Container>
    </Layout>
  );
}

