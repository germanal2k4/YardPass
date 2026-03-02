import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { AdminResidentsPage } from '../AdminResidentsPage';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };
const adminNoBuildingId = { user_id: 1, role: 'admin' as const };

describe('AdminResidentsPage', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows error when buildingId is missing', () => {
    renderWithProviders(<AdminResidentsPage />, {
      auth: { user: adminNoBuildingId },
    });
    expect(screen.getByText(/Не удалось определить ID здания/i)).toBeInTheDocument();
  });

  it('shows loading state then residents table', async () => {
    renderWithProviders(<AdminResidentsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('Иван Петров')).toBeInTheDocument();
    });
    expect(screen.getByText('Мария Иванова')).toBeInTheDocument();
    expect(screen.getByText('Всего жителей: 2')).toBeInTheDocument();
  });

  it('shows table headers', async () => {
    renderWithProviders(<AdminResidentsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('Иван Петров')).toBeInTheDocument();
    });

    expect(screen.getByText('ID')).toBeInTheDocument();
    expect(screen.getByText('Квартира ID')).toBeInTheDocument();
    expect(screen.getByText('Telegram ID')).toBeInTheDocument();
  });

  describe('create form', () => {
    it('creates a resident on valid submit', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      await user.type(screen.getByLabelText(/ID квартиры/i), '201');
      await user.type(screen.getByLabelText(/Telegram ID/i), '999888777');
      await user.type(screen.getByLabelText(/Имя \(опционально\)/i), 'Тестовый Житель');

      await user.click(screen.getByRole('button', { name: /Создать жителя/i }));

      await waitFor(() => {
        expect(screen.getByText(/Житель успешно создан/i)).toBeInTheDocument();
      });
    });

    it('shows phone validation error for invalid phone', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      await user.type(screen.getByLabelText(/ID квартиры/i), '201');
      await user.type(screen.getByLabelText(/Telegram ID/i), '999888777');
      await user.type(screen.getByLabelText(/Телефон/i), 'invalid-phone');

      await user.click(screen.getByRole('button', { name: /Создать жителя/i }));

      await waitFor(() => {
        expect(screen.getByText(/Неверный формат телефона/i)).toBeInTheDocument();
      });
    });
  });

  describe('bulk JSON dialog', () => {
    it('opens dialog and shows error for invalid JSON', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', { name: /Открыть редактор JSON/i }));

      await waitFor(() => {
        expect(screen.getByText('Массовое создание жителей (JSON)')).toBeInTheDocument();
      });

      const textarea = screen.getByRole('textbox');
      await user.clear(textarea);
      await user.click(textarea);
      await user.paste('not valid json');
      await user.click(screen.getByRole('button', { name: /^Создать$/i }));

      await waitFor(() => {
        expect(screen.getByText(/Ошибка парсинга JSON/i)).toBeInTheDocument();
      });
    });

    it('shows error when JSON is not an array', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', { name: /Открыть редактор JSON/i }));

      const textarea = screen.getByRole('textbox');
      await user.clear(textarea);
      // Paste valid JSON that's an object, not an array
      await user.click(textarea);
      await user.paste('{"key":"value"}');

      await user.click(screen.getByRole('button', { name: /^Создать$/i }));

      await waitFor(() => {
        expect(screen.getByText(/JSON должен содержать массив/i)).toBeInTheDocument();
      });
    });
  });

  describe('CSV import dialog', () => {
    it('opens CSV dialog and shows file select button', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', { name: /Загрузить CSV/i }));

      await waitFor(() => {
        expect(screen.getByText('Импорт жителей из CSV')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: /Выбрать файл/i })).toBeInTheDocument();
    });
  });

  describe('delete flow', () => {
    it('opens confirmation dialog and deletes resident', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByLabelText(/Удалить жителя/i);
      await user.click(deleteButtons[0]);

      await waitFor(() => {
        expect(screen.getByText('Подтверждение удаления')).toBeInTheDocument();
      });
      expect(screen.getByText(/Вы уверены/i)).toBeInTheDocument();

      await user.click(screen.getByRole('button', { name: /^Удалить$/i }));

      await waitFor(() => {
        expect(screen.getByText(/Житель успешно удален/i)).toBeInTheDocument();
      });
    });

    it('shows error inside delete dialog on failure', async () => {
      server.use(
        http.delete('http://localhost:8080/api/v1/residents/:id', () => {
          return HttpResponse.json(
            { error: { code: 'FORBIDDEN', message: 'Forbidden' } },
            { status: 403 },
          );
        }),
      );

      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWithProviders(<AdminResidentsPage />, {
        auth: { user: adminUser },
      });

      await waitFor(() => {
        expect(screen.getByText('Иван Петров')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByLabelText(/Удалить жителя/i);
      await user.click(deleteButtons[0]);

      await waitFor(() => {
        expect(screen.getByText('Подтверждение удаления')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', { name: /^Удалить$/i }));

      await waitFor(() => {
        const dialog = screen.getByRole('dialog');
        expect(within(dialog).getByText(/Доступ запрещен/i)).toBeInTheDocument();
      });
    });
  });
});
