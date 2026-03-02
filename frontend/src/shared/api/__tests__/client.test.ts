import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import MockAdapter from 'axios-mock-adapter';
import axios from 'axios';
import { STORAGE_KEYS } from '@/shared/config/constants';

let apiClient: typeof import('../client').apiClient;
let mock: MockAdapter;
let axiosMock: MockAdapter;

// Fresh import of apiClient for each test to reset interceptor state
async function loadClient() {
  // Dynamic import to get a fresh module (interceptor state like isRefreshing/failedQueue
  // lives in module scope). Using vi.resetModules() + import avoids stale state.
  vi.resetModules();
  const mod = await import('../client');
  apiClient = mod.apiClient;
  mock = new MockAdapter(apiClient);
  axiosMock = new MockAdapter(axios);
}

describe('apiClient interceptors', () => {
  const originalLocation = window.location.href;

  beforeEach(async () => {
    localStorage.clear();
    // Mock window.location.href setter
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...window.location, href: originalLocation },
    });
    await loadClient();
  });

  afterEach(() => {
    mock.restore();
    axiosMock.restore();
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: originalLocation },
    });
  });

  it('adds Authorization header from localStorage', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'test-token');
    mock.onGet('/api/v1/test').reply(200, { ok: true });

    await apiClient.get('/api/v1/test');

    const requestHeaders = mock.history.get[0].headers;
    expect(requestHeaders?.Authorization).toBe('Bearer test-token');
  });

  it('does not add Authorization header when no token', async () => {
    mock.onGet('/api/v1/test').reply(200, { ok: true });

    await apiClient.get('/api/v1/test');

    const requestHeaders = mock.history.get[0].headers;
    expect(requestHeaders?.Authorization).toBeUndefined();
  });

  it('retries request after successful token refresh on 401', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'expired-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'valid-refresh');

    // First call returns 401
    mock.onGet('/api/v1/data').replyOnce(401, {
      error: { code: 'UNAUTHORIZED', message: 'Token expired' },
    });
    // After refresh, the retried request succeeds
    mock.onGet('/api/v1/data').replyOnce(200, { result: 'success' });

    // Refresh endpoint (called on raw axios, not apiClient)
    axiosMock.onPost('http://localhost:8080/auth/refresh').reply(200, {
      access_token: 'new-access-token',
      refresh_token: 'new-refresh-token',
    });

    const response = await apiClient.get('/api/v1/data');

    expect(response.data).toEqual({ result: 'success' });
    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBe('new-access-token');
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBe('new-refresh-token');
  });

  it('clears tokens and redirects when no refresh token available', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'expired-token');
    // No refresh token set

    mock.onGet('/api/v1/data').reply(401, {
      error: { code: 'UNAUTHORIZED', message: 'Token expired' },
    });

    await expect(apiClient.get('/api/v1/data')).rejects.toThrow();

    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBeNull();
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBeNull();
    expect(window.location.href).toBe('/login');
  });

  it('clears tokens and redirects when refresh fails', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'expired-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'invalid-refresh');

    mock.onGet('/api/v1/data').reply(401, {
      error: { code: 'UNAUTHORIZED', message: 'Token expired' },
    });

    axiosMock.onPost('http://localhost:8080/auth/refresh').reply(401, {
      error: { code: 'INVALID_REFRESH_TOKEN', message: 'Invalid refresh token' },
    });

    await expect(apiClient.get('/api/v1/data')).rejects.toThrow();

    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBeNull();
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBeNull();
    expect(window.location.href).toBe('/login');
  });

  it('passes non-401 errors through without refresh attempt', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'valid-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'valid-refresh');

    mock.onGet('/api/v1/data').reply(403, {
      error: { code: 'FORBIDDEN', message: 'Forbidden' },
    });

    await expect(apiClient.get('/api/v1/data')).rejects.toThrow();

    // Token should still be in localStorage (no refresh attempted)
    expect(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)).toBe('valid-token');
    expect(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)).toBe('valid-refresh');
  });

  it('queues concurrent 401 requests and refreshes only once', async () => {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, 'expired-token');
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, 'valid-refresh');

    // Both requests return 401 initially
    mock.onGet('/api/v1/first').replyOnce(401);
    mock.onGet('/api/v1/second').replyOnce(401);
    // After refresh, retries succeed
    mock.onGet('/api/v1/first').replyOnce(200, { id: 1 });
    mock.onGet('/api/v1/second').replyOnce(200, { id: 2 });

    axiosMock.onPost('http://localhost:8080/auth/refresh').reply(200, {
      access_token: 'new-token',
      refresh_token: 'new-refresh',
    });

    const [r1, r2] = await Promise.all([
      apiClient.get('/api/v1/first'),
      apiClient.get('/api/v1/second'),
    ]);

    expect(r1.data).toEqual({ id: 1 });
    expect(r2.data).toEqual({ id: 2 });

    // Refresh should have been called exactly once
    const refreshCalls = axiosMock.history.post.filter(
      (req) => req.url === 'http://localhost:8080/auth/refresh',
    );
    expect(refreshCalls).toHaveLength(1);
  });
});
