import { describe, it, expect } from 'vitest';
import { playersCreate } from '../../../src/players/service';
import { matchesCreate } from '../../../src/matches/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); playersCreate('x'); matchesCreate('y'); cleanup(); }); });
