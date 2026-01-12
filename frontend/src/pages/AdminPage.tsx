import { Container, Paper, Typography, Grid, Button, Box } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { Layout } from '@/shared/ui/Layout';
import { APP_ROUTES } from '@/shared/config/constants';
import SettingsIcon from '@mui/icons-material/Settings';
import AssessmentIcon from '@mui/icons-material/Assessment';
import PeopleIcon from '@mui/icons-material/People';

export function AdminPage() {
  const navigate = useNavigate();

  return (
    <Layout title="Панель администратора">
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
          Выберите раздел
        </Typography>

        <Grid container spacing={4} sx={{ mt: 2 }}>
          {/* Правила и настройки */}
          <Grid item xs={12} sm={6} md={4}>
            <Paper
              elevation={4}
              sx={{
                p: 5,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
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
              onClick={() => navigate(APP_ROUTES.ADMIN_RULES)}
            >
              <SettingsIcon 
                sx={{ 
                  fontSize: 100, 
                  color: '#E53935',
                  mb: 3,
                  filter: 'drop-shadow(0 4px 8px rgba(229, 57, 53, 0.3))',
                }} 
              />
              <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
                Правила и настройки
              </Typography>
              <Typography variant="body1" color="text.secondary" align="center" sx={{ mb: 4 }}>
                Настройка тихих часов, лимитов пропусков и других параметров контрольного пункта
              </Typography>
              <Button 
                variant="contained" 
                size="large"
                sx={{ 
                  mt: 'auto',
                  px: 4,
                  py: 1.5,
                  fontSize: '1.1rem',
                  fontWeight: 700,
                }}
              >
                Открыть
              </Button>
            </Paper>
          </Grid>

          {/* Управление жителями */}
          <Grid item xs={12} sm={6} md={4}>
            <Paper
              elevation={4}
              sx={{
                p: 5,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                cursor: 'pointer',
                transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
                border: '2px solid transparent',
                background: 'linear-gradient(135deg, rgba(255, 179, 0, 0.08) 0%, rgba(255, 255, 255, 1) 100%)',
                '&:hover': {
                  transform: 'translateY(-8px)',
                  boxShadow: '0 12px 40px rgba(255, 179, 0, 0.25)',
                  borderColor: '#FFB300',
                },
              }}
              onClick={() => navigate(APP_ROUTES.ADMIN_RESIDENTS)}
            >
              <PeopleIcon 
                sx={{ 
                  fontSize: 100, 
                  color: '#FFB300',
                  mb: 3,
                  filter: 'drop-shadow(0 4px 8px rgba(255, 179, 0, 0.3))',
                }} 
              />
              <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
                Жители
              </Typography>
              <Typography variant="body1" color="text.secondary" align="center" sx={{ mb: 4 }}>
                Управление жителями, добавление новых резидентов и просмотр списка
              </Typography>
              <Button 
                variant="contained"
                sx={{ 
                  mt: 'auto',
                  px: 4,
                  py: 1.5,
                  fontSize: '1.1rem',
                  fontWeight: 700,
                  backgroundColor: '#FFB300',
                  '&:hover': {
                    backgroundColor: '#FFA000',
                  },
                }}
              >
                Открыть
              </Button>
            </Paper>
          </Grid>

          {/* Отчеты */}
          <Grid item xs={12} sm={6} md={4}>
            <Paper
              elevation={4}
              sx={{
                p: 5,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
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
              onClick={() => navigate(APP_ROUTES.ADMIN_REPORTS)}
            >
              <AssessmentIcon 
                sx={{ 
                  fontSize: 100, 
                  color: '#FF6D00',
                  mb: 3,
                  filter: 'drop-shadow(0 4px 8px rgba(255, 109, 0, 0.3))',
                }} 
              />
              <Typography variant="h4" gutterBottom fontWeight="700" color="#263238">
                Отчеты
              </Typography>
              <Typography variant="body1" color="text.secondary" align="center" sx={{ mb: 4 }}>
                Просмотр статистики, журнала событий и выгрузка отчетов в Excel
              </Typography>
              <Button 
                variant="contained"
                color="secondary"
                size="large"
                sx={{ 
                  mt: 'auto',
                  px: 4,
                  py: 1.5,
                  fontSize: '1.1rem',
                  fontWeight: 700,
                }}
              >
                Открыть
              </Button>
            </Paper>
          </Grid>
        </Grid>
      </Container>
    </Layout>
  );
}

