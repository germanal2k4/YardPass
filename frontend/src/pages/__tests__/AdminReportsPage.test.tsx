import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/helpers';
import { AdminReportsPage } from '../AdminReportsPage';
import { reportsApi } from '@/shared/api/reports';

const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };

describe('AdminReportsPage', () => {
  let createObjectURLSpy: ReturnType<typeof vi.spyOn>;
  let revokeObjectURLSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:http://localhost/fake-url');
    revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {});
  });

  afterEach(() => {
    createObjectURLSpy.mockRestore();
    revokeObjectURLSpy.mockRestore();
  });

  it('renders statistics cards', async () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument();
    });
    expect(screen.getByText('120')).toBeInTheDocument();
    expect(screen.getByText('30')).toBeInTheDocument();
    expect(screen.getByText('80')).toBeInTheDocument();
  });

  it('renders stat card labels', async () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('Всего сканирований')).toBeInTheDocument();
    });
    expect(screen.getByText('Действительных')).toBeInTheDocument();
    expect(screen.getByText('Недействительных')).toBeInTheDocument();
    expect(screen.getByText('Уникальных пропусков')).toBeInTheDocument();
  });

  it('renders top reasons chips', async () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('PASS_EXPIRED: 15')).toBeInTheDocument();
    });
    expect(screen.getByText('PASS_NOT_FOUND: 10')).toBeInTheDocument();
  });

  it('renders events table with data', async () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('А123ВС777')).toBeInTheDocument();
    });
    expect(screen.getByText('Гость 1')).toBeInTheDocument();
    expect(screen.getAllByText('guard1').length).toBeGreaterThanOrEqual(1);
  });

  it('renders filter chips', () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    expect(screen.getByText('Все')).toBeInTheDocument();
    expect(screen.getByText('Действительные')).toBeInTheDocument();
    expect(screen.getByText('Недействительные')).toBeInTheDocument();
  });

  it('clicking filter chip updates selection', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const validChip = screen.getByText('Действительные');
    await user.click(validChip);

    // The chip should now be visually active (MUI Chip with color="success")
    expect(validChip.closest('.MuiChip-root')).toHaveClass('MuiChip-colorSuccess');
  });

  it('triggers export and creates download link', async () => {
    const user = userEvent.setup();
    const exportSpy = vi.spyOn(reportsApi, 'exportReport').mockResolvedValueOnce(new Blob(['fake']));
    const appendChildSpy = vi.spyOn(document.body, 'appendChild');
    const removeChildSpy = vi.spyOn(document.body, 'removeChild');

    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('button', { name: /Экспорт/i }));

    await waitFor(() => {
      expect(createObjectURLSpy).toHaveBeenCalled();
    });
    expect(appendChildSpy).toHaveBeenCalled();
    expect(removeChildSpy).toHaveBeenCalled();

    appendChildSpy.mockRestore();
    removeChildSpy.mockRestore();
    exportSpy.mockRestore();
  });

  it('renders page sections', async () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    expect(screen.getByText('Фильтры')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Журнал событий')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByText('Основные причины отказов')).toBeInTheDocument();
    });
  });

  it('renders date filter fields', () => {
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    expect(screen.getByLabelText('Дата от')).toBeInTheDocument();
    expect(screen.getByLabelText('Дата до')).toBeInTheDocument();
    expect(screen.getByLabelText('Номер квартиры')).toBeInTheDocument();
    expect(screen.getByLabelText('Номер авто')).toBeInTheDocument();
  });

  it('changes date from filter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const dateFrom = screen.getByLabelText('Дата от');
    await user.type(dateFrom, '2026-01-01T00:00');
    expect(dateFrom).toHaveValue('2026-01-01T00:00');
  });

  it('changes date to filter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const dateTo = screen.getByLabelText('Дата до');
    await user.type(dateTo, '2026-01-31T23:59');
    expect(dateTo).toHaveValue('2026-01-31T23:59');
  });

  it('changes apartment and car plate filters', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const apartmentInput = screen.getByLabelText('Номер квартиры');
    const carPlateInput = screen.getByLabelText('Номер авто');

    await user.type(apartmentInput, '101');
    await user.type(carPlateInput, 'А123ВС777');

    expect(apartmentInput).toHaveValue('101');
    expect(carPlateInput).toHaveValue('А123ВС777');
  });

  it('reset filters button clears all filter fields', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const dateFrom = screen.getByLabelText('Дата от');
    const dateTo = screen.getByLabelText('Дата до');
    const apartmentInput = screen.getByLabelText('Номер квартиры');
    const carPlateInput = screen.getByLabelText('Номер авто');

    await user.type(dateFrom, '2026-01-01T00:00');
    await user.type(dateTo, '2026-01-31T23:59');
    await user.type(apartmentInput, '101');
    await user.type(carPlateInput, 'А123ВС777');
    await user.click(screen.getByText('Действительные'));

    await user.click(screen.getByRole('button', { name: 'Сбросить фильтры' }));

    expect(dateFrom).toHaveValue('');
    expect(dateTo).toHaveValue('');
    expect(apartmentInput).toHaveValue('');
    expect(carPlateInput).toHaveValue('');
    expect(screen.getByText('Все').closest('.MuiChip-root')).toHaveClass('MuiChip-colorPrimary');
  });

  it('refresh button refetches data', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('button', { name: /Обновить/i }));

    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument();
    });
  });

  it('shows export error on failure', async () => {
    const user = userEvent.setup();
    const exportSpy = vi.spyOn(reportsApi, 'exportReport').mockRejectedValueOnce(new Error('Network error'));

    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await waitFor(() => {
      expect(screen.getByText('150')).toBeInTheDocument();
    });

    await user.click(screen.getByRole('button', { name: /Экспорт/i }));

    await waitFor(() => {
      expect(screen.getByText(/Ошибка при экспорте отчета/)).toBeInTheDocument();
    });

    exportSpy.mockRestore();
  });

  it('clicking invalid filter chip switches active filter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    const invalidChip = screen.getByText('Недействительные');
    await user.click(invalidChip);

    expect(invalidChip.closest('.MuiChip-root')).toHaveClass('MuiChip-colorError');
  });

  it('clicking "Все" filter chip resets filter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AdminReportsPage />, {
      auth: { user: adminUser },
    });

    await user.click(screen.getByText('Действительные'));
    await user.click(screen.getByText('Все'));

    expect(screen.getByText('Все').closest('.MuiChip-root')).toHaveClass('MuiChip-colorPrimary');
  });
});
