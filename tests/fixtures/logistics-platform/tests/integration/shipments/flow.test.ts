import { describe, it, expect } from 'vitest';
import { shipmentsCreate } from '../../../src/shipments/service';
import { trackingGet } from '../../../src/tracking/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('shipments integration', () => {
  it('should flow', () => { connect(); seed(); shipmentsCreate('test'); trackingGet('id_1'); cleanup(); });
});
