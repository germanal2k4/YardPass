import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AuthProvider } from '../AuthProvider';
import { useAuth } from '../useAuth';
import { server } from '@/test/msw/server';
import { handlers, mockUser } from '@/test/msw/handlers';
import { http, HttpResponse } from 'msw';

function TestConsumer() {
  const { user, isLoading, login, logout } = useAuth();
  if (isLoading) return <div>loading</div>;
  if (!user)
    return (
      <div>
        <span>no user</span>
        <button onClick={() => login({ username: 'admin', password: 'password' })}>login</button>
      </div>
    );
  return (
    <div>
      <span>user:{user.role}</span>
      <button onClick={() => void logout()}>logout</button>
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

function clearAllCookies() {
  document.cookie.split(';').forEach((c) => {
    const name = c.split('=')[0]?.trim();
    if (name) {
      document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
    }
  });
}

beforeEach(() => {
  clearAllCookies();
  server.resetHandlers(...handlers);
});

afterEach(() => {
  clearAllCookies();
});

describe('AuthProvider', () => {
  it('shows isLoading=false and user=null when no session', async () => {
    server.use(
      http.get('http://localhost:8080/api/v1/me', () =>
        HttpResponse.json(
          { error: { code: 'UNAUTHORIZED', message: 'Not logged in' } },
          { status: 401 },
        ),
      ),
    );

    renderAuth();
    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });
  });

  it('fetches user when access cookie is present', async () => {
    document.cookie = 'access_token=mock-access-token; Path=/';

    renderAuth();

    expect(screen.getByText('loading')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('user:admin')).toBeInTheDocument();
    });
  });

  it('shows no user when getMe fails', async () => {
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

    document.cookie = 'access_token=invalid-token; Path=/';

    renderAuth();

    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });
  });

  it('login fetches user and navigates', async () => {
    let mePolicy: 'before-login' | 'after-login' = 'before-login';
    server.use(
      http.get('http://localhost:8080/api/v1/me', ({ request }) => {
        if (mePolicy === 'before-login') {
          return HttpResponse.json(
            { error: { code: 'UNAUTHORIZED', message: 'Not logged in' } },
            { status: 401 },
          );
        }
        const cookie = request.headers.get('Cookie') ?? '';
        const hasAccess = /(?:^|;\s)access_token=[^;]+/.test(cookie);
        if (!hasAccess) {
          return HttpResponse.json(
            { error: { code: 'UNAUTHORIZED', message: 'Not logged in' } },
            { status: 401 },
          );
        }
        return HttpResponse.json(mockUser);
      }),
    );

    renderAuth();

    await waitFor(() => {
      expect(screen.getByText('no user')).toBeInTheDocument();
    });

    mePolicy = 'after-login';

    await act(async () => {
      screen.getByText('login').click();
    });

    await waitFor(() => {
      expect(screen.getByText('user:admin')).toBeInTheDocument();
    });
  });

  it('logout clears user', async () => {
    document.cookie = 'access_token=mock-access-token; Path=/';

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
  });
});
