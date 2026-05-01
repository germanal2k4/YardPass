import { describe, it, expect } from 'vitest';
import { rulesApi } from '../rules';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('rulesApi', () => {
  it('get fetches rules with building_id query param', async () => {
    let capturedUrl = '';
    server.use(
      http.get(`${API_BASE}/api/v1/rules`, ({ request }) => {
        capturedUrl = request.url;
        return HttpResponse.json({
          id: 1,
          building_id: 5,
          quiet_hours_start: '22:00',
          quiet_hours_end: '08:00',
          daily_pass_limit_per_apartment: 5,
          max_pass_duration_hours: 24,
        });
      }),
    );

    const result = await rulesApi.get(5);
    expect(capturedUrl).toContain('building_id=5');
    expect(result.quiet_hours_start).toBe('22:00');
    expect(result.daily_pass_limit_per_apartment).toBe(5);
  });

  it('update sends PUT with building_id and data', async () => {
    let capturedUrl = '';
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.put(`${API_BASE}/api/v1/rules`, async ({ request }) => {
        capturedUrl = request.url;
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({
          id: 1,
          building_id: 3,
          ...capturedBody,
        });
      }),
    );

    const data = {
      quiet_hours_start: '23:00',
      quiet_hours_end: '07:00',
      daily_pass_limit_per_apartment: 10,
      max_pass_duration_hours: 48,
    };

    const result = await rulesApi.update(3, data);
    expect(capturedUrl).toContain('building_id=3');
    expect(capturedBody.quiet_hours_start).toBe('23:00');
    expect(result.daily_pass_limit_per_apartment).toBe(10);
  });
});
