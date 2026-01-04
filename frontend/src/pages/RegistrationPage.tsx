import { useState, FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Container,
  Paper,
  TextField,
  Button,
  Typography,
  Box,
  Alert,
  IconButton,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  Link,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { APP_ROUTES } from '@/shared/config/constants';

export function RegistrationPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const preselectedRole = searchParams.get('role') as 'admin' | 'guard' | null;
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<'admin' | 'guard'>(preselectedRole || 'guard');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess(false);

    // Validation
    if (password !== confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    if (password.length < 6) {
      setError('Пароль должен содержать минимум 6 символов');
      return;
    }

    if (username.length < 3) {
      setError('Имя пользователя должно содержать минимум 3 символа');
      return;
    }

    setIsLoading(true);

    try {
      // PLACEHOLDER: Backend endpoint не реализован
      // Требуется: POST /auth/register
      // Body: { username, password, email?, role }
      // Response: { message: "User created", user_id: number }
      
      await new Promise(resolve => setTimeout(resolve, 1000)); // Имитация запроса
      
      // Симуляция успешной регистрации
      setSuccess(true);
      setError('');
      
      // Показываем сообщение и перенаправляем на логин через 2 секунды
      setTimeout(() => {
        navigate(`${APP_ROUTES.LOGIN}?role=${role}`);
      }, 2000);
      
    } catch (err) {
      setError('Ошибка при регистрации. Попробуйте позже.');
    } finally {
      setIsLoading(false);
    }
  };

  const getRoleLabel = (roleValue: string) => {
    return roleValue === 'admin' ? 'Администратор' : 'Охранник';
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          py: 4,
        }}
      >
        <Paper elevation={3} sx={{ p: 4, width: '100%', position: 'relative' }}>
          <IconButton
            onClick={() => navigate(APP_ROUTES.HOME)}
            sx={{ position: 'absolute', top: 16, left: 16 }}
            aria-label="назад"
          >
            <ArrowBackIcon />
          </IconButton>

          <Box sx={{ textAlign: 'center', mb: 3 }}>
            <PersonAddIcon sx={{ fontSize: 48, color: 'primary.main', mb: 1 }} />
            <Typography variant="h4" component="h1" gutterBottom>
              Регистрация
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Создание нового пользователя
            </Typography>
          </Box>

          <Alert severity="warning" sx={{ mb: 3 }}>
            <Typography variant="body2" sx={{ mb: 1 }}>
              <strong>Внимание:</strong> Функция регистрации находится в разработке.
            </Typography>
            <Typography variant="caption">
              Backend endpoint <code>POST /auth/register</code> еще не реализован.
              Форма работает в режиме демонстрации.
            </Typography>
          </Alert>

          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              Пользователь успешно зарегистрирован! Перенаправление на страницу входа...
            </Alert>
          )}

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <form onSubmit={handleSubmit}>
            <TextField
              label="Имя пользователя"
              type="text"
              fullWidth
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              sx={{ mb: 2 }}
              autoFocus
              helperText="Минимум 3 символа"
              disabled={isLoading || success}
            />

            <TextField
              label="Email (опционально)"
              type="email"
              fullWidth
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              sx={{ mb: 2 }}
              disabled={isLoading || success}
            />

            <TextField
              label="Пароль"
              type="password"
              fullWidth
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              sx={{ mb: 2 }}
              helperText="Минимум 6 символов"
              disabled={isLoading || success}
            />

            <TextField
              label="Подтвердите пароль"
              type="password"
              fullWidth
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              sx={{ mb: 2 }}
              disabled={isLoading || success}
            />

            <FormControl fullWidth sx={{ mb: 3 }}>
              <InputLabel>Роль</InputLabel>
              <Select
                value={role}
                label="Роль"
                onChange={(e) => setRole(e.target.value as 'admin' | 'guard')}
                disabled={isLoading || success}
              >
                <MenuItem value="guard">
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Chip label="Охранник" color="primary" size="small" />
                    <Typography variant="body2">Сканирование пропусков</Typography>
                  </Box>
                </MenuItem>
                <MenuItem value="admin">
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Chip label="Администратор" color="secondary" size="small" />
                    <Typography variant="body2">Управление системой</Typography>
                  </Box>
                </MenuItem>
              </Select>
            </FormControl>

            <Button
              type="submit"
              variant="contained"
              fullWidth
              size="large"
              disabled={isLoading || success}
              startIcon={<PersonAddIcon />}
            >
              {isLoading ? 'Регистрация...' : 'Зарегистрироваться'}
            </Button>
          </form>

          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Уже есть аккаунт?{' '}
              <Link
                component="button"
                variant="body2"
                onClick={() => navigate(`${APP_ROUTES.LOGIN}?role=${role}`)}
                sx={{ cursor: 'pointer' }}
              >
                Войти
              </Link>
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
}

