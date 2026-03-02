import { authApi } from '../auth';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('authApi', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe('login', () => {
    it('returns tokens on successful login', async () => {
      const result = await authApi.login({ username: 'admin', password: 'password' });

      expect(result.access_token).toBe('mock-access-token');
      expect(result.refresh_token).toBe('mock-refresh-token');
      expect(result.token_type).toBe('Bearer');
    });

    it('throws on invalid credentials', async () => {
      await expect(
        authApi.login({ username: 'wrong', password: 'wrong' }),
      ).rejects.toThrow();
    });
  });

  describe('getMe', () => {
    it('returns user data when authenticated', async () => {
      localStorage.setItem('access_token', 'mock-token');
      const result = await authApi.getMe();

      expect(result.user_id).toBe(1);
      expect(result.role).toBe('admin');
    });

    it('throws on server error', async () => {
      localStorage.setItem('access_token', 'mock-token');
      server.use(
        http.get(`${API_BASE}/api/v1/me`, () => {
          return HttpResponse.json(
            { error: { code: 'INTERNAL_ERROR', message: 'Server failure' } },
            { status: 500 },
          );
        }),
      );

      await expect(authApi.getMe()).rejects.toThrow();
    });
  });

  describe('logout', () => {
    it('clears tokens from localStorage', () => {
      localStorage.setItem('access_token', 'tok-a');
      localStorage.setItem('refresh_token', 'tok-r');

      authApi.logout();

      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });
  });
});
