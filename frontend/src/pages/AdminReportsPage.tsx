import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Container,
  Paper,
  Typography,
  Alert,
  Box,
  Grid,
  Card,
  CardContent,
  Button,
  TextField,
  Stack,
  Chip,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
} from '@mui/material';
import { Layout } from '@/shared/ui/Layout';
import DownloadIcon from '@mui/icons-material/Download';
import RefreshIcon from '@mui/icons-material/Refresh';
import AssessmentIcon from '@mui/icons-material/Assessment';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';
import { reportsApi } from '@/shared/api/reports';
import type { ScanEventWithDetails } from '@/shared/types/api';

export function AdminReportsPage() {
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(20);
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [resultFilter, setResultFilter] = useState<'valid' | 'invalid' | ''>('');
  const [exportError, setExportError] = useState('');
  const [isExporting, setIsExporting] = useState(false);

  // Fetch statistics
  const { data: stats, isLoading: statsLoading, refetch: refetchStats } = useQuery({
    queryKey: ['statistics', dateFrom, dateTo],
    queryFn: () => {
      // Преобразуем datetime-local формат в RFC3339 (ISO 8601)
      const fromRFC3339 = dateFrom ? new Date(dateFrom).toISOString() : undefined;
      const toRFC3339 = dateTo ? new Date(dateTo).toISOString() : undefined;
      
      return reportsApi.getStatistics({
        from: fromRFC3339,
        to: toRFC3339,
      });
    },
  });

  // Fetch scan events
  const { data: eventsData, isLoading: eventsLoading, isError, refetch: refetchEvents } = useQuery({
    queryKey: ['scanEvents', page, rowsPerPage, dateFrom, dateTo, resultFilter],
    queryFn: () => {
      // Преобразуем datetime-local формат в RFC3339 (ISO 8601)
      const fromRFC3339 = dateFrom ? new Date(dateFrom).toISOString() : undefined;
      const toRFC3339 = dateTo ? new Date(dateTo).toISOString() : undefined;
      
      return reportsApi.getScanEvents({
        limit: rowsPerPage,
        offset: page * rowsPerPage,
        from: fromRFC3339,
        to: toRFC3339,
        result: resultFilter || undefined,
      });
    },
  });

  const handleRefresh = () => {
    // При обновлении возвращаемся на первую страницу
    setPage(0);
    refetchStats();
    refetchEvents();
  };

  const handleExport = async () => {
    setIsExporting(true);
    setExportError('');
    
    try {
      // Преобразуем datetime-local формат в RFC3339 (ISO 8601)
      const fromRFC3339 = dateFrom ? new Date(dateFrom).toISOString() : undefined;
      const toRFC3339 = dateTo ? new Date(dateTo).toISOString() : undefined;
      
      const blob = await reportsApi.exportReport({
        format: 'xlsx',
        from: fromRFC3339,
        to: toRFC3339,
      });

      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `report_${format(new Date(), 'yyyy-MM-dd_HH-mm')}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Error exporting report:', error);
      setExportError('Ошибка при экспорте отчета. Попробуйте еще раз.');
    } finally {
      setIsExporting(false);
    }
  };

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleDateFromChange = (value: string) => {
    setDateFrom(value);
    setPage(0); // Сбрасываем на первую страницу при изменении фильтра
  };

  const handleDateToChange = (value: string) => {
    setDateTo(value);
    setPage(0); // Сбрасываем на первую страницу при изменении фильтра
  };

  const handleResultFilterChange = (value: 'valid' | 'invalid' | '') => {
    setResultFilter(value);
    setPage(0); // Сбрасываем на первую страницу при изменении фильтра
  };

  const events = eventsData?.events || [];

  // Если на текущей странице нет событий и это не первая страница,
  // автоматически вернуться на первую страницу
  useEffect(() => {
    if (!eventsLoading && events.length === 0 && page > 0) {
      setPage(0);
    }
  }, [events.length, eventsLoading, page]);

  // Определяем, есть ли еще страницы (если событий меньше чем rowsPerPage - это последняя страница)
  const hasMorePages = events.length >= rowsPerPage;
  // Для TablePagination: если есть еще страницы, указываем -1, иначе точное количество
  const count = hasMorePages ? -1 : page * rowsPerPage + events.length;

  return (
    <Layout title="Отчеты и статистика">
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {/* Filters */}
        <Paper elevation={2} sx={{ p: 3, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Фильтры
          </Typography>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems="center">
            <TextField
              label="Дата от"
              type="datetime-local"
              value={dateFrom}
              onChange={(e) => handleDateFromChange(e.target.value)}
              InputLabelProps={{ shrink: true }}
              fullWidth
            />
            <TextField
              label="Дата до"
              type="datetime-local"
              value={dateTo}
              onChange={(e) => handleDateToChange(e.target.value)}
              InputLabelProps={{ shrink: true }}
              fullWidth
            />
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              <Chip
                label="Все"
                onClick={() => handleResultFilterChange('')}
                color={resultFilter === '' ? 'primary' : 'default'}
                clickable
              />
              <Chip
                label="Действительные"
                onClick={() => handleResultFilterChange('valid')}
                color={resultFilter === 'valid' ? 'success' : 'default'}
                clickable
              />
              <Chip
                label="Недействительные"
                onClick={() => handleResultFilterChange('invalid')}
                color={resultFilter === 'invalid' ? 'error' : 'default'}
                clickable
              />
            </Box>
            <Box sx={{ display: 'flex', gap: 1, flexShrink: 0 }}>
              <Button
                variant="outlined"
                startIcon={<RefreshIcon />}
                onClick={handleRefresh}
                disabled={statsLoading || eventsLoading}
                sx={{ whiteSpace: 'nowrap' }}
              >
                Обновить
              </Button>
              <Button
                variant="contained"
                startIcon={<DownloadIcon />}
                onClick={handleExport}
                color="success"
                disabled={isExporting}
                sx={{ whiteSpace: 'nowrap' }}
              >
                {isExporting ? 'Экспортируется...' : 'Экспорт'}
              </Button>
            </Box>
          </Stack>
          {exportError && (
            <Alert severity="error" sx={{ mt: 2 }} onClose={() => setExportError('')}>
              {exportError}
            </Alert>
          )}
        </Paper>

        {/* Statistics Cards */}
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <AssessmentIcon sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="subtitle2" color="text.secondary">
                    Всего сканирований
                  </Typography>
                </Box>
                {statsLoading ? (
                  <CircularProgress size={24} />
                ) : (
                  <Typography variant="h4">{stats?.total_scans || 0}</Typography>
                )}
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <CheckCircleIcon sx={{ mr: 1, color: 'success.main' }} />
                  <Typography variant="subtitle2" color="text.secondary">
                    Действительных
                  </Typography>
                </Box>
                {statsLoading ? (
                  <CircularProgress size={24} />
                ) : (
                  <Typography variant="h4" color="success.main">
                    {stats?.valid_scans || 0}
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <CancelIcon sx={{ mr: 1, color: 'error.main' }} />
                  <Typography variant="subtitle2" color="text.secondary">
                    Недействительных
                  </Typography>
                </Box>
                {statsLoading ? (
                  <CircularProgress size={24} />
                ) : (
                  <Typography variant="h4" color="error.main">
                    {stats?.invalid_scans || 0}
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <AssessmentIcon sx={{ mr: 1, color: 'info.main' }} />
                  <Typography variant="subtitle2" color="text.secondary">
                    Уникальных пропусков
                  </Typography>
                </Box>
                {statsLoading ? (
                  <CircularProgress size={24} />
                ) : (
                  <Typography variant="h4">{stats?.unique_passes || 0}</Typography>
                )}
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Top Reasons */}
        {stats?.top_reasons && stats.top_reasons.length > 0 && (
          <Paper elevation={2} sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>
              Основные причины отказов
            </Typography>
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              {stats.top_reasons.map((item, idx) => (
                <Chip
                  key={idx}
                  label={`${item.reason}: ${item.count}`}
                  color="warning"
                  variant="outlined"
                />
              ))}
            </Stack>
          </Paper>
        )}

        {/* Events Table */}
        <Paper elevation={2} sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Журнал событий
          </Typography>

          {isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Ошибка при загрузке данных
            </Alert>
          )}

          {eventsLoading ? (
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
                            {event.ApartmentNumber
                              ? `${event.BuildingName || ''} № ${event.ApartmentNumber}`
                              : '—'}
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
      </Container>
    </Layout>
  );
}


