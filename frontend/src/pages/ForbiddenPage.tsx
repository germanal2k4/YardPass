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
          background: 'linear-gradient(135deg, rgba(229, 57, 53, 0.05) 0%, rgba(255, 109, 0, 0.05) 100%)',
        }}
      >
        <Paper 
          elevation={6} 
          sx={{ 
            p: 6, 
            textAlign: 'center',
            borderRadius: 4,
            background: 'linear-gradient(to bottom, #FFFFFF 0%, #FAFAFA 100%)',
          }}
        >
          <BlockIcon 
            sx={{ 
              fontSize: 120, 
              color: '#E53935',
              mb: 3,
              filter: 'drop-shadow(0 4px 8px rgba(229, 57, 53, 0.3))',
            }} 
          />
          <Typography 
            variant="h3" 
            gutterBottom
            fontWeight="800"
            sx={{
              background: 'linear-gradient(135deg, #E53935 0%, #FF6D00 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            Доступ запрещен
          </Typography>
          <Typography variant="h6" color="text.secondary" sx={{ mb: 4 }}>
            У вас нет прав для просмотра этой страницы
          </Typography>
          <Button
            variant="contained"
            size="large"
            onClick={() => navigate(APP_ROUTES.HOME)}
            sx={{
              px: 4,
              py: 1.5,
              fontSize: '1.1rem',
              fontWeight: 700,
            }}
          >
            На главную
          </Button>
        </Paper>
      </Box>
    </Container>
  );
}

