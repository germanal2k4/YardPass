import { test as base, expect, Page } from '@playwright/test';

const mockAuthMeta = {
  expires_in: 3600,
  token_type: 'Bearer',
};

const mockAdminUser = { user_id: 1, role: 'admin', building_id: 1 };
const mockGuardUser = { user_id: 2, role: 'guard', building_id: 1 };

const mockResidents = [
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

const mockRule = {
  id: 1,
  building_id: 1,
  quiet_hours_start: '22:00',
  quiet_hours_end: '08:00',
  daily_pass_limit_per_apartment: 5,
  max_pass_duration_hours: 24,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

const mockScanEvents = {
  events: [
    {
      ID: 1,
      PassID: 'uuid-1',
      GuardUserID: 2,
      GuardUsername: 'guard1',
      ScannedAt: '2026-01-15T14:30:00Z',
      Result: 'valid',
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
      Result: 'invalid',
      Reason: 'PASS_EXPIRED',
      CarPlate: 'В456ОР50',
    },
  ],
  limit: 20,
  offset: 0,
};

const mockStatistics = {
  total_scans: 150,
  valid_scans: 120,
  invalid_scans: 30,
  unique_passes: 80,
  top_reasons: [
    { reason: 'PASS_EXPIRED', count: 15 },
    { reason: 'PASS_NOT_FOUND', count: 10 },
  ],
};

async function setupApiMocks(page: Page) {
  let loggedInRole: 'admin' | 'guard' | null = null;

  await page.route('**/auth/login', async (route) => {
    const body = route.request().postDataJSON();
    if (body.username === 'admin' && body.password === 'password') {
      loggedInRole = 'admin';
      await route.fulfill({
        status: 200,
        json: mockAuthMeta,
        headers: {
          'Set-Cookie': 'access_token=mock-admin-token; Path=/',
        },
      });
    } else if (body.username === 'guard' && body.password === 'password') {
      loggedInRole = 'guard';
      await route.fulfill({
        status: 200,
        json: mockAuthMeta,
        headers: {
          'Set-Cookie': 'access_token=mock-guard-token; Path=/',
        },
      });
    } else {
      await route.fulfill({
        status: 401,
        json: { error: { code: 'INVALID_CREDENTIALS', message: 'Invalid username or password' } },
      });
    }
  });

  await page.route('**/auth/refresh', async (route) => {
    await route.fulfill({
      status: 200,
      json: mockAuthMeta,
      headers: {
        'Set-Cookie':
          loggedInRole === 'guard'
            ? 'access_token=mock-guard-token; Path=/'
            : 'access_token=mock-admin-token; Path=/',
      },
    });
  });

  await page.route('**/auth/logout', async (route) => {
    loggedInRole = null;
    await route.fulfill({ status: 204, body: '' });
  });

  await page.route('**/api/v1/me', async (route) => {
    if (!loggedInRole) {
      await route.fulfill({
        status: 401,
        json: { error: { code: 'UNAUTHORIZED', message: 'Missing token' } },
      });
      return;
    }
    if (loggedInRole === 'guard') {
      await route.fulfill({ json: mockGuardUser });
    } else {
      await route.fulfill({ json: mockAdminUser });
    }
  });

  await page.route('**/api/v1/passes/validate', async (route) => {
    const body = route.request().postDataJSON();
    if (body.qr_uuid === 'valid-uuid' || body.car_plate === 'А123ВС777') {
      await route.fulfill({
        json: { valid: true, car_plate: 'А123ВС777', apartment: '101', valid_to: '2026-12-31T23:59:59Z' },
      });
    } else if (body.qr_uuid === 'expired-uuid') {
      await route.fulfill({ json: { valid: false, reason: 'PASS_EXPIRED' } });
    } else {
      await route.fulfill({
        status: 404,
        json: { error: { code: 'PASS_NOT_FOUND', message: 'Pass not found' } },
      });
    }
  });

  await page.route('**/api/v1/residents**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    if (url.includes('/residents/bulk') && method === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({ json: { created: body.length, residents: [], errors: [] } });
    } else if (url.includes('/residents/import') && method === 'POST') {
      await route.fulfill({ json: { imported: 2, errors: [] } });
    } else if (method === 'GET') {
      await route.fulfill({ json: { residents: mockResidents } });
    } else if (method === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        json: { id: 3, ...body, status: 'active', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      });
    } else if (method === 'DELETE') {
      await route.fulfill({ json: { message: 'Deleted' } });
    } else {
      await route.continue();
    }
  });

  await page.route('**/api/v1/rules*', async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ json: mockRule });
    } else {
      await route.fulfill({ json: mockRule });
    }
  });

  await page.route('**/api/v1/scan-events*', async (route) => {
    await route.fulfill({ json: mockScanEvents });
  });

  await page.route('**/api/v1/reports/statistics*', async (route) => {
    await route.fulfill({ json: mockStatistics });
  });

  await page.route('**/api/v1/reports/export*', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      body: Buffer.from('fake xlsx content'),
    });
  });
}

export const test = base.extend({
  page: async ({ page }, use) => {
    await setupApiMocks(page);
    await use(page);
  },
});

export { expect };
