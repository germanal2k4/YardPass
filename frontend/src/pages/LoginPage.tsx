import { useState, FormEvent } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import {
  Container,
  Paper,
  TextField,
  Button,
  Typography,
  Box,
  Alert,
  Chip,
  IconButton,
  Link,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import SecurityIcon from '@mui/icons-material/Security';
import { useAuth } from '@/features/auth/useAuth';
import { AxiosError } from 'axios';
import type { ErrorResponse } from '@/shared/types/api';
import { ERROR_MESSAGES, APP_ROUTES } from '@/shared/config/constants';

export function LoginPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const role = searchParams.get('role') as 'admin' | 'guard' | null;
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();

  const getRoleLabel = () => {
    if (role === 'admin') return 'Администратор';
    if (role === 'guard') return 'Охранник';
    return 'Пользователь';
  };

  const getRoleIcon = () => {
    if (role === 'admin') return <AdminPanelSettingsIcon />;
    if (role === 'guard') return <SecurityIcon />;
    return null;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login({ username, password });
    } catch (err) {
      const axiosError = err as AxiosError<ErrorResponse>;
      const errorCode = axiosError.response?.data?.error?.code || 'UNKNOWN_ERROR';
      setError(ERROR_MESSAGES[errorCode] || ERROR_MESSAGES.UNKNOWN_ERROR);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, rgba(229, 57, 53, 0.05) 0%, rgba(255, 109, 0, 0.05) 100%)',
        }}
      >
        <Paper 
          elevation={6} 
          sx={{ 
            p: 5, 
            width: '100%', 
            position: 'relative',
            borderRadius: 4,
            background: 'linear-gradient(to bottom, #FFFFFF 0%, #FAFAFA 100%)',
          }}
        >
          <IconButton
            onClick={() => navigate(APP_ROUTES.HOME)}
            sx={{ 
              position: 'absolute', 
              top: 20, 
              left: 20,
              color: '#E53935',
              '&:hover': {
                backgroundColor: 'rgba(229, 57, 53, 0.08)',
              },
            }}
            aria-label="назад"
          >
            <ArrowBackIcon />
          </IconButton>

          <Box sx={{ textAlign: 'center', mb: 2 }}>
            <Box
              component="img"
              src="/logo.png"
              alt="YardPass Logo"
              sx={{
                height: { xs: 80, sm: 100 },
                width: 'auto',
                mb: 2,
                display: 'inline-block',
                transition: 'transform 0.3s ease',
                '&:hover': {
                  transform: 'scale(1.05)',
                },
              }}
            />
          </Box>

          <Typography 
            variant="h3" 
            component="h1" 
            gutterBottom 
            align="center"
            fontWeight="800"
            sx={{
              background: 'linear-gradient(135deg, #E53935 0%, #FF6D00 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            YardPass
          </Typography>
          
          {role && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mb: 3 }}>
              <Chip
                icon={getRoleIcon()}
                label={getRoleLabel()}
                color={role === 'admin' ? 'primary' : 'secondary'}
                size="medium"
                sx={{
                  fontWeight: 700,
                  px: 2,
                  py: 2.5,
                  fontSize: '1rem',
                }}
              />
            </Box>
          )}
          
          <Typography 
            variant="h6" 
            gutterBottom 
            align="center" 
            color="text.secondary" 
            mb={4}
            fontWeight="600"
          >
            Вход в систему
          </Typography>

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
            />

            <TextField
              label="Пароль"
              type="password"
              fullWidth
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              sx={{ mb: 3 }}
            />

            <Button
              type="submit"
              variant="contained"
              fullWidth
              size="large"
              disabled={isLoading}
              color={role === 'admin' ? 'primary' : 'secondary'}
              sx={{
                py: 1.5,
                fontSize: '1.1rem',
                fontWeight: 700,
              }}
            >
              {isLoading ? 'Вход...' : 'Войти'}
            </Button>
          </form>

          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Нет аккаунта?{' '}
              <Link
                component="button"
                variant="body2"
                onClick={() => navigate(`${APP_ROUTES.REGISTER}${role ? `?role=${role}` : ''}`)}
                sx={{ 
                  cursor: 'pointer',
                  color: '#E53935',
                  fontWeight: 700,
                  '&:hover': {
                    color: '#FF6D00',
                  },
                }}
              >
                Зарегистрироваться
              </Link>
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
}

