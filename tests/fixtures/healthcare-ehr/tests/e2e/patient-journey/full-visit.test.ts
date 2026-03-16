import { describe, it, expect } from 'vitest';
import { registerPatient } from '../../../src/patients/registry';
import { scheduleAppointment } from '../../../src/appointments/scheduler';
import { createPrescription } from '../../../src/prescriptions/prescribe';
import { orderLabTest } from '../../../src/labs/orders';
import { submitClaim } from '../../../src/billing/claims';
import { connectDB, seedPatient, seedDoctor, cleanupDB } from '../../../src/shared/db';
describe('full patient visit e2e', () => {
  it('should complete visit flow', () => {
    connectDB(); seedPatient(); seedDoctor();
    const pat = registerPatient('Jane', '1990-01-01');
    scheduleAppointment(pat.patientId, 'doc_1', '2026-04-01');
    createPrescription(pat.patientId, 'Amoxicillin', '500mg');
    orderLabTest(pat.patientId, 'cbc');
    submitClaim(pat.patientId, 350, 'office_visit');
    cleanupDB();
  });
});
