import { Container, Typography, Box, Button, Paper } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { APP_ROUTES } from '@/shared/config/constants';
import BlockIcon from '@mui/icons-material/Block';

export function ForbiddenPage() {
  const navigate = useNavigate();

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Paper elevation={3} sx={{ p: 4, textAlign: 'center' }}>
          <BlockIcon sx={{ fontSize: 80, color: 'error.main', mb: 2 }} />
          <Typography variant="h4" gutterBottom>
            Доступ запрещен
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
            У вас нет прав для просмотра этой страницы
          </Typography>
          <Button
            variant="contained"
            onClick={() => navigate(APP_ROUTES.HOME)}
          >
            На главную
          </Button>
        </Paper>
      </Box>
    </Container>
  );
}

