import { describe, it, expect } from 'vitest';
import { trackingCreate } from '../../../src/tracking/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('tracking integration', () => { it('should flow', () => { connect(); seed(); trackingCreate('test'); cleanup(); }); });
