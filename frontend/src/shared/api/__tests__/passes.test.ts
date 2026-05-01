import { describe, it, expect } from 'vitest';
import { passesApi } from '../passes';

describe('passesApi', () => {
  it('validate with qr_uuid returns valid pass', async () => {
    const result = await passesApi.validate({ qr_uuid: 'valid-uuid' });
    expect(result.valid).toBe(true);
    expect(result.car_plate).toBe('А123ВС777');
    expect(result.apartment).toBe('101');
  });

  it('validate with car_plate returns valid pass', async () => {
    const result = await passesApi.validate({ car_plate: 'А123ВС777' });
    expect(result.valid).toBe(true);
  });

  it('validate with unknown qr_uuid returns 404 error', async () => {
    await expect(passesApi.validate({ qr_uuid: 'unknown-uuid' })).rejects.toThrow();
  });

  it('validate with expired UUID returns invalid pass', async () => {
    const result = await passesApi.validate({ qr_uuid: 'expired-uuid' });
    expect(result.valid).toBe(false);
    expect(result.reason).toBe('PASS_EXPIRED');
  });

  it('validate with resident personal token returns valid pass', async () => {
    const result = await passesApi.validate({ qr_uuid: 'resident:123:token' });
    expect(result.valid).toBe(true);
  });
});
