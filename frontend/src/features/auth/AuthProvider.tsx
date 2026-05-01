import { createContext, useState, useEffect, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '@/shared/api/auth';
import { APP_ROUTES } from '@/shared/config/constants';
import type { MeResponse, LoginRequest } from '@/shared/types/api';

interface AuthContextType {
  user: MeResponse | null;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const initAuth = async () => {
      try {
        const userData = await authApi.getMe();
        setUser(userData);
      } catch {
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = async (credentials: LoginRequest) => {
    await authApi.login(credentials);

    const userData = await authApi.getMe();
    setUser(userData);

    if (userData.role === 'admin') {
      navigate(APP_ROUTES.ADMIN, { replace: true });
    } else if (userData.role === 'guard') {
      navigate(APP_ROUTES.SECURITY, { replace: true });
    } else {
      navigate(APP_ROUTES.HOME, { replace: true });
    }
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } catch {
      // still leave the app logged out locally
    }
    setUser(null);
    navigate(APP_ROUTES.LOGIN);
  };

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
