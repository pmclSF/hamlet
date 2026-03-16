import { describe, it, expect } from 'vitest';
import { registerPatient, getPatient, updatePatient } from '../../../src/patients/registry';
describe('registerPatient', () => {
  it('should register', () => { expect(registerPatient('Jane', '1990-01-01').status).toBe('active'); });
});
describe('getPatient', () => {
  it('should get', () => { expect(getPatient('pat_1').name).toBe('Jane Doe'); });
});
describe('updatePatient', () => {
  it('should update', () => { expect(updatePatient('pat_1', { name: 'John' }).updated).toBe(true); });
});
