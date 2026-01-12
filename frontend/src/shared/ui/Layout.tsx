import { ReactNode } from 'react';
import { 
  Box, 
  AppBar, 
  Toolbar, 
  Typography,
  Button, 
  Container,
  Chip,
  Tooltip,
} from '@mui/material';
import { useAuth } from '@/features/auth/useAuth';
import LogoutIcon from '@mui/icons-material/Logout';
import SettingsIcon from '@mui/icons-material/Settings';
import AssessmentIcon from '@mui/icons-material/Assessment';
import PeopleIcon from '@mui/icons-material/People';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import SecurityIcon from '@mui/icons-material/Security';
import { useNavigate } from 'react-router-dom';
import { APP_ROUTES } from '@/shared/config/constants';

interface LayoutProps {
  children: ReactNode;
  title?: string;
}

export function Layout({ children, title }: LayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const getRoleIcon = (role: string) => {
    return role === 'admin' ? <AdminPanelSettingsIcon /> : <SecurityIcon />;
  };

  const getRoleName = (role: string) => {
    return role === 'admin' ? 'Администратор' : 'Охрана';
  };

  const getHomeRoute = () => {
    if (!user) return APP_ROUTES.HOME;
    return user.role === 'admin' ? APP_ROUTES.ADMIN : APP_ROUTES.SECURITY;
  };

  return (
    <Box sx={{ minHeight: '100vh', backgroundColor: 'background.default' }}>
      <AppBar 
        position="static" 
        sx={{ 
          boxShadow: '0 4px 20px rgba(229, 57, 53, 0.3)',
          borderRadius: 0,
          background: user?.role === 'admin' 
            ? 'linear-gradient(135deg, #E53935 0%, #FF6D00 50%, #FFB300 100%)'
            : 'linear-gradient(135deg, #FF6D00 0%, #FFB300 50%, #FFC107 100%)',
        }}
      >
        <Toolbar sx={{ minHeight: { xs: 64, sm: 72 } }}>
          {/* Logo - Clickable */}
          <Box sx={{ display: 'flex', alignItems: 'center', flexGrow: 1 }}>
            <Tooltip title="На главную" arrow>
              <Box
                component="img"
                src="/logo.png"
                alt="YardPass"
                sx={{
                  height: { xs: 50, sm: 65 },
                  width: 'auto',
                  cursor: 'pointer',
                  filter: 'brightness(0) invert(1)', // Make logo white
                  transition: 'all 0.3s ease',
                  '&:hover': { 
                    opacity: 0.85,
                    transform: 'scale(1.05)',
                  },
                }}
                onClick={() => navigate(getHomeRoute())}
              />
            </Tooltip>
          </Box>
          
          {user && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              {/* Role Chip */}
              <Chip
                icon={getRoleIcon(user.role)}
                label={getRoleName(user.role)}
                variant="filled"
                sx={{
                  fontWeight: 700,
                  display: { xs: 'none', sm: 'flex' },
                  backgroundColor: 'rgba(255, 255, 255, 0.98)',
                  color: user.role === 'admin' ? '#E53935' : '#FF6D00',
                  boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
                  border: '2px solid rgba(255, 255, 255, 1)',
                  '& .MuiChip-icon': {
                    color: user.role === 'admin' ? '#E53935' : '#FF6D00',
                  },
                }}
              />
              
              {/* Admin Navigation */}
              {user.role === 'admin' && (
                <Box sx={{ display: 'flex', gap: 1 }}>
                  <Tooltip title="Настройка правил">
                    <Button 
                      color="inherit" 
                      onClick={() => navigate(APP_ROUTES.ADMIN_RULES)}
                      startIcon={<SettingsIcon />}
                      sx={{ 
                        fontWeight: 700,
                        '&:hover': { backgroundColor: 'rgba(255,255,255,0.2)' },
                      }}
                    >
                      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
                        Правила
                      </Box>
                    </Button>
                  </Tooltip>
                  <Tooltip title="Управление жителями">
                    <Button 
                      color="inherit" 
                      onClick={() => navigate(APP_ROUTES.ADMIN_RESIDENTS)}
                      startIcon={<PeopleIcon />}
                      sx={{ 
                        fontWeight: 700,
                        '&:hover': { backgroundColor: 'rgba(255,255,255,0.2)' },
                      }}
                    >
                      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
                        Жители
                      </Box>
                    </Button>
                  </Tooltip>
                  <Tooltip title="Отчеты и статистика">
                    <Button 
                      color="inherit" 
                      onClick={() => navigate(APP_ROUTES.ADMIN_REPORTS)}
                      startIcon={<AssessmentIcon />}
                      sx={{ 
                        fontWeight: 700,
                        '&:hover': { backgroundColor: 'rgba(255,255,255,0.2)' },
                      }}
                    >
                      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
                        Отчеты
                      </Box>
                    </Button>
                  </Tooltip>
                </Box>
              )}
              
              {/* Logout Button */}
              <Tooltip title="Выйти из системы">
                <Button 
                  color="inherit" 
                  onClick={logout} 
                  startIcon={<LogoutIcon />}
                  sx={{ 
                    fontWeight: 700,
                    '&:hover': { backgroundColor: 'rgba(255,255,255,0.2)' },
                  }}
                >
                  <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
                    Выход
                  </Box>
                </Button>
              </Tooltip>
            </Box>
          )}
        </Toolbar>
      </AppBar>

      {title && (
        <Box 
          sx={{ 
            background: user?.role === 'admin'
              ? 'linear-gradient(135deg, #E53935 0%, #FF6D00 50%, #FFB300 100%)'
              : 'linear-gradient(135deg, #FF6D00 0%, #FFB300 50%, #FFC107 100%)',
            color: 'white', 
            py: 5,
            boxShadow: '0 4px 20px rgba(229, 57, 53, 0.3)',
          }}
        >
          <Container maxWidth="lg">
            <Typography 
              variant="h3" 
              sx={{ 
                fontWeight: 800,
                color: '#FFFFFF',
                textShadow: '3px 3px 6px rgba(0,0,0,0.3)',
                letterSpacing: 0.5,
              }}
            >
              {title}
            </Typography>
          </Container>
        </Box>
      )}

      <Box>{children}</Box>
    </Box>
  );
}

