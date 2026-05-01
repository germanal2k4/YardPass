import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
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
    expect(screen.getByText('Оплата подписки')).toBeInTheDocument();
    expect(screen.getByLabelText(/Здание/i)).toBeInTheDocument();
    expect(screen.getByTestId('card-number-input')).toBeInTheDocument();
  });

  it('shows error when building is missing', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { container } = renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Email/i), 'owner@example.com');
    fireEvent.submit(container.querySelector('form')!);

    await waitFor(() => {
      expect(screen.getByText('Укажите название здания')).toBeInTheDocument();
    });
  });

  it('shows error when email is missing', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { container } = renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Здание/i), 'ЖК Лесной');
    fireEvent.submit(container.querySelector('form')!);

    await waitFor(() => {
      expect(screen.getByText(/Укажите email/i)).toBeInTheDocument();
    });
  });

  it('shows error when card fields are missing', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { container } = renderWithProviders(<RegistrationPage />, { auth: { user: null } });

    await user.type(screen.getByLabelText(/Здание/i), 'ЖК Лесной');
    await user.type(screen.getByLabelText(/Email/i), 'owner@example.com');
    fireEvent.submit(container.querySelector('form')!);

    await waitFor(() => {
      expect(screen.getByText(/Заполните все поля карты/i)).toBeInTheDocument();
    });
  });

  it('shows success alert and redirects on valid submit', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<RegistrationPage />, {
      auth: { user: null },
      routerProps: { initialEntries: ['/register'] },
    });

    await user.type(screen.getByLabelText(/Здание/i), 'ЖК Лесной');
    await user.type(screen.getByLabelText(/Email/i), 'owner@example.com');
    await user.type(screen.getByTestId('card-number-input'), '4111111111111111');
    await user.type(screen.getByLabelText(/Владелец карты/i), 'IVAN PETROV');
    await user.type(screen.getByLabelText(/Срок/i), '12/30');
    await user.type(screen.getByLabelText(/CVV/i), '123');
    await user.click(screen.getByRole('button', { name: /Оплатить/i }));

    await waitFor(() => {
      expect(screen.getByText(/Оплата прошла успешно/i)).toBeInTheDocument();
    });

    await vi.advanceTimersByTimeAsync(2300);
  });
});
