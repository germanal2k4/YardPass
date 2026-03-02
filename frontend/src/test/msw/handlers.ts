import { http, HttpResponse } from 'msw';

// Default mock data
export const mockUser = {
  user_id: 1,
  role: 'admin' as const,
  building_id: 1,
};

export const mockGuardUser = {
  user_id: 2,
  role: 'guard' as const,
  building_id: 1,
};

export const mockTokens = {
  access_token: 'mock-access-token',
  refresh_token: 'mock-refresh-token',
  expires_in: 3600,
  token_type: 'Bearer' as const,
};

export const mockResidents = [
  {
    id: 1,
    apartment_id: 101,
    telegram_id: 123456789,
    chat_id: 123456789,
    name: 'Иван Петров',
    phone: '+79001234567',
    status: 'active',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 2,
    apartment_id: 102,
    telegram_id: 111222333,
    chat_id: 111222333,
    name: 'Мария Иванова',
    status: 'active',
    created_at: '2026-01-11T10:00:00Z',
    updated_at: '2026-01-11T10:00:00Z',
  },
];

export const mockRule = {
  id: 1,
  building_id: 1,
  quiet_hours_start: '22:00',
  quiet_hours_end: '08:00',
  daily_pass_limit_per_apartment: 5,
  max_pass_duration_hours: 24,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

export const mockScanEvents = {
  events: [
    {
      ID: 1,
      PassID: 'uuid-1',
      GuardUserID: 2,
      GuardUsername: 'guard1',
      ScannedAt: '2026-01-15T14:30:00Z',
      Result: 'valid' as const,
      CarPlate: 'А123ВС777',
      GuestName: 'Гость 1',
      ApartmentNumber: '101',
      BuildingID: 1,
      BuildingName: 'Дом 1',
    },
    {
      ID: 2,
      PassID: 'uuid-2',
      GuardUserID: 2,
      GuardUsername: 'guard1',
      ScannedAt: '2026-01-15T15:00:00Z',
      Result: 'invalid' as const,
      Reason: 'PASS_EXPIRED',
      CarPlate: 'В456ОР50',
    },
  ],
  limit: 20,
  offset: 0,
};

export const mockStatistics = {
  total_scans: 150,
  valid_scans: 120,
  invalid_scans: 30,
  unique_passes: 80,
  top_reasons: [
    { reason: 'PASS_EXPIRED', count: 15 },
    { reason: 'PASS_NOT_FOUND', count: 10 },
  ],
};

const API_BASE = 'http://localhost:8080';

export const handlers = [
  // Auth
  http.post(`${API_BASE}/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { username: string; password: string };
    if (body.username === 'admin' && body.password === 'password') {
      return HttpResponse.json(mockTokens);
    }
    if (body.username === 'guard' && body.password === 'password') {
      return HttpResponse.json(mockTokens);
    }
    return HttpResponse.json(
      { error: { code: 'INVALID_CREDENTIALS', message: 'Invalid username or password' } },
      { status: 401 },
    );
  }),

  http.post(`${API_BASE}/auth/refresh`, () => {
    return HttpResponse.json(mockTokens);
  }),

  http.get(`${API_BASE}/api/v1/me`, ({ request }) => {
    const auth = request.headers.get('Authorization');
    if (!auth) {
      return HttpResponse.json(
        { error: { code: 'UNAUTHORIZED', message: 'Invalid or missing token' } },
        { status: 401 },
      );
    }
    return HttpResponse.json(mockUser);
  }),

  // Passes
  http.post(`${API_BASE}/api/v1/passes/validate`, async ({ request }) => {
    const body = (await request.json()) as { qr_uuid?: string; car_plate?: string };
    if (body.qr_uuid === 'valid-uuid' || body.car_plate === 'А123ВС777') {
      return HttpResponse.json({
        valid: true,
        car_plate: 'А123ВС777',
        apartment: '101',
        valid_to: '2026-12-31T23:59:59Z',
      });
    }
    if (body.qr_uuid === 'expired-uuid') {
      return HttpResponse.json({
        valid: false,
        reason: 'PASS_EXPIRED',
      });
    }
    return HttpResponse.json(
      { error: { code: 'PASS_NOT_FOUND', message: 'Pass not found' } },
      { status: 404 },
    );
  }),

  // Residents
  http.get(`${API_BASE}/api/v1/residents`, () => {
    return HttpResponse.json({ residents: mockResidents });
  }),

  http.post(`${API_BASE}/api/v1/residents`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      id: 3,
      ...body,
      status: 'active',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });
  }),

  http.post(`${API_BASE}/api/v1/residents/bulk`, async ({ request }) => {
    const body = (await request.json()) as unknown[];
    return HttpResponse.json({
      created: body.length,
      residents: [],
      errors: [],
    });
  }),

  http.post(`${API_BASE}/api/v1/residents/import`, () => {
    return HttpResponse.json({ imported: 2, errors: [] });
  }),

  http.delete(`${API_BASE}/api/v1/residents/:id`, () => {
    return HttpResponse.json({ message: 'Deleted' });
  }),

  // Rules
  http.get(`${API_BASE}/api/v1/rules`, () => {
    return HttpResponse.json(mockRule);
  }),

  http.put(`${API_BASE}/api/v1/rules`, () => {
    return HttpResponse.json(mockRule);
  }),

  // Reports
  http.get(`${API_BASE}/api/v1/scan-events`, () => {
    return HttpResponse.json(mockScanEvents);
  }),

  http.get(`${API_BASE}/api/v1/reports/statistics`, () => {
    return HttpResponse.json(mockStatistics);
  }),

  http.get(`${API_BASE}/api/v1/reports/export`, () => {
    return new HttpResponse(new Blob(['fake-xlsx-data']), {
      headers: { 'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' },
    });
  }),
];
