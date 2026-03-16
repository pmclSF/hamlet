import { describe, it, expect } from 'vitest';
import { registerPatient } from '../../../src/patients/registry';
import { scheduleAppointment } from '../../../src/appointments/scheduler';
import { orderLabTest } from '../../../src/labs/orders';
import { submitClaim } from '../../../src/billing/claims';
import { connectDB, seedPatient, seedDoctor, cleanupDB } from '../../../src/shared/db';
describe('checkup flow e2e', () => {
  it('should complete checkup', () => {
    connectDB(); seedPatient(); seedDoctor();
    const pat = registerPatient('Jane', '1990-01-01');
    scheduleAppointment(pat.patientId, 'doc_1', '2026-04-01');
    orderLabTest(pat.patientId, 'blood_panel');
    submitClaim(pat.patientId, 200, 'checkup');
    cleanupDB();
  });
});
