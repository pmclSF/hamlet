import { describe, it, expect } from 'vitest';
import { shipmentsCreate } from '../../../src/shipments/service';
import { trackingCreate } from '../../../src/tracking/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); shipmentsCreate('x'); trackingCreate('y'); cleanup(); }); });
