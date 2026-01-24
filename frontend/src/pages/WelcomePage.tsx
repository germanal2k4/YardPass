import { useNavigate } from 'react-router-dom';
import {
  Container,
  Paper,
  Typography,
  Box,
  Button,
  Grid,
  Divider,
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
    <Container maxWidth="lg">
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
            p: 6, 
            width: '100%',
            borderRadius: 4,
            background: 'linear-gradient(to bottom, #FFFFFF 0%, #FAFAFA 100%)',
          }}
        >
          <Box sx={{ textAlign: 'center', mb: 6 }}>
            <Box
              component="img"
              src="/logo.png"
              alt="YardPass Logo"
              sx={{
                height: { xs: 100, sm: 130, md: 150 },
                width: 'auto',
                mb: 3,
                display: 'inline-block',
                transition: 'transform 0.3s ease',
                '&:hover': {
                  transform: 'scale(1.05)',
                },
              }}
            />
            <Typography 
              variant="h2" 
              component="h1" 
              gutterBottom 
              fontWeight="800"
              sx={{
                background: 'linear-gradient(135deg, #E53935 0%, #FF6D00 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                mb: 2,
              }}
            >
              YardPass
            </Typography>
            <Typography 
              variant="h5" 
              color="text.secondary" 
              gutterBottom
              fontWeight="600"
            >
              Система управления пропусками
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mt: 3, fontSize: '1.1rem' }}>
              Выберите роль для входа в систему
            </Typography>
          </Box>

          <Grid container spacing={4}>
            <Grid item xs={12} md={6}>
              <Paper
                elevation={3}
                sx={{
                  p: 5,
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
                  border: '2px solid transparent',
                  background: 'linear-gradient(135deg, rgba(255, 109, 0, 0.08) 0%, rgba(255, 255, 255, 1) 100%)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 12px 40px rgba(255, 109, 0, 0.25)',
                    borderColor: '#FF6D00',
                  },
                }}
                onClick={() => handleRoleSelect('guard')}
              >
                <SecurityIcon 
                  sx={{ 
                    fontSize: 100, 
                    color: '#FF6D00',
                    mb: 3,
                    filter: 'drop-shadow(0 4px 8px rgba(255, 109, 0, 0.3))',
                  }} 
                />
                <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
                  Охрана
                </Typography>
                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                  Сканирование и проверка QR-кодов пропусков
                </Typography>
                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  color="secondary"
                  startIcon={<SecurityIcon />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRoleSelect('guard');
                  }}
                  sx={{
                    py: 1.5,
                    fontSize: '1.1rem',
                    fontWeight: 700,
                  }}
                >
                  Войти как охранник
                </Button>
              </Paper>
            </Grid>

            <Grid item xs={12} md={6}>
              <Paper
                elevation={3}
                sx={{
                  p: 5,
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
                  border: '2px solid transparent',
                  background: 'linear-gradient(135deg, rgba(229, 57, 53, 0.08) 0%, rgba(255, 255, 255, 1) 100%)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 12px 40px rgba(229, 57, 53, 0.25)',
                    borderColor: '#E53935',
                  },
                }}
                onClick={() => handleRoleSelect('admin')}
              >
                <AdminPanelSettingsIcon 
                  sx={{ 
                    fontSize: 100, 
                    color: '#E53935',
                    mb: 3,
                    filter: 'drop-shadow(0 4px 8px rgba(229, 57, 53, 0.3))',
                  }} 
                />
                <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
                  Администратор
                </Typography>
                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                  Настройка правил, просмотр отчетов и статистики
                </Typography>
                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  color="primary"
                  startIcon={<AdminPanelSettingsIcon />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRoleSelect('admin');
                  }}
                  sx={{
                    py: 1.5,
                    fontSize: '1.1rem',
                    fontWeight: 700,
                  }}
                >
                  Войти как администратор
                </Button>
              </Paper>
            </Grid>
          </Grid>

          <Divider sx={{ my: 5 }}>
            <Typography variant="body1" color="text.secondary" fontWeight="600">
              или
            </Typography>
          </Divider>

          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="h6" gutterBottom fontWeight="600" color="text.primary">
              Нет аккаунта?
            </Typography>
            <Button
              variant="outlined"
              size="large"
              startIcon={<PersonAddIcon />}
              onClick={handleRegister}
              fullWidth
              sx={{ 
                mb: 2,
                py: 1.5,
                fontSize: '1.1rem',
                fontWeight: 700,
                borderWidth: 2,
                borderColor: '#E53935',
                color: '#E53935',
                '&:hover': {
                  borderWidth: 2,
                  borderColor: '#FF6D00',
                  color: '#FF6D00',
                  backgroundColor: 'rgba(255, 109, 0, 0.08)',
                },
              }}
            >
              Зарегистрироваться
            </Button>
            <Typography variant="body2" color="text.secondary">
              Создайте новый аккаунт охранника или администратора
            </Typography>
          </Box>

          <Box sx={{ mt: 5, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Версия 1.0.0 | © 2026 YardPass
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
}

