import { useState, useRef } from 'react';
import {
  Container,
  Paper,
  Typography,
  TextField,
  Button,
  Box,
  Alert,
  CircularProgress,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Divider,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import { Layout } from '@/shared/ui/Layout';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { residentsApi } from '@/shared/api/residents';
import { useAuth } from '@/features/auth/useAuth';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { CreateResidentRequest, Resident } from '@/shared/types/api';
import { AxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';
import { formatErrorMessage, formatBulkError } from '@/shared/utils/errors';
import RefreshIcon from '@mui/icons-material/Refresh';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import DeleteIcon from '@mui/icons-material/Delete';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

// Regex для валидации телефонных номеров (российский формат)
// Поддерживает форматы: +7XXXXXXXXXX, 8XXXXXXXXXX, +7 (XXX) XXX-XX-XX и т.д.
const PHONE_REGEX = /^(\+7|7|8)?[\s\-]?\(?[489][0-9]{2}\)?[\s\-]?[0-9]{3}[\s\-]?[0-9]{2}[\s\-]?[0-9]{2}$/;

const residentSchema = z.object({
  apartment_id: z.union([
    z.string().min(1, 'ID квартиры обязателен').transform((val) => parseInt(val, 10)),
    z.number().int().min(1, 'ID квартиры обязателен'),
  ]),
  telegram_id: z.union([
    z.string().min(1, 'Telegram ID обязателен').transform((val) => parseInt(val, 10)),
    z.number().int().min(1, 'Telegram ID обязателен'),
  ]),
  name: z.string().optional(),
  phone: z.string()
    .optional()
    .refine(
      (val) => !val || val.trim() === '' || PHONE_REGEX.test(val),
      {
        message: 'Неверный формат телефона. Используйте формат: +7 (XXX) XXX-XX-XX или 8XXXXXXXXXX',
      }
    ),
});

type ResidentFormData = z.infer<typeof residentSchema>;

export function AdminResidentsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const [statusFilter, setStatusFilter] = useState<'active' | 'inactive' | 'all'>('all');
  
  // Bulk import state
  const [bulkDialogOpen, setBulkDialogOpen] = useState(false);
  const [bulkJson, setBulkJson] = useState('');
  const [bulkResult, setBulkResult] = useState<{ created: number; errors: any[] } | null>(null);
  
  // CSV import state
  const [csvDialogOpen, setCsvDialogOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [csvResult, setCsvResult] = useState<{ imported: number; errors: any[] } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Delete state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [residentToDelete, setResidentToDelete] = useState<Resident | null>(null);
  const [deleteError, setDeleteError] = useState('');

  // Get building_id from user
  const buildingId = user?.building_id;

  // Fetch residents
  const { data: residents, isLoading, error } = useQuery({
    queryKey: ['residents', statusFilter],
    queryFn: () => residentsApi.getAll({
      status: statusFilter === 'all' ? undefined : statusFilter,
    }),
  });

  // Form
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<ResidentFormData>({
    resolver: zodResolver(residentSchema),
    defaultValues: {
      apartment_id: '' as any, // Пустая строка для начального состояния
      telegram_id: '' as any,
      name: '',
      phone: '',
    },
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateResidentRequest) => residentsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['residents'] });
      setSuccessMsg('Житель успешно создан');
      setErrorMsg('');
      reset();
      setTimeout(() => setSuccessMsg(''), 3000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      setErrorMsg(formatErrorMessage(error));
      setSuccessMsg('');
    },
  });

  // Bulk create mutation
  const bulkCreateMutation = useMutation({
    mutationFn: (data: CreateResidentRequest[]) => residentsApi.createBulk(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['residents'] });
      setBulkResult(result);
      setSuccessMsg(`Массовое создание завершено: создано ${result.created} жителей`);
      setErrorMsg('');
      setTimeout(() => setSuccessMsg(''), 5000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      setErrorMsg(formatErrorMessage(error));
      setSuccessMsg('');
      setBulkResult(null);
    },
  });

  // CSV import mutation
  const csvImportMutation = useMutation({
    mutationFn: (params: { file: File; buildingId: number }) => 
      residentsApi.importFromCSV(params.file, params.buildingId),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['residents'] });
      setCsvResult(result);
      setSuccessMsg(`Импорт завершен: импортировано ${result.imported} жителей`);
      setErrorMsg('');
      setSelectedFile(null);
      setTimeout(() => setSuccessMsg(''), 5000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      setErrorMsg(formatErrorMessage(error));
      setSuccessMsg('');
      setCsvResult(null);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => residentsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['residents'] });
      setSuccessMsg('Житель успешно удален');
      setErrorMsg('');
      setDeleteError('');
      setDeleteDialogOpen(false);
      setResidentToDelete(null);
      setTimeout(() => setSuccessMsg(''), 3000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      const errorMessage = formatErrorMessage(error);
      setErrorMsg(errorMessage);
      setDeleteError(errorMessage);
      setSuccessMsg('');
    },
  });

  const onSubmit = (data: ResidentFormData) => {
    const createData: CreateResidentRequest = {
      apartment_id: data.apartment_id as number,
      telegram_id: data.telegram_id as number,
      chat_id: data.telegram_id as number, // Chat ID равен Telegram ID
      name: data.name?.trim() || undefined,
      phone: data.phone?.trim() || undefined,
    };
    createMutation.mutate(createData);
  };

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['residents'] });
  };

  const getStatusColor = (status: string): 'success' | 'default' => {
    return status === 'active' ? 'success' : 'default';
  };

  const formatDate = (dateString: string) => {
    try {
      return format(new Date(dateString), 'dd.MM.yyyy HH:mm', { locale: ru });
    } catch {
      return dateString;
    }
  };

  // Bulk import handlers
  const handleBulkSubmit = () => {
    try {
      const data = JSON.parse(bulkJson);
      if (!Array.isArray(data)) {
        setErrorMsg('JSON должен содержать массив объектов');
        return;
      }
      bulkCreateMutation.mutate(data);
    } catch (err) {
      setErrorMsg('Ошибка парсинга JSON: ' + (err as Error).message);
    }
  };

  const handleCloseBulkDialog = () => {
    setBulkDialogOpen(false);
    setBulkJson('');
    setBulkResult(null);
  };

  // CSV import handlers
  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!file.name.endsWith('.csv')) {
        setErrorMsg('Пожалуйста, выберите CSV файл');
        return;
      }
      setSelectedFile(file);
    }
  };

  const handleCsvSubmit = () => {
    if (!selectedFile) {
      setErrorMsg('Пожалуйста, выберите файл');
      return;
    }
    if (!buildingId) {
      setErrorMsg('Не удалось определить ID здания. Обратитесь к администратору.');
      return;
    }
    csvImportMutation.mutate({ file: selectedFile, buildingId });
  };

  const handleCloseCsvDialog = () => {
    setCsvDialogOpen(false);
    setSelectedFile(null);
    setCsvResult(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Delete handlers
  const handleDeleteClick = (resident: Resident) => {
    setResidentToDelete(resident);
    setDeleteError('');
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = () => {
    if (residentToDelete) {
      setDeleteError('');
      deleteMutation.mutate(residentToDelete.id);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    setResidentToDelete(null);
    setDeleteError('');
  };

  // Проверка наличия building_id
  if (!buildingId) {
    return (
      <Layout title="Управление жителями">
        <Container maxWidth="xl" sx={{ py: 4 }}>
          <Alert severity="error">
            Не удалось определить ID здания. Пожалуйста, обратитесь к администратору системы.
          </Alert>
        </Container>
      </Layout>
    );
  }

  return (
    <Layout title="Управление жителями">
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {/* Global Success/Error Messages */}
        {successMsg && (
          <Alert severity="success" sx={{ mb: 3 }}>
            {successMsg}
          </Alert>
        )}

        {errorMsg && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {errorMsg}
          </Alert>
        )}

        {/* Create Form */}
        <Paper elevation={2} sx={{ p: 4, mb: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
            <PersonAddIcon sx={{ fontSize: 32, mr: 2, color: 'primary.main' }} />
            <Typography variant="h5">
              Добавить нового жителя
            </Typography>
          </Box>
          
          <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
            Заполните форму для создания нового жителя в системе
          </Typography>

          <form onSubmit={handleSubmit(onSubmit)}>
            <Grid container spacing={3}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="apartment_id"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="ID квартиры *"
                      type="number"
                      fullWidth
                      error={!!errors.apartment_id}
                      helperText={errors.apartment_id?.message}
                      onChange={(e) => field.onChange(e.target.value || '')}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="telegram_id"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Telegram ID *"
                      type="number"
                      fullWidth
                      error={!!errors.telegram_id}
                      helperText={errors.telegram_id?.message}
                      onChange={(e) => field.onChange(e.target.value || '')}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="name"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Имя (опционально)"
                      fullWidth
                      error={!!errors.name}
                      helperText={errors.name?.message}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="phone"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Телефон (опционально)"
                      fullWidth
                      error={!!errors.phone}
                      helperText={
                        errors.phone?.message || 
                        'Формат: +7 (XXX) XXX-XX-XX, 8XXXXXXXXXX или 7XXXXXXXXXX'
                      }
                      placeholder="+7 (900) 123-45-67"
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12}>
                <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end' }}>
                  <Button
                    variant="outlined"
                    onClick={() => reset()}
                    disabled={!isDirty || createMutation.isPending}
                  >
                    Очистить
                  </Button>
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={!isDirty || createMutation.isPending}
                    startIcon={<PersonAddIcon />}
                  >
                    {createMutation.isPending ? 'Создание...' : 'Создать жителя'}
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </form>
        </Paper>

        {/* Bulk Import Section */}
        <Paper elevation={2} sx={{ p: 4, mb: 4, mt: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
            <GroupAddIcon sx={{ fontSize: 32, mr: 2, color: 'warning.main' }} />
            <Typography variant="h5">
              Массовое создание и импорт
            </Typography>
          </Box>
          
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Создайте несколько жителей одновременно через JSON или импортируйте из CSV файла
          </Typography>

          <Grid container spacing={3}>
            <Grid item xs={12} sm={6}>
              <Paper
                elevation={1}
                sx={{
                  p: 3,
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  border: '2px dashed',
                  borderColor: 'warning.light',
                  backgroundColor: 'rgba(255, 179, 0, 0.05)',
                }}
              >
                <GroupAddIcon sx={{ fontSize: 60, color: 'warning.main', mb: 2 }} />
                <Typography variant="h6" gutterBottom align="center">
                  Массовое создание (JSON)
                </Typography>
                <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
                  Вставьте JSON массив с данными жителей
                </Typography>
                <Button
                  variant="contained"
                  color="warning"
                  startIcon={<GroupAddIcon />}
                  onClick={() => setBulkDialogOpen(true)}
                  sx={{ mt: 'auto' }}
                >
                  Открыть редактор JSON
                </Button>
              </Paper>
            </Grid>

            <Grid item xs={12} sm={6}>
              <Paper
                elevation={1}
                sx={{
                  p: 3,
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  border: '2px dashed',
                  borderColor: 'success.light',
                  backgroundColor: 'rgba(76, 175, 80, 0.05)',
                }}
              >
                <UploadFileIcon sx={{ fontSize: 60, color: 'success.main', mb: 2 }} />
                <Typography variant="h6" gutterBottom align="center">
                  Импорт из CSV
                </Typography>
                <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
                  Загрузите CSV файл с данными жителей
                </Typography>
                <Button
                  variant="contained"
                  color="success"
                  startIcon={<UploadFileIcon />}
                  onClick={() => setCsvDialogOpen(true)}
                  sx={{ mt: 'auto' }}
                >
                  Загрузить CSV
                </Button>
              </Paper>
            </Grid>
          </Grid>
        </Paper>

        <Divider sx={{ my: 4 }} />

        {/* Residents List */}
        <Paper elevation={2} sx={{ p: 4 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Typography variant="h5">
              Список жителей
            </Typography>
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
              <FormControl size="small" sx={{ minWidth: 150 }}>
                <InputLabel>Статус</InputLabel>
                <Select
                  value={statusFilter}
                  label="Статус"
                  onChange={(e) => setStatusFilter(e.target.value as 'active' | 'inactive' | 'all')}
                >
                  <MenuItem value="all">Все</MenuItem>
                  <MenuItem value="active">Активные</MenuItem>
                  <MenuItem value="inactive">Неактивные</MenuItem>
                </Select>
              </FormControl>
              <Tooltip title="Обновить список">
                <IconButton onClick={handleRefresh} color="primary">
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>

          {isLoading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          )}

          {error && (
            <Alert severity="error">
              Ошибка загрузки жителей: {(error as Error).message}
            </Alert>
          )}

          {!isLoading && !error && residents && (
            <>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Всего жителей: {residents.length}
              </Typography>
              
              {residents.length === 0 ? (
                <Alert severity="info">
                  Жители не найдены. Создайте первого жителя используя форму выше.
                </Alert>
              ) : (
                <TableContainer>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell><strong>ID</strong></TableCell>
                        <TableCell><strong>Квартира ID</strong></TableCell>
                        <TableCell><strong>Имя</strong></TableCell>
                        <TableCell><strong>Телефон</strong></TableCell>
                        <TableCell><strong>Telegram ID</strong></TableCell>
                        <TableCell><strong>Статус</strong></TableCell>
                        <TableCell><strong>Создан</strong></TableCell>
                        <TableCell align="right"><strong>Действия</strong></TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {residents.map((resident: Resident) => (
                        <TableRow key={resident.id} hover>
                          <TableCell>{resident.id}</TableCell>
                          <TableCell>{resident.apartment_id}</TableCell>
                          <TableCell>{resident.name || '—'}</TableCell>
                          <TableCell>{resident.phone || '—'}</TableCell>
                          <TableCell>{resident.telegram_id}</TableCell>
                          <TableCell>
                            <Chip 
                              label={resident.status} 
                              color={getStatusColor(resident.status)} 
                              size="small" 
                            />
                          </TableCell>
                          <TableCell>{formatDate(resident.created_at)}</TableCell>
                          <TableCell align="right">
                            <Tooltip title="Удалить жителя">
                              <IconButton
                                color="error"
                                size="small"
                                onClick={() => handleDeleteClick(resident)}
                                disabled={deleteMutation.isPending}
                              >
                                <DeleteIcon />
                              </IconButton>
                            </Tooltip>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
            </>
          )}
        </Paper>

        {/* Bulk JSON Dialog */}
        <Dialog 
          open={bulkDialogOpen} 
          onClose={handleCloseBulkDialog}
          maxWidth="md"
          fullWidth
        >
          <DialogTitle>
            Массовое создание жителей (JSON)
          </DialogTitle>
          <DialogContent>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Вставьте JSON массив с объектами резидентов. Пример:
            </Typography>
            <Paper elevation={0} sx={{ p: 2, backgroundColor: '#f5f5f5', mb: 2 }}>
              <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>
{`[
  {
    "apartment_id": 101,
    "telegram_id": 123456789,
    "name": "Иван Петров",
    "phone": "+7 900 123 45 67"
  },
  {
    "apartment_id": 102,
    "telegram_id": 111222333,
    "name": "Мария Иванова",
    "phone": "89001234567"
  }
]`}
              </Typography>
            </Paper>
            
            <TextField
              fullWidth
              multiline
              rows={12}
              value={bulkJson}
              onChange={(e) => setBulkJson(e.target.value)}
              placeholder='[{"apartment_id": 101, "telegram_id": 123456789, ...}]'
              sx={{ fontFamily: 'monospace' }}
            />

            {bulkResult && (
              <Alert severity={bulkResult.errors.length > 0 ? 'warning' : 'success'} sx={{ mt: 2 }}>
                <Typography variant="body2">
                  Создано: {bulkResult.created}
                </Typography>
                {bulkResult.errors.length > 0 && (
                  <>
                    <Typography variant="body2" sx={{ mt: 1 }}>
                      Ошибки: {bulkResult.errors.length}
                    </Typography>
                    <Box sx={{ mt: 1, maxHeight: 200, overflow: 'auto' }}>
                      {bulkResult.errors.map((err, idx) => (
                        <Typography key={idx} variant="caption" display="block">
                          • {formatBulkError(err)}
                        </Typography>
                      ))}
                    </Box>
                  </>
                )}
              </Alert>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseBulkDialog}>
              Закрыть
            </Button>
            <Button 
              onClick={handleBulkSubmit}
              variant="contained"
              color="warning"
              disabled={!bulkJson || bulkCreateMutation.isPending}
            >
              {bulkCreateMutation.isPending ? 'Создание...' : 'Создать'}
            </Button>
          </DialogActions>
        </Dialog>

        {/* CSV Import Dialog */}
        <Dialog 
          open={csvDialogOpen} 
          onClose={handleCloseCsvDialog}
          maxWidth="sm"
          fullWidth
        >
          <DialogTitle>
            Импорт жителей из CSV
          </DialogTitle>
          <DialogContent>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              CSV файл должен содержать следующие столбцы:
            </Typography>
            <Paper elevation={0} sx={{ p: 2, backgroundColor: '#f5f5f5', mb: 3 }}>
              <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>
{`apartment_id,telegram_id,name,phone
101,123456789,Иван Петров,+79001234567
102,111222333,Мария Иванова,89001234567`}
              </Typography>
            </Paper>

            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              onChange={handleFileSelect}
              style={{ display: 'none' }}
            />

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <Button
                variant="outlined"
                startIcon={<UploadFileIcon />}
                onClick={() => fileInputRef.current?.click()}
              >
                Выбрать файл
              </Button>

              {selectedFile && (
                <Alert severity="info">
                  Выбран файл: {selectedFile.name}
                </Alert>
              )}

              {csvResult && (
                <Alert severity={csvResult.errors.length > 0 ? 'warning' : 'success'}>
                  <Typography variant="body2">
                    Импортировано: {csvResult.imported}
                  </Typography>
                  {csvResult.errors.length > 0 && (
                    <>
                      <Typography variant="body2" sx={{ mt: 1 }}>
                        Ошибки: {csvResult.errors.length}
                      </Typography>
                      <Box sx={{ mt: 1, maxHeight: 150, overflow: 'auto' }}>
                        {csvResult.errors.map((err, idx) => (
                          <Typography key={idx} variant="caption" display="block">
                            • {formatBulkError(err)}
                          </Typography>
                        ))}
                      </Box>
                    </>
                  )}
                </Alert>
              )}
            </Box>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseCsvDialog}>
              Закрыть
            </Button>
            <Button 
              onClick={handleCsvSubmit}
              variant="contained"
              color="success"
              disabled={!selectedFile || csvImportMutation.isPending}
            >
              {csvImportMutation.isPending ? 'Импорт...' : 'Импортировать'}
            </Button>
          </DialogActions>
        </Dialog>

        {/* Delete Confirmation Dialog */}
        <Dialog
          open={deleteDialogOpen}
          onClose={handleDeleteCancel}
          maxWidth="sm"
          fullWidth
        >
          <DialogTitle>
            Подтверждение удаления
          </DialogTitle>
          <DialogContent>
            {residentToDelete && (
              <>
                <Typography variant="body1" gutterBottom>
                  Вы уверены, что хотите удалить жителя?
                </Typography>
                <Box sx={{ mt: 2, p: 2, backgroundColor: '#f5f5f5', borderRadius: 1 }}>
                  <Typography variant="body2">
                    <strong>ID:</strong> {residentToDelete.id}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Имя:</strong> {residentToDelete.name || '—'}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Квартира:</strong> {residentToDelete.apartment_id}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Telegram ID:</strong> {residentToDelete.telegram_id}
                  </Typography>
                </Box>
                <Alert severity="warning" sx={{ mt: 2 }}>
                  Это действие нельзя отменить. Житель будет удален из системы.
                </Alert>
                {deleteError && (
                  <Alert severity="error" sx={{ mt: 2 }}>
                    {deleteError}
                  </Alert>
                )}
              </>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={handleDeleteCancel} disabled={deleteMutation.isPending}>
              Отмена
            </Button>
            <Button
              onClick={handleDeleteConfirm}
              variant="contained"
              color="error"
              disabled={deleteMutation.isPending}
              startIcon={<DeleteIcon />}
            >
              {deleteMutation.isPending ? 'Удаление...' : 'Удалить'}
            </Button>
          </DialogActions>
        </Dialog>
      </Container>
    </Layout>
  );
}

