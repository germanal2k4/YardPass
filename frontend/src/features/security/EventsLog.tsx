import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
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
  Chip,
  CircularProgress,
  Alert,
  TablePagination,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';
import { reportsApi } from '@/shared/api/reports';
import type { ScanEventWithDetails } from '@/shared/types/api';

export function EventsLog() {
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(20);

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['scanEvents', page, rowsPerPage],
    queryFn: () =>
      reportsApi.getScanEvents({
        limit: rowsPerPage,
        offset: page * rowsPerPage,
      }),
  });

  const events = data?.events || [];

  // Если на текущей странице нет событий и это не первая страница,
  // автоматически вернуться на первую страницу
  useEffect(() => {
    if (!isLoading && events.length === 0 && page > 0) {
      setPage(0);
    }
  }, [events.length, isLoading, page]);

  const handleRefresh = () => {
    // При обновлении возвращаемся на первую страницу
    setPage(0);
    refetch();
  };

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  // Определяем, есть ли еще страницы (если событий меньше чем rowsPerPage - это последняя страница)
  const hasMorePages = events.length >= rowsPerPage;
  // Для TablePagination: если есть еще страницы, указываем -1, иначе точное количество
  const count = hasMorePages ? -1 : page * rowsPerPage + events.length;

  return (
    <Paper elevation={2} sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">Журнал событий</Typography>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={handleRefresh}
          disabled={isLoading}
        >
          Обновить
        </Button>
      </Box>

      {isError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Ошибка при загрузке данных: {error instanceof Error ? error.message : 'Неизвестная ошибка'}
        </Alert>
      )}

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </Box>
      ) : events.length === 0 && page === 0 ? (
        <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>
          События отсутствуют
        </Typography>
      ) : events.length > 0 ? (
        <>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Время</TableCell>
                  <TableCell>Статус</TableCell>
                  <TableCell>Номер авто</TableCell>
                  <TableCell>Гость</TableCell>
                  <TableCell>Квартира</TableCell>
                  <TableCell>Охранник</TableCell>
                  <TableCell>Причина</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {events.map((event: ScanEventWithDetails) => {
                  const scannedDate = event.ScannedAt ? new Date(event.ScannedAt) : null;
                  const isValidDate = scannedDate && !isNaN(scannedDate.getTime());
                  
                  return (
                    <TableRow key={event.ID}>
                      <TableCell>
                        {isValidDate
                          ? format(scannedDate, 'dd.MM.yyyy HH:mm:ss', { locale: ru })
                          : '—'}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={event.Result === 'valid' ? 'Действителен' : 'Недействителен'}
                          color={event.Result === 'valid' ? 'success' : 'error'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{event.CarPlate || '—'}</TableCell>
                      <TableCell>{event.GuestName || '—'}</TableCell>
                      <TableCell>
                        {event.ApartmentNumber ? `${event.BuildingName || ''} № ${event.ApartmentNumber}` : '—'}
                      </TableCell>
                      <TableCell>{event.GuardUsername || '—'}</TableCell>
                      <TableCell>{event.Reason || '—'}</TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
          <TablePagination
            component="div"
            count={count}
            page={page}
            onPageChange={handleChangePage}
            rowsPerPage={rowsPerPage}
            onRowsPerPageChange={handleChangeRowsPerPage}
            rowsPerPageOptions={[10, 20, 50, 100]}
            labelRowsPerPage="Строк на странице:"
            labelDisplayedRows={({ from, to }) => `${from}–${to}`}
          />
        </>
      ) : null}
    </Paper>
  );
}


