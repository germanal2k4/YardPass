import { useState } from 'react';
import {
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Box,
  Alert,
  Chip,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

// NOTE: Backend currently does NOT have an endpoint for fetching scan events
// This is a placeholder implementation that would need the backend to add:
// GET /api/v1/scan-events?limit=20&offset=0
// For now, showing placeholder message

interface ScanEventDisplay {
  id: number;
  scanned_at: string;
  result: 'valid' | 'invalid';
  car_plate?: string;
  reason?: string;
}

export function EventsLog() {
  const [events] = useState<ScanEventDisplay[]>([]);
  const [isLoading] = useState(false);

  // const handleRefresh = () => {
  //   // TODO: Implement when backend endpoint is available
  //   // GET /api/v1/scan-events
  // };

  return (
    <Paper elevation={2} sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">
          Журнал событий
        </Typography>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => {}}
          disabled={isLoading}
        >
          Обновить
        </Button>
      </Box>

      <Alert severity="info" sx={{ mb: 2 }}>
        <strong>Примечание:</strong> Для отображения журнала событий необходимо добавить endpoint в backend:
        <code style={{ display: 'block', marginTop: 8 }}>
          GET /api/v1/scan-events?limit=20&offset=0
        </code>
      </Alert>

      {events.length === 0 ? (
        <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>
          События отсутствуют
        </Typography>
      ) : (
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Время</TableCell>
                <TableCell>Статус</TableCell>
                <TableCell>Номер авто</TableCell>
                <TableCell>Причина</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {events.map((event) => (
                <TableRow key={event.id}>
                  <TableCell>
                    {format(new Date(event.scanned_at), 'dd.MM.yyyy HH:mm:ss', { locale: ru })}
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={event.result === 'valid' ? 'Действителен' : 'Недействителен'}
                      color={event.result === 'valid' ? 'success' : 'error'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{event.car_plate || '—'}</TableCell>
                  <TableCell>{event.reason || '—'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Paper>
  );
}

