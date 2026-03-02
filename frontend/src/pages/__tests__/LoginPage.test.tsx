import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { LoginPage } from '../LoginPage';

describe('LoginPage', () => {
  it('renders login form', () => {
    renderWithProviders(<LoginPage />, { auth: { user: null } });
    expect(screen.getByText('YardPass')).toBeInTheDocument();
    expect(screen.getByText('Вход в систему')).toBeInTheDocument();
    expect(screen.getByLabelText(/Имя пользователя/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Пароль/i)).toBeInTheDocument();
  });

  it('shows admin chip when ?role=admin', () => {
    renderWithProviders(<LoginPage />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/login?role=admin'] },
    });
    expect(screen.getByText('Администратор')).toBeInTheDocument();
  });

  it('shows guard chip when ?role=guard', () => {
    renderWithProviders(<LoginPage />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/login?role=guard'] },
    });
    expect(screen.getByText('Охранник')).toBeInTheDocument();
  });

  it('calls login on form submit', async () => {
    const user = userEvent.setup();
    const login = vi.fn().mockResolvedValue(undefined);

    renderWithProviders(<LoginPage />, {
      auth: { user: null, login },
    });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'admin');
    await user.type(screen.getByLabelText(/Пароль/i), 'password');
    await user.click(screen.getByRole('button', { name: /Войти/i }));

    expect(login).toHaveBeenCalledWith({ username: 'admin', password: 'password' });
  });

  it('shows error when login fails', async () => {
    const user = userEvent.setup();
    const login = vi.fn().mockRejectedValue({
      response: {
        status: 401,
        data: {
          error: {
            code: 'INVALID_CREDENTIALS',
            message: 'Invalid username or password',
          },
        },
      },
      isAxiosError: true,
    });

    renderWithProviders(<LoginPage />, {
      auth: { user: null, login },
    });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'wrong');
    await user.type(screen.getByLabelText(/Пароль/i), 'wrong');
    await user.click(screen.getByRole('button', { name: /Войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('preserves role in register link', () => {
    renderWithProviders(<LoginPage />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/login?role=admin'] },
    });
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument();
  });
});
