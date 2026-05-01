import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import MockAdapter from 'axios-mock-adapter';

let apiClient: typeof import('../client').apiClient;
let mock: MockAdapter;

async function loadClient() {
  vi.resetModules();
  const mod = await import('../client');
  apiClient = mod.apiClient;
  mock = new MockAdapter(apiClient);
}

describe('apiClient interceptors', () => {
  const originalLocation = window.location.href;

  beforeEach(async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...window.location, href: originalLocation },
    });
    await loadClient();
  });

  afterEach(() => {
    mock.restore();
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: originalLocation },
    });
  });

  it('sends credentials on requests', async () => {
    mock.onGet('/api/v1/test').reply(200, { ok: true });

    await apiClient.get('/api/v1/test');

    expect(mock.history.get[0].withCredentials).toBe(true);
  });

  it('retries request after successful token refresh on 401', async () => {
    mock.onGet('/api/v1/data').replyOnce(401, {
      error: { code: 'UNAUTHORIZED', message: 'Token expired' },
    });
    mock.onGet('/api/v1/data').replyOnce(200, { result: 'success' });

    mock.onPost('/auth/refresh').reply(200, { expires_in: 3600, token_type: 'Bearer' });

    const response = await apiClient.get('/api/v1/data');

    expect(response.data).toEqual({ result: 'success' });
  });

  it('redirects to login when refresh fails', async () => {
    mock.onGet('/api/v1/data').reply(401, {
      error: { code: 'UNAUTHORIZED', message: 'Token expired' },
    });

    mock.onPost('/auth/refresh').reply(401, {
      error: { code: 'INVALID_REFRESH_TOKEN', message: 'Invalid refresh token' },
    });

    await expect(apiClient.get('/api/v1/data')).rejects.toThrow();

    expect(window.location.href).toBe('/login');
  });

  it('passes non-401 errors through without refresh attempt', async () => {
    mock.onGet('/api/v1/data').reply(403, {
      error: { code: 'FORBIDDEN', message: 'Forbidden' },
    });

    await expect(apiClient.get('/api/v1/data')).rejects.toThrow();
  });

  it('queues concurrent 401 requests and refreshes only once', async () => {
    mock.onGet('/api/v1/first').replyOnce(401);
    mock.onGet('/api/v1/second').replyOnce(401);
    mock.onGet('/api/v1/first').replyOnce(200, { id: 1 });
    mock.onGet('/api/v1/second').replyOnce(200, { id: 2 });

    mock.onPost('/auth/refresh').reply(200, { expires_in: 3600, token_type: 'Bearer' });

    const [r1, r2] = await Promise.all([
      apiClient.get('/api/v1/first'),
      apiClient.get('/api/v1/second'),
    ]);

    expect(r1.data).toEqual({ id: 1 });
    expect(r2.data).toEqual({ id: 2 });

    const refreshCalls = mock.history.post.filter((req) => req.url?.includes('/auth/refresh'));
    expect(refreshCalls).toHaveLength(1);
  });
});
