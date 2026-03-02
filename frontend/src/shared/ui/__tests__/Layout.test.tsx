import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Layout } from '../Layout';
import { renderWithProviders } from '@/test/helpers';

const adminUser = { user_id: 1, role: 'admin' as const, building_id: 1 };
const guardUser = { user_id: 2, role: 'guard' as const, building_id: 1 };

describe('Layout', () => {
  it('renders children', () => {
    renderWithProviders(
      <Layout><div>Hello</div></Layout>,
      { auth: { user: adminUser } },
    );
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });

  it('renders title when provided', () => {
    renderWithProviders(
      <Layout title="Dashboard"><div /></Layout>,
      { auth: { user: adminUser } },
    );
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('shows admin navigation buttons for admin role', () => {
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: adminUser } },
    );
    expect(screen.getByText('Правила')).toBeInTheDocument();
    expect(screen.getByText('Жители')).toBeInTheDocument();
    expect(screen.getByText('Отчеты')).toBeInTheDocument();
  });

  it('hides admin navigation for guard role', () => {
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: guardUser } },
    );
    expect(screen.queryByText('Правила')).not.toBeInTheDocument();
    expect(screen.queryByText('Жители')).not.toBeInTheDocument();
    expect(screen.queryByText('Отчеты')).not.toBeInTheDocument();
  });

  it('shows role chip for admin', () => {
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: adminUser } },
    );
    expect(screen.getByText('Администратор')).toBeInTheDocument();
  });

  it('shows role chip for guard', () => {
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: guardUser } },
    );
    expect(screen.getByText('Охрана')).toBeInTheDocument();
  });

  it('calls logout when Выход button is clicked', async () => {
    const user = userEvent.setup();
    const logout = vi.fn();
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: adminUser, logout } },
    );

    await user.click(screen.getByText('Выход'));
    expect(logout).toHaveBeenCalled();
  });

  it('does not show user controls when no user', () => {
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: null } },
    );
    expect(screen.queryByText('Выход')).not.toBeInTheDocument();
    expect(screen.queryByText('Администратор')).not.toBeInTheDocument();
  });

  it('navigates to admin home when logo is clicked (admin)', async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: adminUser } },
    );

    const logo = screen.getByAltText('YardPass');
    await user.click(logo);
  });

  it('navigates to security home when logo is clicked (guard)', async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: guardUser } },
    );

    const logo = screen.getByAltText('YardPass');
    await user.click(logo);
  });

  it('navigates to home route "/" when no user and logo clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: null } },
    );

    const logo = screen.getByAltText('YardPass');
    await user.click(logo);
  });

  it('admin nav buttons are clickable', async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <Layout><div /></Layout>,
      { auth: { user: adminUser } },
    );

    await user.click(screen.getByTestId('nav-rules'));
    await user.click(screen.getByTestId('nav-residents'));
    await user.click(screen.getByTestId('nav-reports'));
  });

  it('renders title banner with guard gradient', () => {
    renderWithProviders(
      <Layout title="Guard View"><div /></Layout>,
      { auth: { user: guardUser } },
    );
    expect(screen.getByText('Guard View')).toBeInTheDocument();
  });
});
