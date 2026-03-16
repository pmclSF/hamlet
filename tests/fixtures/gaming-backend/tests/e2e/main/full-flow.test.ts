import { describe, it, expect } from 'vitest';
import { playersCreate } from '../../../src/players/service';
import { matchesCreate } from '../../../src/matches/service';
import { leaderboardsCreate } from '../../../src/leaderboards/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); playersCreate('a'); matchesCreate('b'); leaderboardsCreate('c'); cleanup(); }); });
