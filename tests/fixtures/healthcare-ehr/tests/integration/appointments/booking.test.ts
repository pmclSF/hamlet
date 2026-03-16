import { describe, it, expect } from 'vitest';
import { scheduleAppointment } from '../../../src/appointments/scheduler';
import { registerPatient } from '../../../src/patients/registry';
import { connectDB, seedPatient, seedDoctor, cleanupDB } from '../../../src/shared/db';
describe('booking integration', () => {
  it('should register patient and schedule', () => {
    connectDB(); seedPatient(); seedDoctor();
    const pat = registerPatient('Jane', '1990-01-01');
    const apt = scheduleAppointment(pat.patientId, 'doc_1', '2026-04-01');
    expect(apt.status).toBe('scheduled');
    cleanupDB();
  });
});
