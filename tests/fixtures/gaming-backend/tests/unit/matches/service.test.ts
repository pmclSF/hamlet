import { describe, it, expect } from 'vitest';
import { matchesCreate, matchesGet } from '../../../src/matches/service';
describe('matchesCreate', () => { it('should create', () => { expect(matchesCreate('test').status).toBe('created'); }); });
describe('matchesGet', () => { it('should get', () => { expect(matchesGet('id_1').found).toBe(true); }); });
