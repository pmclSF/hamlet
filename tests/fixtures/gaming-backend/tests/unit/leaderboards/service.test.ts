import { describe, it, expect } from 'vitest';
import { leaderboardsCreate, leaderboardsGet } from '../../../src/leaderboards/service';
describe('leaderboardsCreate', () => { it('should create', () => { expect(leaderboardsCreate('test').status).toBe('created'); }); });
describe('leaderboardsGet', () => { it('should get', () => { expect(leaderboardsGet('id_1').found).toBe(true); }); });
