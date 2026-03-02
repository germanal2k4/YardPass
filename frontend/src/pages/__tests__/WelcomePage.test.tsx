import { screen, fireEvent } from '@testing-library/react';
import { renderWithProviders } from '@/test/helpers';
import { WelcomePage } from '../WelcomePage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('WelcomePage', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it('renders header, role cards and registration section', () => {
    renderWithProviders(<WelcomePage />);

    expect(screen.getByText('YardPass')).toBeInTheDocument();
    expect(screen.getByText('Система управления пропусками')).toBeInTheDocument();
    expect(screen.getByText('Охрана')).toBeInTheDocument();
    expect(screen.getByText('Администратор')).toBeInTheDocument();
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument();
  });

  it('navigates to login with role=guard when guard card button is clicked', () => {
    renderWithProviders(<WelcomePage />);

    fireEvent.click(screen.getByRole('button', { name: /Войти как охранник/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/login?role=guard');
  });

  it('navigates to login with role=admin when admin card button is clicked', () => {
    renderWithProviders(<WelcomePage />);

    fireEvent.click(screen.getByRole('button', { name: /Войти как администратор/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/login?role=admin');
  });

  it('navigates to login with role=guard when guard card area is clicked', () => {
    renderWithProviders(<WelcomePage />);

    fireEvent.click(screen.getByText('Сканирование и проверка QR-кодов пропусков'));
    expect(mockNavigate).toHaveBeenCalledWith('/login?role=guard');
  });

  it('navigates to login with role=admin when admin card area is clicked', () => {
    renderWithProviders(<WelcomePage />);

    fireEvent.click(screen.getByText('Настройка правил, просмотр отчетов и статистики'));
    expect(mockNavigate).toHaveBeenCalledWith('/login?role=admin');
  });

  it('navigates to registration page', () => {
    renderWithProviders(<WelcomePage />);

    fireEvent.click(screen.getByRole('button', { name: /Зарегистрироваться/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/register');
  });

  it('displays version footer', () => {
    renderWithProviders(<WelcomePage />);
    expect(screen.getByText(/Версия 1.0.0/)).toBeInTheDocument();
  });
});
