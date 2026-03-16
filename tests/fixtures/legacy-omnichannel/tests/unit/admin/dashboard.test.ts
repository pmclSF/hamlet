import { describe, it, expect } from 'vitest';
import { getAdminStats, searchOrders } from '../../../src/admin/dashboard';

describe('getAdminStats', () => {
  it('should return stats', () => {
    const stats = getAdminStats();
    expect(stats.totalOrders).toBeGreaterThan(0);
  });
});

describe('searchOrders', () => {
  it('should search', () => {
    expect(searchOrders('laptop').total).toBe(0);
  });
});
