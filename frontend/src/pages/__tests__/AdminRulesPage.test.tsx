import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { AdminRulesPage } from '../AdminRulesPage';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };
const adminNoBuildingId = { user_id: 1, role: 'admin' as const };

describe('AdminRulesPage', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows error when buildingId is missing', () => {
    renderWithProviders(<AdminRulesPage />, {
      auth: { user: adminNoBuildingId },
    });
    expect(screen.getByText(/Не удалось определить ID здания/i)).toBeInTheDocument();
  });

  it('loads and displays rules data in form', async () => {
    renderWithProviders(<AdminRulesPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue('22:00')).toBeInTheDocument();
    });
    expect(screen.getByDisplayValue('08:00')).toBeInTheDocument();
    expect(screen.getByDisplayValue('5')).toBeInTheDocument();
    expect(screen.getByDisplayValue('24')).toBeInTheDocument();
  });

  it('shows building ID in subtitle', async () => {
    renderWithProviders(<AdminRulesPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText(/Здание ID: 1/i)).toBeInTheDocument();
    });
  });

  it('submits updated rules and shows success', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderWithProviders(<AdminRulesPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue('22:00')).toBeInTheDocument();
    });

    const startInput = screen.getByDisplayValue('22:00');
    await user.clear(startInput);
    await user.type(startInput, '23:00');

    await user.click(screen.getByRole('button', { name: /Сохранить изменения/i }));

    await waitFor(() => {
      expect(screen.getByText(/Правила успешно обновлены/i)).toBeInTheDocument();
    });
  });

  it('shows error state when API fails', async () => {
    server.use(
      http.get('http://localhost:8080/api/v1/rules', () => {
        return HttpResponse.json(
          { error: { code: 'INTERNAL_ERROR', message: 'Internal server error' } },
          { status: 500 },
        );
      }),
    );

    renderWithProviders(<AdminRulesPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText(/Ошибка загрузки правил/i)).toBeInTheDocument();
    });
  });
});
