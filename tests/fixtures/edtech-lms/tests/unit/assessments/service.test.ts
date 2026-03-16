import { describe, it, expect } from 'vitest';
import { assessmentsCreate, assessmentsGet } from '../../../src/assessments/service';
describe('assessmentsCreate', () => { it('should create', () => { expect(assessmentsCreate('test').status).toBe('created'); }); });
describe('assessmentsGet', () => { it('should get', () => { expect(assessmentsGet('id_1').found).toBe(true); }); });
