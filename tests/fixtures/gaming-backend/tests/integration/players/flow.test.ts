import { describe, it, expect } from 'vitest';
import { playersCreate } from '../../../src/players/service';
import { matchesGet } from '../../../src/matches/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('players integration', () => {
  it('should flow', () => { connect(); seed(); playersCreate('test'); matchesGet('id_1'); cleanup(); });
});
