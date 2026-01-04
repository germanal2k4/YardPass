import { useState, useEffect } from 'react';
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
} from '@mui/material';
import { Layout } from '@/shared/ui/Layout';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rulesApi } from '@/shared/api/rules';
import { config } from '@/shared/config/env';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { UpdateRuleRequest } from '@/shared/types/api';
import { AxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';
import { ERROR_MESSAGES } from '@/shared/config/constants';

const ruleSchema = z.object({
  quiet_hours_start: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/, 'Формат: HH:mm').optional().or(z.literal('')),
  quiet_hours_end: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/, 'Формат: HH:mm').optional().or(z.literal('')),
  daily_pass_limit_per_apartment: z.number().int().min(1).max(100),
  max_pass_duration_hours: z.number().int().min(1).max(168),
});

type RuleFormData = z.infer<typeof ruleSchema>;

export function AdminRulesPage() {
  const queryClient = useQueryClient();
  const buildingId = config.defaultBuildingId;
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const { data: rule, isLoading, error } = useQuery({
    queryKey: ['rules', buildingId],
    queryFn: () => rulesApi.get(buildingId),
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<RuleFormData>({
    resolver: zodResolver(ruleSchema),
    defaultValues: {
      quiet_hours_start: '',
      quiet_hours_end: '',
      daily_pass_limit_per_apartment: 5,
      max_pass_duration_hours: 24,
    },
  });

  useEffect(() => {
    if (rule) {
      reset({
        quiet_hours_start: rule.quiet_hours_start || '',
        quiet_hours_end: rule.quiet_hours_end || '',
        daily_pass_limit_per_apartment: rule.daily_pass_limit_per_apartment,
        max_pass_duration_hours: rule.max_pass_duration_hours,
      });
    }
  }, [rule, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: UpdateRuleRequest) => rulesApi.update(buildingId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules', buildingId] });
      setSuccessMsg('Правила успешно обновлены');
      setErrorMsg('');
      setTimeout(() => setSuccessMsg(''), 3000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      const errorCode = error.response?.data?.error?.code || 'UNKNOWN_ERROR';
      setErrorMsg(ERROR_MESSAGES[errorCode] || ERROR_MESSAGES.UNKNOWN_ERROR);
      setSuccessMsg('');
    },
  });

  const onSubmit = (data: RuleFormData) => {
    const updateData: UpdateRuleRequest = {
      quiet_hours_start: data.quiet_hours_start || undefined,
      quiet_hours_end: data.quiet_hours_end || undefined,
      daily_pass_limit_per_apartment: data.daily_pass_limit_per_apartment,
      max_pass_duration_hours: data.max_pass_duration_hours,
    };
    updateMutation.mutate(updateData);
  };

  if (isLoading) {
    return (
      <Layout title="Правила и настройки">
        <Container maxWidth="md" sx={{ py: 4, display: 'flex', justifyContent: 'center' }}>
          <CircularProgress />
        </Container>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout title="Правила и настройки">
        <Container maxWidth="md" sx={{ py: 4 }}>
          <Alert severity="error">
            Ошибка загрузки правил: {(error as Error).message}
          </Alert>
        </Container>
      </Layout>
    );
  }

  return (
    <Layout title="Правила и настройки">
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Paper elevation={2} sx={{ p: 4 }}>
          <Typography variant="h5" gutterBottom>
            Настройка правил контрольного пункта
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
            Здание ID: {buildingId}
          </Typography>

          {successMsg && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {successMsg}
            </Alert>
          )}

          {errorMsg && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {errorMsg}
            </Alert>
          )}

          <form onSubmit={handleSubmit(onSubmit)}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <Typography variant="h6" gutterBottom>
                  Тихие часы
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  Время, когда создание новых пропусков запрещено
                </Typography>
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="quiet_hours_start"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Начало (HH:mm)"
                      placeholder="22:00"
                      fullWidth
                      error={!!errors.quiet_hours_start}
                      helperText={errors.quiet_hours_start?.message}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="quiet_hours_end"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Конец (HH:mm)"
                      placeholder="08:00"
                      fullWidth
                      error={!!errors.quiet_hours_end}
                      helperText={errors.quiet_hours_end?.message}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12}>
                <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>
                  Лимиты
                </Typography>
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="daily_pass_limit_per_apartment"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Лимит пропусков в день на квартиру"
                      type="number"
                      fullWidth
                      error={!!errors.daily_pass_limit_per_apartment}
                      helperText={errors.daily_pass_limit_per_apartment?.message}
                      onChange={(e) => field.onChange(parseInt(e.target.value, 10))}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <Controller
                  name="max_pass_duration_hours"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Макс. срок действия пропуска (часы)"
                      type="number"
                      fullWidth
                      error={!!errors.max_pass_duration_hours}
                      helperText={errors.max_pass_duration_hours?.message}
                      onChange={(e) => field.onChange(parseInt(e.target.value, 10))}
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12}>
                <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end' }}>
                  <Button
                    variant="outlined"
                    onClick={() => reset()}
                    disabled={!isDirty || updateMutation.isPending}
                  >
                    Отменить
                  </Button>
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={!isDirty || updateMutation.isPending}
                  >
                    {updateMutation.isPending ? 'Сохранение...' : 'Сохранить изменения'}
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </form>
        </Paper>
      </Container>
    </Layout>
  );
}

