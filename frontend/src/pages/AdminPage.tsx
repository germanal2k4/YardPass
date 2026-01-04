import { Container, Paper, Typography, Grid, Button, Box } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { Layout } from '@/shared/ui/Layout';
import { APP_ROUTES } from '@/shared/config/constants';
import SettingsIcon from '@mui/icons-material/Settings';
import AssessmentIcon from '@mui/icons-material/Assessment';

export function AdminPage() {
  const navigate = useNavigate();

  return (
    <Layout title="Панель администратора">
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Typography variant="h5" gutterBottom>
          Выберите раздел
        </Typography>

        <Grid container spacing={3} sx={{ mt: 2 }}>
          <Grid item xs={12} md={6}>
            <Paper
              elevation={2}
              sx={{
                p: 4,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                cursor: 'pointer',
                '&:hover': {
                  backgroundColor: 'action.hover',
                },
              }}
              onClick={() => navigate(APP_ROUTES.ADMIN_RULES)}
            >
              <SettingsIcon sx={{ fontSize: 80, color: 'primary.main', mb: 2 }} />
              <Typography variant="h5" gutterBottom>
                Правила и настройки
              </Typography>
              <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 2 }}>
                Настройка тихих часов, лимитов пропусков и других параметров контрольного пункта
              </Typography>
              <Button variant="contained" sx={{ mt: 'auto' }}>
                Открыть
              </Button>
            </Paper>
          </Grid>

          <Grid item xs={12} md={6}>
            <Paper
              elevation={2}
              sx={{
                p: 4,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                cursor: 'pointer',
                '&:hover': {
                  backgroundColor: 'action.hover',
                },
              }}
              onClick={() => navigate(APP_ROUTES.ADMIN_REPORTS)}
            >
              <AssessmentIcon sx={{ fontSize: 80, color: 'primary.main', mb: 2 }} />
              <Typography variant="h5" gutterBottom>
                Отчеты
              </Typography>
              <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 2 }}>
                Просмотр статистики, журнала событий и выгрузка отчетов в Excel
              </Typography>
              <Button variant="contained" sx={{ mt: 'auto' }}>
                Открыть
              </Button>
            </Paper>
          </Grid>
        </Grid>
      </Container>
    </Layout>
  );
}

