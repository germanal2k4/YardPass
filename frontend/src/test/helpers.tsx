import { ReactNode } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { MemoryRouter, MemoryRouterProps } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '@/shared/ui/theme';
import { AuthContext } from '@/features/auth/AuthProvider';
import type { MeResponse, LoginRequest } from '@/shared/types/api';

export interface AuthOverrides {
  user?: MeResponse | null;
  isLoading?: boolean;
  login?: (credentials: LoginRequest) => Promise<void>;
  logout?: () => void | Promise<void>;
}

interface WrapperOptions {
  auth?: AuthOverrides;
  routerProps?: MemoryRouterProps;
  queryClient?: QueryClient;
}

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

function createWrapper({ auth, routerProps, queryClient }: WrapperOptions = {}) {
  const qc = queryClient ?? createTestQueryClient();

  const authValue = {
    user: auth?.user ?? null,
    isLoading: auth?.isLoading ?? false,
    login: auth?.login ?? vi.fn(),
    logout: auth?.logout ?? vi.fn(),
  };

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={qc}>
        <ThemeProvider theme={theme}>
          <MemoryRouter {...routerProps}>
            <AuthContext.Provider value={authValue}>
              {children}
            </AuthContext.Provider>
          </MemoryRouter>
        </ThemeProvider>
      </QueryClientProvider>
    );
  };
}

export function renderWithProviders(
  ui: React.ReactElement,
  options?: WrapperOptions & Omit<RenderOptions, 'wrapper'>,
) {
  const { auth, routerProps, queryClient: passedQC, ...renderOptions } = options ?? {};
  const queryClient = passedQC ?? createTestQueryClient();
  const Wrapper = createWrapper({ auth, routerProps, queryClient });
  return {
    ...render(ui, { wrapper: Wrapper, ...renderOptions }),
    queryClient,
  };
}
