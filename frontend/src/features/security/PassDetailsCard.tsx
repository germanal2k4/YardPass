import { Paper, Typography, Box, Chip, Alert } from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';
import type { ValidatePassResponse } from '@/shared/types/api';
import { ERROR_MESSAGES } from '@/shared/config/constants';

interface PassDetailsCardProps {
  result: ValidatePassResponse;
}

export function PassDetailsCard({ result }: PassDetailsCardProps) {
  const isValid = result.valid;

  return (
    <Paper
      elevation={3}
      sx={{
        p: 4,
        backgroundColor: isValid ? 'success.light' : 'error.light',
        border: 3,
        borderColor: isValid ? 'success.main' : 'error.main',
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        {isValid ? (
          <CheckCircleIcon sx={{ fontSize: 60, color: 'success.dark', mr: 2 }} />
        ) : (
          <ErrorIcon sx={{ fontSize: 60, color: 'error.dark', mr: 2 }} />
        )}
        <Box>
          <Typography variant="h4" fontWeight="bold">
            {isValid ? 'Пропуск действителен' : 'Пропуск недействителен'}
          </Typography>
          {!isValid && result.reason && (
            <Typography variant="body1" color="error.dark" sx={{ mt: 1 }}>
              {ERROR_MESSAGES[result.reason] || result.reason}
            </Typography>
          )}
        </Box>
      </Box>

      {isValid && (
        <Box sx={{ mt: 3 }}>
          <Box sx={{ mb: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Номер автомобиля
            </Typography>
            <Typography variant="h5" fontWeight="bold">
              {result.car_plate}
            </Typography>
          </Box>

          <Box sx={{ mb: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Квартира
            </Typography>
            <Typography variant="h6">
              {result.apartment}
            </Typography>
          </Box>

          {result.valid_to && (
            <Box>
              <Typography variant="body2" color="text.secondary">
                Действителен до
              </Typography>
              <Typography variant="h6">
                {format(new Date(result.valid_to), 'dd MMMM yyyy, HH:mm', { locale: ru })}
              </Typography>
            </Box>
          )}
        </Box>
      )}
    </Paper>
  );
}

