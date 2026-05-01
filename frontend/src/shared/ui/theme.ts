import { createTheme } from '@mui/material/styles';

// YardPass color scheme: White-Red-Orange
export const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#E53935', // Deep Red (like in logo)
      light: '#FF6F60',
      dark: '#B71C1C',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#FF6D00', // Vibrant Orange
      light: '#FF9E40',
      dark: '#C43E00',
      contrastText: '#FFFFFF',
    },
    success: {
      main: '#43A047', // Keep green for success states
    },
    error: {
      main: '#D32F2F', // Red for errors
    },
    warning: {
      main: '#FF6D00', // Orange for warnings
    },
    background: {
      default: '#FAFAFA', // Light gray (almost white)
      paper: '#FFFFFF',
    },
    text: {
      primary: '#263238', // Dark gray for text
      secondary: '#546E7A', // Medium gray
    },
  },
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    h1: {
      fontSize: '2.5rem',
      fontWeight: 700,
      color: '#263238',
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 700,
      color: '#263238',
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 700,
      color: '#263238',
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 600,
      color: '#263238',
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 600,
      color: '#263238',
    },
    h6: {
      fontSize: '1rem',
      fontWeight: 600,
      color: '#263238',
    },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          borderRadius: 0,
          background: 'linear-gradient(135deg, #E53935 0%, #FF6D00 100%)',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          borderRadius: 12,
          fontWeight: 600,
        },
        sizeLarge: {
          padding: '14px 28px',
          fontSize: '1.1rem',
        },
        contained: {
          boxShadow: '0 4px 12px rgba(229, 57, 53, 0.3)',
          '&:hover': {
            boxShadow: '0 6px 20px rgba(229, 57, 53, 0.4)',
          },
        },
        containedSecondary: {
          boxShadow: '0 4px 12px rgba(255, 109, 0, 0.3)',
          '&:hover': {
            boxShadow: '0 6px 20px rgba(255, 109, 0, 0.4)',
          },
        },
      },
    },
    MuiTextField: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          boxShadow: '0 4px 20px rgba(0, 0, 0, 0.08)',
        },
        elevation3: {
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 600,
          borderRadius: 8,
        },
        colorPrimary: {
          background: 'linear-gradient(135deg, #E53935 0%, #FF6F60 100%)',
        },
        colorSecondary: {
          background: 'linear-gradient(135deg, #FF6D00 0%, #FF9E40 100%)',
        },
      },
    },
  },
});

