import { describe, it, expect } from 'vitest';
import { matchesCreate } from '../../../src/matches/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('matches integration', () => { it('should flow', () => { connect(); seed(); matchesCreate('test'); cleanup(); }); });
