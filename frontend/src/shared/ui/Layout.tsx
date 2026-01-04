import { ReactNode } from 'react';
import { 
  Box, 
  AppBar, 
  Toolbar, 
  Typography, 
  Button, 
  Container,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import { useAuth } from '@/features/auth/useAuth';
import LogoutIcon from '@mui/icons-material/Logout';
import HomeIcon from '@mui/icons-material/Home';
import SettingsIcon from '@mui/icons-material/Settings';
import AssessmentIcon from '@mui/icons-material/Assessment';
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

  const getRoleColor = (role: string) => {
    return role === 'admin' ? 'secondary' : 'primary';
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
          boxShadow: 3,
          borderRadius: 0,
          background: user?.role === 'admin' 
            ? 'linear-gradient(45deg, #1976d2 30%, #42a5f5 90%)'
            : 'linear-gradient(45deg, #2196f3 30%, #21cbf3 90%)',
        }}
      >
        <Toolbar sx={{ minHeight: { xs: 56, sm: 64 } }}>
          {/* Logo/Brand */}
          <Box sx={{ display: 'flex', alignItems: 'center', flexGrow: 1 }}>
            <Tooltip title="На главную">
              <IconButton
                color="inherit"
                onClick={() => navigate(getHomeRoute())}
                sx={{ mr: 1 }}
              >
                <HomeIcon />
              </IconButton>
            </Tooltip>
            <Typography 
              variant="h6" 
              component="div" 
              sx={{ 
                fontWeight: 700,
                letterSpacing: 1,
                cursor: 'pointer',
                '&:hover': { opacity: 0.8 },
              }} 
              onClick={() => navigate(getHomeRoute())}
            >
              YardPass
            </Typography>
          </Box>
          
          {user && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              {/* Role Chip */}
              <Chip
                icon={getRoleIcon(user.role)}
                label={getRoleName(user.role)}
                color={getRoleColor(user.role)}
                variant="filled"
                sx={{
                  fontWeight: 600,
                  display: { xs: 'none', sm: 'flex' },
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
                        fontWeight: 600,
                        '&:hover': { backgroundColor: 'rgba(255,255,255,0.2)' },
                      }}
                    >
                      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
                        Правила
                      </Box>
                    </Button>
                  </Tooltip>
                  <Tooltip title="Отчеты и статистика">
                    <Button 
                      color="inherit" 
                      onClick={() => navigate(APP_ROUTES.ADMIN_REPORTS)}
                      startIcon={<AssessmentIcon />}
                      sx={{ 
                        fontWeight: 600,
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
                    fontWeight: 600,
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
              ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
              : 'linear-gradient(135deg, #667eea 0%, #42a5f5 100%)',
            color: 'white', 
            py: 4,
            boxShadow: 2,
          }}
        >
          <Container maxWidth="lg">
            <Typography 
              variant="h4" 
              sx={{ 
                fontWeight: 700,
                textShadow: '2px 2px 4px rgba(0,0,0,0.2)',
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

