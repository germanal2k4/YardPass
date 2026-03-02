import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AuthProvider } from '../AuthProvider';
import { useAuth } from '../useAuth';
import { STORAGE_KEYS } from '@/shared/config/constants';

function TestConsumer() {
  const { user, isLoading, login, logout } = useAuth();
  if (isLoading) return <div>loading</div>;
  if (!user) return (
    <div>
      <span>no user</span>
      <button onClick={() => login({ username: 'admin', password: 'password' })}>login</button>
    </div>
  );
  return (
    <div>
      <span>user:{user.role}</span>
      <button onClick={logout}>logout</button>
    </div>
  );
}

function renderAuth() {
  return render(
    <MemoryRouter>
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  localStorage.clear();
});

describe('AuthProvider', () => {
  it('shows isLoading=false and user=null when no token', async () => {
    renderAuth();
    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });
  });

  it('fetches user when access token exists', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'mock-access-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'mock-refresh-token');

    renderAuth();

    // Should show loading first
    expect(screen.getByText('loading')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('user:admin')).toBeInTheDocument();
    });
  });

  it('clears tokens when getMe fails', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'invalid-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'invalid-refresh');

    const { server } = await import('@/test/msw/server');
    const { http, HttpResponse } = await import('msw');
    server.use(
      http.get('http://localhost:8080/api/v1/me', () => {
        return HttpResponse.json(
          { error: { code: 'INVALID_TOKEN', message: 'Invalid or missing token' } },
          { status: 401 },
        );
      }),
    );

    renderAuth();

    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });

    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBeNull();
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBeNull();
  });

  it('login sets tokens, fetches user and navigates', async () => {
    renderAuth();

    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });

    await act(async () => {
      screen.getByText('login').click();
    });

    await waitFor(() => {
      expect(screen.getByText('user:admin')).toBeInTheDocument();
    });

    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBe('mock-access-token');
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBe('mock-refresh-token');
  });

  it('logout clears user and tokens', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'mock-access-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'mock-refresh-token');

    renderAuth();

    await waitFor(() => {
      expect(screen.getByText('user:admin')).toBeInTheDocument();
    });

    await act(async () => {
      screen.getByText('logout').click();
    });

    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });

    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBeNull();
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBeNull();
  });
});
