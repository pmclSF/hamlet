import { describe, it, expect } from 'vitest';
import { leadsCreate, leadsGet } from '../../../src/leads/service';
describe('leadsCreate', () => { it('should create', () => { expect(leadsCreate('test').status).toBe('created'); }); });
describe('leadsGet', () => { it('should get', () => { expect(leadsGet('id_1').found).toBe(true); }); });
