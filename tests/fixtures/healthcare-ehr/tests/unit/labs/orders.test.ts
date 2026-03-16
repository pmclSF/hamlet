import { describe, it, expect } from 'vitest';
import { orderLabTest, getLabResults } from '../../../src/labs/orders';
describe('orderLabTest', () => {
  it('should order', () => { expect(orderLabTest('pat_1', 'blood_panel').status).toBe('ordered'); });
});
describe('getLabResults', () => {
  it('should get results', () => { expect(getLabResults('lab_1').status).toBe('pending'); });
});
