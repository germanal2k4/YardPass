import { screen, fireEvent } from '@testing-library/react';
import { renderWithProviders } from '@/test/helpers';
import { ForbiddenPage } from '../ForbiddenPage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('ForbiddenPage', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it('renders forbidden message', () => {
    renderWithProviders(<ForbiddenPage />);

    expect(screen.getByText('Доступ запрещен')).toBeInTheDocument();
    expect(screen.getByText('У вас нет прав для просмотра этой страницы')).toBeInTheDocument();
  });

  it('renders "На главную" button', () => {
    renderWithProviders(<ForbiddenPage />);
    expect(screen.getByRole('button', { name: /На главную/i })).toBeInTheDocument();
  });

  it('navigates home when button is clicked', () => {
    renderWithProviders(<ForbiddenPage />);

    fireEvent.click(screen.getByRole('button', { name: /На главную/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });
});
