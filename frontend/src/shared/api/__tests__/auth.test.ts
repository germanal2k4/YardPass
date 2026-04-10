import { authApi } from '../auth';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('authApi', () => {
  beforeEach(() => {
    document.cookie = 'access_token=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/';
  });

  describe('login', () => {
    it('returns metadata on successful login and sets access cookie (mock)', async () => {
      const result = await authApi.login({ username: 'admin', password: 'password' });

      expect(result.expires_in).toBe(3600);
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
      document.cookie = 'access_token=mock-token; Path=/';
      const result = await authApi.getMe();

      expect(result.user_id).toBe(1);
      expect(result.role).toBe('admin');
    });

    it('throws on server error', async () => {
      document.cookie = 'access_token=mock-token; Path=/';
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
    it('calls logout endpoint', async () => {
      await expect(authApi.logout()).resolves.toBeUndefined();
    });
  });
});
