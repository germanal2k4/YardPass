import { screen, fireEvent } from '@testing-library/react';
import { renderWithProviders } from '@/test/helpers';
import { AdminPage } from '../AdminPage';
import { mockUser } from '@/test/msw/handlers';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('AdminPage', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  const adminAuth = { user: mockUser };

  it('renders three section cards', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });

    expect(screen.getByText('Правила и настройки')).toBeInTheDocument();
    expect(screen.getByText(/Настройка тихих часов/)).toBeInTheDocument();
    expect(screen.getByText(/Управление жителями/)).toBeInTheDocument();
    expect(screen.getByText(/Просмотр статистики/)).toBeInTheDocument();
  });

  it('renders section descriptions', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });

    expect(screen.getByText(/Настройка тихих часов/)).toBeInTheDocument();
    expect(screen.getByText(/Управление жителями, добавление новых/)).toBeInTheDocument();
    expect(screen.getByText(/Просмотр статистики, журнала событий/)).toBeInTheDocument();
  });

  it('navigates to rules page when rules card is clicked', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });

    fireEvent.click(screen.getByText('Правила и настройки'));
    expect(mockNavigate).toHaveBeenCalledWith('/admin/rules');
  });

  it('navigates to residents page when residents card description is clicked', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });

    fireEvent.click(screen.getByText(/Управление жителями, добавление новых/));
    expect(mockNavigate).toHaveBeenCalledWith('/admin/residents');
  });

  it('navigates to reports page when reports card description is clicked', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });

    fireEvent.click(screen.getByText(/Просмотр статистики, журнала событий/));
    expect(mockNavigate).toHaveBeenCalledWith('/admin/reports');
  });

  it('renders layout with admin title', () => {
    renderWithProviders(<AdminPage />, { auth: adminAuth });
    expect(screen.getByText('Панель администратора')).toBeInTheDocument();
  });
});
