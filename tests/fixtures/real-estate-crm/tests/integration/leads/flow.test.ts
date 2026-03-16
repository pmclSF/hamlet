import { describe, it, expect } from 'vitest';
import { leadsCreate } from '../../../src/leads/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('leads integration', () => { it('should flow', () => { connect(); seed(); leadsCreate('test'); cleanup(); }); });
