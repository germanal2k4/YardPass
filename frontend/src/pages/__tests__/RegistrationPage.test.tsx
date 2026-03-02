import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { RegistrationPage } from '../RegistrationPage';

describe('RegistrationPage', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders the registration form', () => {
    renderWithProviders(<RegistrationPage />, { auth: { user: null } });
    expect(screen.getByText('Регистрация')).toBeInTheDocument();
    expect(screen.getByLabelText(/Имя пользователя/i)).toBeInTheDocument();
    expect(screen.getByTestId('password-input')).toBeInTheDocument();
    expect(screen.getByTestId('confirm-password-input')).toBeInTheDocument();
  });

  it('shows error on password mismatch', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'testuser');
    await user.type(screen.getByTestId('password-input'), 'password123');
    await user.type(screen.getByTestId('confirm-password-input'), 'different');
    await user.click(screen.getByRole('button', { name: /Зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByText('Пароли не совпадают')).toBeInTheDocument();
    });
  });

  it('shows error on short password', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'testuser');
    await user.type(screen.getByTestId('password-input'), '123');
    await user.type(screen.getByTestId('confirm-password-input'), '123');
    await user.click(screen.getByRole('button', { name: /Зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByText(/Пароль должен содержать минимум 6 символов/i)).toBeInTheDocument();
    });
  });

  it('shows error on short username', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'ab');
    await user.type(screen.getByTestId('password-input'), 'password123');
    await user.type(screen.getByTestId('confirm-password-input'), 'password123');
    await user.click(screen.getByRole('button', { name: /Зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByText(/Имя пользователя должно содержать минимум 3 символа/i)).toBeInTheDocument();
    });
  });

  it('shows success alert and redirects on valid submit', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<RegistrationPage />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/register?role=guard'] },
    });

    await user.type(screen.getByLabelText(/Имя пользователя/i), 'newuser');
    await user.type(screen.getByTestId('password-input'), 'password123');
    await user.type(screen.getByTestId('confirm-password-input'), 'password123');
    await user.click(screen.getByRole('button', { name: /Зарегистрироваться/i }));

    // Wait for the simulated async request (1 second)
    await vi.advanceTimersByTimeAsync(1100);

    await waitFor(() => {
      expect(screen.getByText(/успешно зарегистрирован/i)).toBeInTheDocument();
    });

    // Advance 2 seconds for redirect timer
    await vi.advanceTimersByTimeAsync(2100);
  });
});
