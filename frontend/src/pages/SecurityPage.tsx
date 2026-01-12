import { useState, useRef, useEffect, KeyboardEvent } from 'react';
import {
  Container,
  Paper,
  TextField,
  Typography,
  Box,
  Alert,
  CircularProgress,
  IconButton,
  Grid,
  Button,
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import QrCodeScannerIcon from '@mui/icons-material/QrCodeScanner';
import DirectionsCarIcon from '@mui/icons-material/DirectionsCar';
import { Layout } from '@/shared/ui/Layout';
import { useMutation } from '@tanstack/react-query';
import { passesApi } from '@/shared/api/passes';
import { PassDetailsCard } from '@/features/security/PassDetailsCard';
import { EventsLog } from '@/features/security/EventsLog';
import { CarPlateInput } from '@/features/security/CarPlateInput';
import type { ValidatePassResponse } from '@/shared/types/api';
import { formatErrorMessage } from '@/shared/utils/errors';
import { AxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';

export function SecurityPage() {
  const [qrInput, setQrInput] = useState('');
  const [carPlateInput, setCarPlateInput] = useState('');
  const [validationResult, setValidationResult] = useState<ValidatePassResponse | null>(null);
  const [errorMsg, setErrorMsg] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const validateMutation = useMutation({
    mutationFn: (params: { qr_uuid?: string; car_plate?: string }) => passesApi.validate(params),
    onSuccess: (data) => {
      setValidationResult(data);
      setErrorMsg('');
      setQrInput('');
      setCarPlateInput('');
      // Play success or error sound based on result
      playFeedbackSound(data.valid);
      // Return focus to input after a delay
      setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      setErrorMsg(formatErrorMessage(error));
      setValidationResult(null);
      setQrInput('');
      setCarPlateInput('');
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
      // Clear previous error when starting new scan
      setErrorMsg('');
      setValidationResult(null);
      validateMutation.mutate({ qr_uuid: qrInput.trim() });
    }
  };

  const handleCarPlateSubmit = () => {
    if (carPlateInput.trim()) {
      setErrorMsg('');
      setValidationResult(null);
      validateMutation.mutate({ car_plate: carPlateInput.trim() });
    }
  };

  useEffect(() => {
    // Auto-focus input on mount and keep focused
    inputRef.current?.focus();
  }, []);

  const handleCloseError = () => {
    // Clear error when user explicitly closes it
    setErrorMsg('');
    inputRef.current?.focus();
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQrInput(e.target.value);
    // Only clear results when user starts typing (not just focusing)
    if (e.target.value.length > 0 && validationResult) {
      setValidationResult(null);
    }
  };

  return (
    <Layout title="Сканирование пропусков">
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Grid container spacing={3} sx={{ mb: 4 }}>
          {/* QR Code Section */}
          <Grid item xs={12} md={6}>
            <Paper 
              elevation={4} 
              sx={{ 
                p: 4, 
                height: '100%',
                borderRadius: 3,
                background: 'linear-gradient(135deg, rgba(255, 109, 0, 0.05) 0%, rgba(255, 255, 255, 1) 100%)',
                border: '2px solid',
                borderColor: 'rgba(255, 109, 0, 0.2)',
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <QrCodeScannerIcon sx={{ fontSize: 40, mr: 2, color: '#FF6D00' }} />
                <Typography variant="h5" fontWeight="700" color="#263238">
                  Проверка QR-кода
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Отсканируйте QR-код пропуска или введите UUID вручную
              </Typography>

              <TextField
            inputRef={inputRef}
            value={qrInput}
            onChange={handleInputChange}
            onKeyPress={handleKeyPress}
            placeholder="Сканируйте QR-код или введите UUID..."
            fullWidth
            size="medium"
            autoComplete="off"
            disabled={validateMutation.isPending}
            InputProps={{
              endAdornment: validateMutation.isPending && <CircularProgress size={24} sx={{ color: '#FF6D00' }} />,
              style: { fontSize: '1.1rem', padding: '16px' },
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
            </Paper>
          </Grid>

          {/* Car Plate Section */}
          <Grid item xs={12} md={6}>
            <Paper 
              elevation={4} 
              sx={{ 
                p: 4, 
                height: '100%',
                borderRadius: 3,
                background: 'linear-gradient(135deg, rgba(33, 150, 243, 0.05) 0%, rgba(255, 255, 255, 1) 100%)',
                border: '2px solid',
                borderColor: 'rgba(33, 150, 243, 0.2)',
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <DirectionsCarIcon sx={{ fontSize: 40, mr: 2, color: '#2196F3' }} />
                <Typography variant="h5" fontWeight="700" color="#263238">
                  Проверка по номеру
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Введите государственный номер автомобиля
              </Typography>

              <CarPlateInput
                value={carPlateInput}
                onChange={setCarPlateInput}
                onSubmit={handleCarPlateSubmit}
                disabled={validateMutation.isPending}
              />

              <Button
                variant="contained"
                fullWidth
                size="large"
                onClick={handleCarPlateSubmit}
                disabled={!carPlateInput.trim() || validateMutation.isPending}
                sx={{
                  mt: 2,
                  py: 2,
                  fontSize: '1.1rem',
                  fontWeight: 600,
                  backgroundColor: '#2196F3',
                  '&:hover': {
                    backgroundColor: '#1976D2',
                  },
                  '&:disabled': {
                    backgroundColor: '#CCCCCC',
                  },
                }}
                startIcon={validateMutation.isPending ? <CircularProgress size={20} color="inherit" /> : <DirectionsCarIcon />}
              >
                {validateMutation.isPending ? 'Проверка...' : 'Проверить номер'}
              </Button>
            </Paper>
          </Grid>
        </Grid>

        {/* Error Message */}
        {errorMsg && (
          <Alert 
            severity="error" 
            sx={{ 
              mb: 3,
              fontSize: '1.1rem',
              fontWeight: 600,
              animation: 'shake 0.5s',
              '@keyframes shake': {
                '0%, 100%': { transform: 'translateX(0)' },
                '10%, 30%, 50%, 70%, 90%': { transform: 'translateX(-5px)' },
                '20%, 40%, 60%, 80%': { transform: 'translateX(5px)' },
              },
            }}
            action={
              <IconButton
                aria-label="close"
                color="inherit"
                size="small"
                onClick={handleCloseError}
                sx={{
                  '&:hover': {
                    backgroundColor: 'rgba(0, 0, 0, 0.1)',
                  },
                }}
              >
                <CloseIcon fontSize="inherit" />
              </IconButton>
            }
          >
            {errorMsg}
          </Alert>
        )}

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

