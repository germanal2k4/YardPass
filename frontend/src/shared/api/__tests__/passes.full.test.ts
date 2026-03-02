import { passesApi } from '../passes';
import { server } from '@/test/msw/server';
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080';

describe('passesApi - extended', () => {
  beforeEach(() => {
    localStorage.setItem('access_token', 'mock-token');
  });

  describe('validate with car_plate', () => {
    it('returns valid result for known car plate', async () => {
      const result = await passesApi.validate({ car_plate: 'А123ВС777' });
      expect(result.valid).toBe(true);
      expect(result.car_plate).toBe('А123ВС777');
    });

    it('returns 404 for unknown car plate', async () => {
      await expect(
        passesApi.validate({ car_plate: 'UNKNOWN' }),
      ).rejects.toThrow();
    });
  });

  describe('getActive', () => {
    it('returns active passes for apartment', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/passes/active`, () => {
          return HttpResponse.json({
            passes: [
              {
                id: 'uuid-a',
                apartment_id: 101,
                car_plate: 'А111АА77',
                valid_from: '2026-01-01T00:00:00Z',
                valid_to: '2026-12-31T23:59:59Z',
                status: 'active',
                created_at: '2026-01-01T00:00:00Z',
                updated_at: '2026-01-01T00:00:00Z',
              },
            ],
          });
        }),
      );

      const result = await passesApi.getActive(101);
      expect(result.passes).toHaveLength(1);
      expect(result.passes[0].status).toBe('active');
    });

    it('returns empty array when no active passes', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/passes/active`, () => {
          return HttpResponse.json({ passes: [] });
        }),
      );

      const result = await passesApi.getActive(999);
      expect(result.passes).toHaveLength(0);
    });
  });

  describe('revoke', () => {
    it('revokes a pass successfully', async () => {
      server.use(
        http.post(`${API_BASE}/api/v1/passes/uuid-1/revoke`, () => {
          return HttpResponse.json({ message: 'Pass revoked', pass_id: 'uuid-1' });
        }),
      );

      const result = await passesApi.revoke('uuid-1');
      expect(result.message).toBe('Pass revoked');
      expect(result.pass_id).toBe('uuid-1');
    });

    it('throws on revocation of non-existent pass', async () => {
      server.use(
        http.post(`${API_BASE}/api/v1/passes/nonexistent/revoke`, () => {
          return HttpResponse.json(
            { error: { code: 'PASS_NOT_FOUND', message: 'Not found' } },
            { status: 404 },
          );
        }),
      );

      await expect(passesApi.revoke('nonexistent')).rejects.toThrow();
    });
  });
});
