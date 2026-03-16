import { describe, it, expect } from 'vitest';
import { submitClaim, processClaim } from '../../../src/billing/claims';
import { scheduleAppointment } from '../../../src/appointments/scheduler';
import { connectDB, seedPatient, cleanupDB } from '../../../src/shared/db';
describe('claims integration', () => {
  it('should submit and process claim', () => {
    connectDB(); seedPatient();
    const apt = scheduleAppointment('pat_test', 'doc_1', '2026-04-01');
    const claim = submitClaim('pat_test', 200, 'visit');
    const processed = processClaim(claim.claimId);
    expect(processed.status).toBe('processed');
    cleanupDB();
  });
});
