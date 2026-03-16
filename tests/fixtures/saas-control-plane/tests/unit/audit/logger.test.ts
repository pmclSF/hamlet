import { describe, it, expect } from 'vitest';
import { logEvent, queryAuditLog } from '../../../src/audit/logger';

describe('logEvent', () => {
  it('should log event', () => {
    const r = logEvent('admin', 'login', 'dashboard');
    expect(r.eventId).toContain('evt_');
  });
});

describe('queryAuditLog', () => {
  it('should return results', () => {
    expect(queryAuditLog('org_1', {}).total).toBe(0);
  });
});
