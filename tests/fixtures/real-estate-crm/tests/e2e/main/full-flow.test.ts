import { describe, it, expect } from 'vitest';
import { listingsCreate } from '../../../src/listings/service';
import { leadsCreate } from '../../../src/leads/service';
import { agentsCreate } from '../../../src/agents/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); listingsCreate('a'); leadsCreate('b'); agentsCreate('c'); cleanup(); }); });
