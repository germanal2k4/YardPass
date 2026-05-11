import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { renderWithProviders } from '@/test/helpers';
import { server } from '@/test/msw/server';
import { AdminGuardsPage } from '../AdminGuardsPage';

const API_BASE = 'http://localhost:8080';
const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };

describe('AdminGuardsPage', () => {
  it('shows only guards from admin building', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/users`, () => {
        return HttpResponse.json({
          users: [
            {
              id: 2,
              username: 'guard1',
              role: 'guard',
              status: 'active',
              created_at: '2026-01-01T00:00:00Z',
              updated_at: '2026-01-01T00:00:00Z',
            },
          ],
        });
      }),
    );

    renderWithProviders(<AdminGuardsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('guard1')).toBeInTheDocument();
    });
    expect(screen.queryByText('guard_other_building')).not.toBeInTheDocument();
  });

  it('shows row-level status when guard credentials are updated', async () => {
    const user = userEvent.setup();
    const updateSpy = vi.fn();
    server.use(
      http.get(`${API_BASE}/api/v1/users`, () => {
        return HttpResponse.json({
          users: [
            {
              id: 2,
              username: 'guard1',
              role: 'guard',
              status: 'active',
              created_at: '2026-01-01T00:00:00Z',
              updated_at: '2026-01-01T00:00:00Z',
            },
          ],
        });
      }),
      http.put(`${API_BASE}/api/v1/users/:id/credentials`, async ({ request }) => {
        updateSpy(await request.json());
        return HttpResponse.json({
          id: 2,
          username: 'guard1_new',
          role: 'guard',
          status: 'active',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: new Date().toISOString(),
        });
      }),
    );

    renderWithProviders(<AdminGuardsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('guard1')).toBeInTheDocument();
    });
    await user.clear(screen.getByLabelText('Новый логин охранника'));
    await user.type(screen.getByLabelText('Новый логин охранника'), 'guard1_new');
    await user.type(screen.getByLabelText('Новый пароль охранника'), 'new_password');
    await user.click(screen.getByRole('button', { name: 'Обновить логин/пароль' }));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalled();
      expect(screen.getByText('Данные успешно обновлены')).toBeInTheDocument();
    });
  });
});
