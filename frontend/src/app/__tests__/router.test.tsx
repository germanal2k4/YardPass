import { describe, it, expect } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders } from '@/test/helpers';
import { AppRouter } from '../router';

const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };
const guardUser = { user_id: 2, role: 'guard' as const, building_id: 1 };

describe('ProtectedRoute', () => {
  it('redirects to /login when user is not authenticated', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/security'] },
    });
    expect(screen.getByText('Вход в систему')).toBeInTheDocument();
  });

  it('redirects to /forbidden when role does not match', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: guardUser },
      routerProps: { initialEntries: ['/admin'] },
    });
    expect(screen.getByText(/запрещен|Forbidden|Доступ/i)).toBeInTheDocument();
  });

  it('renders guard page for guard user', async () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: guardUser },
      routerProps: { initialEntries: ['/security'] },
    });
    await waitFor(() => {
      expect(screen.getByText('Сканирование пропусков')).toBeInTheDocument();
    });
  });

  it('shows loading indicator when isLoading is true', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null, isLoading: true },
      routerProps: { initialEntries: ['/security'] },
    });
    expect(screen.getByText('Загрузка...')).toBeInTheDocument();
  });
});

describe('HomeRedirect', () => {
  it('shows WelcomePage for unauthenticated user', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/'] },
    });
    expect(screen.queryByText('Загрузка...')).not.toBeInTheDocument();
  });

  it('redirects admin to /admin', async () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: adminUser },
      routerProps: { initialEntries: ['/'] },
    });
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Панель администратора/i })).toBeInTheDocument();
    });
  });

  it('redirects guard to /security', async () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: guardUser },
      routerProps: { initialEntries: ['/'] },
    });
    await waitFor(() => {
      expect(screen.getByText('Сканирование пропусков')).toBeInTheDocument();
    });
  });

  it('shows loading when isLoading', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null, isLoading: true },
      routerProps: { initialEntries: ['/'] },
    });
    expect(screen.getByText('Загрузка...')).toBeInTheDocument();
  });
});

describe('Public routes', () => {
  it('renders LoginPage at /login', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/login'] },
    });
    expect(screen.getByText('Вход в систему')).toBeInTheDocument();
  });

  it('renders RegistrationPage at /register', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/register'] },
    });
    expect(screen.getByText('Оплата подписки')).toBeInTheDocument();
  });

  it('renders ForbiddenPage at /forbidden', () => {
    renderWithProviders(<AppRouter />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/forbidden'] },
    });
    expect(screen.getByText(/запрещен|Forbidden|Доступ/i)).toBeInTheDocument();
  });
});
