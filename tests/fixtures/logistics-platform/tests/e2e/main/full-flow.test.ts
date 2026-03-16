import { describe, it, expect } from 'vitest';
import { shipmentsCreate } from '../../../src/shipments/service';
import { trackingCreate } from '../../../src/tracking/service';
import { warehousesCreate } from '../../../src/warehouses/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); shipmentsCreate('a'); trackingCreate('b'); warehousesCreate('c'); cleanup(); }); });
