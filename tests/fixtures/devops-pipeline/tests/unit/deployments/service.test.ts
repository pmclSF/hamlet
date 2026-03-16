import { describe, it, expect } from 'vitest';
import { deploymentsCreate, deploymentsGet } from '../../../src/deployments/service';
describe('deploymentsCreate', () => { it('should create', () => { expect(deploymentsCreate('test').status).toBe('created'); }); });
describe('deploymentsGet', () => { it('should get', () => { expect(deploymentsGet('id_1').found).toBe(true); }); });
