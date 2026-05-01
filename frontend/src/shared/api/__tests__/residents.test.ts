import { describe, it, expect, vi } from 'vitest';
import { residentsApi } from '../residents';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('residentsApi', () => {
  it('getAll returns residents array', async () => {
    const result = await residentsApi.getAll();
    expect(result).toHaveLength(2);
    expect(result[0].name).toBe('Иван Петров');
  });

  it('getAll passes status filter as query param', async () => {
    let capturedUrl = '';
    server.use(
      http.get(`${API_BASE}/api/v1/residents`, ({ request }) => {
        capturedUrl = request.url;
        return HttpResponse.json({ residents: [] });
      }),
    );

    await residentsApi.getAll({ status: 'active' });
    expect(capturedUrl).toContain('status=active');
  });

  it('create sends POST with resident data', async () => {
    const data = {
      apartment_number: '101',
      telegram_id: 999,
      chat_id: 999,
      name: 'Test',
    };

    const result = await residentsApi.create(data);
    expect(result).toHaveProperty('id');
    expect(result.telegram_id).toBe(999);
  });

  it('createBulk sends POST to /residents/bulk', async () => {
    const data = [
      { apartment_number: '101', telegram_id: 111, chat_id: 111 },
      { apartment_number: '102', telegram_id: 222, chat_id: 222 },
    ];

    const result = await residentsApi.createBulk(data);
    expect(result).toHaveProperty('created');
    expect(result.created).toBe(2);
  });

  it('importFromCSV constructs FormData with file and building_id', async () => {
    const { apiClient } = await import('../client');
    const postSpy = vi.spyOn(apiClient, 'post').mockResolvedValue({
      data: { created: 3, errors: [] },
    });

    const file = new File(['col1,col2\n1,2'], 'test.csv', { type: 'text/csv' });
    const result = await residentsApi.importFromCSV(file, 42);

    expect(postSpy).toHaveBeenCalledWith(
      '/api/v1/residents/import',
      expect.any(FormData),
      expect.objectContaining({
        params: { building_id: 42 },
        headers: { 'Content-Type': 'multipart/form-data' },
      }),
    );
    expect(result.created).toBe(3);

    postSpy.mockRestore();
  });

  it('delete sends DELETE to /residents/:id', async () => {
    const result = await residentsApi.delete(1);
    expect(result).toHaveProperty('message');
  });
});
