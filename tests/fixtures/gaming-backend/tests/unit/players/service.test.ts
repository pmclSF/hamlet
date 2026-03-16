import { describe, it, expect } from 'vitest';
import { playersCreate, playersGet } from '../../../src/players/service';
describe('playersCreate', () => { it('should create', () => { expect(playersCreate('test').status).toBe('created'); }); });
describe('playersGet', () => { it('should get', () => { expect(playersGet('id_1').found).toBe(true); }); });
