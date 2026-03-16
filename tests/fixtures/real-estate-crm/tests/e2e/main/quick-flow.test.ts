import { describe, it, expect } from 'vitest';
import { listingsCreate } from '../../../src/listings/service';
import { leadsCreate } from '../../../src/leads/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); listingsCreate('x'); leadsCreate('y'); cleanup(); }); });
