import { describe, it, expect } from 'vitest';
import { listingsCreate } from '../../../src/listings/service';
import { leadsGet } from '../../../src/leads/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('listings integration', () => {
  it('should flow', () => { connect(); seed(); listingsCreate('test'); leadsGet('id_1'); cleanup(); });
});
