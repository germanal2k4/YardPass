import { useState } from 'react';
import { Alert, Box, Button, Container, Paper, TextField, Typography } from '@mui/material';
import { useMutation } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { Layout } from '@/shared/ui/Layout';
import { usersApi } from '@/shared/api/users';
import { useAuth } from '@/features/auth/useAuth';
import type { ErrorResponse } from '@/shared/types/api';
import { formatErrorMessage } from '@/shared/utils/errors';

const MIN_USERNAME_LENGTH = 4;
const MIN_PASSWORD_LENGTH = 6;

function validateCredentials(username: string, password: string): string | null {
  const normalizedUsername = username.trim();
  const normalizedPassword = password.trim();

  if (!normalizedUsername && !normalizedPassword) {
    return 'Укажите новый логин и/или пароль.';
  }
  if (normalizedUsername && normalizedUsername.length < MIN_USERNAME_LENGTH) {
    return 'Логин должен быть не короче 4 символов.';
  }
  if (normalizedPassword && normalizedPassword.length < MIN_PASSWORD_LENGTH) {
    return 'Пароль должен быть не короче 6 символов.';
  }
  return null;
}

export function AdminAccountPage() {
  const { user } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const updateMutation = useMutation({
    mutationFn: () => {
      if (!user) {
        return Promise.reject(new Error('Пользователь не найден'));
      }
      return usersApi.updateCredentials(user.user_id, {
        username: username.trim() || undefined,
        password: password.trim() || undefined,
      });
    },
    onSuccess: () => {
      setSuccessMsg('Данные администратора обновлены');
      setErrorMsg('');
      setPassword('');
      setTimeout(() => setSuccessMsg(''), 3000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      setErrorMsg(formatErrorMessage(error));
      setSuccessMsg('');
    },
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const validationError = validateCredentials(username, password);
    if (validationError) {
      setErrorMsg(validationError);
      setSuccessMsg('');
      return;
    }
    updateMutation.mutate();
  };

  return (
    <Layout title="Мой аккаунт">
      <Container maxWidth="sm" sx={{ py: 4 }}>
        <Paper elevation={2} sx={{ p: 4 }}>
          <Typography variant="h5" gutterBottom>
            Изменение логина и пароля
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Логин минимум 4 символа, пароль минимум 6 символов.
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

          <Box component="form" onSubmit={handleSubmit} sx={{ display: 'grid', gap: 2 }}>
            <TextField
              label="Новый логин"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <TextField
              label="Новый пароль"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
            />
            <Button type="submit" variant="contained" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'Сохранение...' : 'Сохранить'}
            </Button>
          </Box>
        </Paper>
      </Container>
    </Layout>
  );
}
