import { Routes, Route, Navigate } from 'react-router-dom';
import { APP_ROUTES } from '@/shared/config/constants';
import { useAuth } from '@/features/auth/useAuth';
import { WelcomePage } from '@/pages/WelcomePage';
import { LoginPage } from '@/pages/LoginPage';
import { RegistrationPage } from '@/pages/RegistrationPage';
import { SecurityPage } from '@/pages/SecurityPage';
import { AdminPage } from '@/pages/AdminPage';
import { AdminRulesPage } from '@/pages/AdminRulesPage';
import { AdminReportsPage } from '@/pages/AdminReportsPage';
import { ForbiddenPage } from '@/pages/ForbiddenPage';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: 'guard' | 'admin';
}

function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return <div>Загрузка...</div>;
  }

  if (!user) {
    return <Navigate to={APP_ROUTES.LOGIN} replace />;
  }

  if (requiredRole && user.role !== requiredRole) {
    return <Navigate to={APP_ROUTES.FORBIDDEN} replace />;
  }

  return <>{children}</>;
}

function HomeRedirect() {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return <div>Загрузка...</div>;
  }

  if (!user) {
    return <WelcomePage />;
  }

  if (user.role === 'admin') {
    return <Navigate to={APP_ROUTES.ADMIN} replace />;
  }

  return <Navigate to={APP_ROUTES.SECURITY} replace />;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path={APP_ROUTES.HOME} element={<HomeRedirect />} />
      <Route path={APP_ROUTES.LOGIN} element={<LoginPage />} />
      <Route path={APP_ROUTES.REGISTER} element={<RegistrationPage />} />
      
      <Route
        path={APP_ROUTES.SECURITY}
        element={
          <ProtectedRoute requiredRole="guard">
            <SecurityPage />
          </ProtectedRoute>
        }
      />
      
      <Route
        path={APP_ROUTES.ADMIN}
        element={
          <ProtectedRoute requiredRole="admin">
            <AdminPage />
          </ProtectedRoute>
        }
      />
      
      <Route
        path={APP_ROUTES.ADMIN_RULES}
        element={
          <ProtectedRoute requiredRole="admin">
            <AdminRulesPage />
          </ProtectedRoute>
        }
      />
      
      <Route
        path={APP_ROUTES.ADMIN_REPORTS}
        element={
          <ProtectedRoute requiredRole="admin">
            <AdminReportsPage />
          </ProtectedRoute>
        }
      />
      
      <Route path={APP_ROUTES.FORBIDDEN} element={<ForbiddenPage />} />
    </Routes>
  );
}

