import { useEffect, useState } from 'react';
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
import { formatErrorMessage } from '@/shared/utils/errors';

type GuardCredentialsDraft = {
  username: string;
  password: string;
};

type GuardStatus = {
  type: 'success' | 'error';
  message: string;
};

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

export function AdminGuardsPage() {
  const queryClient = useQueryClient();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [guardDrafts, setGuardDrafts] = useState<Record<number, GuardCredentialsDraft>>({});
  const [guardStatuses, setGuardStatuses] = useState<Record<number, GuardStatus>>({});
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
      setErrorMsg(formatErrorMessage(error));
      setSuccessMsg('');
    },
  });

  const updateGuardCredentialsMutation = useMutation({
    mutationFn: ({ guardId, draft }: { guardId: number; draft: GuardCredentialsDraft }) =>
      usersApi.updateCredentials(guardId, {
        username: draft.username.trim() || undefined,
        password: draft.password.trim() || undefined,
      }),
    onSuccess: (_updatedUser, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users', 'guards'] });
      setGuardStatuses((prev) => ({
        ...prev,
        [variables.guardId]: { type: 'success', message: 'Данные успешно обновлены' },
      }));
    },
    onError: (error: AxiosError<ErrorResponse>, variables) => {
      setGuardStatuses((prev) => ({
        ...prev,
        [variables.guardId]: { type: 'error', message: formatErrorMessage(error) },
      }));
    },
  });

  useEffect(() => {
    if (!guardsQuery.data) return;
    const nextDrafts: Record<number, GuardCredentialsDraft> = {};
    guardsQuery.data.users.forEach((guard) => {
      nextDrafts[guard.id] = guardDrafts[guard.id] || {
        username: guard.username,
        password: '',
      };
    });
    setGuardDrafts(nextDrafts);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [guardsQuery.data]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const validationError = validateCredentials(username, password);
    if (validationError) {
      setErrorMsg(validationError);
      setSuccessMsg('');
      return;
    }

    createMutation.mutate({
      username: username.trim(),
      email: email.trim() || undefined,
      password: password.trim(),
      role: 'guard',
    });
  };

  const handleGuardDraftChange = (guardId: number, field: keyof GuardCredentialsDraft, value: string) => {
    setGuardStatuses((prev) => {
      if (!prev[guardId]) return prev;
      const { [guardId]: _removed, ...rest } = prev;
      return rest;
    });
    setGuardDrafts((prev) => ({
      ...prev,
      [guardId]: {
        ...(prev[guardId] || { username: '', password: '' }),
        [field]: value,
      },
    }));
  };

  const handleGuardCredentialsUpdate = (guardId: number) => {
    const draft = guardDrafts[guardId];
    const validationError = validateCredentials(draft?.username ?? '', draft?.password ?? '');
    if (validationError) {
      setGuardStatuses((prev) => ({
        ...prev,
        [guardId]: { type: 'error', message: validationError },
      }));
      return;
    }
    updateGuardCredentialsMutation.mutate({ guardId, draft });
    setGuardDrafts((prev) => ({
      ...prev,
      [guardId]: {
        ...prev[guardId],
        password: '',
      },
    }));
  };

  return (
    <Layout title="Охранники">
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Paper elevation={2} sx={{ p: 4, mb: 3 }}>
          <Typography variant="h5" gutterBottom>
            Добавить аккаунт охранника
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
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Показаны только аккаунты охраны вашего здания.
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
                  <Box sx={{ width: '100%', display: 'grid', gap: 1.5 }}>
                    <ListItemText
                      primary={user.username}
                      secondary={user.email ? `${user.email} · ${user.status}` : user.status}
                    />
                    {guardStatuses[user.id] && (
                      <Alert severity={guardStatuses[user.id].type}>
                        {guardStatuses[user.id].message}
                      </Alert>
                    )}
                    <TextField
                      label="Новый логин охранника"
                      value={guardDrafts[user.id]?.username ?? ''}
                      onChange={(e) => handleGuardDraftChange(user.id, 'username', e.target.value)}
                      size="small"
                    />
                    <TextField
                      label="Новый пароль охранника"
                      value={guardDrafts[user.id]?.password ?? ''}
                      onChange={(e) => handleGuardDraftChange(user.id, 'password', e.target.value)}
                      type="password"
                      size="small"
                    />
                    <Button
                      variant="outlined"
                      onClick={() => handleGuardCredentialsUpdate(user.id)}
                      disabled={updateGuardCredentialsMutation.isPending}
                    >
                      {updateGuardCredentialsMutation.isPending ? 'Сохранение...' : 'Обновить логин/пароль'}
                    </Button>
                  </Box>
                </ListItem>
              ))}
            </List>
          )}
        </Paper>
      </Container>
    </Layout>
  );
}
