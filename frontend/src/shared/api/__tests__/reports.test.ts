import { describe, it, expect } from 'vitest';
import { reportsApi } from '../reports';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('reportsApi', () => {
  describe('getScanEvents', () => {
    it('returns events from default handler', async () => {
      const result = await reportsApi.getScanEvents();
      expect(result.events).toHaveLength(2);
      expect(result.events[0].CarPlate).toBe('А123ВС777');
    });

    it('builds query string with all params', async () => {
      let capturedUrl = '';
      server.use(
        http.get(`${API_BASE}/api/v1/scan-events`, ({ request }) => {
          capturedUrl = request.url;
          return HttpResponse.json({ events: [], limit: 10, offset: 0 });
        }),
      );

      await reportsApi.getScanEvents({
        limit: 10,
        offset: 20,
        from: '2026-01-01T00:00:00Z',
        to: '2026-12-31T23:59:59Z',
        result: 'valid',
      });

      expect(capturedUrl).toContain('limit=10');
      expect(capturedUrl).toContain('offset=20');
      expect(capturedUrl).toContain('from=2026-01-01T00%3A00%3A00Z');
      expect(capturedUrl).toContain('result=valid');
    });

    it('omits undefined params from URL', async () => {
      let capturedUrl = '';
      server.use(
        http.get(`${API_BASE}/api/v1/scan-events`, ({ request }) => {
          capturedUrl = request.url;
          return HttpResponse.json({ events: [], limit: 20, offset: 0 });
        }),
      );

      await reportsApi.getScanEvents({ limit: 20 });

      expect(capturedUrl).toContain('limit=20');
      expect(capturedUrl).not.toContain('offset');
      expect(capturedUrl).not.toContain('from');
      expect(capturedUrl).not.toContain('result');
    });
  });

  describe('getStatistics', () => {
    it('returns statistics from default handler', async () => {
      const result = await reportsApi.getStatistics();
      expect(result.total_scans).toBe(150);
      expect(result.valid_scans).toBe(120);
      expect(result.invalid_scans).toBe(30);
      expect(result.unique_passes).toBe(80);
    });

    it('passes from/to as query params', async () => {
      let capturedUrl = '';
      server.use(
        http.get(`${API_BASE}/api/v1/reports/statistics`, ({ request }) => {
          capturedUrl = request.url;
          return HttpResponse.json({
            total_scans: 0,
            valid_scans: 0,
            invalid_scans: 0,
            unique_passes: 0,
            top_reasons: [],
          });
        }),
      );

      await reportsApi.getStatistics({
        from: '2026-03-01T00:00:00Z',
        to: '2026-03-02T00:00:00Z',
      });

      expect(capturedUrl).toContain('from=2026-03-01T00%3A00%3A00Z');
      expect(capturedUrl).toContain('to=2026-03-02T00%3A00%3A00Z');
    });
  });

  describe('exportReport', () => {
    it('returns blob for xlsx export', async () => {
      const result = await reportsApi.exportReport({ format: 'xlsx' });
      expect(result).toBeInstanceOf(Blob);
    });

    it('builds correct URL with format and date params', async () => {
      let capturedUrl = '';
      server.use(
        http.get(`${API_BASE}/api/v1/reports/export`, ({ request }) => {
          capturedUrl = request.url;
          return new HttpResponse(new Blob(['data']), {
            headers: { 'Content-Type': 'application/octet-stream' },
          });
        }),
      );

      await reportsApi.exportReport({
        format: 'xlsx',
        from: '2026-01-01T00:00:00Z',
        to: '2026-12-31T23:59:59Z',
      });

      expect(capturedUrl).toContain('format=xlsx');
      expect(capturedUrl).toContain('from=2026-01-01T00%3A00%3A00Z');
    });
  });
});
