import { describe, it, expect } from 'vitest';
import { createPrescription, renewPrescription } from '../../../src/prescriptions/prescribe';
describe('createPrescription', () => {
  it('should create', () => { expect(createPrescription('pat_1', 'Aspirin', '100mg').status).toBe('active'); });
});
describe('renewPrescription', () => {
  it('should renew', () => { expect(renewPrescription('rx_1').status).toBe('renewed'); });
});
