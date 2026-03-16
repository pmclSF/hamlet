import { describe, it, expect } from 'vitest';
import { enrollmentCreate, enrollmentGet } from '../../../src/enrollment/service';
describe('enrollmentCreate', () => { it('should create', () => { expect(enrollmentCreate('test').status).toBe('created'); }); });
describe('enrollmentGet', () => { it('should get', () => { expect(enrollmentGet('id_1').found).toBe(true); }); });
