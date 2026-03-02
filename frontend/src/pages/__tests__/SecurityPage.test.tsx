import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { SecurityPage } from '../SecurityPage';

const guardUser = { user_id: 2, role: 'guard' as const, building_id: 1 };

describe('SecurityPage', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders QR and car plate sections', () => {
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });
    expect(screen.getByText('Проверка QR-кода')).toBeInTheDocument();
    expect(screen.getByText('Проверка по номеру')).toBeInTheDocument();
  });

  it('validates QR code on Enter', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });

    const input = screen.getByPlaceholderText(/Сканируйте QR-код/i);
    await user.type(input, 'valid-uuid{Enter}');

    await waitFor(() => {
      expect(screen.getByText('Пропуск действителен')).toBeInTheDocument();
    });
  });

  it('shows error for unknown QR', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });

    const input = screen.getByPlaceholderText(/Сканируйте QR-код/i);
    await user.type(input, 'unknown-uuid{Enter}');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('clears error on close button', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });

    const input = screen.getByPlaceholderText(/Сканируйте QR-код/i);
    await user.type(input, 'unknown-uuid{Enter}');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    const closeBtn = screen.getByLabelText('close');
    await user.click(closeBtn);

    await waitFor(() => {
      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
  });

  it('shows invalid pass result for expired UUID', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });

    const input = screen.getByPlaceholderText(/Сканируйте QR-код/i);
    await user.type(input, 'expired-uuid{Enter}');

    await waitFor(() => {
      expect(screen.getByText('Пропуск недействителен')).toBeInTheDocument();
    });
  });

  it('renders car plate input', () => {
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });
    expect(screen.getByText('Проверить номер')).toBeInTheDocument();
  });

  it('car plate button is disabled when empty', () => {
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });
    const btn = screen.getByRole('button', { name: /Проверить номер/i });
    expect(btn).toBeDisabled();
  });

  it('clears validation result when user starts typing again', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<SecurityPage />, { auth: { user: guardUser } });

    const input = screen.getByPlaceholderText(/Сканируйте QR-код/i);
    await user.type(input, 'valid-uuid{Enter}');

    await waitFor(() => {
      expect(screen.getByText('Пропуск действителен')).toBeInTheDocument();
    });

    await user.type(input, 'a');

    await waitFor(() => {
      expect(screen.queryByText('Пропуск действителен')).not.toBeInTheDocument();
    });
  });
});
