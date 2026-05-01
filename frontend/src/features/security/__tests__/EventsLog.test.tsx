import { screen, waitFor, fireEvent } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { renderWithProviders } from '@/test/helpers';
import { EventsLog } from '../EventsLog';
import { server } from '@/test/msw/server';

const API_BASE = 'http://localhost:8080';

describe('EventsLog', () => {
  it('renders header and refresh button', async () => {
    renderWithProviders(<EventsLog />);

    expect(screen.getByText('Журнал событий')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Обновить/i })).toBeInTheDocument();
  });

  it('shows loading state then displays events', async () => {
    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getAllByText('guard1').length).toBeGreaterThanOrEqual(1);
    });

    expect(screen.getByText('Действителен')).toBeInTheDocument();
    expect(screen.getByText('Недействителен')).toBeInTheDocument();
    expect(screen.getByText('А123ВС777')).toBeInTheDocument();
    expect(screen.getByText('Гость 1')).toBeInTheDocument();
  });

  it('shows apartment info with building name', async () => {
    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getByText(/Дом 1 № 101/)).toBeInTheDocument();
    });
  });

  it('shows "no events" when data is empty', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/scan-events`, () => {
        return HttpResponse.json({ events: [], limit: 20, offset: 0 });
      }),
    );

    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getByText('События отсутствуют')).toBeInTheDocument();
    });
  });

  it('shows error alert on fetch failure', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/scan-events`, () => {
        return HttpResponse.json({ error: 'Server error' }, { status: 500 });
      }),
    );

    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getByText(/Ошибка при загрузке данных/)).toBeInTheDocument();
    });
  });

  it('calls refetch when refresh button is clicked', async () => {
    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getAllByText('guard1').length).toBeGreaterThanOrEqual(1);
    });

    const refreshBtn = screen.getByRole('button', { name: /Обновить/i });
    fireEvent.click(refreshBtn);

    await waitFor(() => {
      expect(screen.getAllByText('guard1').length).toBeGreaterThanOrEqual(1);
    });
  });

  it('displays dash for missing optional fields', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/scan-events`, () => {
        return HttpResponse.json({
          events: [
            {
              ID: 10,
              PassID: 'uuid-10',
              GuardUserID: 2,
              GuardUsername: 'guard2',
              ScannedAt: '2026-01-16T10:00:00Z',
              Result: 'invalid',
            },
          ],
          limit: 20,
          offset: 0,
        });
      }),
    );

    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getByText('guard2')).toBeInTheDocument();
    });

    const dashes = screen.getAllByText('—');
    expect(dashes.length).toBeGreaterThanOrEqual(3);
  });

  it('renders pagination controls', async () => {
    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getAllByText('guard1').length).toBeGreaterThanOrEqual(1);
    });

    expect(screen.getByText('Строк на странице:')).toBeInTheDocument();
  });

  it('handles invalid date gracefully', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/scan-events`, () => {
        return HttpResponse.json({
          events: [
            {
              ID: 11,
              PassID: 'uuid-11',
              GuardUserID: 2,
              GuardUsername: 'guard3',
              ScannedAt: 'invalid-date',
              Result: 'valid',
            },
          ],
          limit: 20,
          offset: 0,
        });
      }),
    );

    renderWithProviders(<EventsLog />);

    await waitFor(() => {
      expect(screen.getByText('guard3')).toBeInTheDocument();
    });

    const dateCells = screen.getAllByText('—');
    expect(dateCells.length).toBeGreaterThanOrEqual(1);
  });
});
