import { describe, it, expect } from 'vitest';
import { agentsCreate, agentsGet } from '../../../src/agents/service';
describe('agentsCreate', () => { it('should create', () => { expect(agentsCreate('test').status).toBe('created'); }); });
describe('agentsGet', () => { it('should get', () => { expect(agentsGet('id_1').found).toBe(true); }); });
