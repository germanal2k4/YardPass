import { useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Container,
  List,
  ListItem,
  ListItemText,
  Paper,
  TextField,
  Typography,
} from '@mui/material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { Layout } from '@/shared/ui/Layout';
import { usersApi } from '@/shared/api/users';
import type { ErrorResponse, RegisterUserRequest } from '@/shared/types/api';
import { ERROR_MESSAGES } from '@/shared/config/constants';

export function AdminGuardsPage() {
  const queryClient = useQueryClient();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const guardsQuery = useQuery({
    queryKey: ['users', 'guards'],
    queryFn: usersApi.listGuards,
  });

  const createMutation = useMutation({
    mutationFn: (payload: RegisterUserRequest) => usersApi.register(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', 'guards'] });
      setUsername('');
      setEmail('');
      setPassword('');
      setErrorMsg('');
      setSuccessMsg('Аккаунт охранника создан');
      setTimeout(() => setSuccessMsg(''), 3000);
    },
    onError: (error: AxiosError<ErrorResponse>) => {
      const errorCode = error.response?.data?.error?.code || 'UNKNOWN_ERROR';
      setErrorMsg(ERROR_MESSAGES[errorCode] || error.response?.data?.error?.message || ERROR_MESSAGES.UNKNOWN_ERROR);
      setSuccessMsg('');
    },
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!username.trim() || !password.trim()) {
      setErrorMsg('Логин и пароль обязательны');
      return;
    }

    createMutation.mutate({
      username: username.trim(),
      email: email.trim() || undefined,
      password,
      role: 'guard',
    });
  };

  return (
    <Layout title="Охранники">
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Paper elevation={2} sx={{ p: 4, mb: 3 }}>
          <Typography variant="h5" gutterBottom>
            Добавить аккаунт охранника
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
              label="Логин"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
            <TextField
              label="Email (опционально)"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              type="email"
            />
            <TextField
              label="Пароль"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
              required
            />
            <Button type="submit" variant="contained" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Создание...' : 'Создать охранника'}
            </Button>
          </Box>
        </Paper>

        <Paper elevation={2} sx={{ p: 4 }}>
          <Typography variant="h6" gutterBottom>
            Список охранников
          </Typography>

          {guardsQuery.isLoading && <Typography>Загрузка...</Typography>}

          {guardsQuery.isError && (
            <Alert severity="error">Не удалось загрузить список охранников</Alert>
          )}

          {guardsQuery.data && guardsQuery.data.users.length === 0 && (
            <Typography color="text.secondary">Охранники пока не добавлены</Typography>
          )}

          {guardsQuery.data && guardsQuery.data.users.length > 0 && (
            <List>
              {guardsQuery.data.users.map((user) => (
                <ListItem key={user.id} divider>
                  <ListItemText
                    primary={user.username}
                    secondary={user.email ? `${user.email} · ${user.status}` : user.status}
                  />
                </ListItem>
              ))}
            </List>
          )}
        </Paper>
      </Container>
    </Layout>
  );
}
