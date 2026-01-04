import { useNavigate } from 'react-router-dom';
import {
  Container,
  Paper,
  Typography,
  Box,
  Button,
  Grid,
  Divider,
  Link,
} from '@mui/material';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import SecurityIcon from '@mui/icons-material/Security';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { APP_ROUTES } from '@/shared/config/constants';

export function WelcomePage() {
  const navigate = useNavigate();

  const handleRoleSelect = (role: 'admin' | 'guard') => {
    navigate(`${APP_ROUTES.LOGIN}?role=${role}`);
  };

  const handleRegister = () => {
    navigate(APP_ROUTES.REGISTER);
  };

  return (
    <Container maxWidth="md">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Paper elevation={3} sx={{ p: 5, width: '100%' }}>
          <Box sx={{ textAlign: 'center', mb: 5 }}>
            <Typography variant="h3" component="h1" gutterBottom fontWeight="bold">
              YardPass
            </Typography>
            <Typography variant="h6" color="text.secondary" gutterBottom>
              Система управления пропусками
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
              Выберите роль для входа в систему
            </Typography>
          </Box>

          <Grid container spacing={4}>
            <Grid item xs={12} md={6}>
              <Paper
                elevation={2}
                sx={{
                  p: 4,
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.3s',
                  '&:hover': {
                    backgroundColor: 'action.hover',
                    transform: 'translateY(-4px)',
                    boxShadow: 6,
                  },
                }}
                onClick={() => handleRoleSelect('guard')}
              >
                <SecurityIcon sx={{ fontSize: 80, color: 'primary.main', mb: 2 }} />
                <Typography variant="h5" gutterBottom fontWeight="600">
                  Охрана
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                  Сканирование и проверка QR-кодов пропусков
                </Typography>
                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  startIcon={<SecurityIcon />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRoleSelect('guard');
                  }}
                >
                  Войти как охранник
                </Button>
              </Paper>
            </Grid>

            <Grid item xs={12} md={6}>
              <Paper
                elevation={2}
                sx={{
                  p: 4,
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.3s',
                  '&:hover': {
                    backgroundColor: 'action.hover',
                    transform: 'translateY(-4px)',
                    boxShadow: 6,
                  },
                }}
                onClick={() => handleRoleSelect('admin')}
              >
                <AdminPanelSettingsIcon sx={{ fontSize: 80, color: 'secondary.main', mb: 2 }} />
                <Typography variant="h5" gutterBottom fontWeight="600">
                  Администратор
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                  Настройка правил, просмотр отчетов и статистики
                </Typography>
                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  color="secondary"
                  startIcon={<AdminPanelSettingsIcon />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRoleSelect('admin');
                  }}
                >
                  Войти как администратор
                </Button>
              </Paper>
            </Grid>
          </Grid>

          <Divider sx={{ my: 4 }}>
            <Typography variant="body2" color="text.secondary">
              или
            </Typography>
          </Divider>

          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="body1" gutterBottom>
              Нет аккаунта?
            </Typography>
            <Button
              variant="outlined"
              size="large"
              startIcon={<PersonAddIcon />}
              onClick={handleRegister}
              fullWidth
              sx={{ mb: 2 }}
            >
              Зарегистрироваться
            </Button>
            <Typography variant="caption" color="text.secondary">
              Создайте новый аккаунт охранника или администратора
            </Typography>
          </Box>

          <Box sx={{ mt: 4, textAlign: 'center' }}>
            <Typography variant="caption" color="text.secondary">
              Версия 1.0.0 | © 2025 YardPass
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
}

